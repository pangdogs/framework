package gtp_gate

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/binaryutil"
	"kit.golaxy.org/plugins/gtp/codec"
	"kit.golaxy.org/plugins/gtp/method"
	"kit.golaxy.org/plugins/gtp/transport"
	"kit.golaxy.org/plugins/internal"
	"math/big"
	"net"
	"strings"
)

// handshake 握手过程
func (acc *_Acceptor) handshake(conn net.Conn) (*_Session, error) {
	// 编解码器
	acc.encoder = &codec.Encoder{}
	acc.decoder = &codec.Decoder{MsgCreator: acc.Options.DecoderMsgCreator}

	// 握手协议
	handshake := &transport.HandshakeProtocol{
		Transceiver: &transport.Transceiver{
			Conn:    conn,
			Encoder: acc.encoder,
			Decoder: acc.decoder,
			Timeout: acc.Options.IOTimeout,
			Buffer:  &transport.UnsequencedBuffer{},
		},
		RetryTimes: acc.Options.IORetryTimes,
	}
	defer handshake.Transceiver.Clean()

	var cs gtp.CipherSuite
	var cm gtp.Compression
	var cliRandom, servRandom []byte
	var cliHelloBytes, servHelloBytes []byte
	var continueFlow, encryptionFlow, authFlow bool
	var session *_Session

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

	// 与客户端互相hello
	err := handshake.ServerHello(func(cliHello transport.Event[*gtp.MsgHello]) (transport.Event[*gtp.MsgHello], error) {
		// 检查协议版本
		if cliHello.Msg.Version != gtp.Version_V1_0 {
			return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
				Code:    gtp.Code_VersionError,
				Message: fmt.Sprintf("version %q not support", cliHello.Msg.Version),
			}
		}

		// 检查客户端要求的会话是否存在，已存在需要走断线重连流程
		if cliHello.Msg.SessionId != "" {
			v, ok := acc.Gate.loadSession(cliHello.Msg.SessionId)
			if !ok {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_SessionNotFound,
					Message: fmt.Sprintf("session %q not exist", cliHello.Msg.SessionId),
				}
			}

			session = v
			continueFlow = true
		} else {
			v, err := acc.newSession(conn)
			if err != nil {
				return transport.Event[*gtp.MsgHello]{}, err
			}

			// 调整会话状态为握手中
			v.setState(SessionState_Handshake)

			session = v
			continueFlow = false
		}

		// 检查是否同意使用客户端建议的加密方案
		if acc.Options.AgreeClientEncryptionProposal {
			cs = cliHello.Msg.CipherSuite
		} else {
			cs = acc.Options.EncCipherSuite
		}

		// 检查是否同意使用客户端建议的压缩方案
		if acc.Options.AgreeClientCompressionProposal {
			cm = cliHello.Msg.Compression
		} else {
			cm = acc.Options.Compression
		}

		// 开启加密时，需要交换随机数
		if cs.SecretKeyExchange != gtp.SecretKeyExchange_None {
			// 记录客户端随机数
			if len(cliHello.Msg.Random) < 0 {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: "client Hello 'random' is empty",
				}
			}
			cliRandom = binaryutil.BytesPool.Get(len(cliHello.Msg.Random))
			copy(cliRandom, cliHello.Msg.Random)

			// 生成服务端随机数
			n, err := rand.Prime(rand.Reader, 256)
			if err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			servRandom = binaryutil.BytesPool.Get(n.BitLen() / 8)
			n.FillBytes(servRandom)

			encryptionFlow = true
		}

		// 返回服务端Hello
		servHello := transport.Event[*gtp.MsgHello]{
			Flags: gtp.Flags(gtp.Flag_HelloDone),
			Msg: &gtp.MsgHello{
				Version:     gtp.Version_V1_0,
				SessionId:   session.GetId(),
				Random:      servRandom,
				CipherSuite: cs,
				Compression: cm,
			},
		}

		authFlow = len(acc.Options.ClientAuthHandlers) > 0

		// 标记是否开启加密
		servHello.Flags.Set(gtp.Flag_Encryption, encryptionFlow)
		// 标记是否开启鉴权
		servHello.Flags.Set(gtp.Flag_Auth, authFlow)
		// 标记是否走断线重连流程
		servHello.Flags.Set(gtp.Flag_Continue, continueFlow)

		// 开启加密时，记录双方hello数据，用于ecdh后加密验证
		if encryptionFlow {
			cliHelloBytes = binaryutil.BytesPool.Get(cliHello.Msg.Size())
			if _, err := cliHello.Msg.Read(cliHelloBytes); err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}

			servHelloBytes = binaryutil.BytesPool.Get(servHello.Msg.Size())
			if _, err := servHello.Msg.Read(servHelloBytes); err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
		}

		return servHello, nil
	})
	if err != nil {
		return nil, err
	}

	// 开启加密时，与客户端交换秘钥
	if encryptionFlow {
		err = acc.secretKeyExchange(handshake, cs, cm, cliRandom, servRandom, cliHelloBytes, servHelloBytes, session.GetId())
		if err != nil {
			return nil, err
		}
	}

	// 安装压缩模块
	err = acc.setupCompressionModule(cm)
	if err != nil {
		return nil, err
	}

	var token string

	// 开启鉴权时，鉴权客户端
	if authFlow {
		err = handshake.ServerAuth(func(e transport.Event[*gtp.MsgAuth]) error {
			// 断线重连流程，检查会话Id与token是否匹配，防止hack客户端猜测会话Id，恶意通过断线重连登录
			if continueFlow {
				if e.Msg.Token != session.GetToken() {
					return &transport.RstError{
						Code:    gtp.Code_AuthFailed,
						Message: "incorrect token",
					}
				}
			}

			for i := range acc.Options.ClientAuthHandlers {
				handler := acc.Options.ClientAuthHandlers[i]
				if handler == nil {
					continue
				}
				err := internal.Call(func() error { return handler(acc.Gate.ctx, conn, e.Msg.Token, e.Msg.Extensions) })
				if err != nil {
					return &transport.RstError{
						Code:    gtp.Code_AuthFailed,
						Message: err.Error(),
					}
				}
			}

			token = strings.Clone(e.Msg.Token)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	// 暂停会话的收发消息io，等握手结束后恢复
	session.pauseIO()
	defer session.continueIO()

	var sendSeq, recvSeq uint32

	// 断线重连流程，需要交换序号，检测是否能补发消息
	if continueFlow {
		err = handshake.ServerContinue(func(e transport.Event[*gtp.MsgContinue]) error {
			// 刷新会话
			sendSeq, recvSeq, err = session.renew(handshake.Transceiver.Conn, e.Msg.RecvSeq)
			if err != nil {
				return &transport.RstError{
					Code:    gtp.Code_ContinueFailed,
					Message: err.Error(),
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 初始化会话
		sendSeq, recvSeq = session.init(handshake.Transceiver.Conn,
			handshake.Transceiver.Encoder,
			handshake.Transceiver.Decoder,
			token)
	}

	// 通知客户端握手结束
	err = handshake.ServerFinished(transport.Event[*gtp.MsgFinished]{
		Flags: gtp.Flags_None().
			Setd(gtp.Flag_EncryptOK, encryptionFlow).
			Setd(gtp.Flag_AuthOK, authFlow).
			Setd(gtp.Flag_ContinueOK, continueFlow),
		Msg: &gtp.MsgFinished{
			SendSeq: sendSeq,
			RecvSeq: recvSeq,
		},
	})
	if err != nil {
		return nil, err
	}

	if continueFlow {
		// 检测会话有效性
		if !acc.Gate.validateSession(session) {
			err = &transport.RstError{
				Code:    gtp.Code_ContinueFailed,
				Message: fmt.Sprintf("session %q has expired", session.GetId()),
			}

			ctrl := transport.CtrlProtocol{
				Transceiver: handshake.Transceiver,
				RetryTimes:  handshake.RetryTimes,
			}
			ctrl.SendRst(err)

			return nil, err
		}
	} else {
		// 存储会话
		acc.Gate.storeSession(session)

		// 调整会话状态为已确认
		session.setState(SessionState_Confirmed)

		// 运行会话
		go session.run()
	}

	return session, nil
}

// secretKeyExchange 秘钥交换过程
func (acc *_Acceptor) secretKeyExchange(handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom, cliHelloBytes, servHelloBytes []byte, sessionId string) (err error) {
	// 控制协议
	ctrl := transport.CtrlProtocol{
		Transceiver: handshake.Transceiver,
		RetryTimes:  handshake.RetryTimes,
	}

	// 是否已发送rst
	rstSent := false

	defer func() {
		if err != nil && !rstSent {
			ctrl.SendRst(&transport.RstError{
				Code:    gtp.Code_EncryptFailed,
				Message: err.Error(),
			})
		}
	}()

	// 选择秘钥交换函数，并与客户端交换秘钥
	switch cs.SecretKeyExchange {
	case gtp.SecretKeyExchange_ECDHE:
		// 创建曲线
		curve, err := method.NewNamedCurve(acc.Options.EncECDHENamedCurve)
		if err != nil {
			return err
		}

		// 生成服务端临时私钥
		servPriv, err := curve.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}

		// 生成服务端临时公钥
		servPub := servPriv.PublicKey()
		servPubBytes := servPub.Bytes()

		// 签名数据
		signature, err := acc.sign(cs, cm, cliRandom, servRandom, sessionId, servPubBytes)
		if err != nil {
			return err
		}

		// 编码器与解码器的加密模块
		encEncryptionModule := &codec.EncryptionModule{}
		decEncryptionModule := &codec.EncryptionModule{}

		defer func() {
			encEncryptionModule.GC()
			decEncryptionModule.GC()
		}()

		// 设置分组对齐填充方案
		if encEncryptionModule.Padding, err = acc.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
			return err
		}
		if decEncryptionModule.Padding, err = acc.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
			return err
		}

		// 创建iv值
		iv, err := acc.makeIV(cs.SymmetricEncryption, cs.BlockCipherMode)
		if err != nil {
			return err
		}

		// 创建nonce值
		nonce, err := acc.makeNonce(cs.SymmetricEncryption, cs.BlockCipherMode)
		if err != nil {
			return err
		}

		var ivBytes, nonceBytes, nonceStepBytes []byte

		if iv != nil {
			ivBytes = iv.Bytes()
		}

		if nonce != nil {
			nonceBytes = nonce.Bytes()
			nonceStepBytes = acc.Options.EncNonceStep.Bytes()
			encEncryptionModule.FetchNonce = acc.makeFetchNonce(nonce, acc.Options.EncNonceStep)
			decEncryptionModule.FetchNonce = acc.makeFetchNonce(nonce, acc.Options.EncNonceStep)
		}

		// 临时共享秘钥
		var sharedKeyBytes []byte

		// 与客户端交换秘钥
		err = handshake.ServerECDHESecretKeyExchange(
			transport.Event[*gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: &gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         acc.Options.EncECDHENamedCurve,
					PublicKey:          servPubBytes,
					IV:                 ivBytes,
					Nonce:              nonceBytes,
					NonceStep:          nonceStepBytes,
					SignatureAlgorithm: acc.Options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			},
			func(cliECDHE transport.Event[*gtp.MsgECDHESecretKeyExchange]) (transport.Event[*gtp.MsgChangeCipherSpec], error) {
				// 检查客户端曲线类型
				if cliECDHE.Msg.NamedCurve != acc.Options.EncECDHENamedCurve {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'NamedCurve' %d is incorrect", cliECDHE.Msg.NamedCurve),
					}
				}

				// 验证客户端签名
				if acc.Options.EncVerifyClientSignature {
					if !cliECDHE.Flags.Is(gtp.Flag_Signature) {
						return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
							Code:    gtp.Code_EncryptFailed,
							Message: "no client signature",
						}
					}

					if err := acc.verify(cliECDHE.Msg.SignatureAlgorithm, cliECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, cliECDHE.Msg.PublicKey); err != nil {
						return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
							Code:    gtp.Code_EncryptFailed,
							Message: err.Error(),
						}
					}
				}

				// 客户端临时公钥
				cliPub, err := curve.NewPublicKey(cliECDHE.Msg.PublicKey)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err),
					}
				}

				// 临时共享秘钥
				sharedKeyBytes, err = servPriv.ECDH(cliPub)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("ECDH failed, %s", err),
					}
				}

				// 创建并设置加解密流
				encryptor, decrypter, err := method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, ivBytes)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("new cipher stream failed, %s", err),
					}
				}
				encEncryptionModule.Cipher = encryptor
				decEncryptionModule.Cipher = decrypter

				// 加密hello消息
				encryptedHello, err := encEncryptionModule.Transforming(nil, servHelloBytes)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("encrypt hello failed, %s", err),
					}
				}

				return transport.Event[*gtp.MsgChangeCipherSpec]{
					Flags: gtp.Flags(gtp.Flag_VerifyEncryption),
					Msg: &gtp.MsgChangeCipherSpec{
						EncryptedHello: encryptedHello,
					},
				}, nil
			}, func(cliChangeCipherSpec transport.Event[*gtp.MsgChangeCipherSpec]) error {
				// 客户端要求不验证加密
				if !cliChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption) {
					return nil
				}

				// 验证加密是否正确
				decryptedHello, err := decEncryptionModule.Transforming(nil, cliChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("decrypt hello failed, %s", err),
					}
				}

				if bytes.Compare(decryptedHello, cliHelloBytes) != 0 {
					return &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: "verify hello failed",
					}
				}

				return nil
			})
		if err != nil {
			rstSent = true
			return err
		}

		// 安装加密模块
		acc.setupEncryptionModule(encEncryptionModule, decEncryptionModule)

		// 安装MAC模块
		return acc.setupMACModule(cs.MACHash, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange %d not support", cs.SecretKeyExchange)
	}

	return nil
}

