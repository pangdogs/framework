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
	"io"
	"math/big"
	"net"
	"strings"

	"git.golaxy.org/core/utils/uid"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/codec"
	"git.golaxy.org/framework/net/gtp/method"
	"git.golaxy.org/framework/net/gtp/transport"
	"git.golaxy.org/framework/utils/binaryutil"
)

// handshake 完成服务端握手，并返回新建或迁移后的会话。
func (acc *_Acceptor) handshake(ctx context.Context, conn net.Conn) (*_Session, bool, error) {
	// 每条待握手连接使用独立编解码器，协商出的模块随后随连接移交给会话。
	acc.encoder = codec.NewEncoder()
	acc.decoder = codec.NewDecoder(acc.options.MsgCreator)

	// 包长限制从第一条明文握手消息起生效。
	acc.decoder.SetMaxPacketSize(acc.options.MaxPacketSize)

	// 握手阶段尚无时序状态，使用不带序号缓存的同步器。
	handshake := &transport.HandshakeProtocol{
		Transceiver: &transport.Transceiver{
			Conn:         conn,
			Encoder:      acc.encoder,
			Decoder:      acc.decoder,
			Timeout:      acc.options.IOTimeout,
			Synchronizer: transport.NewUnsequencedSynchronizer(),
		},
		RetryTimes: acc.options.IORetryTimes,
	}
	defer handshake.Transceiver.Dispose()

	var cs gtp.CipherSuite
	var cm gtp.Compression
	var cliRandom, servRandom []byte
	var cliHelloHash, servHelloHash [sha256.Size]byte
	var continueFlow, encryptionFlow, authFlow bool
	var sessionID uid.ID
	var userID, token string
	var extensions []byte

	defer func() {
		if cliRandom != nil {
			binaryutil.BytesPool.Put(cliRandom)
		}
		if servRandom != nil {
			binaryutil.BytesPool.Put(servRandom)
		}
	}()

	// Hello 阶段完成版本检查、会话续接判断及密码和压缩方案选择。
	err := handshake.ServerHello(ctx, func(cliHello transport.Event[*gtp.MsgHello]) (transport.Event[*gtp.MsgHello], error) {
		// 检查协议版本
		if cliHello.Msg.Version != gtp.Version_V1_0 {
			return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
				Code:    gtp.Code_VersionError,
				Message: fmt.Sprintf("version %q not supported", cliHello.Msg.Version),
			}
		}

		// 检查客户端要求的会话是否存在，已存在需要走断线重连流程
		continueFlow = cliHello.Msg.SessionID != ""
		if continueFlow {
			session, ok := acc.getSession(uid.From(cliHello.Msg.SessionID))
			if !ok {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_SessionNotFound,
					Message: fmt.Sprintf("session %q not exist", cliHello.Msg.SessionID),
				}
			}
			sessionID, userID, token = session.ID(), session.UserID(), session.Token()
		} else {
			sessionID = acc.genSessionID()
		}

		// 检查是否同意使用客户端建议的加密方案
		if acc.options.AgreeClientEncryptionProposal {
			cs = cliHello.Msg.CipherSuite
		} else {
			cs = acc.options.EncCipherSuite
		}

		// 检查是否同意使用客户端建议的压缩方案
		if acc.options.AgreeClientCompressionProposal {
			cm = cliHello.Msg.Compression
		} else {
			cm = acc.options.Compression
		}

		// 需要密钥交换时保留客户端随机数，并生成服务端随机数。
		if cs.SecretKeyExchange != gtp.SecretKeyExchange_None {
			// MsgHello 的切片属于解码结果，复制到握手期池化缓冲区。
			if len(cliHello.Msg.Random) <= 0 {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: "client Hello 'random' is empty",
				}
			}
			cliRandom = binaryutil.BytesPool.Get(len(cliHello.Msg.Random))
			copy(cliRandom, cliHello.Msg.Random)

			// 随机数参与签名和共享密钥派生。
			n, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 256))
			if err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			servRandom = binaryutil.BytesPool.Get(len(n.Bytes()))
			n.FillBytes(servRandom)

			encryptionFlow = true
		}

		// 服务端 Hello 回显最终采用的协议参数及新建或续接的会话 ID。
		servHello := transport.Event[*gtp.MsgHello]{
			Flags: gtp.Flags(gtp.Flag_HelloDone),
			Msg: &gtp.MsgHello{
				Version:     gtp.Version_V1_0,
				SessionID:   sessionID.String(),
				Random:      servRandom,
				CipherSuite: cs,
				Compression: cm,
			},
		}

		authFlow = len(acc.options.Authenticator) > 0

		// 标志位告知客户端后续需要执行的可选握手阶段。
		servHello.Flags.Set(gtp.Flag_Encryption, encryptionFlow)
		servHello.Flags.Set(gtp.Flag_Auth, authFlow)
		servHello.Flags.Set(gtp.Flag_Continue, continueFlow)

		// Hello 摘要绑定协商参数，供 ECDH 签名及切换密码后的验证使用。
		if encryptionFlow {
			h := sha256.New()

			hashBuff := binaryutil.BytesPool.Get(4 * 1024)
			defer binaryutil.BytesPool.Put(hashBuff)

			h.Reset()
			_, err := io.CopyBuffer(h, cliHello.Msg, hashBuff)
			if err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			h.Sum(cliHelloHash[:0])

			h.Reset()
			_, err = io.CopyBuffer(h, servHello.Msg, hashBuff)
			if err != nil {
				return transport.Event[*gtp.MsgHello]{}, &transport.RstError{
					Code:    gtp.Code_EncryptFailed,
					Message: err.Error(),
				}
			}
			h.Sum(servHelloHash[:0])
		}

		return servHello, nil
	})
	if err != nil {
		return nil, false, err
	}

	// 密钥交换完成后，握手 Transceiver 已安装协商出的加密和认证模块。
	if encryptionFlow {
		err = acc.secretKeyExchange(ctx, handshake, cs, cm, cliRandom, servRandom, cliHelloHash, servHelloHash, sessionID)
		if err != nil {
			return nil, false, err
		}
	}

	// 压缩在密钥交换之后启用，避免改变参与签名的 Hello 表示。
	err = acc.setupCompression(cm)
	if err != nil {
		return nil, false, err
	}

	// 新会话保存鉴权信息；续接会话还必须与原会话身份和令牌一致。
	if authFlow {
		err = handshake.ServerAuth(ctx, func(e transport.Event[*gtp.MsgAuth]) error {
			// 续接时校验身份与令牌，拒绝伪造的会话续接请求。
			if continueFlow {
				if e.Msg.UserID != userID || e.Msg.Token != token {
					return &transport.RstError{
						Code:    gtp.Code_AuthFailed,
						Message: "incorrect token",
					}
				}
			}

			err := acc.options.Authenticator.UnsafeCall(func(err, _ error) bool {
				return err != nil
			}, acc._Gate, conn, e.Msg.UserID, e.Msg.Token, e.Msg.Extensions)
			if err != nil {
				return &transport.RstError{
					Code:    gtp.Code_AuthFailed,
					Message: err.Error(),
				}
			}

			if !continueFlow {
				userID = strings.Clone(e.Msg.UserID)
				token = strings.Clone(e.Msg.Token)
				extensions = bytes.Clone(e.Msg.Extensions)
			}

			return nil
		})
		if err != nil {
			return nil, false, err
		}
	}

	var session *_Session
	var sendSeq, recvSeq uint32

	// 续接连接交换收发序号，并在旧会话的 Transceiver 中定位补发起点。
	if continueFlow {
		err = handshake.ServerContinue(ctx, func(e transport.Event[*gtp.MsgContinue]) error {
			// Hello 后会话仍可能因超时并发删除，因此再次查询。
			var ok bool
			session, ok = acc.getSession(sessionID)
			if !ok {
				return &transport.RstError{
					Code:    gtp.Code_ContinueFailed,
					Message: "session has expired",
				}
			}

			// 迁移成功后旧连接关闭，新连接接管未确认帧。
			var err error
			sendSeq, recvSeq, err = session.migrateConn(handshake.Transceiver.Conn, e.Msg.RecvSeq)
			if err != nil {
				return &transport.RstError{
					Code:    gtp.Code_ContinueFailed,
					Message: err.Error(),
				}
			}
			return nil
		})
		if err != nil {
			return nil, false, err
		}
	} else {
		// 新会话从随机序号开始，并接管握手使用的连接和编解码器。
		session = acc.newSession(sessionID, userID, token, extensions)
		sendSeq, recvSeq = session.initConn(handshake.Transceiver.Conn, handshake.Transceiver.Encoder, handshake.Transceiver.Decoder)
	}

	// Finished 确认各可选阶段成功，并返回服务端当前收发序号。
	err = handshake.ServerFinished(ctx, transport.Event[*gtp.MsgFinished]{
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
		return nil, false, err
	}

	sendRst := func(code gtp.Code, message string) error {
		err := &transport.RstError{
			Code:    code,
			Message: message,
		}
		ctrl := transport.CtrlProtocol{
			Transceiver: handshake.Transceiver,
			RetryTimes:  handshake.RetryTimes,
		}
		ctrl.SendRst(err)
		return err
	}

	if continueFlow {
		// Finished 发送期间会话仍可能过期，返回前再验证一次注册身份。
		if !acc.validateSession(session) {
			return nil, false, sendRst(gtp.Code_Reject, "session has expired")
		}
	} else {
		// 新会话必须先加入关闭屏障和会话表，主循环退出时负责 Done。
		if !acc.barrier.Join(1) {
			return nil, false, sendRst(gtp.Code_Shutdown, "service shutdown")
		}
		if !acc.addSession(session) {
			acc.barrier.Done()
			return nil, false, sendRst(gtp.Code_Reject, "session can't be confirmed")
		}
		session.setState(SessionState_Confirmed)
		go session.mainLoop()
	}

	return session, continueFlow, nil
}

