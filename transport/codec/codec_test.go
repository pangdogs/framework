package codec

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"kit.golaxy.org/plugins/transport"
	"kit.golaxy.org/plugins/transport/method"
	"testing"
)

func TestCodec(t *testing.T) {
	key, _ := rand.Prime(rand.Reader, 256)
	//iv, _ := rand.Prime(rand.Reader, chacha20.NonceSize*8)

	encrypter, decrypter, err := method.NewCipher(transport.SymmetricEncryption_AES, transport.BlockCipherMode_GCM, key.Bytes(), nil)
	if err != nil {
		panic(err)
	}

	nonce, _ := rand.Prime(rand.Reader, encrypter.NonceSize()*8)

	compressionStream, err := method.NewCompressionStream(transport.Compression_Brotli)
	if err != nil {
		panic(err)
	}

	//padding, err := method.NewPadding(transport.PaddingMode_X923)
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

		err = encoder.Stuff(transport.Flags_None(), &transport.MsgHello{
			Version:   transport.Version(i),
			SessionId: sessionId.String(),
			Random:    random.Bytes(),
			CipherSuite: transport.CipherSuite{
				SecretKeyExchange:   transport.SecretKeyExchange_ECDHE,
				SymmetricEncryption: transport.SymmetricEncryption_AES,
				BlockCipherMode:     transport.BlockCipherMode_CFB,
				MACHash:             transport.Hash_Fnv1a32,
			},
		})
		if err != nil {
			panic(err)
		}
	}

	decoder := Decoder{
		MsgCreator: DefaultMsgCreator(),
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
		mp, err := decoder.Fetch()
		if err != nil {
			if errors.Is(err, ErrBufferNotEnough) {
				return
			}
			panic(err)
		}
		v, _ := json.Marshal(mp)
		fmt.Printf("%s\n", v)
	}
}
