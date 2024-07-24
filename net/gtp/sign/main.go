package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"time"
)

func main() {
	cmd := &cobra.Command{
		Short: "生成签名工具。",
		PreRun: func(cmd *cobra.Command, args []string) {
			viper.BindPFlags(cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			bits := sha256.Size * viper.GetInt("sha256_len")
			priKey, err := rsa.GenerateKey(rand.Reader, bits)
			if err != nil {
				panic(err)
			}

			nowStr := time.Now().Format("2006-01-02T15_04_05")

			priKeyFile, err := os.Create(fmt.Sprintf("rsa-%d-%s.pem", bits, nowStr))
			if err != nil {
				panic(err)
			}
			defer priKeyFile.Close()

			err = pem.Encode(priKeyFile, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(priKey),
			})
			if err != nil {
				panic(err)
			}

			pubKeyFile, err := os.Create(fmt.Sprintf("ras-%d-%s.pub", bits, nowStr))
			if err != nil {
				panic(err)
			}
			defer pubKeyFile.Close()

			err = pem.Encode(pubKeyFile, &pem.Block{
				Type:  "PUBLIC KEY",
				Bytes: x509.MarshalPKCS1PublicKey(&priKey.PublicKey),
			})
			if err != nil {
				panic(err)
			}

			fmt.Printf("saved to %s, %s\n", priKeyFile.Name(), pubKeyFile.Name())
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   true,
			DisableNoDescFlag:   true,
			DisableDescriptions: true,
		},
	}
	cmd.Flags().Int("sha256_len", 64, "sha256 hash length")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