// setupCompressionModule 安装压缩模块
func (acc *_Acceptor) setupCompressionModule(cm gtp.Compression) (err error) {
	if acc.encoder.CompressionModule, acc.encoder.CompressedSize, err = acc.makeCompressionModule(cm); err != nil {
		return err
	}
	if acc.decoder.CompressionModule, _, err = acc.makeCompressionModule(cm); err != nil {
		return err
	}
	return nil
}

// setupEncryptionModule 安装加密模块
func (acc *_Acceptor) setupEncryptionModule(encEncryptionModule, decEncryptionModule *codec.EncryptionModule) {
	acc.encoder.EncryptionModule = encEncryptionModule
	acc.encoder.Encryption = true
	acc.decoder.EncryptionModule = decEncryptionModule
}

// setupMACModule 安装MAC模块
func (acc *_Acceptor) setupMACModule(hash gtp.Hash, sharedKeyBytes []byte) (err error) {
	if acc.encoder.MACModule, acc.encoder.PatchMAC, err = acc.makeMACModule(hash, sharedKeyBytes); err != nil {
		return err
	}
	if acc.decoder.MACModule, _, err = acc.makeMACModule(hash, sharedKeyBytes); err != nil {
		return err
	}
	return nil
}

// makeIV 构造iv值
func (acc *_Acceptor) makeIV(se gtp.SymmetricEncryption, bcm gtp.BlockCipherMode) (*big.Int, error) {
	size, ok := se.IV()
	if !ok {
		if !se.BlockCipherMode() || !bcm.IV() {
			return nil, nil
		}
		size, ok = se.BlockSize()
		if !ok {
			return nil, fmt.Errorf("CipherSuite.BlockCipherMode %d needs IV, but CipherSuite.SymmetricEncryption %d lacks a fixed block size", bcm, se)
		}
	}

	iv, err := rand.Prime(rand.Reader, size*8)
	if err != nil {
		return nil, err
	}

	return iv, nil
}

