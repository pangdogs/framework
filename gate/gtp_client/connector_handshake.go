package gtp_client

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/method"
	"kit.golaxy.org/plugins/transport/protocol"
	"math/big"
	"net"
	"strings"
)

// handshake 握手过程
func (ctor *_Connector) handshake(conn net.Conn, client *Client) error {
	// 握手协议
	handshake := &protocol.HandshakeProtocol{
		Transceiver: &protocol.Transceiver{
			Conn:    conn,
			Encoder: &codec.Encoder{},
			Decoder: &codec.Decoder{MsgCreator: ctor.Options.DecoderMsgCreator},
			Timeout: ctor.Options.IOTimeout,
		},
		RetryTimes: ctor.Options.IORetryTimes,
	}
	handshake.Transceiver.SequencedBuff.Reset(0, 0, ctor.Options.IOSequencedBuffCap)

	var sessionId string
	cs := ctor.Options.EncCipherSuite
	cm := ctor.Options.Compression
	var cliRandom, servRandom []byte
	var cliHelloBytes, servHelloBytes []byte
	var continueFlow, encryptionFlow, authFlow bool

	defer func() {
		if cliRandom != nil {
			codec.BytesPool.Put(cliRandom)
		}
		if servRandom != nil {
			codec.BytesPool.Put(servRandom)
		}
		if cliHelloBytes != nil {
			codec.BytesPool.Put(cliHelloBytes)
		}
		if servHelloBytes != nil {
			codec.BytesPool.Put(servHelloBytes)
		}
	}()

	// 生成客户端随机数
	n, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		return err
	}
	servRandom = codec.BytesPool.Get(n.BitLen() / 8)
	n.FillBytes(servRandom)

	cliHello := protocol.Event[*transport.MsgHello]{
		Msg: &transport.MsgHello{
			Version:     transport.Version_V1_0,
			SessionId:   client.GetSessionId(),
			Random:      cliRandom,
			CipherSuite: cs,
			Compression: cm,
		},
	}

	// 与服务端互相hello
	err = handshake.ClientHello(cliHello,
		func(servHello protocol.Event[*transport.MsgHello]) error {
			// 检查HelloDone标记
			if !servHello.Flags.Is(transport.Flag_HelloDone) {
				return fmt.Errorf("the expected msg-hello-flag (0x%x) was not received", transport.Flag_HelloDone)
			}

			// 检查协议版本
			if servHello.Msg.Version != transport.Version_V1_0 {
				return fmt.Errorf("version %q not support", servHello.Msg.Version)
			}

			// 记录握手参数
			sessionId = strings.Clone(servHello.Msg.SessionId)
			cs = servHello.Msg.CipherSuite
			cm = servHello.Msg.Compression
			continueFlow = servHello.Flags.Is(transport.Flag_Continue)
			encryptionFlow = servHello.Flags.Is(transport.Flag_Encryption)
			authFlow = servHello.Flags.Is(transport.Flag_Auth)

			// 开启加密流程
			if encryptionFlow {
				// 记录服务端随机数
				if len(servHello.Msg.Random) < 0 {
					return errors.New("server Hello 'random' is empty")
				}
				servRandom = codec.BytesPool.Get(len(servHello.Msg.Random))
				copy(servRandom, servHello.Msg.Random)

				// 记录双方hello数据，用于ecdh后加密验证
				cliHelloBytes = codec.BytesPool.Get(cliHello.Msg.Size())
				if _, err := cliHello.Msg.Read(cliHelloBytes); err != nil {
					return err
				}

				servHelloBytes = codec.BytesPool.Get(servHello.Msg.Size())
				if _, err := servHello.Msg.Read(servHelloBytes); err != nil {
					return err
				}
			}

			return nil
		})
	if err != nil {
		return err
	}

	// 开启加密时，与服务端交换秘钥
	if encryptionFlow {
		err = ctor.secretKeyExchange(handshake, cs, cm, cliRandom, servRandom, cliHelloBytes, servHelloBytes, sessionId)
		if err != nil {
			return err
		}
	}

	// 开启鉴权时，向服务端发起鉴权
	if authFlow {
		err = handshake.ClientAuth(protocol.Event[*transport.MsgAuth]{
			Msg: &transport.MsgAuth{
				Token:      ctor.Options.AuthToken,
				Extensions: ctor.Options.AuthExtensions,
			},
		})
		if err != nil {
			return err
		}
	}

	// 暂停客户端的收发消息io，等握手结束后恢复
	client.pauseIO()
	defer client.continueIO()

	// 断线重连流程，需要交换序号，检测是否能补发消息
	if continueFlow {
		err = handshake.ClientContinue(protocol.Event[*transport.MsgContinue]{
			Msg: &transport.MsgContinue{
				SendSeq: client.transceiver.SequencedBuff.SendSeq,
				RecvSeq: client.transceiver.SequencedBuff.RecvSeq,
			},
		})
		if err != nil {
			return err
		}
	}

	var remoteSendSeq, remoteRecvSeq uint32

	// 等待服务端通知握手结束
	err = handshake.ClientFinished(func(finished protocol.Event[*transport.MsgFinished]) error {
		if encryptionFlow && !finished.Flags.Is(transport.Flag_EncryptOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", transport.Flag_EncryptOK)
		}

		if authFlow && !finished.Flags.Is(transport.Flag_AuthOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", transport.Flag_AuthOK)
		}

		if continueFlow && !finished.Flags.Is(transport.Flag_ContinueOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", transport.Flag_ContinueOK)
		}

		remoteSendSeq = finished.Msg.SendSeq
		remoteRecvSeq = finished.Msg.RecvSeq
		return nil
	})
	if err != nil {
		return err
	}

	if continueFlow {
		// 刷新客户端
		_, _, err = client.renew(conn, remoteRecvSeq)
		if err != nil {
			return err
		}
	} else {
		// 重置缓存
		handshake.Transceiver.SequencedBuff.Reset(remoteRecvSeq, remoteSendSeq, ctor.Options.IOSequencedBuffCap)

		// 初始化客户端
		client.init(handshake.Transceiver, sessionId)
	}

	return nil
}

