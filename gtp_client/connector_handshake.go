package gtp_client

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/method"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/util/binaryutil"
	"math/big"
	"net"
	"strings"
)

// handshake 握手过程
func (ctor *_Connector) handshake(conn net.Conn, client *Client) error {
	// 编解码器
	ctor.encoder = &codec.Encoder{}
	ctor.decoder = &codec.Decoder{MsgCreator: ctor.options.DecoderMsgCreator}

	// 握手协议
	handshake := &transport.HandshakeProtocol{
		Transceiver: &transport.Transceiver{
			Conn:    conn,
			Encoder: ctor.encoder,
			Decoder: ctor.decoder,
			Timeout: ctor.options.IOTimeout,
			Buffer:  &transport.UnsequencedBuffer{},
		},
		RetryTimes: ctor.options.IORetryTimes,
	}
	defer handshake.Transceiver.Clean()

	var sessionId string
	cs := ctor.options.EncCipherSuite
	cm := ctor.options.Compression
	var cliRandom, servRandom []byte
	var cliHelloBytes, servHelloBytes []byte
	var continueFlow, encryptionFlow, authFlow bool

	defer func() {
		if cliRandom != nil {
			binaryutil.BytesPool.Put(cliRandom)
		}
		if servRandom != nil {
			binaryutil.BytesPool.Put(servRandom)
		}
		if cliHelloBytes != nil {
			binaryutil.BytesPool.Put(cliHelloBytes)
		}
		if servHelloBytes != nil {
			binaryutil.BytesPool.Put(servHelloBytes)
		}
	}()

	// 生成客户端随机数
	n, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		return err
	}
	servRandom = binaryutil.BytesPool.Get(n.BitLen() / 8)
	n.FillBytes(servRandom)

	cliHello := transport.Event[gtp.MsgHello]{
		Msg: gtp.MsgHello{
			Version:     gtp.Version_V1_0,
			SessionId:   client.GetSessionId(),
			Random:      cliRandom,
			CipherSuite: cs,
			Compression: cm,
		},
	}

	// 与服务端互相hello
	err = handshake.ClientHello(cliHello,
		func(servHello transport.Event[gtp.MsgHello]) error {
			// 检查HelloDone标记
			if !servHello.Flags.Is(gtp.Flag_HelloDone) {
				return fmt.Errorf("the expected msg-hello-flag (0x%x) was not received", gtp.Flag_HelloDone)
			}

			// 检查协议版本
			if servHello.Msg.Version != gtp.Version_V1_0 {
				return fmt.Errorf("version %q not support", servHello.Msg.Version)
			}

			// 记录握手参数
			sessionId = strings.Clone(servHello.Msg.SessionId)
			cs = servHello.Msg.CipherSuite
			cm = servHello.Msg.Compression
			continueFlow = servHello.Flags.Is(gtp.Flag_Continue)
			encryptionFlow = servHello.Flags.Is(gtp.Flag_Encryption)
			authFlow = servHello.Flags.Is(gtp.Flag_Auth)

			// 开启加密流程
			if encryptionFlow {
				// 记录服务端随机数
				if len(servHello.Msg.Random) < 0 {
					return errors.New("server Hello 'random' is empty")
				}
				servRandom = binaryutil.BytesPool.Get(len(servHello.Msg.Random))
				copy(servRandom, servHello.Msg.Random)

				// 记录双方hello数据，用于ecdh后加密验证
				cliHelloBytes = binaryutil.BytesPool.Get(cliHello.Msg.Size())
				if _, err := cliHello.Msg.Read(cliHelloBytes); err != nil {
					return err
				}

				servHelloBytes = binaryutil.BytesPool.Get(servHello.Msg.Size())
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

	// 安装压缩模块
	err = ctor.setupCompressionModule(cm)
	if err != nil {
		return err
	}

	// 开启鉴权时，向服务端发起鉴权
	if authFlow {
		err = handshake.ClientAuth(transport.Event[gtp.MsgAuth]{
			Msg: gtp.MsgAuth{
				UserId:     ctor.options.AuthUserId,
				Token:      ctor.options.AuthToken,
				Extensions: ctor.options.AuthExtensions,
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
		err = handshake.ClientContinue(transport.Event[gtp.MsgContinue]{
			Msg: gtp.MsgContinue{
				SendSeq: client.transceiver.Buffer.SendSeq(),
				RecvSeq: client.transceiver.Buffer.RecvSeq(),
			},
		})
		if err != nil {
			return err
		}
	}

	var remoteSendSeq, remoteRecvSeq uint32

	// 等待服务端通知握手结束
	err = handshake.ClientFinished(func(finished transport.Event[gtp.MsgFinished]) error {
		if encryptionFlow && !finished.Flags.Is(gtp.Flag_EncryptOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", gtp.Flag_EncryptOK)
		}

		if authFlow && !finished.Flags.Is(gtp.Flag_AuthOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", gtp.Flag_AuthOK)
		}

		if continueFlow && !finished.Flags.Is(gtp.Flag_ContinueOK) {
			return fmt.Errorf("the expected msg-finished-flag (0x%x) was not received", gtp.Flag_ContinueOK)
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
		// 初始化客户端
		client.init(handshake.Transceiver.Conn,
			handshake.Transceiver.Encoder,
			handshake.Transceiver.Decoder,
			remoteSendSeq,
			remoteRecvSeq,
			sessionId)
	}

	return nil
}

// secretKeyExchange 秘钥交换过程
func (ctor *_Connector) secretKeyExchange(handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom, cliHelloBytes, servHelloBytes []byte, sessionId string) error {
	// 选择秘钥交换函数，并与客户端交换秘钥
	switch cs.SecretKeyExchange {
	case gtp.SecretKeyExchange_ECDHE:
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
		err := handshake.ClientSecretKeyExchange(func(e transport.Event[gtp.Msg]) (transport.Event[gtp.MsgReader], error) {
			// 解包ECDHESecretKeyExchange消息事件
			switch e.Msg.MsgId() {
			case gtp.MsgId_ECDHESecretKeyExchange:
				break
			default:
				return transport.Event[gtp.MsgReader]{}, fmt.Errorf("%w (%d)", transport.ErrUnexpectedMsg, e.Msg.MsgId())
			}
			servECDHE := transport.UnpackEvent[gtp.MsgECDHESecretKeyExchange](e)

			// 验证服务端签名
			if ctor.options.EncVerifyServerSignature {
				if !servECDHE.Flags.Is(gtp.Flag_Signature) {
					return transport.Event[gtp.MsgReader]{}, errors.New("no server signature")
				}

				if err := ctor.verify(servECDHE.Msg.SignatureAlgorithm, servECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, servECDHE.Msg.PublicKey); err != nil {
					return transport.Event[gtp.MsgReader]{}, err
				}
			}

			// 创建曲线
			curve, err := method.NewNamedCurve(servECDHE.Msg.NamedCurve)
			if err != nil {
				return transport.Event[gtp.MsgReader]{}, err
			}

			// 生成客户端临时私钥
			cliPriv, err := curve.GenerateKey(rand.Reader)
			if err != nil {
				return transport.Event[gtp.MsgReader]{}, err
			}

			// 生成客户端临时公钥
			cliPub := cliPriv.PublicKey()
			cliPubBytes := cliPub.Bytes()

			// 服务端临时公钥
			servPub, err := curve.NewPublicKey(servECDHE.Msg.PublicKey)
			if err != nil {
				return transport.Event[gtp.MsgReader]{}, fmt.Errorf("server ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err)
			}

			// 临时共享秘钥
			sharedKeyBytes, err = cliPriv.ECDH(servPub)
			if err != nil {
				return transport.Event[gtp.MsgReader]{}, fmt.Errorf("ECDH failed, %s", err)
			}

			// 签名数据
			signature, err := ctor.sign(cs, cm, cliRandom, servRandom, sessionId, cliPubBytes)
			if err != nil {
				return transport.Event[gtp.MsgReader]{}, err
			}

			// 设置分组对齐填充方案
			if encEncryptionModule.Padding, err = ctor.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return transport.Event[gtp.MsgReader]{}, err
			}
			if decEncryptionModule.Padding, err = ctor.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return transport.Event[gtp.MsgReader]{}, err
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
				return transport.Event[gtp.MsgReader]{}, fmt.Errorf("new cipher stream failed, %s", err)
			}
			encEncryptionModule.Cipher = encryptor
			decEncryptionModule.Cipher = decrypter

			cliECDHE := transport.Event[gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         servECDHE.Msg.NamedCurve,
					PublicKey:          cliPubBytes,
					SignatureAlgorithm: ctor.options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			}

			return cliECDHE.Pack(), nil

		}, func(servChangeCipherSpec transport.Event[gtp.MsgChangeCipherSpec]) (transport.Event[gtp.MsgChangeCipherSpec], error) {
			verifyEncryption := servChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption)

			// 验证加密是否正确
			if verifyEncryption {
				decryptedHello, err := decEncryptionModule.Transforming(nil, servChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, fmt.Errorf("decrypt hello failed, %s", err)
				}

				if bytes.Compare(decryptedHello, servHelloBytes) != 0 {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, errors.New("verify hello failed")
				}
			}

			cliChangeCipherSpec := transport.Event[gtp.MsgChangeCipherSpec]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_VerifyEncryption, verifyEncryption),
			}

			// 加密hello消息
			if verifyEncryption {
				encryptedHello, err := encEncryptionModule.Transforming(nil, cliHelloBytes)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, fmt.Errorf("encrypt hello failed, %s", err)
				}
				cliChangeCipherSpec.Msg.EncryptedHello = encryptedHello
			}

			return cliChangeCipherSpec, nil
		})
		if err != nil {
			return err
		}

		// 安装加密模块
		ctor.setupEncryptionModule(encEncryptionModule, decEncryptionModule)

		// 安装MAC模块
		return ctor.setupMACModule(cs.MACHash, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange %d not support", cs.SecretKeyExchange)
	}

	return nil
}

// setupCompressionModule 安装压缩模块
func (ctor *_Connector) setupCompressionModule(cm gtp.Compression) (err error) {
	if ctor.encoder.CompressionModule, ctor.encoder.CompressedSize, err = ctor.makeCompressionModule(cm); err != nil {
		return err
	}
	if ctor.decoder.CompressionModule, _, err = ctor.makeCompressionModule(cm); err != nil {
		return err
	}
	return nil
}

// setupEncryptionModule 安装加密模块
func (ctor *_Connector) setupEncryptionModule(encEncryptionModule, decEncryptionModule *codec.EncryptionModule) {
	ctor.encoder.EncryptionModule = encEncryptionModule
	ctor.encoder.Encryption = true
	ctor.decoder.EncryptionModule = decEncryptionModule
}

// setupMACModule 安装MAC模块
func (ctor *_Connector) setupMACModule(hash gtp.Hash, sharedKeyBytes []byte) (err error) {
	if ctor.encoder.MACModule, ctor.encoder.PatchMAC, err = ctor.makeMACModule(hash, sharedKeyBytes); err != nil {
		return err
	}
	if ctor.decoder.MACModule, _, err = ctor.makeMACModule(hash, sharedKeyBytes); err != nil {
		return err
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
func (ctor *_Connector) makePaddingMode(bcm gtp.BlockCipherMode, paddingMode gtp.PaddingMode) (method.Padding, error) {
	if !bcm.Padding() {
		return nil, nil
	}

	if paddingMode == gtp.PaddingMode_None {
		return nil, fmt.Errorf("CipherSuite.BlockCipherMode %d, plaintext padding is necessary", bcm)
	}

	padding, err := method.NewPadding(paddingMode)
	if err != nil {
		return nil, err
	}

	return padding, nil
}

// makeMACModule 构造MAC模块
func (ctor *_Connector) makeMACModule(hash gtp.Hash, sharedKeyBytes []byte) (codec.IMACModule, bool, error) {
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
func (ctor *_Connector) makeCompressionModule(compression gtp.Compression) (codec.ICompressionModule, int, error) {
	if compression == gtp.Compression_None {
		return nil, 0, nil
	}

	compressionStream, err := method.NewCompressionStream(compression)
	if err != nil {
		return nil, 0, err
	}

	compressionModule := &codec.CompressionModule{
		CompressionStream: compressionStream,
	}

	return compressionModule, ctor.options.CompressedSize, err
}

// sign 签名
func (ctor *_Connector) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId string, cliPubBytes []byte) ([]byte, error) {
	if ctor.options.EncSignatureAlgorithm.AsymmetricEncryption == gtp.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if ctor.options.EncSignaturePrivateKey == nil {
		return nil, errors.New("option EncSignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		ctor.options.EncSignatureAlgorithm.AsymmetricEncryption,
		ctor.options.EncSignatureAlgorithm.PaddingMode,
		ctor.options.EncSignatureAlgorithm.Hash)
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
	signature, err := signer.Sign(ctor.options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 验证签名
func (ctor *_Connector) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId string, servPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if ctor.options.EncVerifySignaturePublicKey == nil {
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

	return signer.Verify(ctor.options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
