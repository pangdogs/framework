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

package gtp

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"os"
)

// LoadPublicKeyFile 加载公钥文件
func LoadPublicKeyFile(filePath string) (*rsa.PublicKey, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadPublicKey(f)
}

// LoadPrivateKeyFile 加载私钥文件
func LoadPrivateKeyFile(filePath string) (*rsa.PrivateKey, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ReadPrivateKey(f)
}

// ReadPublicKey 读取公钥
func ReadPublicKey(reader io.Reader) (*rsa.PublicKey, error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bs)

	pubKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return pubKey, nil
}

// ReadPrivateKey 读取私钥
func ReadPrivateKey(reader io.Reader) (*rsa.PrivateKey, error) {
	bs, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(bs)

	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}
