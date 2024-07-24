package codec

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"git.golaxy.org/framework/net/gtp"
	"git.golaxy.org/framework/net/gtp/method"
	"hash/fnv"
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

	encoder := CreateEncoder().
		SetupEncryptionModule(NewEncryptionModule(encrypter, nil, func() ([]byte, error) { return nonce.Bytes(), nil })).
		SetupMACModule(NewMAC64Module(fnv.New64a(), key.Bytes())).
		SetupCompressionModule(NewCompressionModule(compressionStream), 1).
		Spawn()

	decoder := CreateDecoder(gtp.DefaultMsgCreator()).
		SetupEncryptionModule(NewEncryptionModule(decrypter, nil, func() ([]byte, error) { return nonce.Bytes(), nil })).
		SetupMACModule(NewMAC64Module(fnv.New64a(), key.Bytes())).
		SetupCompressionModule(NewCompressionModule(compressionStream)).
		Spawn()

	for i := 0; i < 10; i++ {
		sessionId, _ := rand.Prime(rand.Reader, 1024)
		random, _ := rand.Prime(rand.Reader, 1024)

		bs, err := encoder.Encode(gtp.Flags_None(), &gtp.MsgHello{
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

		mp, _, err := decoder.Decode(bs.Data())
		if err != nil {
			panic(err)
		}
		v, _ := json.Marshal(mp)
		fmt.Printf("%s\n", v)

		bs.Release()
	}
}
