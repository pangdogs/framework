package codec

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"kit.golaxy.org/plugins/gtp"
	"kit.golaxy.org/plugins/gtp/method"
	"testing"
)

func TestCodec(t *testing.T) {
	key, _ := rand.Prime(rand.Reader, 256)
	//iv, _ := rand.Prime(rand.Reader, chacha20.NonceSize*8)

	encrypter, decrypter, err := method.NewCipher(gtp.SymmetricEncryption_AES, gtp.BlockCipherMode_GCM, key.Bytes(), nil)
	if err != nil {
		panic(err)
	}

	nonce, _ := rand.Prime(rand.Reader, encrypter.NonceSize()*8)

	compressionStream, err := method.NewCompressionStream(gtp.Compression_Brotli)
	if err != nil {
		panic(err)
	}

	//padding, err := method.NewPadding(gtp.PaddingMode_X923)
	//if err != nil {
	//	panic(err)
	//}

	encoder := Encoder{
		EncryptionModule: &EncryptionModule{
			Cipher: encrypter,
			//Padding:      padding,
			FetchNonce: func() ([]byte, error) {
				return nonce.Bytes(), nil
			},
		},
		MACModule: &MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressionModule: &CompressionModule{
			CompressionStream: compressionStream,
		},
		Encryption:     true,
		PatchMAC:       true,
		CompressedSize: 1,
	}

	for i := 0; i < 5; i++ {
		sessionId, _ := rand.Prime(rand.Reader, 1024)
		random, _ := rand.Prime(rand.Reader, 1024)

		err = encoder.Encode(gtp.Flags_None(), &gtp.MsgHello{
			Version:   gtp.Version(i),
			SessionId: sessionId.String(),
			Random:    random.Bytes(),
			CipherSuite: gtp.CipherSuite{
				SecretKeyExchange:   gtp.SecretKeyExchange_ECDHE,
				SymmetricEncryption: gtp.SymmetricEncryption_AES,
				BlockCipherMode:     gtp.BlockCipherMode_CFB,
				MACHash:             gtp.Hash_Fnv1a32,
			},
		})
		if err != nil {
			panic(err)
		}
	}

	decoder := Decoder{
		MsgCreator: gtp.DefaultMsgCreator(),
		EncryptionModule: &EncryptionModule{
			Cipher: decrypter,
			//Padding:      padding,
			FetchNonce: func() ([]byte, error) {
				return nonce.Bytes(), nil
			},
		},
		MACModule: &MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressionModule: &CompressionModule{
			CompressionStream: compressionStream,
		},
	}

	for {
		_, err = decoder.ReadFrom(&encoder)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			panic(err)
		}
	}

	for {
		mp, err := decoder.Decode()
		if err != nil {
			if errors.Is(err, ErrDataNotEnough) {
				return
			}
			panic(err)
		}
		v, _ := json.Marshal(mp)
		fmt.Printf("%s\n", v)
	}
}
