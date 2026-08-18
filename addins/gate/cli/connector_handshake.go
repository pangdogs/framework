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

package cli

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

// handshake 完成客户端握手，并初始化或迁移客户端连接。
func (ctor *_Connector) handshake(ctx context.Context, conn net.Conn, client *Client) error {
	// 每次连接使用独立编解码器，协商出的模块随后随连接移交给客户端。
	ctor.encoder = codec.NewEncoder()
	ctor.decoder = codec.NewDecoder(ctor.options.MsgCreator)

	// 包长限制从第一条明文握手消息起生效。
	ctor.decoder.SetMaxPacketSize(ctor.options.MaxPacketSize)

	// 握手阶段尚无时序状态，使用不带序号缓存的同步器。
	handshake := &transport.HandshakeProtocol{
		Transceiver: &transport.Transceiver{
			Conn:         conn,
			Encoder:      ctor.encoder,
			Decoder:      ctor.decoder,
			Timeout:      ctor.options.IOTimeout,
			Synchronizer: transport.NewUnsequencedSynchronizer(),
		},
		RetryTimes: ctor.options.IORetryTimes,
	}
	defer handshake.Transceiver.Dispose()

	var sessionID uid.ID
	cs := ctor.options.EncCipherSuite
	cm := ctor.options.Compression
	var cliRandom, servRandom []byte
	var cliHelloHash, servHelloHash [sha256.Size]byte
	var continueFlow, encryptionFlow, authFlow bool

	defer func() {
		if cliRandom != nil {
			binaryutil.BytesPool.Put(cliRandom)
		}
		if servRandom != nil {
			binaryutil.BytesPool.Put(servRandom)
		}
	}()

	// 客户端随机数参与签名和共享密钥派生。
	n, err := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 256))
	if err != nil {
		return err
	}
	cliRandom = binaryutil.BytesPool.Get(len(n.Bytes()))
	n.FillBytes(cliRandom)

	cliHello := transport.Event[*gtp.MsgHello]{
		Msg: &gtp.MsgHello{
			Version:     gtp.Version_V1_0,
			SessionID:   client.SessionID().String(),
			Random:      cliRandom,
			CipherSuite: cs,
			Compression: cm,
		},
	}

	// Hello 阶段提交期望参数，并接收服务端最终选择和可选阶段标志。
	err = handshake.ClientHello(ctx, cliHello,
		func(servHello transport.Event[*gtp.MsgHello]) error {
			// 检查 HelloDone 标记
			if !servHello.Flags.Is(gtp.Flag_HelloDone) {
				return fmt.Errorf("cli: the expected msg-hello-flag (0x%x) was not received", gtp.Flag_HelloDone)
			}

			// 检查协议版本
			if servHello.Msg.Version != gtp.Version_V1_0 {
				return fmt.Errorf("cli: version %q not supported", servHello.Msg.Version)
			}

			// 后续阶段严格采用服务端确认的参数和会话 ID。
			sessionID = uid.From(strings.Clone(servHello.Msg.SessionID))
			cs = servHello.Msg.CipherSuite
			cm = servHello.Msg.Compression
			continueFlow = servHello.Flags.Is(gtp.Flag_Continue)
			encryptionFlow = servHello.Flags.Is(gtp.Flag_Encryption)
			authFlow = servHello.Flags.Is(gtp.Flag_Auth)

			// 加密流程需要复制服务端随机数并绑定双方 Hello 摘要。
			if encryptionFlow {
				// MsgHello 的切片属于解码结果，复制到握手期池化缓冲区。
				if len(servHello.Msg.Random) <= 0 {
					return errors.New("cli: server Hello 'random' is empty")
				}
				servRandom = binaryutil.BytesPool.Get(len(servHello.Msg.Random))
				copy(servRandom, servHello.Msg.Random)

				// 摘要供 ECDH 签名及切换密码后的验证使用。
				h := sha256.New()

				hashBuff := binaryutil.BytesPool.Get(4 * 1024)
				defer binaryutil.BytesPool.Put(hashBuff)

				h.Reset()
				_, err := io.CopyBuffer(h, cliHello.Msg, hashBuff)
				if err != nil {
					return err
				}
				h.Sum(cliHelloHash[:0])

				h.Reset()
				_, err = io.CopyBuffer(h, servHello.Msg, hashBuff)
				if err != nil {
					return err
				}
				h.Sum(servHelloHash[:0])
			}

			return nil
		})
	if err != nil {
		return err
	}

	// 密钥交换完成后，握手 Transceiver 已安装协商出的加密和认证模块。
	if encryptionFlow {
		err = ctor.secretKeyExchange(ctx, handshake, cs, cm, cliRandom, servRandom, cliHelloHash, servHelloHash, sessionID)
		if err != nil {
			return err
		}
	}

	// 压缩在密钥交换之后启用，避免改变参与签名的 Hello 表示。
	err = ctor.setupCompression(cm)
	if err != nil {
		return err
	}

	// 服务端要求鉴权时提交配置的用户、令牌及扩展数据。
	if authFlow {
		err = handshake.ClientAuth(ctx, transport.Event[*gtp.MsgAuth]{
			Msg: &gtp.MsgAuth{
				UserID:     ctor.options.AuthUserID,
				Token:      ctor.options.AuthToken,
				Extensions: ctor.options.AuthExtensions,
			},
		})
		if err != nil {
			return err
		}
	}

	// 续接连接提交旧连接的收发序号，供服务端定位补发起点。
	if continueFlow {
		err = handshake.ClientContinue(ctx, transport.Event[*gtp.MsgContinue]{
			Msg: &gtp.MsgContinue{
				SendSeq: client.transceiver.Synchronizer.SendSeq(),
				RecvSeq: client.transceiver.Synchronizer.RecvSeq(),
			},
		})
		if err != nil {
			return err
		}
	}

	var remoteSendSeq, remoteRecvSeq uint32

	// Finished 必须确认所有已协商阶段，并携带服务端当前收发序号。
	err = handshake.ClientFinished(ctx, func(finished transport.Event[*gtp.MsgFinished]) error {
		if encryptionFlow && !finished.Flags.Is(gtp.Flag_EncryptOK) {
			return fmt.Errorf("cli: the expected msg-finished-flag (0x%x) was not received", gtp.Flag_EncryptOK)
		}

		if authFlow && !finished.Flags.Is(gtp.Flag_AuthOK) {
			return fmt.Errorf("cli: the expected msg-finished-flag (0x%x) was not received", gtp.Flag_AuthOK)
		}

		if continueFlow && !finished.Flags.Is(gtp.Flag_ContinueOK) {
			return fmt.Errorf("cli: the expected msg-finished-flag (0x%x) was not received", gtp.Flag_ContinueOK)
		}

		remoteSendSeq = finished.Msg.SendSeq
		remoteRecvSeq = finished.Msg.RecvSeq
		return nil
	})
	if err != nil {
		return err
	}

	if continueFlow {
		// 续接时切换旧客户端的连接，并保留未确认发送帧。
		_, _, err = client.migrateConn(conn, remoteRecvSeq)
		if err != nil {
			return err
		}
	} else {
		// 首次连接接管握手使用的连接、编解码器和服务端分配的会话 ID。
		client.initConn(handshake.Transceiver.Conn, handshake.Transceiver.Encoder, handshake.Transceiver.Decoder, remoteSendSeq, remoteRecvSeq, sessionID)
	}

	return nil
}

