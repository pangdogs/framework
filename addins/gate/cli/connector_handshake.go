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
func (ctor *_Connector) handshake(ctx context.Context, conn net.Conn, client *Client) error {
	// 编解码器构建器
	ctor.encoder = codec.NewEncoder()
	ctor.decoder = codec.NewDecoder(ctor.options.DecoderMsgCreator)

	// 握手协议
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
	defer handshake.Transceiver.Clean()

	var sessionId uid.Id
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

	// 生成客户端随机数
	n, err := rand.Prime(rand.Reader, 256)
	if err != nil {
		return err
	}
	cliRandom = binaryutil.BytesPool.Get(n.BitLen() / 8)
	n.FillBytes(cliRandom)

	cliHello := transport.Event[gtp.MsgHello]{
		Msg: gtp.MsgHello{
			Version:     gtp.Version_V1_0,
			SessionId:   client.GetSessionId().String(),
			Random:      cliRandom,
			CipherSuite: cs,
			Compression: cm,
		},
	}

	// 与服务端互相hello
	err = handshake.ClientHello(ctx, cliHello,
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
			sessionId = uid.From(strings.Clone(servHello.Msg.SessionId))
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
				h := sha256.New()

				hashBuff := binaryutil.BytesPool.Get(4 * 1024)
				defer binaryutil.BytesPool.Put(hashBuff)

				h.Reset()
				_, err := io.CopyBuffer(h, cliHello.Msg, hashBuff)
				if err != nil {
					return err
				}
				copy(cliHelloHash[:], h.Sum(nil))

				h.Reset()
				_, err = io.CopyBuffer(h, servHello.Msg, hashBuff)
				if err != nil {
					return err
				}
				copy(servHelloHash[:], h.Sum(nil))
			}

			return nil
		})
	if err != nil {
		return err
	}

	// 开启加密时，与服务端交换秘钥
	if encryptionFlow {
		err = ctor.secretKeyExchange(ctx, handshake, cs, cm, cliRandom, servRandom, cliHelloHash, servHelloHash, sessionId)
		if err != nil {
			return err
		}
	}

	// 安装压缩模块
	err = ctor.setupCompression(cm)
	if err != nil {
		return err
	}

	// 开启鉴权时，向服务端发起鉴权
	if authFlow {
		err = handshake.ClientAuth(ctx, transport.Event[gtp.MsgAuth]{
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
		err = handshake.ClientContinue(ctx, transport.Event[gtp.MsgContinue]{
			Msg: gtp.MsgContinue{
				SendSeq: client.transceiver.Synchronizer.SendSeq(),
				RecvSeq: client.transceiver.Synchronizer.RecvSeq(),
			},
		})
		if err != nil {
			return err
		}
	}

	var remoteSendSeq, remoteRecvSeq uint32

	// 等待服务端通知握手结束
	err = handshake.ClientFinished(ctx, func(finished transport.Event[gtp.MsgFinished]) error {
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
func (ctor *_Connector) secretKeyExchange(ctx context.Context, handshake *transport.HandshakeProtocol, cs gtp.CipherSuite, cm gtp.Compression,
	cliRandom, servRandom []byte, cliHelloHash, servHelloHash [sha256.Size]byte, sessionId uid.Id) error {
	// 选择秘钥交换函数，并与客户端交换秘钥
	switch cs.SecretKeyExchange {
	case gtp.SecretKeyExchange_ECDHE:
		// 临时共享秘钥
		var sharedKeyBytes []byte

		// 加密后的hello消息
		var encryptedHello binaryutil.RecycleBytes
		defer encryptedHello.Release()

		// 加密参数
		var padding [2]method.Padding
		var fetchNonce [2]codec.FetchNonce
		var cipher [2]method.Cipher
		var encryption [2]codec.IEncryption

		// 与服务端交换秘钥
		err := handshake.ClientSecretKeyExchange(ctx, func(e transport.IEvent) (transport.IEvent, error) {
			// 解包ECDHESecretKeyExchange消息事件
			switch e.Msg.MsgId() {
			case gtp.MsgId_ECDHESecretKeyExchange:
				break
			default:
				return transport.IEvent{}, fmt.Errorf("%w (%d)", transport.ErrUnexpectedMsg, e.Msg.MsgId())
			}
			servECDHE := transport.EventT[gtp.MsgECDHESecretKeyExchange](e)

			// 验证服务端签名
			if ctor.options.EncVerifyServerSignature {
				if !servECDHE.Flags.Is(gtp.Flag_Signature) {
					return transport.IEvent{}, errors.New("no server signature")
				}

				if err := ctor.verify(servECDHE.Msg.SignatureAlgorithm, servECDHE.Msg.Signature, cs, cm, cliRandom, servRandom, sessionId, servECDHE.Msg.PublicKey); err != nil {
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

			// 临时共享秘钥
			sharedKeyBytes, err = cliPriv.ECDH(servPub)
			if err != nil {
				return transport.IEvent{}, fmt.Errorf("ECDH failed, %s", err)
			}

			// 签名数据
			signature, err := ctor.sign(cs, cm, cliRandom, servRandom, sessionId, cliPubBytes)
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

			// 设置nonce值
			if len(servECDHE.Msg.Nonce) > 0 && len(servECDHE.Msg.NonceStep) > 0 {
				nonce := big.NewInt(0).SetBytes(servECDHE.Msg.Nonce)
				nonceStep := big.NewInt(0).SetBytes(servECDHE.Msg.NonceStep)
				fetchNonce[0] = ctor.newFetchNonce(nonce, nonceStep)
				fetchNonce[1] = ctor.newFetchNonce(nonce, nonceStep)
			}

			// 创建并设置加解密流
			cipher[0], cipher[1], err = method.NewCipher(cs.SymmetricEncryption, cs.BlockCipherMode, sharedKeyBytes, servECDHE.Msg.IV)
			if err != nil {
				return transport.IEvent{}, fmt.Errorf("new cipher stream failed, %s", err)
			}

			cliECDHE := transport.Event[gtp.MsgECDHESecretKeyExchange]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_Signature, len(signature) > 0),
				Msg: gtp.MsgECDHESecretKeyExchange{
					NamedCurve:         servECDHE.Msg.NamedCurve,
					PublicKey:          cliPubBytes,
					SignatureAlgorithm: ctor.options.EncSignatureAlgorithm,
					Signature:          signature,
				},
			}

			return cliECDHE.Interface(), nil

		}, func(servChangeCipherSpec transport.Event[gtp.MsgChangeCipherSpec]) (transport.Event[gtp.MsgChangeCipherSpec], error) {
			verifyEncryption := servChangeCipherSpec.Flags.Is(gtp.Flag_VerifyEncryption)

			// 加解密模块
			encryption[0] = codec.NewEncryption(cipher[0], padding[0], fetchNonce[0])
			encryption[1] = codec.NewEncryption(cipher[1], padding[1], fetchNonce[1])

			// 验证加密是否正确
			if verifyEncryption {
				decryptedHello, err := encryption[1].Transforming(nil, servChangeCipherSpec.Msg.EncryptedHello)
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, fmt.Errorf("decrypt hello failed, %s", err)
				}
				defer decryptedHello.Release()

				if bytes.Compare(decryptedHello.Data(), servHelloHash[:]) != 0 {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, errors.New("verify hello failed")
				}
			}

			cliChangeCipherSpec := transport.Event[gtp.MsgChangeCipherSpec]{
				Flags: gtp.Flags_None().Setd(gtp.Flag_VerifyEncryption, verifyEncryption),
			}

			// 加密hello消息
			if verifyEncryption {
				var err error
				encryptedHello, err = encryption[0].Transforming(nil, cliHelloHash[:])
				if err != nil {
					return transport.Event[gtp.MsgChangeCipherSpec]{}, fmt.Errorf("encrypt hello failed, %s", err)
				}

				cliChangeCipherSpec.Msg.EncryptedHello = encryptedHello.Data()
			}

			return cliChangeCipherSpec, nil
		})
		if err != nil {
			return err
		}

		// 安装加密模块
		ctor.setupEncryption(encryption)

		// 安装MAC模块
		return ctor.setupMAC(cs.MACHash, sharedKeyBytes)

	default:
		return fmt.Errorf("CipherSuite.SecretKeyExchange %d not support", cs.SecretKeyExchange)
	}

	return nil
}

// setupCompression 安装压缩模块
func (ctor *_Connector) setupCompression(cm gtp.Compression) error {
	compression, compressedSize, err := ctor.newCompression(cm)
	if err != nil {
		return err
	}
	ctor.encoder.SetCompression(compression, compressedSize)

	compression, _, err = ctor.newCompression(cm)
	if err != nil {
		return err
	}
	ctor.decoder.SetCompression(compression)

	return nil
}

// setupEncryption 安装加密模块
func (ctor *_Connector) setupEncryption(encryption [2]codec.IEncryption) {
	ctor.encoder.SetEncryption(encryption[0])
	ctor.decoder.SetEncryption(encryption[1])
}

// setupMAC 安装MAC模块
func (ctor *_Connector) setupMAC(hash gtp.Hash, sharedKeyBytes []byte) error {
	mac, err := ctor.newMAC(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	ctor.encoder.SetMAC(mac)

	mac, err = ctor.newMAC(hash, sharedKeyBytes)
	if err != nil {
		return err
	}
	ctor.decoder.SetMAC(mac)

	return nil
}

// newFetchNonce 构造获取nonce值函数
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

// newPaddingMode 构造填充方案
func (ctor *_Connector) newPaddingMode(bcm gtp.BlockCipherMode, paddingMode gtp.PaddingMode) (method.Padding, error) {
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

// newMAC 构造MAC模块
func (ctor *_Connector) newMAC(hash gtp.Hash, sharedKeyBytes []byte) (codec.IMAC, error) {
	if hash.Bits() <= 0 {
		return nil, nil
	}

	var mac codec.IMAC

	switch hash.Bits() {
	case 32:
		macHash, err := method.NewHash32(hash)
		if err != nil {
			return nil, err
		}
		mac = codec.NewMAC32(macHash, sharedKeyBytes)
	case 64:
		macHash, err := method.NewHash64(hash)
		if err != nil {
			return nil, err
		}
		mac = codec.NewMAC64(macHash, sharedKeyBytes)
	default:
		macHash, err := method.NewHash(hash)
		if err != nil {
			return nil, err
		}
		mac = codec.NewMAC(macHash, sharedKeyBytes)
	}

	return mac, nil
}

// newCompression 构造压缩模块
func (ctor *_Connector) newCompression(compression gtp.Compression) (codec.ICompression, int, error) {
	if compression == gtp.Compression_None {
		return nil, 0, nil
	}

	compressionStream, err := method.NewCompressionStream(compression)
	if err != nil {
		return nil, 0, err
	}

	return codec.NewCompression(compressionStream), ctor.options.CompressedSize, err
}

// sign 签名
func (ctor *_Connector) sign(cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId uid.Id, cliPubBytes []byte) ([]byte, error) {
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
	signBuf.WriteString(sessionId.String())
	signBuf.Write(cliPubBytes)

	// 生成签名
	signature, err := signer.Sign(ctor.options.EncSignaturePrivateKey, signBuf.Bytes())
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// verify 验证签名
func (ctor *_Connector) verify(signatureAlgorithm gtp.SignatureAlgorithm, signature []byte, cs gtp.CipherSuite, cm gtp.Compression, cliRandom, servRandom []byte, sessionId uid.Id, servPubBytes []byte) error {
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
	signBuf.WriteString(sessionId.String())
	signBuf.Write(servPubBytes)

	return signer.Verify(ctor.options.EncVerifySignaturePublicKey, signBuf.Bytes(), signature)
}
