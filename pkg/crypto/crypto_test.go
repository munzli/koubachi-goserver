package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestDecrypt(t *testing.T) {

	key := "00112233445566778899aabbccddeeff"
	data := "6cc6527f1d3d56c79d6b130beb76fe90cf170663be65a0952fc3ec7c280a8512" +
		"c989288a55d64514663c85725aff0224633301b7c48bc9d1d14b8b77c77c9920"

	k, err := hex.DecodeString(key)
	if err != nil {
		t.Errorf("hex decode of key resulted in eror: %s", err)
	}
	v, err := hex.DecodeString(data)
	if err != nil {
		t.Errorf("hex decode of data resulted in eror: %s", err)
	}

	expected := []byte("just some random boring test data")
	result := Decrypt(k,v)
	if bytes.Compare(result, expected) != 0 {
		t.Errorf("received \"%s\" , expected \"%s\"", result, expected)
	}
}

func TestEncrypt(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	data := []byte("another chunk of boring test data for encryption, long enough to fill multiple blocks")

	encrypted := Encrypt(key, data)
	decrypted := Decrypt(key, encrypted)

	if bytes.Compare(decrypted, data) != 0 {
		t.Errorf("received \"%s\" , expected \"%s\"", decrypted, data)
	}
}