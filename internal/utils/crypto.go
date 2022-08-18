//Package utils provides helper functions for all app.
package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

// Crypto defines object structure and its attributes.
type Crypto struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

func NewCrypto(keyStr string) *Crypto {
	key := sha256.Sum256([]byte(keyStr))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		panic("failed to create Crypto: " + err.Error())
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		panic("failed to create Crypto: " + err.Error())
	}
	nonce := key[len(key)-aesgcm.NonceSize():]
	return &Crypto{
		aesgcm: aesgcm,
		nonce:  nonce,
	}
}

func (s *Crypto) Encode(data string) string {
	encoded := s.aesgcm.Seal(nil, s.nonce, []byte(data), nil)
	return hex.EncodeToString(encoded)
}

func (s *Crypto) Decode(msg string) (string, error) {
	msgBytes, err := hex.DecodeString(msg)
	if err != nil {
		return "", err
	}
	decoded, err := s.aesgcm.Open(nil, s.nonce, msgBytes, nil)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
