/*
 * This file is part of Golaxy Distributed Service Development Framework.
 *
 * Golaxy Distributed Service Development Framework is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 2.1 of the License, or
 * (at your option) any later version.
 *
 * Golaxy Distributed Service Development Framework is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with Golaxy Distributed Service Development Framework. If not, see <http://www.gnu.org/licenses/>.
 *
 * Copyright (c) 2024 pangdogs.
 */

package gate

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
	"io"
	"math/big"
	"net"
	"strings"
)

// handshake 握手过程
func (acc *_Acceptor) handshake(ctx context.Context, conn net.Conn) (*_Session, error) {
	// 编解码器构建器
	acc.encoderCreator = codec.BuildEncoder()
	acc.decoderCreator = codec.BuildDecoder(acc.gate.options.DecoderMsgCreator)

	// 握手协议
	handshake := &transport.HandshakeProtocol{
		Transceiver: &transport.Transceiver{
			Conn:         conn,
			Encoder:      acc.encoderCreator.Make(),
			Decoder:      acc.decoderCreator.Make(),
			Timeout:      acc.gate.options.IOTimeout,
			Synchronizer: transport.NewUnsequencedSynchronizer(),
		},
		RetryTimes: acc.gate.options.IORetryTimes,
	}
	defer handshake.Transceiver.Clean()

	var cs gtp.CipherSuite
	var cm gtp.Compression
	var cliRandom, servRandom []byte
	var cliHelloHash, servHelloHash [sha256.Size]byte
	var continueFlow, encryptionFlow, authFlow bool
	var session *_Session

	defer func() {
		if cliRandom != nil {
			binaryutil.BytesPool.Put(cliRandom)
		}
		if servRandom != nil {
			binaryutil.BytesPool.Put(servRandom)
		}
	}()

	// 与客户端互相hello
	err := handshake.ServerHello(ctx, func(cliHello transport.Event[gtp.MsgHello]) (transport.Event[gtp.MsgHello], error) {
		// 检查协议版本
		if cliHello.Msg.Version != gtp.Version_V1_0 {
			return transport.Event[gtp.MsgHello]{}, &transport.RstError{
				Code:    gtp.Code_VersionError,
				Message: fmt.Sprintf("version %q not support", cliHello.Msg.Version),
			}
		}

		// 检查客户端要求的会话是否存在，已存在需要走断线重连流程
		if cliHello.Msg.SessionId != "" {
			v, ok := acc.gate.getSession(uid.From(cliHello.Msg.SessionId))
			if !ok {
				return transport.Event[gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_SessionNotFound,
					Message: fmt.Sprintf("session %q not exist", cliHello.Msg.SessionId),
				}
			}

			session = v
			continueFlow = true
		} else {
			v, err := acc.newSession(conn)
			if err != nil {
				return transport.Event[gtp.MsgHello]{}, err
			}

			// 调整会话状态为握手中
			v.setState(SessionState_Handshake)

			session = v
			continueFlow = false
		}

		// 检查是否同意使用客户端建议的加密方案
		if acc.gate.options.AgreeClientEncryptionProposal {
			cs = cliHello.Msg.CipherSuite
		} else {
			cs = acc.gate.options.EncCipherSuite
		}

		// 检查是否同意使用客户端建议的压缩方案
		if acc.gate.options.AgreeClientCompressionProposal {
			cm = cliHello.Msg.Compression
		} else {
			cm = acc.gate.options.Compression
		}

		// 开启加密时，需要交换随机数
		if cs.SecretKeyExchange != gtp.SecretKeyExchange_None {
			// 记录客户端随机数
			if len(cliHello.Msg.Random) < 0 {
				return transport.Event[gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: "client Hello 'random' is empty",
				}
			}
			cliRandom = binaryutil.BytesPool.Get(len(cliHello.Msg.Random))
			copy(cliRandom, cliHello.Msg.Random)

			// 生成服务端随机数
			n, err := rand.Prime(rand.Reader, 256)
			if err != nil {
				return transport.Event[gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			servRandom = binaryutil.BytesPool.Get(n.BitLen() / 8)
			n.FillBytes(servRandom)

			encryptionFlow = true
		}

		// 返回服务端Hello
		servHello := transport.Event[gtp.MsgHello]{
			Flags: gtp.Flags(gtp.Flag_HelloDone),
			Msg: gtp.MsgHello{
				Version:     gtp.Version_V1_0,
				SessionId:   session.GetId().String(),
				Random:      servRandom,
				CipherSuite: cs,
				Compression: cm,
			},
		}

		authFlow = len(acc.gate.options.Authenticator) > 0

		// 标记是否开启加密
		servHello.Flags.Set(gtp.Flag_Encryption, encryptionFlow)
		// 标记是否开启鉴权
		servHello.Flags.Set(gtp.Flag_Auth, authFlow)
		// 标记是否走断线重连流程
		servHello.Flags.Set(gtp.Flag_Continue, continueFlow)

		// 开启加密时，记录双方hello数据，用于ecdh后加密验证
		if encryptionFlow {
			h := sha256.New()

			hashBuff := binaryutil.BytesPool.Get(4 * 1024)
			defer binaryutil.BytesPool.Put(hashBuff)

			h.Reset()
			_, err := io.CopyBuffer(h, cliHello.Msg, hashBuff)
			if err != nil {
				return transport.Event[gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			copy(cliHelloHash[:], h.Sum(nil))

			h.Reset()
			_, err = io.CopyBuffer(h, servHello.Msg, hashBuff)
			if err != nil {
				return transport.Event[gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			copy(servHelloHash[:], h.Sum(nil))
		}

		return servHello, nil
	})
	if err != nil {
		return nil, err
	}

	// 开启加密时，与客户端交换秘钥
	if encryptionFlow {
		err = acc.secretKeyExchange(ctx, handshake, cs, cm, cliRandom, servRandom, cliHelloHash, servHelloHash, session.GetId())
		if err != nil {
			return nil, err
		}
	}

	// 安装压缩模块
	err = acc.setupCompressionModule(cm)
	if err != nil {
		return nil, err
	}

	var userId, token string

	// 开启鉴权时，鉴权客户端
	if authFlow {
		err = handshake.ServerAuth(ctx, func(e transport.Event[gtp.MsgAuth]) error {
			// 断线重连流程，检查会话Id与token是否匹配，防止hack客户端猜测会话Id，恶意通过断线重连登录
			if continueFlow {
				if e.Msg.UserId != session.GetUserId() || e.Msg.Token != session.GetToken() {
					return &transport.RstError{
						Code:    gtp.Code_AuthFailed,
						Message: "incorrect token",
					}
				}
			}

			err := acc.gate.options.Authenticator.UnsafeCall(func(err, _ error) bool {
				return err != nil
			}, acc.gate, conn, e.Msg.UserId, e.Msg.Token, e.Msg.Extensions)
			if err != nil {
				return &transport.RstError{
					Code:    gtp.Code_AuthFailed,
					Message: err.Error(),
				}
			}

			userId = strings.Clone(e.Msg.UserId)
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
		err = handshake.ServerContinue(ctx, func(e transport.Event[gtp.MsgContinue]) error {
			// 刷新会话
			var err error
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
		sendSeq, recvSeq = session.init(handshake.Transceiver.Conn, handshake.Transceiver.Encoder,
			handshake.Transceiver.Decoder, userId, token)
	}

	// 通知客户端握手结束
	err = handshake.ServerFinished(ctx, transport.Event[gtp.MsgFinished]{
		Flags: gtp.Flags_None().
			Setd(gtp.Flag_EncryptOK, encryptionFlow).
			Setd(gtp.Flag_AuthOK, authFlow).
			Setd(gtp.Flag_ContinueOK, continueFlow),
		Msg: gtp.MsgFinished{
			SendSeq: sendSeq,
			RecvSeq: recvSeq,
		},
	})
	if err != nil {
		return nil, err
	}

	if continueFlow {
		// 检测会话有效性
		if !acc.gate.validateSession(session) {
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
		acc.gate.storeSession(session)

		// 调整会话状态为已确认
		session.setState(SessionState_Confirmed)

		// 运行会话
		session.gate.wg.Add(1)
		go session.mainLoop()
	}

	return session, nil
}

// secretKeyExchange 秘钥交换过程
func (acc *_Acceptor) secretKeyExchange(ctx context.Context, handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom []byte, cliHelloHash, servHelloHash [sha256.Size]byte, sessionId uid.Id) (err error) {
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
		curve, err := method.NewNamedCurve(acc.gate.options.EncECDHENamedCurve)
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

		// 加密参数
		var padding [2]method.Padding
		var fetchNonce [2]codec.FetchNonce
		var cipher [2]method.Cipher
		var encryptionModule [2]codec.IEncryptionModule

		// 设置分组对齐填充方案
		if padding[0], err = acc.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
			return err
		}
		if padding[1], err = acc.makePaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
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
			nonceStepBytes = acc.gate.options.EncNonceStep.Bytes()
			fetchNonce[0] = acc.makeFetchNonce(nonce, acc.gate.options.EncNonceStep)
			fetchNonce[1] = acc.makeFetchNonce(nonce, acc.gate.options.EncNonceStep)
		}

		// 临时共享秘钥
		var sharedKeyBytes []byte

		// 加密后的hello消息
		var encryptedHello binaryutil.RecycleBytes
		defer encryptedHello.Release()

		// 与客户端交换秘钥
		err = handshake.ServerECDHESecretKeyExchange(ctx,
			transport.Event[gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         acc.gate.options.EncECDHENamedCurve,
					PublicKey:          servPubBytes,
					IV:                 ivBytes,
					Nonce:              nonceBytes,
					NonceStep:          nonceStepBytes,
					SignatureAlgorithm: acc.gate.options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			},
			func(cliECDHE transport.Event[gtp.MsgECDHESecretKeyExchange]) (transport.Event[gtp.MsgChangeCipherSpec], error) {
				// 检查客户端曲线类型
				if cliECDHE.Msg.NamedCurve != acc.gate.options.EncECDHENamedCurve {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'NamedCurve' %d is incorrect", cliECDHE.Msg.NamedCurve),
					}
				}

				// 验证客户端签名
				if acc.gate.options.EncVerifyClientSignature {
					if !cliECDHE.Flags.Is(gtp.Flag_Signature) {
						return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
							Code:    gtp.Code_EncryptFailed,
							Message: "no client signature",
						}
					}

					if err := acc.verify(cliECDHE.Msg.SignatureAlgorithm, cliECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, cliECDHE.Msg.PublicKey); err != nil {
						return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
							Code:    gtp.Code_EncryptFailed,
							Message: err.Error(),
						}
					}
				}

				// 客户端临时公钥
				cliPub, err := curve.NewPublicKey(cliECDHE.Msg.PublicKey)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err),
					}
				}

				// 临时共享秘钥
				sharedKeyBytes, err = servPriv.ECDH(cliPub)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("ECDH failed, %s", err),
					}
				}

				// 创建并设置加解密流
				cipher[0], cipher[1], err = method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, ivBytes)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("new cipher stream failed, %s", err),
					}
				}

				// 加密模块
				encryptionModule[0] = codec.NewEncryptionModule(cipher[0], padding[0], fetchNonce[0])

				// 加密hello消息
				encryptedHello, err = encryptionModule[0].Transforming(nil, servHelloHash[:])
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("encrypt hello failed, %s", err),
					}
				}

				return transport.Event[gtp.MsgChangeCipherSpec]{
					Flags: gtp.Flags(gtp.Flag_VerifyEncryption),
					Msg: gtp.MsgChangeCipherSpec{
						EncryptedHello: encryptedHello.Data(),
					},
				}, nil
			}, func(cliChangeCipherSpec transport.Event[gtp.MsgChangeCipherSpec]) error {
				// 解密模块
				encryptionModule[1] = codec.NewEncryptionModule(cipher[1], padding[1], fetchNonce[1])

				// 客户端要求不验证加密
				if !cliChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption) {
					return nil
				}

				// 验证加密是否正确
				decryptedHello, err := encryptionModule[1].Transforming(nil, cliChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("decrypt hello failed, %s", err),
					}
				}
				defer decryptedHello.Release()

				if bytes.Compare(decryptedHello.Data(), cliHelloHash[:]) != 0 {
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
		acc.setupEncryptionModule(encryptionModule)

		// 安装MAC模块
		return acc.setupMACModule(cs.MACHash, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange %d not support", cs.SecretKeyExchange)
	}

	return nil
}

