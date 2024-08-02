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
	"log"
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

			log.Printf("saved to %s, %s", priKeyFile.Name(), pubKeyFile.Name())
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
