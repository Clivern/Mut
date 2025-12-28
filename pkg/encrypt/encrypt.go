// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package encrypt

import (
	"github.com/clivern/mut/pkg/util"
	"github.com/spf13/viper"
)

// EncryptionKey returns the app encryption key.
func EncryptionKey() string {
	return viper.GetString("app.encryption.key")
}

// Encrypt seals plaintext with the app encryption key.
func Encrypt(plaintext string) (string, error) {
	return util.Encrypt(EncryptionKey(), plaintext)
}

// Decrypt opens ciphertext sealed by Encrypt.
func Decrypt(ciphertext string) (string, error) {
	return util.Decrypt(EncryptionKey(), ciphertext)
}
