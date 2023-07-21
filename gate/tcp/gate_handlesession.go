package tcp

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"kit.golaxy.org/golaxy/util"
	"kit.golaxy.org/plugins/logger"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/codec"
	"kit.golaxy.org/plugins/transport/method"
	"kit.golaxy.org/plugins/transport/protocol"
	"math/big"
	math_rand "math/rand"
	"net"
	"strings"
)

func (g *_TcpGate) handleSession(conn net.Conn) {
	var err error

	defer func() {
		if panicErr := util.Panic2Err(); panicErr != nil {
			err = panicErr
		}
		if err != nil {
			logger.Errorf(g.ctx, "create session failed, listener %q client %q, %s", conn.LocalAddr(), conn.RemoteAddr(), err)
			conn.Close()
			g.wg.Done()
		}
	}()

	// 与客户端握手
	session, err := g.handshake(conn)
	if err != nil {
		return
	}

	_ = session

	logger.Debugf(g.ctx, "create session success, listener %q client %q", conn.LocalAddr(), conn.RemoteAddr())
}

func (g *_TcpGate) handshake(conn net.Conn) (*_TcpSession, error) {
	// 握手协议
	handshake := &protocol.HandshakeProtocol{
		Transceiver: &protocol.Transceiver{
			Conn:          conn,
			Encoder:       &codec.Encoder{},
			Decoder:       &codec.Decoder{MsgCreator: g.options.MsgCreator},
			Timeout:       g.options.Timeout,
			SequencedBuff: protocol.SequencedBuff{Cap: g.options.SequencedBuffCap},
		},
	}
	handshake.Transceiver.SequencedBuff.Reset(math_rand.Uint32(), math_rand.Uint32())

	var cs transport.CipherSuite
	var cm transport.Compression
	var cliRandom, servRandom []byte
	var cliHelloBytes, servHelloBytes []byte
	var continueFlow, encryptionFlow, authFlow bool
	var session *_TcpSession

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

	// 与客户端互相hello
	err := handshake.ServerHello(func(cliHello protocol.Event[*transport.MsgHello]) (protocol.Event[*transport.MsgHello], error) {
		// 检查协议版本
		if cliHello.Msg.Version != transport.Version_V1_0 {
			return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
				Code:    transport.Code_VersionError,
				Message: fmt.Sprintf("version %q not support", cliHello.Msg.Version),
			}
		}

		// 检查客户端要求的会话是否存在，已存在需要走断线重连流程
		if cliHello.Msg.SessionId != "" {
			v, ok := g.sessionMap.Load(cliHello.Msg.SessionId)
			if !ok {
				return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
					Code:    transport.Code_SessionNotFound,
					Message: fmt.Sprintf("session %q not exist", cliHello.Msg.SessionId),
				}
			}
			session = v.(*_TcpSession)
			continueFlow = true
		} else {
			session = newTcpSession(g)
			continueFlow = false
		}

		// 检查是否同意使用客户端建议的加密与压缩方案
		if g.options.AgreeClientProposal {
			cs = cliHello.Msg.CipherSuite
			cm = cliHello.Msg.Compression
		} else {
			cs = g.options.CipherSuite
			cm = g.options.Compression
		}

		// 开启加密时，需要交换随机数
		if cs.SecretKeyExchange != transport.SecretKeyExchange_None {
			// 记录客户端随机数
			if len(cliHello.Msg.Random) < 0 {
				return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
					Code:    transport.Code_EncryptFailed,
					Message: "client Hello 'random' is empty",
				}
			}
			cliRandom = codec.BytesPool.Get(len(cliHello.Msg.Random))
			copy(cliRandom, cliHello.Msg.Random)

			// 生成服务端随机数
			n, err := rand.Prime(rand.Reader, 256)
			if err != nil {
				return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
					Code:    transport.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			servRandom = codec.BytesPool.Get(n.BitLen() / 8)
			n.FillBytes(servRandom)

			encryptionFlow = true
		}

		// 返回服务端Hello
		servHello := protocol.Event[*transport.MsgHello]{
			Flags: transport.Flags(transport.Flag_HelloDone),
			Msg: &transport.MsgHello{
				Version:     transport.Version_V1_0,
				SessionId:   session.GetId(),
				Random:      servRandom,
				CipherSuite: cs,
				Compression: cm,
			},
		}

		authFlow = g.options.Auth != nil

		// 标记是否开启加密
		servHello.Flags.Set(transport.Flag_Encryption, encryptionFlow)
		// 标记是否开启鉴权
		servHello.Flags.Set(transport.Flag_Auth, authFlow)
		// 标记是否走断线重连流程
		servHello.Flags.Set(transport.Flag_Continue, continueFlow)

		// 开启加密时，记录双方hello数据，用于ecdh后加密验证
		if encryptionFlow {
			cliHelloBytes = codec.BytesPool.Get(cliHello.Msg.Size())
			if _, err := cliHello.Msg.Read(cliHelloBytes); err != nil {
				return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
					Code:    transport.Code_EncryptFailed,
					Message: err.Error(),
				}
			}

			servHelloBytes = codec.BytesPool.Get(servHello.Msg.Size())
			if _, err := servHello.Msg.Read(servHelloBytes); err != nil {
				return protocol.Event[*transport.MsgHello]{}, &protocol.RstError{
					Code:    transport.Code_EncryptFailed,
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
		err = g.secretKeyExchange(handshake, cs, cm, cliRandom, servRandom, cliHelloBytes, servHelloBytes, session.GetId())
		if err != nil {
			return nil, err
		}
	}

	var token string

	// 开启鉴权时，鉴权客户端
	if authFlow {
		err = handshake.ServerAuth(func(e protocol.Event[*transport.MsgAuth]) error {
			// 断线重连流程，检查会话Id与token是否匹配，防止hack客户端猜测会话Id，恶意通过断线重连登录
			if continueFlow {
				if e.Msg.Token != session.GetToken() {
					return &protocol.RstError{
						Code:    transport.Code_AuthFailed,
						Message: "incorrect token",
					}
				}
			}

			err := g.options.Auth(e.Msg.Token, e.Msg.Extensions)
			if err != nil {
				return &protocol.RstError{
					Code:    transport.Code_AuthFailed,
					Message: err.Error(),
				}
			}

			token = strings.Clone(e.Msg.Token)

			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	//// 使用token加分布式锁
	//if token != "" {
	//	mutex := dsync.NewDMutex(g.ctx, fmt.Sprintf("session-token-%s", token), dsync.WithOption{}.Expiry(handshake.Transceiver.Timeout))
	//	if err := mutex.Lock(context.Background()); err != nil {
	//		return nil, err
	//	}
	//	defer mutex.Unlock(context.Background())
	//}

	var sendSeq, recvSeq uint32

	// 断线重连流程，需要交换序号，检测是否能补发消息
	if continueFlow {
		err = handshake.ServerContinue(func(e protocol.Event[*transport.MsgContinue]) error {
			// 刷新会话
			sendSeq, recvSeq, err = session.Renew(handshake.Transceiver.Conn, e.Msg.RecvSeq)
			if err != nil {
				return &protocol.RstError{
					Code:    transport.Code_ContinueFailed,
					Message: err.Error(),
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// 消息序号
		sendSeq = handshake.Transceiver.SequencedBuff.SendSeq
		recvSeq = handshake.Transceiver.SequencedBuff.RecvSeq

		// 初始化会话
		session.Init(*handshake.Transceiver, token)
	}

	// 通知客户端握手结束
	err = handshake.ServerFinished(protocol.Event[*transport.MsgFinished]{
		Flags: transport.Flags_None().
			Setd(transport.Flag_EncryptOK, encryptionFlow).
			Setd(transport.Flag_AuthOK, authFlow).
			Setd(transport.Flag_ContinueOK, continueFlow),
		Msg: &transport.MsgFinished{
			SendSeq: sendSeq,
			RecvSeq: recvSeq,
		},
	})
	if err != nil {
		return nil, err
	}

	// 存储会话
	actual, loaded := g.sessionMap.LoadOrStore(session.GetId(), session)
	if loaded && actual != session {

	}

	return session, nil
}

func (g *_TcpGate) secretKeyExchange(handshake *protocol.HandshakeProtocol, cs transport.CipherSuite, cm transport.Compression,
	cliRandom, servRandom, cliHelloBytes, servHelloBytes []byte, sessionId string) (err error) {
	// 控制协议
	ctrl := protocol.CtrlProtocol{
		Transceiver: handshake.Transceiver,
		RetryTimes:  handshake.RetryTimes,
	}

	// 是否已发送rst
	rstSent := false

	defer func() {
		if err != nil && !rstSent {
			ctrl.SendRst(&protocol.RstError{
				Code:    transport.Code_EncryptFailed,
				Message: err.Error(),
			})
		}
	}()

	// 选择秘钥交换函数，并与客户端交换秘钥
	switch cs.SecretKeyExchange {
	case transport.SecretKeyExchange_None:
		break
	case transport.SecretKeyExchange_ECDHE:
		// 创建曲线
		curve, err := method.NewNamedCurve(g.options.ECDHENamedCurve)
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
		signature, err := g.sign(cs, cm, cliRandom, servRandom, sessionId, servPubBytes)
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
		if cs.BlockCipherMode.Padding() {
			if encEncryptionModule.Padding, err = g.makePaddingMode(cs.PaddingMode); err != nil {
				return err
			}
			if decEncryptionModule.Padding, err = g.makePaddingMode(cs.PaddingMode); err != nil {
				return err
			}
		}

		// 创建iv值
		iv, err := g.makeIV(cs.SymmetricEncryption, len(servPubBytes))
		if err != nil {
			return err
		}

		// 创建nonce值
		nonce, err := g.makeNonce(cs.SymmetricEncryption, len(servPubBytes))
		if err != nil {
			return err
		}

		var ivBytes, nonceBytes []byte

		if iv != nil {
			ivBytes = iv.Bytes()
		}

		if nonce != nil {
			nonceBytes = nonce.Bytes()
			encEncryptionModule.FetchNonce = g.makeFetchNonce(nonce)
			decEncryptionModule.FetchNonce = g.makeFetchNonce(nonce)
		}

		// 临时共享秘钥
		var sharedKeyBytes []byte

		// 与客户端交换秘钥
		err = handshake.ServerECDHESecretKeyExchange(
			protocol.Event[*transport.MsgECDHESecretKeyExchange]{
				Flags: transport.Flags_None().Setd(transport.Flag_Signature, len(signature) > 0),
				Msg: &transport.MsgECDHESecretKeyExchange{
					NamedCurve:         g.options.ECDHENamedCurve,
					PublicKey:          servPubBytes,
					IV:                 ivBytes,
					Nonce:              nonceBytes,
					SignatureAlgorithm: g.options.SignatureAlgorithm,
					Signature:          signature,
				},
			},
			func(cliECDHESecretKeyExchange protocol.Event[*transport.MsgECDHESecretKeyExchange]) (protocol.Event[*transport.MsgChangeCipherSpec], error) {
				// 检查客户端曲线类型
				if cliECDHESecretKeyExchange.Msg.NamedCurve != g.options.ECDHENamedCurve {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'NamedCurve' %d is incorrect", cliECDHESecretKeyExchange.Msg.NamedCurve),
					}
				}

				// 客户端临时公钥
				cliPub, err := curve.NewPublicKey(cliECDHESecretKeyExchange.Msg.PublicKey)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err),
					}
				}

				// 临时共享秘钥
				sharedKeyBytes, err = servPriv.ECDH(cliPub)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("ECDH failed, %s", err),
					}
				}

				// 创建并设置加解密流
				encrypter, decrypter, err := method.NewCipherStream(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, ivBytes)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("new cipher stream failed, %s", err),
					}
				}
				encEncryptionModule.CipherStream = encrypter
				decEncryptionModule.CipherStream = decrypter

				// 验证客户端签名
				if g.options.VerifyClientSignature {
					if !cliECDHESecretKeyExchange.Flags.Is(transport.Flag_Signature) {
						return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
							Code:    transport.Code_EncryptFailed,
							Message: "no client signature",
						}
					}

					if err := g.verify(cliECDHESecretKeyExchange.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, cliECDHESecretKeyExchange.Msg.PublicKey); err != nil {
						return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
							Code:    transport.Code_EncryptFailed,
							Message: err.Error(),
						}
					}
				}

				// 加密hello消息
				encryptedHello, err := encEncryptionModule.Transforming(nil, servHelloBytes)
				if err != nil {
					return protocol.Event[*transport.MsgChangeCipherSpec]{}, &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("encrypt hello failed, %s", err),
					}
				}

				return protocol.Event[*transport.MsgChangeCipherSpec]{
					Flags: transport.Flags(transport.Flag_VerifyEncryption),
					Msg: &transport.MsgChangeCipherSpec{
						EncryptedHello: encryptedHello,
					},
				}, nil
			}, func(cliChangeCipherSpec protocol.Event[*transport.MsgChangeCipherSpec]) error {
				// 客户端要求不验证加密
				if !cliChangeCipherSpec.Flags.Is(transport.Flag_VerifyEncryption) {
					return nil
				}

				// 验证加密是否正确
				decryptedHello, err := decEncryptionModule.Transforming(nil, cliChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: fmt.Sprintf("decrypt hello failed, %s", err),
					}
				}

				if bytes.Compare(decryptedHello, cliHelloBytes) != 0 {
					return &protocol.RstError{
						Code:    transport.Code_EncryptFailed,
						Message: "verify hello failed",
					}
				}

				return nil
			})
		if err != nil {
			rstSent = true
			return err
		}

		// 编码器
		encoder := &codec.Encoder{
			EncryptionModule: encEncryptionModule,
			Encryption:       true,
		}

		if encoder.MACModule, encoder.PatchMAC, err = g.makeMACModule(cs.MACHash, sharedKeyBytes); err != nil {
			return err
		}

		if encoder.CompressionModule, encoder.CompressedSize, err = g.makeCompressionModule(cm); err != nil {
			return err
		}

		// 解码器
		decoder := &codec.Decoder{
			MsgCreator:       handshake.Transceiver.Decoder.GetMsgCreator(),
			EncryptionModule: decEncryptionModule,
		}

		if decoder.MACModule, _, err = g.makeMACModule(cs.MACHash, sharedKeyBytes); err != nil {
			return err
		}

		if decoder.CompressionModule, _, err = g.makeCompressionModule(cm); err != nil {
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

func (g *_TcpGate) makeIV(se transport.SymmetricEncryption, pubSize int) (*big.Int, error) {
	size, ok := se.IV()
	if !ok {
		return nil, nil
	}

	if size <= 0 {
		size = pubSize
	}

	iv, err := rand.Prime(rand.Reader, size*8)
	if err != nil {
		return nil, err
	}

	return iv, nil
}

func (g *_TcpGate) makeNonce(se transport.SymmetricEncryption, pubSize int) (*big.Int, error) {
	size, ok := se.Nonce()
	if !ok {
		return nil, nil
	}

	if size <= 0 {
		size = pubSize
	}

	nonce, err := rand.Prime(rand.Reader, size*8)
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

func (g *_TcpGate) makeFetchNonce(nonce *big.Int) codec.FetchNonce {
	if nonce == nil {
		return nil
	}

	encryptionNonce := new(big.Int).Set(nonce)
	encryptionNonceNonceBuff := encryptionNonce.Bytes()

	bits := nonce.BitLen()

	return func() ([]byte, error) {
		if g.options.NonceStep == nil || g.options.NonceStep.Sign() == 0 {
			return encryptionNonceNonceBuff, nil
		}

		encryptionNonce.Add(encryptionNonce, g.options.NonceStep)
		if encryptionNonce.BitLen() > bits {
			encryptionNonce.SetInt64(0)
		}
		encryptionNonce.FillBytes(encryptionNonceNonceBuff)

		return encryptionNonceNonceBuff, nil
	}
}

func (g *_TcpGate) makePaddingMode(paddingMode transport.PaddingMode) (method.Padding, error) {
	if paddingMode == transport.PaddingMode_None {
		return nil, nil
	}

	padding, err := method.NewPadding(paddingMode)
	if err != nil {
		return nil, err
	}

	return padding, nil
}

func (g *_TcpGate) makeMACModule(hash transport.Hash, sharedKeyBytes []byte) (codec.IMACModule, bool, error) {
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

func (g *_TcpGate) makeCompressionModule(compression transport.Compression) (codec.ICompressionModule, int, error) {
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

	return compressionModule, g.options.CompressedSize, err
}

func (g *_TcpGate) sign(cs transport.CipherSuite, cm transport.Compression, cliRandom, servRandom []byte, sessionId string, servPubBytes []byte) ([]byte, error) {
	if g.options.SignatureAlgorithm.AsymmetricEncryption == transport.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if g.options.SignaturePrivateKey == nil {
		return nil, errors.New("option SignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		g.options.SignatureAlgorithm.AsymmetricEncryption,
		g.options.SignatureAlgorithm.PaddingMode,
		g.options.SignatureAlgorithm.Hash)
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
	signature, err := signer.Sign(g.options.SignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func (g *_TcpGate) verify(signature []byte, cs transport.CipherSuite, cm transport.Compression, cliRandom, servRandom []byte, sessionId string, cliPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if g.options.VerifySignaturePublicKey == nil {
		return errors.New("option VerifySignaturePublicKey is nil, unable to perform the verify signature operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		g.options.SignatureAlgorithm.AsymmetricEncryption,
		g.options.SignatureAlgorithm.PaddingMode,
		g.options.SignatureAlgorithm.Hash)
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

	return signer.Verify(g.options.VerifySignaturePublicKey, signBuf.Bytes(), signature)
}