// secretKeyExchange 秘钥交换过程
func (ctor *_Connector) secretKeyExchange(handshake *protocol.HandshakeProtocol, cs transport.CipherSuite, cm transport.Compression,
	cliRandom, servRandom, cliHelloBytes, servHelloBytes []byte, sessionId string) error {
	// 选择秘钥交换函数，并与客户端交换秘钥
	switch cs.SecretKeyExchange {
	case transport.SecretKeyExchange_ECDHE:
		// 临时共享秘钥
		var sharedKeyBytes []byte

		// 编码器与解码器的加密模块
		encEncryptionModule := &codec.EncryptionModule{}
		decEncryptionModule := &codec.EncryptionModule{}

		defer func() {
			encEncryptionModule.GC()
			decEncryptionModule.GC()
		}()

		// 与服务端交换秘钥
		err := handshake.ClientSecretKeyExchange(func(e protocol.Event[transport.Msg]) (protocol.Event[transport.Msg], error) {
			// 解包ECDHESecretKeyExchange消息事件
			switch e.Msg.MsgId() {
			case transport.MsgId_ECDHESecretKeyExchange:
				break
			default:
				return protocol.Event[transport.Msg]{}, fmt.Errorf("%w: %d", protocol.ErrUnexpectedMsg, e.Msg.MsgId())
			}
			servECDHE := protocol.UnpackEvent[*transport.MsgECDHESecretKeyExchange](e)

			// 验证服务端签名
			if ctor.Options.EncVerifyServerSignature {
				if !servECDHE.Flags.Is(transport.Flag_Signature) {
					return protocol.Event[transport.Msg]{}, errors.New("no server signature")
				}

				if err := ctor.verify(servECDHE.Msg.SignatureAlgorithm, servECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, servECDHE.Msg.PublicKey); err != nil {
					return protocol.Event[transport.Msg]{}, err
				}
			}

			// 创建曲线
			curve, err := method.NewNamedCurve(servECDHE.Msg.NamedCurve)
			if err != nil {
				return protocol.Event[transport.Msg]{}, err
			}

			// 生成客户端临时私钥
			cliPriv, err := curve.GenerateKey(rand.Reader)
			if err != nil {
				return protocol.Event[transport.Msg]{}, err
			}

			// 生成客户端临时公钥
			cliPub := cliPriv.PublicKey()
			cliPubBytes := cliPub.Bytes()

			// 服务端临时公钥
			servPub, err := curve.NewPublicKey(servECDHE.Msg.PublicKey)
			if err != nil {
				return protocol.Event[transport.Msg]{}, fmt.Errorf("server ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err)
			}

			// 临时共享秘钥
			sharedKeyBytes, err = cliPriv.ECDH(servPub)
			if err != nil {
				return protocol.Event[transport.Msg]{}, fmt.Errorf("ECDH failed, %s", err)
			}

			// 签名数据
			signature, err := ctor.sign(cs, cm, cliRandom, servRandom, sessionId, cliPubBytes)
			if err != nil {
				return protocol.Event[transport.Msg]{}, err
			}

			// 设置分组对齐填充方案
			if encEncryptionModule.Padding, err = ctor.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return protocol.Event[transport.Msg]{}, err
			}
			if decEncryptionModule.Padding, err = ctor.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return protocol.Event[transport.Msg]{}, err
			}

			// 设置nonce值
			if len(servECDHE.Msg.Nonce) > 0 && len(servECDHE.Msg.NonceStep) > 0 {
				var nonce, nonceStep big.Int
				nonce.SetBytes(servECDHE.Msg.Nonce)
				nonceStep.SetBytes(servECDHE.Msg.NonceStep)
				encEncryptionModule.FetchNonce = ctor.makeFetchNonce(&nonce, &nonceStep)
				decEncryptionModule.FetchNonce = ctor.makeFetchNonce(&nonce, &nonceStep)
			}

			// 创建并设置加解密流
			encryptor, decrypter, err := method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, servECDHE.Msg.IV)
			if err != nil {
				return protocol.Event[transport.Msg]{}, fmt.Errorf("new cipher stream failed, %s", err)
			}
			encEncryptionModule.Cipher = encryptor
			decEncryptionModule.Cipher = decrypter

			cliECDHE := protocol.Event[*transport.MsgECDHESecretKeyExchange]{
				Flags: transport.Flags_None().Setd(transport.Flag_Signature, len(signature) > 0),
				Msg: &transport.MsgECDHESecretKeyExchange{
					NamedCurve:         servECDHE.Msg.NamedCurve,
					PublicKey:          cliPubBytes,
					SignatureAlgorithm: ctor.Options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			}

			return protocol.PackEvent(cliECDHE), nil

		}, func(servChangeCipherSpec protocol.Event[*transport.MsgChangeCipherSpec]) (protocol.Event[*transport.MsgChangeCipherSpec], error) {
			verifyEncryption := servChangeCipherSpec.Flags.Is(transport.Flag_VerifyEncryption)

			// 验证加密是否正确
			if verifyEncryption {
				decryptedHello, err := decEncryptionModule.Transforming(nil, servChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, fmt.Errorf("decrypt hello failed, %s", err)
				}

				if bytes.Compare(decryptedHello, servHelloBytes) != 0 {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, errors.New("verify hello failed")
				}
			}

			cliChangeCipherSpec := protocol.Event[*transport.MsgChangeCipherSpec]{
				Flags: transport.Flags_None().Setd(transport.Flag_VerifyEncryption, verifyEncryption),
				Msg:   &transport.MsgChangeCipherSpec{},
			}

			// 加密hello消息
			if verifyEncryption {
				encryptedHello, err := encEncryptionModule.Transforming(nil, cliHelloBytes)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, fmt.Errorf("encrypt hello failed, %s", err)
				}
				cliChangeCipherSpec.Msg.EncryptedHello = encryptedHello
			}

			return cliChangeCipherSpec, nil
		})
		if err != nil {
			return err
		}

		// 编码器
		encoder := &codec.Encoder{
			EncryptionModule: encEncryptionModule,
			Encryption:       true,
		}

		if encoder.MACModule, encoder.PatchMAC, err = ctor.makeMACModule(cs.MACHash, sharedKeyBytes); err != nil {
			return err
		}

		if encoder.CompressionModule, encoder.CompressedSize, err = ctor.makeCompressionModule(cm); err != nil {
			return err
		}

		// 解码器
		decoder := &codec.Decoder{
			MsgCreator:       handshake.Transceiver.Decoder.GetMsgCreator(),
			EncryptionModule: decEncryptionModule,
		}

		if decoder.MACModule, _, err = ctor.makeMACModule(cs.MACHash, sharedKeyBytes); err != nil {
			return err
		}

		if decoder.CompressionModule, _, err = ctor.makeCompressionModule(cm); err != nil {
			return err
		}

		// 更换解编码器
		handshake.Transceiver.Encoder = encoder
		handshake.Transceiver.Decoder = decoder

		return nil
	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange %d not support", cs.SecretKeyExchange)
	}

	return nil
}

