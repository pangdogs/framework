package codec

import (
	"crypto/aes"
	"crypto/cipher"
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

	block, err := aes.NewCipher(key.Bytes())
	if err != nil {
		panic(err)
	}

	iv, _ := rand.Prime(rand.Reader, aes.BlockSize*8)

	modeEncrypt := cipher.NewCTR(block, iv.Bytes())
	if err != nil {
		panic(err)
	}

	modeDecrypt := cipher.NewCTR(block, iv.Bytes())
	if err != nil {
		panic(err)
	}

	encodeCS, err := method.NewCompressionStream(transport.CompressionMethod_Brotli)
	if err != nil {
		panic(err)
	}

	decodeCS, err := method.NewCompressionStream(transport.CompressionMethod_Brotli)
	if err != nil {
		panic(err)
	}

	encoder := Encoder{
		EncryptionModule: &EncryptionModule{
			CipherStream: modeEncrypt,
		},
		MACModule: &MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressionModule: &CompressionModule{
			CompressionStream: encodeCS,
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
				SecretKeyExchangeMethod: transport.SecretKeyExchangeMethod_ECDHE,
				SymmetricEncryptMethod:  transport.SymmetricEncryptMethod_AES256,
				BlockCipherMode:         transport.BlockCipherMode_CFB,
				HashMethod:              transport.HashMethod_Fnv1a32,
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
			CipherStream: modeDecrypt,
		},
		MACModule: &MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressionModule: &CompressionModule{
			CompressionStream: decodeCS,
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
