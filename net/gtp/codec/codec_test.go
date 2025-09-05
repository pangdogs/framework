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

package codec

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
	"math/big"
	"testing"
)

func TestCodec(t *testing.T) {
	key, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 256))
	//iv, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), chacha20.NonceSize*8))

	encrypter, decrypter, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key.Bytes(), nil, nil)
	if err != nil {
		panic(err)
	}

	nonce, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), uint(encrypter.NonceSize())*8))

	compressionStream, err := method.NewCompressionStream(gtp.Compression_Brotli)
	if err != nil {
		panic(err)
	}

	//padding, err := method.NewPadding(gtp.PaddingMode_X923)
	//if err != nil {
	//	panic(err)
	//}

	hmac, err := method.NewHMAC(gtp.Hash_BLAKE2b256, key.Bytes())
	if err != nil {
		panic(err)
	}

	encoder := NewEncoder().
		SetEncryption(NewEncryption(encrypter, nil, func() ([]byte, error) { return nonce.Bytes(), nil })).
		SetAuthentication(NewAuthentication(hmac)).
		SetCompression(NewCompression(compressionStream), 1)

	decoder := NewDecoder(gtp.DefaultMsgCreator()).
		SetEncryption(NewEncryption(decrypter, nil, func() ([]byte, error) { return nonce.Bytes(), nil })).
		SetAuthentication(NewAuthentication(hmac)).
		SetCompression(NewCompression(compressionStream))

	for i := 0; i < 10; i++ {
		sessionId, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 1024))
		random, _ := rand.Int(rand.Reader, big.NewInt(0).Lsh(big.NewInt(1), 1024))

		bs, err := encoder.Encode(gtp.Flags_None(), &gtp.MsgHello{
			Version:   gtp.Version(i),
			SessionId: sessionId.String(),
			Random:    random.Bytes(),
			CipherSuite: gtp.CipherSuite{
				SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
				SymmetricEncryption: gtp.SymmetricEncryption_AES,
				BlockCipherMode:     gtp.BlockCipherMode_CFB,
				HMAC:                gtp.Hash_BLAKE2b256,
			},
		})
		if err != nil {
			panic(err)
		}

		mp, _, err := decoder.Decode(bs.Data(), nil)
		if err != nil {
			panic(err)
		}
		v, _ := json.Marshal(mp)
		fmt.Printf("%s\n", v)

		bs.Release()
	}
}
