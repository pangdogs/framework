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
			shaBits := viper.GetInt("sha_bits")
			var rsaBits int

			switch shaBits {
			case 256:
				rsaBits = 2048
			case 384:
				rsaBits = 3072
			case 512:
				rsaBits = 4096
			default:
				panic("sha_bits must be 256, 384 or 512")
			}

			priKey, err := rsa.GenerateKey(rand.Reader, rsaBits)
			if err != nil {
				panic(err)
			}

			nowStr := time.Now().Format("2006-01-02T15_04_05")

			priKeyFile, err := os.Create(fmt.Sprintf("rsa-%d-%s.pem", rsaBits, nowStr))
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

			pubKeyFile, err := os.Create(fmt.Sprintf("rsa-%d-%s.pub", rsaBits, nowStr))
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
	cmd.Flags().Int("sha_bits", 512, "sha hash bits")

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
