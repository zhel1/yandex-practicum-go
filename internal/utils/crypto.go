package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/hex"
)

// Secretary defines object structure and its attributes.
type Crypto struct {
	aesgcm cipher.AEAD
	nonce  []byte
}

func NewCrypto(keyStr string) (*Crypto, error) {
	key := sha256.Sum256([]byte(keyStr))
	aesblock, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	aesgcm, err := cipher.NewGCM(aesblock)
	if err != nil {
		return nil, err
	}
	nonce := key[len(key)-aesgcm.NonceSize():]
	return &Crypto{
		aesgcm: aesgcm,
		nonce:  nonce,
	}, nil
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