// secretKeyExchange 与客户端完成协商密码套件的密钥交换流程。
func (acc *_Acceptor) secretKeyExchange(ctx context.Context, handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom []byte, cliHelloHash, servHelloHash [sha256.Size]byte, sessionID uid.ID) (err error) {
	// 密钥交换失败时通过独立控制门面尽力发送一次 RST。
	ctrl := transport.CtrlProtocol{
		Transceiver: handshake.Transceiver,
		RetryTimes:  handshake.RetryTimes,
	}

	// HandshakeProtocol 内部已发送 RST 时避免在 defer 中重复发送。
	rstSent := false

	defer func() {
		if err != nil && !rstSent {
			ctrl.SendRst(&transport.RstError{
				Code:    gtp.Code_EncryptFailed,
				Message: err.Error(),
			})
		}
	}()

	// 选择密钥交换算法，并与客户端交换密钥
	switch cs.SecretKeyExchange {
	case gtp.SecretKeyExchange_ECDHE:
		// 创建曲线
		curve, err := method.NewNamedCurve(acc.options.EncECDHENamedCurve)
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
		signature, err := acc.sign(cs, cm, cliRandom, servRandom, sessionID, servPubBytes)
		if err != nil {
			return err
		}

		// 加密参数
		var padding [2]method.Padding
		var fetchNonce [2]codec.FetchNonce
		var cipher [2]method.Cipher
		var encryption [2]codec.IEncryption

		// 设置分组对齐填充方案
		if padding[0], err = acc.newPaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
			return err
		}
		if padding[1], err = acc.newPaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
			return err
		}

		// 创建 IV
		iv, err := acc.newIV(cs.SymmetricEncryption, cs.BlockCipherMode)
		if err != nil {
			return err
		}

		// 创建 nonce
		nonce, err := acc.newNonce(cs.SymmetricEncryption, cs.BlockCipherMode)
		if err != nil {
			return err
		}

		var ivBytes, nonceBytes, nonceStepBytes []byte

		if iv != nil {
			ivBytes = iv.Bytes()
		}

		if nonce != nil {
			nonceBytes = nonce.Bytes()
			nonceStepBytes = acc.options.EncNonceStep.Bytes()
			fetchNonce[0] = acc.newFetchNonce(nonce, acc.options.EncNonceStep)
			fetchNonce[1] = acc.newFetchNonce(nonce, acc.options.EncNonceStep)
		}

		// 临时共享密钥
		var sharedKeyBytes []byte

		// 加密后的 Hello 消息
		var encryptedHello binaryutil.Bytes
		defer encryptedHello.Release()

		// 与客户端交换密钥
		err = handshake.ServerECDHESecretKeyExchange(ctx,
			transport.Event[*gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: &gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         acc.options.EncECDHENamedCurve,
					PublicKey:          servPubBytes,
					IV:                 ivBytes,
					Nonce:              nonceBytes,
					NonceStep:          nonceStepBytes,
					SignatureAlgorithm: acc.options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			},
			func(cliECDHE transport.Event[*gtp.MsgECDHESecretKeyExchange]) (transport.Event[*gtp.MsgChangeCipherSpec], error) {
				// 检查客户端曲线类型
				if cliECDHE.Msg.NamedCurve != acc.options.EncECDHENamedCurve {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("client ECDHESecretKeyExchange 'NamedCurve(%s)' is incorrect", cliECDHE.Msg.NamedCurve),
					}
				}

				// 验证客户端签名
				if acc.options.EncVerifyClientSignature {
					if !cliECDHE.Flags.Is(gtp.Flag_Signature) {
						return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
							Code:    gtp.Code_EncryptFailed,
							Message: "no client signature",
						}
					}

					if err := acc.verify(cliECDHE.Msg.SignatureAlgorithm, cliECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionID, cliECDHE.Msg.PublicKey); err != nil {
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

				// 临时共享密钥
				sharedKeyBytes, err = servPriv.ECDH(cliPub)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("ECDH failed, %s", err),
					}
				}

				// 创建并设置加解密流
				cipher[0], cipher[1], err = method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, ivBytes, nonceBytes)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("new cipher stream failed, %s", err),
					}
				}

				// 加密模块
				encryption[0] = codec.NewEncryption(cipher[0], padding[0], fetchNonce[0])

				// 加密 Hello 消息
				encryptedHello, err = encryption[0].Transforming(nil, servHelloHash[:])
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("encrypt hello failed, %s", err),
					}
				}

				return transport.Event[*gtp.MsgChangeCipherSpec]{
					Flags: gtp.Flags(gtp.Flag_VerifyEncryption),
					Msg: &gtp.MsgChangeCipherSpec{
						EncryptedHello: encryptedHello.Payload(),
					},
				}, nil
			}, func(cliChangeCipherSpec transport.Event[*gtp.MsgChangeCipherSpec]) error {
				// 解密模块
				encryption[1] = codec.NewEncryption(cipher[1], padding[1], fetchNonce[1])

				// 客户端要求不验证加密
				if !cliChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption) {
					return nil
				}

				// 验证加密是否正确
				decryptedHello, err := encryption[1].Transforming(nil, cliChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return &transport.RstError{
						Code:    gtp.Code_EncryptFailed,
						Message: fmt.Sprintf("decrypt hello failed, %s", err),
					}
				}
				defer decryptedHello.Release()

				if bytes.Compare(decryptedHello.Payload(), cliHelloHash[:]) != 0 {
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
		acc.setupEncryption(encryption)

		// 安装认证模块
		return acc.setupAuthentication(cs.HMAC, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange(%s) not supported", cs.SecretKeyExchange)
	}
}

