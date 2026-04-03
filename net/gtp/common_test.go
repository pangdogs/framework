package gtp

import "testing"

func TestFlagsSetAndSetd(t *testing.T) {
	var flags Flags

	flags.Set(Flag_Encrypted, true).Set(Flag_Signed, true)
	if !flags.Is(Flag_Encrypted) || !flags.Is(Flag_Signed) {
		t.Fatalf("expected flags to be set: %08b", flags)
	}

	flags.Set(Flag_Signed, false)
	if flags.Is(Flag_Signed) || !flags.Is(Flag_Encrypted) {
		t.Fatalf("unexpected flags after clear: %08b", flags)
	}

	clone := flags.Setd(Flag_Compressed, true)
	if !clone.Is(Flag_Compressed) {
		t.Fatal("expected cloned flags to include compressed")
	}
	if flags.Is(Flag_Compressed) {
		t.Fatal("expected Setd to leave original unchanged")
	}
}

func TestParseAndStringHelpers(t *testing.T) {
	tcs := []struct {
		name    string
		parse   func(string) (string, error)
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "secret key exchange",
			input: "ECDHE",
			want:  "ecdhe",
			parse: func(s string) (string, error) {
				v, err := ParseSecretKeyExchange(s)
				return v.String(), err
			},
		},
		{
			name:  "asymmetric encryption",
			input: "RSA",
			want:  "rsa",
			parse: func(s string) (string, error) {
				v, err := ParseAsymmetricEncryption(s)
				return v.String(), err
			},
		},
		{
			name:  "symmetric encryption",
			input: "XChaCha20_Poly1305",
			want:  "xchacha20_poly1305",
			parse: func(s string) (string, error) {
				v, err := ParseSymmetricEncryption(s)
				return v.String(), err
			},
		},
		{
			name:  "padding mode",
			input: "Pkcs1v15",
			want:  "pkcs1v15",
			parse: func(s string) (string, error) {
				v, err := ParsePaddingMode(s)
				return v.String(), err
			},
		},
		{
			name:  "block cipher mode",
			input: "GCM",
			want:  "gcm",
			parse: func(s string) (string, error) {
				v, err := ParseBlockCipherMode(s)
				return v.String(), err
			},
		},
		{
			name:  "hash",
			input: "BLAKE2b512",
			want:  "blake2b512",
			parse: func(s string) (string, error) {
				v, err := ParseHash(s)
				return v.String(), err
			},
		},
		{
			name:  "named curve",
			input: "P384",
			want:  "p384",
			parse: func(s string) (string, error) {
				v, err := ParseNamedCurve(s)
				return v.String(), err
			},
		},
		{
			name:  "compression",
			input: "Snappy",
			want:  "snappy",
			parse: func(s string) (string, error) {
				v, err := ParseCompression(s)
				return v.String(), err
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.parse(tc.input)
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if got != tc.want {
				t.Fatalf("unexpected string value: got %q want %q", got, tc.want)
			}
		})
	}

	invalid := []struct {
		name  string
		check func(string) error
	}{
		{name: "secret key exchange", check: func(s string) error { _, err := ParseSecretKeyExchange(s); return err }},
		{name: "asymmetric encryption", check: func(s string) error { _, err := ParseAsymmetricEncryption(s); return err }},
		{name: "symmetric encryption", check: func(s string) error { _, err := ParseSymmetricEncryption(s); return err }},
		{name: "padding mode", check: func(s string) error { _, err := ParsePaddingMode(s); return err }},
		{name: "block cipher mode", check: func(s string) error { _, err := ParseBlockCipherMode(s); return err }},
		{name: "hash", check: func(s string) error { _, err := ParseHash(s); return err }},
		{name: "named curve", check: func(s string) error { _, err := ParseNamedCurve(s); return err }},
		{name: "compression", check: func(s string) error { _, err := ParseCompression(s); return err }},
	}

	for _, tc := range invalid {
		t.Run(tc.name+" invalid", func(t *testing.T) {
			if err := tc.check("bad-value"); err == nil {
				t.Fatal("expected invalid parse error")
			}
		})
	}
}

func TestCipherSuiteAndSignatureAlgorithmHelpers(t *testing.T) {
	cs, err := ParseCipherSuite("ecdhe-aes-cbc-pkcs7-sha256")
	if err != nil {
		t.Fatalf("ParseCipherSuite failed: %v", err)
	}
	if cs.String() != "ecdhe-aes-cbc-pkcs7-sha256" {
		t.Fatalf("unexpected cipher suite string: %q", cs.String())
	}

	sa, err := ParseSignatureAlgorithm("ecdsa-none-sha384")
	if err != nil {
		t.Fatalf("ParseSignatureAlgorithm failed: %v", err)
	}
	if sa.String() != "ecdsa-none-sha384" {
		t.Fatalf("unexpected signature algorithm string: %q", sa.String())
	}

	if _, err := ParseCipherSuite("ecdhe-aes-cbc-pkcs7-bad"); err == nil {
		t.Fatal("expected ParseCipherSuite error")
	}
	if _, err := ParseSignatureAlgorithm("ecdsa-bad-sha256"); err == nil {
		t.Fatal("expected ParseSignatureAlgorithm error")
	}
}

func TestSymmetricEncryptionCapabilities(t *testing.T) {
	if blockSize, ok := SymmetricEncryption_AES.BlockSize(); !ok || blockSize <= 0 {
		t.Fatalf("expected AES block size, got %d %v", blockSize, ok)
	}
	if nonce, ok := SymmetricEncryption_ChaCha20.Nonce(); !ok || nonce <= 0 {
		t.Fatalf("expected ChaCha20 nonce size, got %d %v", nonce, ok)
	}
	if !SymmetricEncryption_AES.BlockCipherMode() || SymmetricEncryption_AES.StreamCipherMode() {
		t.Fatal("unexpected AES mode flags")
	}
	if SymmetricEncryption_ChaCha20.BlockCipherMode() || !SymmetricEncryption_ChaCha20.StreamCipherMode() {
		t.Fatal("unexpected ChaCha20 mode flags")
	}
	if _, ok := SymmetricEncryption_None.BlockSize(); ok {
		t.Fatal("expected no block size for none")
	}
	if _, ok := SymmetricEncryption_None.Nonce(); ok {
		t.Fatal("expected no nonce size for none")
	}
}

func TestBlockCipherModeCapabilities(t *testing.T) {
	if !BlockCipherMode_CTR.IV() || BlockCipherMode_CTR.Nonce() || BlockCipherMode_CTR.Padding() {
		t.Fatal("unexpected CTR capabilities")
	}
	if !BlockCipherMode_GCM.Nonce() || BlockCipherMode_GCM.IV() || BlockCipherMode_GCM.Padding() {
		t.Fatal("unexpected GCM capabilities")
	}
	if !BlockCipherMode_CBC.Padding() || !BlockCipherMode_CBC.IV() {
		t.Fatal("unexpected CBC capabilities")
	}
}