// setupCompressionModule 安装压缩模块
func (acc *_Acceptor) setupCompressionModule(cm gtp.Compression) error {
	compressionModule, compressedSize, err := acc.makeCompressionModule(cm)
	if err != nil {
		return err
	}
	acc.encoderCreator.SetupCompressionModule(compressionModule, compressedSize)

	compressionModule, _, err = acc.makeCompressionModule(cm)
	if err != nil {
		return err
	}
	acc.decoderCreator.SetupCompressionModule(compressionModule)

	return nil
}

// setupEncryptionModule 安装加密模块
func (acc *_Acceptor) setupEncryptionModule(encryptionModule [2]codec.IEncryptionModule) {
	acc.encoderCreator.SetupEncryptionModule(encryptionModule[0])
	acc.decoderCreator.SetupEncryptionModule(encryptionModule[1])
}

// setupMACModule 安装MAC模块
func (acc *_Acceptor) setupMACModule(hash gtp.Hash, sharedKeyBytes []byte) error {
	macModule, err := acc.makeMACModule(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	acc.encoderCreator.SetupMACModule(macModule)

	macModule, err = acc.makeMACModule(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	acc.decoderCreator.SetupMACModule(macModule)

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

	encryptionNonce := big.NewInt(0).Set(nonce)
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
func (acc *_Acceptor) makeMACModule(hash gtp.Hash, sharedKeyBytes []byte) (codec.IMACModule, error) {
	if hash.Bits() <= 0 {
		return nil, nil
	}

	var macModule codec.IMACModule

	switch hash.Bits() {
	case 32:
		macHash, err := method.NewHash32(hash)
		if err != nil {
			return nil, err
		}
		macModule = codec.NewMAC32Module(macHash, sharedKeyBytes)
	case 64:
		macHash, err := method.NewHash64(hash)
		if err != nil {
			return nil, err
		}
		macModule = codec.NewMAC64Module(macHash, sharedKeyBytes)
	default:
		macHash, err := method.NewHash(hash)
		if err != nil {
			return nil, err
		}
		macModule = codec.NewMACModule(macHash, sharedKeyBytes)
	}

	return macModule, nil
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

	return codec.NewCompressionModule(compressionStream), acc.gate.options.CompressedSize, err
}

// sign 签名
func (acc *_Acceptor) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId uid.Id, servPubBytes []byte) ([]byte, error) {
	// 无需签名
	if acc.gate.options.EncSignatureAlgorithm.AsymmetricEncryption == gtp.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if acc.gate.options.EncSignaturePrivateKey == nil {
		return nil, errors.New("option EncSignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		acc.gate.options.EncSignatureAlgorithm.AsymmetricEncryption,
		acc.gate.options.EncSignatureAlgorithm.PaddingMode,
		acc.gate.options.EncSignatureAlgorithm.Hash)
	if err != nil {
		return nil, err
	}

	// 签名数据
	signBuf := bytes.NewBuffer(nil)
	io.CopyN(signBuf, cs, int64(cs.Size()))
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionId.String())
	signBuf.Write(servPubBytes)

	// 生成签名
	signature, err := signer.Sign(acc.gate.options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 验证签名
func (acc *_Acceptor) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId uid.Id, cliPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if acc.gate.options.EncVerifySignaturePublicKey == nil {
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
	io.CopyN(signBuf, cs, int64(cs.Size()))
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionId.String())
	signBuf.Write(cliPubBytes)

	return signer.Verify(acc.gate.options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