// secretKeyExchange 与服务端完成协商密码套件的密钥交换流程。
func (ctor *_Connector) secretKeyExchange(ctx context.Context, handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom []byte, cliHelloHash, servHelloHash [sha256.Size]byte, sessionID uid.ID) error {
	// 选择密钥交换算法，并与服务端交换密钥
	switch cs.SecretKeyExchange {
	case gtp.SecretKeyExchange_ECDHE:
		// 临时共享密钥
		var sharedKeyBytes []byte

		// 加密后的 Hello 消息
		var encryptedHello binaryutil.Bytes
		defer encryptedHello.Release()

		// 加密参数
		var padding [2]method.Padding
		var fetchNonce [2]codec.FetchNonce
		var cipher [2]method.Cipher
		var encryption [2]codec.IEncryption

		// 与服务端交换密钥
		err := handshake.ClientSecretKeyExchange(ctx, func(e transport.IEvent) (transport.IEvent, error) {
			// 解包 ECDHESecretKeyExchange 事件
			switch e.Msg.MsgID() {
			case gtp.MsgID_ECDHESecretKeyExchange:
				break
			default:
				return transport.IEvent{}, fmt.Errorf("%w (%d)", transport.ErrUnexpectedMsg, e.Msg.MsgID())
			}
			servECDHE := transport.AssertEvent[*gtp.MsgECDHESecretKeyExchange](e)

			// 验证服务端签名
			if ctor.options.EncVerifyServerSignature {
				if !servECDHE.Flags.Is(gtp.Flag_Signature) {
					return transport.IEvent{}, errors.New("no server signature")
				}

				if err := ctor.verify(servECDHE.Msg.SignatureAlgorithm, servECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionID, servECDHE.Msg.PublicKey); err != nil {
					return transport.IEvent{}, err
				}
			}

			// 创建曲线
			curve, err := method.NewNamedCurve(servECDHE.Msg.NamedCurve)
			if err != nil {
				return transport.IEvent{}, err
			}

			// 生成客户端临时私钥
			cliPriv, err := curve.GenerateKey(rand.Reader)
			if err != nil {
				return transport.IEvent{}, err
			}

			// 生成客户端临时公钥
			cliPub := cliPriv.PublicKey()
			cliPubBytes := cliPub.Bytes()

			// 服务端临时公钥
			servPub, err := curve.NewPublicKey(servECDHE.Msg.PublicKey)
			if err != nil {
				return transport.IEvent{}, fmt.Errorf("server ECDHESecretKeyExchange 'PublicKey' is invalid, %s", err)
			}

			// 临时共享密钥
			sharedKeyBytes, err = cliPriv.ECDH(servPub)
			if err != nil {
				return transport.IEvent{}, fmt.Errorf("ECDH failed, %s", err)
			}

			// 签名数据
			signature, err := ctor.sign(cs, cm, cliRandom, servRandom, sessionID, cliPubBytes)
			if err != nil {
				return transport.IEvent{}, err
			}

			// 设置分组对齐填充方案
			if padding[0], err = ctor.newPaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return transport.IEvent{}, err
			}
			if padding[1], err = ctor.newPaddingMode(cs.BlockCipherMode, cs.PaddingMode); err != nil {
				return transport.IEvent{}, err
			}

			// 设置 nonce
			if len(servECDHE.Msg.Nonce) > 0 && len(servECDHE.Msg.NonceStep) > 0 {
				nonce := big.NewInt(0).SetBytes(servECDHE.Msg.Nonce)
				nonceStep := big.NewInt(0).SetBytes(servECDHE.Msg.NonceStep)
				fetchNonce[0] = ctor.newFetchNonce(nonce, nonceStep)
				fetchNonce[1] = ctor.newFetchNonce(nonce, nonceStep)
			}

			// 创建并设置加解密流
			cipher[0], cipher[1], err = method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, servECDHE.Msg.IV, servECDHE.Msg.Nonce)
			if err != nil {
				return transport.IEvent{}, fmt.Errorf("new cipher stream failed, %s", err)
			}

			cliECDHE := transport.Event[*gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: &gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         servECDHE.Msg.NamedCurve,
					PublicKey:          cliPubBytes,
					SignatureAlgorithm: ctor.options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			}

			return cliECDHE.Interface(), nil

		}, func(servChangeCipherSpec transport.Event[*gtp.MsgChangeCipherSpec]) (transport.Event[*gtp.MsgChangeCipherSpec], error) {
			verifyEncryption := servChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption)

			// 加解密模块
			encryption[0] = codec.NewEncryption(cipher[0], padding[0], fetchNonce[0])
			encryption[1] = codec.NewEncryption(cipher[1], padding[1], fetchNonce[1])

			// 验证加密是否正确
			if verifyEncryption {
				decryptedHello, err := encryption[1].Transforming(nil, servChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, fmt.Errorf("decrypt hello failed, %s", err)
				}
				defer decryptedHello.Release()

				if bytes.Compare(decryptedHello.Payload(), servHelloHash[:]) != 0 {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, errors.New("verify hello failed")
				}
			}

			cliChangeCipherSpec := transport.Event[*gtp.MsgChangeCipherSpec]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_VerifyEncryption, verifyEncryption),
				Msg:   &gtp.MsgChangeCipherSpec{},
			}

			// 加密 Hello 消息
			if verifyEncryption {
				var err error
				encryptedHello, err = encryption[0].Transforming(nil, cliHelloHash[:])
				if err != nil {
					return transport.Event[*gtp.MsgChangeCipherSpec]{}, fmt.Errorf("encrypt hello failed, %s", err)
				}

				cliChangeCipherSpec.Msg.EncryptedHello = encryptedHello.Payload()
			}

			return cliChangeCipherSpec, nil
		})
		if err != nil {
			return err
		}

		// 安装加密模块
		ctor.setupEncryption(encryption)

		// 安装认证模块
		return ctor.setupAuthentication(cs.HMAC, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange(%s) not supported", cs.SecretKeyExchange)
	}
}

