package codec

import (
	"crypto/aes"
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
	iv, _ := rand.Prime(rand.Reader, aes.BlockSize)

	encrypter, decrypter, err := method.NewCipherStream(transport.SymmetricEncryption_AES, transport.BlockCipherMode_GCM, key.Bytes(), iv.Bytes())
	if err != nil {
		panic(err)
	}

	compressionStream, err := method.NewCompressionStream(transport.Compression_Brotli)
	if err != nil {
		panic(err)
	}

	encoder := Encoder{
		EncryptionModule: &EncryptionModule{
			CipherStream: encrypter,
		},
		MACModule: &MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressionModule: &CompressionModule{
			CompressionStream: compressionStream,
		},
		Encryption: true,
		PatchMAC:   true,
		Compressed: 1,
	}

	for i := 0; i < 5; i++ {
		sessionId, _ := rand.Prime(rand.Reader, 1024)
		random, _ := rand.Prime(rand.Reader, 1024)
		extensions, _ := rand.Prime(rand.Reader, 2048)

		err = encoder.Stuff(0, &transport.MsgHello{
			Version:   transport.Version(i),
			SessionId: sessionId.String(),
			Random:    random.Bytes(),
			CipherSuite: transport.CipherSuite{
				SecretKeyExchange:   transport.SecretKeyExchange_ECDHE,
				SymmetricEncryption: transport.SymmetricEncryption_AES,
				BlockCipherMode:     transport.BlockCipherMode_CFB,
				Hash:                transport.Hash_Fnv1a32,
			},
			Extensions: extensions.Bytes(),
		})
		if err != nil {
			panic(err)
		}
	}

	decoder := Decoder{
		MsgCreator: DefaultMsgCreator(),
		EncryptionModule: &EncryptionModule{
			CipherStream: decrypter,
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

	for decoder.Fetch(func(mp transport.MsgPacket) {
		v, _ := json.Marshal(mp)
		fmt.Printf("%s\n", v)
	}) == nil {
	}
}
