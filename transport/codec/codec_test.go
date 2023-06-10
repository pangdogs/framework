package codec

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"hash/fnv"
	"io"
	"kit.golaxy.org/plugins/transport"
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

	encoder := Encoder{
		CipherModule: CipherModule{
			StreamCipher: modeEncrypt,
		},
		MACModule: MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressModule: CompressModule{
			NewReader: func(reader io.Reader) (io.Reader, error) {
				return gzip.NewReader(reader)
			},
			NewWriter: func(writer io.Writer) (io.WriteCloser, error) {
				return gzip.NewWriterLevel(writer, gzip.BestCompression)
			},
		},
		Encryption: true,
		PatchMAC:   true,
		Compressed: 0,
	}

	err = encoder.Stuff(0, &transport.MsgHello{
		Version:   10,
		SessionId: []byte("abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"),
		Random:    []byte("efgdfffffffff222222222333333333334abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111"),
		CipherSuite: transport.CipherSuite{
			SecretKeyExchangeMethod: transport.SecretKeyExchangeMethod_ECDHE,
			SymmetricEncryptMethod:  transport.SymmetricEncryptMethod_AES256,
			BlockCipherMode:         transport.BlockCipherMode_CFB,
			HashMethod:              transport.HashMethod_Fnv1a32,
		},
		Extensions: []byte("abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111efgdfffffffff222222222333333333334abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa1111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111abcaaaaaaaaaaaaaaaaaaaaa11111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111111112"),
	})
	if err != nil {
		panic(err)
	}

	decoder := Decoder{
		MsgCreator: DefaultMsgCreator(),
		CipherModule: CipherModule{
			StreamCipher: modeDecrypt,
		},
		MACModule: MAC64Module{
			Hash:       fnv.New64a(),
			PrivateKey: key.Bytes(),
		},
		CompressModule: CompressModule{
			NewReader: func(reader io.Reader) (io.Reader, error) {
				return gzip.NewReader(reader)
			},
			NewWriter: func(writer io.Writer) (io.WriteCloser, error) {
				return gzip.NewWriterLevel(writer, gzip.BestCompression)
			},
		},
	}

	_, err = decoder.ReadFrom(&encoder)
	if err != nil {
		panic(err)
	}

	for err = decoder.Fetch(func(mp MsgPacket) {
		fmt.Printf("%+v", mp)
	}); err == nil; {
	}

}
