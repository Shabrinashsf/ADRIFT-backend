package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrAESKeyNotSet         = errors.New("AES_KEY must be set")
	ErrAESKeyInvalidHex     = errors.New("AES_KEY must be a valid hex string")
	ErrAESKeyInvalidLength  = errors.New("AES_KEY must decode to 16, 24, or 32 bytes")
	ErrAESInvalidCiphertext = errors.New("invalid encrypted payload")
	ErrAESDecryptFailed     = errors.New("failed to decrypt payload")
)

func getAESKeyFromEnv() ([]byte, error) {
	rawKey := strings.TrimSpace(os.Getenv("AES_KEY"))
	if rawKey == "" {
		return nil, ErrAESKeyNotSet
	}

	key, err := hex.DecodeString(rawKey)
	if err != nil {
		return nil, ErrAESKeyInvalidHex
	}

	keyLen := len(key)
	if keyLen != 16 && keyLen != 24 && keyLen != 32 {
		return nil, ErrAESKeyInvalidLength
	}

	return key, nil
}

func AESEncrypt(stringToEncrypt string) (encryptedString string, err error) {
	key, err := getAESKeyFromEnv()
	if err != nil {
		return "", err
	}

	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext), nil
}

func AESDecrypt(encryptedString string) (decryptedString string, err error) {
	key, err := getAESKeyFromEnv()
	if err != nil {
		return "", err
	}

	enc, err := hex.DecodeString(strings.TrimSpace(encryptedString))
	if err != nil {
		return "", ErrAESInvalidCiphertext
	}

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()
	if len(enc) < nonceSize {
		return "", ErrAESInvalidCiphertext
	}

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrAESDecryptFailed
	}

	return string(plaintext), nil
}