// makeNonce 构造nonce值
func (acc *_Acceptor) makeNonce(se gtp.SymmetricEncryption, bcm gtp.BlockCipherMode) (*big.Int, error) {
	size, ok := se.Nonce()
	if !ok {
		if !se.BlockCipherMode() || !bcm.Nonce() {
			return nil, nil
		}
		size, ok = se.BlockSize()
		if !ok {
			return nil, fmt.Errorf("CipherSuite.BlockCipherMode %d needs Nonce, but CipherSuite.SymmetricEncryption %d lacks a fixed block size", bcm, se)
		}
	}

	nonce, err := rand.Prime(rand.Reader, size*8)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

// makeFetchNonce 构造获取nonce值函数
func (acc *_Acceptor) makeFetchNonce(nonce, nonceStep *big.Int) codec.FetchNonce {
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
func (acc *_Acceptor) makePaddingMode(bcm gtp.BlockCipherMode, paddingMode gtp.PaddingMode) (method.Padding, error) {
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
func (acc *_Acceptor) makeMACModule(hash gtp.Hash, sharedKeyBytes []byte) (codec.IMACModule, bool, error) {
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
func (acc *_Acceptor) makeCompressionModule(compression gtp.Compression) (codec.ICompressionModule, int, error) {
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

	return compressionModule, acc.Options.CompressedSize, err
}

// sign 签名
func (acc *_Acceptor) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId string, servPubBytes []byte) ([]byte, error) {
	// 无需签名
	if acc.Options.EncSignatureAlgorithm.AsymmetricEncryption == gtp.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if acc.Options.EncSignaturePrivateKey == nil {
		return nil, errors.New("option EncSignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		acc.Options.EncSignatureAlgorithm.AsymmetricEncryption,
		acc.Options.EncSignatureAlgorithm.PaddingMode,
		acc.Options.EncSignatureAlgorithm.Hash)
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
	signBuf.Write(servPubBytes)

	// 生成签名
	signature, err := signer.Sign(acc.Options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 验证签名
func (acc *_Acceptor) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId string, cliPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if acc.Options.EncVerifySignaturePublicKey == nil {
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
	signBuf.Write(cliPubBytes)

	return signer.Verify(acc.Options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
