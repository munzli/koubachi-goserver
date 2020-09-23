package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"hash/crc32"
	"io"
	"log"
	"strconv"
)

func padding(src []byte) []byte {
	padding := aes.BlockSize - (len(src) + crc32.Size) % aes.BlockSize
	padText := bytes.Repeat([]byte{0}, padding)
	return append(src, padText...)
}

func unpadding(src []byte) []byte {
	return bytes.TrimRight(src, string([]byte{0}))
}

func createCrc(data []byte) []byte {
	crc, _ := hex.DecodeString(strconv.FormatUint(uint64(crc32.ChecksumIEEE(data)), 16))
	return crc
}

func Encrypt(key, plaintext []byte) []byte {
	// CBC mode works on blocks so plaintext may need to be padded to the
	// next whole block.
	plaintext = padding(plaintext)

	// append crc
	plaintext = append(plaintext, createCrc(plaintext)...)

	if len(plaintext)%aes.BlockSize != 0 {
		log.Panic("error: plaintext is not a multiple of the block size")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panicf("error: %v", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the persist.
	data := make([]byte, aes.BlockSize+len(plaintext))
	iv := data[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Panicf("error: %v", err)
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(data[aes.BlockSize:], plaintext)

	return data
}

func Decrypt(key, data []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		log.Panicf("error: %v", err)
	}

	// The IV needs to be unique, but not secure. Therefore it's common to
	// include it at the beginning of the ciphertext.
	if len(data) < aes.BlockSize {
		log.Panic("error: persist too short")
	}
	iv := data[:aes.BlockSize]
	data = data[aes.BlockSize:]

	// CBC mode always works in whole blocks.
	if len(data)%aes.BlockSize != 0 {
		log.Panic("persist is not a multiple of the block size")
	}

	mode := cipher.NewCBCDecrypter(block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(data, data)

	// check crc
	plaintext, crc := data[:len(data) - crc32.Size], data[len(data) - crc32.Size:]
	if !bytes.Equal(crc, createCrc(plaintext)) {
		log.Panic("invalid checksum")
	}

	// trim padding
	plaintext = unpadding(plaintext)
	return plaintext
}