// setupCompression 安装协商得到的压缩模块。
func (acc *_Acceptor) setupCompression(cm gtp.Compression) error {
	compression, compressionThreshold, _, err := acc.newCompression(cm)
	if err != nil {
		return err
	}
	acc.encoder.SetCompression(compression, compressionThreshold)

	compression, _, maxUncompressedSize, err := acc.newCompression(cm)
	if err != nil {
		return err
	}
	acc.decoder.SetCompression(compression, maxUncompressedSize)

	return nil
}

// setupEncryption 安装发送与接收方向的加密模块。
func (acc *_Acceptor) setupEncryption(encryption [2]codec.IEncryption) {
	acc.encoder.SetEncryption(encryption[0])
	acc.decoder.SetEncryption(encryption[1])
}

// setupAuthentication 安装发送与接收方向的消息认证模块。
func (acc *_Acceptor) setupAuthentication(hash gtp.Hash, sharedKeyBytes []byte) error {
	authentication, err := acc.newAuthentication(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	acc.encoder.SetAuthentication(authentication)

	authentication, err = acc.newAuthentication(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	acc.decoder.SetAuthentication(authentication)

	return nil
}

// newIV 构造配置长度的随机 IV。
func (acc *_Acceptor) newIV(se gtp.SymmetricEncryption, bcm gtp.BlockCipherMode) (*big.Int, error) {
	if !se.BlockCipherMode() || !bcm.IV() {
		return nil, nil
	}

	size, ok := se.BlockSize()
	if !ok {
		return nil, fmt.Errorf("CipherSuite.BlockCipherMode(%s) needs IV, but CipherSuite.SymmetricEncryption(%s) lacks a fixed block size", bcm, se)
	}

	iv, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), uint(size)*8))
	if err != nil {
		return nil, err
	}

	return iv, nil
}