// makeFetchNonce 构造获取nonce值函数
func (ctor *_Connector) makeFetchNonce(nonce, nonceStep *big.Int) codec.FetchNonce {
	if nonce == nil {
		return nil
	}

	encryptionNonce := new(big.Int).Set(nonce)
	encryptionNonceNonceBuff := encryptionNonce.Bytes()

	bits := nonce.BitLen()

	return func() ([]byte, error) {
		if nonceStep == nil || nonceStep.Sign() == 0 {
			return encryptionNonceNonceBuff, nil
		}

		encryptionNonce.Add(encryptionNonce, nonceStep)
		if encryptionNonce.BitLen() > bits {
			encryptionNonce.SetInt64(0)
		}
		encryptionNonce.FillBytes(encryptionNonceNonceBuff)

		return encryptionNonceNonceBuff, nil
	}
}

// makePaddingMode 构造填充方案
func (ctor *_Connector) makePaddingMode(bcm transport.BlockCipherMode, paddingMode transport.PaddingMode) (method.Padding, error) {
	if !bcm.Padding() {
		return nil, nil
	}

	if paddingMode == transport.PaddingMode_None {
		return nil, fmt.Errorf("CipherSuite.BlockCipherMode %d, plaintext padding is necessary", bcm)
	}

	padding, err := method.NewPadding(paddingMode)
	if err != nil {
		return nil, err
	}

	return padding, nil
}

