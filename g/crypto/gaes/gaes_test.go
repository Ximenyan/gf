// Copyright 2019 gf Author(https://github.com/gogf/gf). All Rights Reserved.
//
// This Source Code Form is subject to the terms of the MIT License.
// If a copy of the MIT was not distributed with this file,
// You can obtain one at https://github.com/gogf/gf.

// go test *.go -bench=".*"

package gaes_test

import (
	"github.com/gogf/gf/g/encoding/gbase64"
	"testing"

	"github.com/gogf/gf/g/crypto/gaes"
	"github.com/gogf/gf/g/test/gtest"
)

var (
	content          = []byte("pibigstar")
	content_16, _    = gbase64.Decode("v1jqsGHId/H8onlVHR8Vaw==")
	content_24, _    = gbase64.Decode("0TXOaj5KMoLhNWmJ3lxY1A==")
	content_32, _    = gbase64.Decode("qM/Waw1kkWhrwzek24rCSA==")
	content_16_iv, _ = gbase64.Decode("DqQUXiHgW/XFb6Qs98+hrA==")
	content_32_iv, _ = gbase64.Decode("ZuLgAOii+lrD5KJoQ7yQ8Q==")
	// iv 长度必须等于blockSize，只能为16
	iv         = []byte("Hello My GoFrame")
	key_16     = []byte("1234567891234567")
	key_17     = []byte("12345678912345670")
	key_24     = []byte("123456789123456789123456")
	key_32     = []byte("12345678912345678912345678912345")
	keys       = []byte("12345678912345678912345678912346")
	key_err    = []byte("1234")
	key_32_err = []byte("1234567891234567891234567891234 ")

	// cfb模式blockSize补位长度, add by zseeker
	padding_size      = 16 - len(content)
	content_16_cfb, _ = gbase64.Decode("oSmget3aBDT1nJnBp8u6kA==")
)

func TestEncrypt(t *testing.T) {
	gtest.Case(t, func() {
		data, err := gaes.Encrypt(content, key_16)
		gtest.Assert(err, nil)
		gtest.Assert(data, []byte(content_16))
		data, err = gaes.Encrypt(content, key_24)
		gtest.Assert(err, nil)
		gtest.Assert(data, []byte(content_24))
		data, err = gaes.Encrypt(content, key_32)
		gtest.Assert(err, nil)
		gtest.Assert(data, []byte(content_32))
		data, err = gaes.Encrypt(content, key_16, iv)
		gtest.Assert(err, nil)
		gtest.Assert(data, []byte(content_16_iv))
		data, err = gaes.Encrypt(content, key_32, iv)
		gtest.Assert(err, nil)
		gtest.Assert(data, []byte(content_32_iv))
	})
}

func TestDecrypt(t *testing.T) {
	gtest.Case(t, func() {
		decrypt, err := gaes.Decrypt([]byte(content_16), key_16)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)

		decrypt, err = gaes.Decrypt([]byte(content_24), key_24)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)

		decrypt, err = gaes.Decrypt([]byte(content_32), key_32)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)

		decrypt, err = gaes.Decrypt([]byte(content_16_iv), key_16, iv)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)

		decrypt, err = gaes.Decrypt([]byte(content_32_iv), key_32, iv)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)

		decrypt, err = gaes.Decrypt([]byte(content_32_iv), keys, iv)
		gtest.Assert(err, "invalid padding")
	})
}

func TestEncryptErr(t *testing.T) {
	gtest.Case(t, func() {
		// encrypt key error
		_, err := gaes.Encrypt(content, key_err)
		gtest.AssertNE(err, nil)
	})
}

func TestDecryptErr(t *testing.T) {
	gtest.Case(t, func() {
		// decrypt key error
		encrypt, err := gaes.Encrypt(content, key_16)
		_, err = gaes.Decrypt(encrypt, key_err)
		gtest.AssertNE(err, nil)

		// decrypt content too short error
		_, err = gaes.Decrypt([]byte("test"), key_16)
		gtest.AssertNE(err, nil)

		// decrypt content size error
		_, err = gaes.Decrypt(key_17, key_16)
		gtest.AssertNE(err, nil)
	})
}

func TestPKCS5UnPaddingErr(t *testing.T) {
	gtest.Case(t, func() {
		// PKCS5UnPadding blockSize zero
		_, err := gaes.PKCS5UnPadding(content, 0)
		gtest.AssertNE(err, nil)

		// PKCS5UnPadding src len zero
		_, err = gaes.PKCS5UnPadding([]byte(""), 16)
		gtest.AssertNE(err, nil)

		// PKCS5UnPadding src len > blockSize
		_, err = gaes.PKCS5UnPadding(key_17, 16)
		gtest.AssertNE(err, nil)

		// PKCS5UnPadding src len > blockSize
		_, err = gaes.PKCS5UnPadding(key_32_err, 32)
		gtest.AssertNE(err, nil)
	})
}

func TestEncryptCFB(t *testing.T) {
	gtest.Case(t, func() {
		var padding int = 0
		data, err := gaes.EncryptCFB(content, key_16, &padding, iv)
		gtest.Assert(err, nil)
		gtest.Assert(padding, padding_size)
		gtest.Assert(data, []byte(content_16_cfb))
	})
}

func TestDecryptCFB(t *testing.T) {
	gtest.Case(t, func() {
		decrypt, err := gaes.DecryptCFB([]byte(content_16_cfb), key_16, padding_size, iv)
		gtest.Assert(err, nil)
		gtest.Assert(decrypt, content)
	})
}
