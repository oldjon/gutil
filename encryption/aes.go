package encryption

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// 16 bytes means using AES-128
var defaultKey = []byte{0x18, 0x60, 0xa2, 0x34, 0xcf, 0x6f, 0x8b, 0x85, 0x3d, 0xf6, 0x90, 0x34, 0x4b, 0xe7, 0x91, 0xdd}

// iv must be the same length as the Block's BlockSize (16 bytes)
// random iv is more secure, yet we decide to use fixed iv for now
var defaultIV = []byte{0x02, 0xf5, 0x98, 0x9e, 0xc0, 0x26, 0xb2, 0xdf, 0xb2, 0x44, 0x21, 0xad, 0x1a, 0x9f, 0x90, 0x8d}

var (
	// ErrInvalidBlockSize indicates block size <= 0.
	ErrInvalidBlockSize = errors.New("invalid block size")

	// ErrInvalidPKCS7Data indicates bad input to PKCS7 pad or unpad.
	ErrInvalidPKCS7Data = errors.New("invalid PKCS7 data (empty or not padded)")

	// ErrInvalidPKCS7Padding indicates PKCS7 unpad fails.
	ErrInvalidPKCS7Padding = errors.New("invalid padding on input")

	// ErrEmptyCiphertext indicates empty ciphertext.
	ErrEmptyCiphertext = errors.New("empty ciphertext")

	// ErrInvalidCiphertextSize indicates that the size of ciphertext is not a multiple of blocksize.
	ErrInvalidCiphertextSize = errors.New("invalid ciphertext size")

	// ErrInvalidKeySize indicates that the size of key is not 16 bytes
	ErrInvalidKeySize = errors.New("invalid key size")

	// ErrInvalidIVSize indicates that the size of iv is not 16 bytes
	ErrInvalidIVSize = errors.New("invalid iv size")
)

// AESEncrypt returns ciphertext which is encrypted with AES from plaintext
// mode: CBC
// padding: PKCS7
// key size: 16 bytes
// iv size: 16 bytes
func AESEncrypt(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	key, iv, err := CheckKeyAndIVSize(key, iv)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// padding with the standard of PKCS7
	plaintext, err = PKCS7Padding(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(plaintext))
	encrypter := cipher.NewCBCEncrypter(c, iv)
	encrypter.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// AESDecrypt decrypts ciphertext with AES and returns plaintext
func AESDecrypt(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	key, iv, err := CheckKeyAndIVSize(key, iv)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) == 0 {
		return nil, ErrEmptyCiphertext
	}
	// the size of ciphertext should be a multiple of block size
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, ErrInvalidCiphertextSize
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plaintext := make([]byte, len(ciphertext))
	decrypter := cipher.NewCBCDecrypter(c, iv)
	decrypter.CryptBlocks(plaintext, ciphertext)

	// unpadding
	plaintext, err = PKCS7Unpadding(plaintext, aes.BlockSize)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// PKCS7Padding pads bytes to the right of plaintext
// so that the length of plaintext after padding will
// be a multiple of blocksize(16 bytes for AES).
func PKCS7Padding(plaintext []byte, blockSize int) ([]byte, error) {
	if blockSize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	if plaintext == nil {
		return nil, ErrInvalidPKCS7Data
	}

	padding := blockSize - len(plaintext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(plaintext, padText...), nil
}

// PKCS7Unpadding validates and unpads bytes from decrypted result
// b is the original data decrypted, including the padding bytes
func PKCS7Unpadding(b []byte, blocksize int) ([]byte, error) {
	if blocksize <= 0 {
		return nil, ErrInvalidBlockSize
	}
	length := len(b)
	if b == nil || length == 0 {
		return nil, ErrInvalidPKCS7Data
	}
	if length%blocksize != 0 {
		return nil, ErrInvalidPKCS7Padding
	}
	lastByte := b[length-1] // the last byte must be a padding byte
	n := int(lastByte)      // n : how many bytes are padded
	if n == 0 || n > length {
		return nil, ErrInvalidPKCS7Padding
	}
	// each padding byte should be exactly the same with the last byte
	for i := 0; i < n; i++ {
		if b[length-n+i] != lastByte {
			return nil, ErrInvalidPKCS7Padding
		}
	}

	// drop the padding bytes
	return b[:length-n], nil
}

// CheckKeyAndIVSize requires that both key's size and iv's size must be 16 bytes
func CheckKeyAndIVSize(key []byte, iv []byte) ([]byte, []byte, error) {
	if len(key) == 0 {
		key = defaultKey
	}
	if len(iv) == 0 {
		iv = defaultIV
	}
	if len(key) != 16 {
		return nil, nil, ErrInvalidKeySize
	}
	if len(iv) != 16 {
		return nil, nil, ErrInvalidIVSize
	}
	return key, iv, nil
}