// setupCompression 安装协商得到的压缩模块。
func (ctor *_Connector) setupCompression(cm gtp.Compression) error {
	compression, compressionThreshold, _, err := ctor.newCompression(cm)
	if err != nil {
		return err
	}
	ctor.encoder.SetCompression(compression, compressionThreshold)

	compression, _, maxUncompressedSize, err := ctor.newCompression(cm)
	if err != nil {
		return err
	}
	ctor.decoder.SetCompression(compression, maxUncompressedSize)

	return nil
}

// setupEncryption 安装发送与接收方向的加密模块。
func (ctor *_Connector) setupEncryption(encryption [2]codec.IEncryption) {
	ctor.encoder.SetEncryption(encryption[0])
	ctor.decoder.SetEncryption(encryption[1])
}

// setupAuthentication 安装发送与接收方向的消息认证模块。
func (ctor *_Connector) setupAuthentication(hash gtp.Hash, sharedKeyBytes []byte) error {
	authentication, err := ctor.newAuthentication(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	ctor.encoder.SetAuthentication(authentication)

	authentication, err = ctor.newAuthentication(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	ctor.decoder.SetAuthentication(authentication)

	return nil
}

// newFetchNonce 构造每次取值后按步长递增的 nonce 函数。
func (ctor *_Connector) newFetchNonce(nonce, nonceStep *big.Int) codec.FetchNonce {
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
func (ctor *_Connector) newPaddingMode(bcm gtp.BlockCipherMode, paddingMode gtp.PaddingMode) (method.Padding, error) {
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
func (ctor *_Connector) newAuthentication(hash gtp.Hash, sharedKeyBytes []byte) (codec.IAuthentication, error) {
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
func (ctor *_Connector) newCompression(compression gtp.Compression) (codec.ICompression, int, int, error) {
	if compression == gtp.Compression_None {
		return nil, 0, 0, nil
	}

	compressionStream, err := method.NewCompressionStream(compression)
	if err != nil {
		return nil, 0, 0, err
	}

	return codec.NewCompression(compressionStream), ctor.options.CompressionThreshold, ctor.options.MaxUncompressedSize, err
}

// sign 使用客户端私钥签署握手参数；未配置签名算法时返回空签名。
func (ctor *_Connector) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionID uid.ID, cliPubBytes []byte) ([]byte, error) {
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
	io.CopyN(signBuf, cs, int64(cs.Size()))
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionID.String())
	signBuf.Write(cliPubBytes)

	// 生成签名
	signature, err := signer.Sign(ctor.options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 使用服务端公钥验证握手参数签名。
func (ctor *_Connector) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionID uid.ID, servPubBytes []byte) error {
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
	io.CopyN(signBuf, cs, int64(cs.Size()))
	signBuf.WriteByte(uint8(cm))
	signBuf.Write(cliRandom)
	signBuf.Write(servRandom)
	signBuf.WriteString(sessionID.String())
	signBuf.Write(servPubBytes)

	return signer.Verify(ctor.options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
