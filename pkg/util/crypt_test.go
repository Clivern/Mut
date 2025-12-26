// Copyright 2025 Clivern. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package util

import "testing"

func TestEncryptDecrypt(t *testing.T) {
	cipher, err := Encrypt("test-secret", "ghs_plain_token")
	if err != nil {
		t.Fatal(err)
	}
	if cipher == "ghs_plain_token" {
		t.Fatal("expected ciphertext")
	}

	plain, err := Decrypt("test-secret", cipher)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "ghs_plain_token" {
		t.Fatalf("got %q", plain)
	}
}