// newNonce 构造配置长度的随机 nonce。
func (acc *_Acceptor) newNonce(se gtp.SymmetricEncryption, bcm gtp.BlockCipherMode) (*big.Int, error) {
	size, ok := se.Nonce()
	if !ok {
		if !se.BlockCipherMode() || !bcm.Nonce() {
			return nil, nil
		}
		size, ok = se.BlockSize()
		if !ok {
			return nil, fmt.Errorf("CipherSuite.BlockCipherMode(%s) needs Nonce, but CipherSuite.SymmetricEncryption(%s) lacks a fixed block size", bcm, se)
		}
	}

	nonce, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), uint(size)*8))
	if err != nil {
		return nil, err
	}

	return nonce, nil
}

// newFetchNonce 构造每次取值后按步长递增的 nonce 函数。
func (acc *_Acceptor) newFetchNonce(nonce, nonceStep *big.Int) codec.FetchNonce {
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

// newPaddingMode 根据分组模式和协商配置构造填充方案。
func (acc *_Acceptor) newPaddingMode(bcm gtp.BlockCipherMode, paddingMode gtp.PaddingMode) (method.Padding, error) {
	if !bcm.Padding() {
		return nil, nil
	}

	if paddingMode == gtp.PaddingMode_None {
		return nil, fmt.Errorf("CipherSuite.BlockCipherMode(%s), plaintext padding is necessary", bcm)
	}

	padding, err := method.NewPadding(paddingMode)
	if err != nil {
		return nil, err
	}

	return padding, nil
}

// newAuthentication 根据协商摘要和共享密钥构造消息认证模块。
func (acc *_Acceptor) newAuthentication(hash gtp.Hash, sharedKeyBytes []byte) (codec.IAuthentication, error) {
	if hash == gtp.Hash_None {
		return nil, nil
	}

	hmac, err := method.NewHMAC(hash, sharedKeyBytes)
	if err != nil {
		return nil, err
	}

	return codec.NewAuthentication(hmac), nil
}

// newCompression 构造协商得到的压缩模块。
func (acc *_Acceptor) newCompression(compression gtp.Compression) (codec.ICompression, int, int, error) {
	if compression == gtp.Compression_None {
		return nil, 0, 0, nil
	}

	compressionStream, err := method.NewCompressionStream(compression)
	if err != nil {
		return nil, 0, 0, err
	}

	return codec.NewCompression(compressionStream), acc.options.CompressionThreshold, acc.options.MaxUncompressedSize, err
}

// sign 使用服务端私钥签署握手参数；未配置签名算法时返回空签名。
func (acc *_Acceptor) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionID uid.ID, servPubBytes []byte) ([]byte, error) {
	// 无需签名
	if acc.options.EncSignatureAlgorithm.AsymmetricEncryption == gtp.AsymmetricEncryption_None {
		return nil, nil
	}

	// 必须设置私钥才能签名
	if acc.options.EncSignaturePrivateKey == nil {
		return nil, errors.New("option EncSignaturePrivateKey is nil, unable to perform the signing operation")
	}

	// 创建签名器
	signer, err := method.NewSigner(
		acc.options.EncSignatureAlgorithm.AsymmetricEncryption,
		acc.options.EncSignatureAlgorithm.PaddingMode,
		acc.options.EncSignatureAlgorithm.Hash)
	if err != nil {
		return nil, err
	}

	// 签名数据
	signBuf := bytes.NewBuffer(nil)
	io.CopyN(signBuf, cs, int64(cs.Size()))
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionID.String())
	signBuf.Write(servPubBytes)

	// 生成签名
	signature, err := signer.Sign(acc.options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 使用客户端公钥验证握手参数签名。
func (acc *_Acceptor) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionID uid.ID, cliPubBytes []byte) error {
	// 必须设置公钥才能验证签名
	if acc.options.EncVerifySignaturePublicKey == nil {
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
	signBuf.WriteString(sessionID.String())
	signBuf.Write(cliPubBytes)

	return signer.Verify(acc.options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