// makeMACModule 构造MAC模块
func (ctor *_Connector) makeMACModule(hash transport.Hash, sharedKeyBytes []byte) (codec.IMACModule, bool, error) {
	if hash.Bits() <= 0 {
		return nil, false, nil
	}

	var macModule codec.IMACModule

	switch hash.Bits() {
	case 32:
		macHash, err := method.NewHash32(hash)
		if err != nil {
			return nil, false, err
		}
		macModule = &codec.MAC32Module{
			Hash:       macHash,
			PrivateKey: sharedKeyBytes,
		}
	case 64:
		macHash, err := method.NewHash64(hash)
		if err != nil {
			return nil, false, err
		}
		macModule = &codec.MAC64Module{
			Hash:       macHash,
			PrivateKey: sharedKeyBytes,
		}
	default:
		macHash, err := method.NewHash(hash)
		if err != nil {
			return nil, false, err
		}
		macModule = &codec.MACModule{
			Hash:       macHash,
			PrivateKey: sharedKeyBytes,
		}
	}
	return macModule, true, nil
}

// makeCompressionModule 构造压缩模块
func (ctor *_Connector) makeCompressionModule(compression transport.Compression) (codec.ICompressionModule, int, error) {
	if compression == transport.Compression_None {
		return nil, 0, nil
	}

	compressionStream, err := method.NewCompressionStream(compression)
	if err != nil {
		return nil, 0, err
	}

	compressionModule := &codec.CompressionModule{
		CompressionStream: compressionStream,
	}

	return compressionModule, ctor.Options.CompressedSize, err
}

// sign 签名
func (ctor *_Connector) sign(cs transport.CipherSuite, cm transport.Compression, cliRandom, servRandom []byte, sessionId string, cliPubBytes []byte) ([]byte, error) {
	if ctor.Options.EncSignatureAlgorithm.AsymmetricEncryption == transport.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if ctor.Options.EncSignaturePrivateKey == nil {
		return nil, errors.New("option EncSignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		ctor.Options.EncSignatureAlgorithm.AsymmetricEncryption,
		ctor.Options.EncSignatureAlgorithm.PaddingMode,
		ctor.Options.EncSignatureAlgorithm.Hash)
	if err != nil {
		return nil, err
	}

	// 签名数据
	signBuf := bytes.NewBuffer(nil)
	signBuf.ReadFrom(&cs)
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionId)
	signBuf.Write(cliPubBytes)

	// 生成签名
	signature, err := signer.Sign(ctor.Options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 验证签名
func (ctor *_Connector) verify(signatureAlgorithm transport.SignatureAlgorithm, signature []byte, cs transport.CipherSuite, cm transport.Compression, cliRandom, servRandom []byte, sessionId string, servPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if ctor.Options.EncVerifySignaturePublicKey == nil {
		return errors.New("option EncVerifySignaturePublicKey is nil, unable to perform the verify signature operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		signatureAlgorithm.AsymmetricEncryption,
		signatureAlgorithm.PaddingMode,
		signatureAlgorithm.Hash)
	if err != nil {
		return err
	}

	// 签名数据
	signBuf := bytes.NewBuffer(nil)
	signBuf.ReadFrom(&cs)
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionId)
	signBuf.Write(servPubBytes)

	return signer.Verify(ctor.Options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
