package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

const ciphertextVersion byte = 1

type Cipher struct {
	aead cipher.AEAD
}

func NewCipher(key []byte) (Cipher, error) {
	if len(key) != 32 {
		return Cipher{}, fmt.Errorf("secret key must be exactly 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return Cipher{}, fmt.Errorf("create AES cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return Cipher{}, fmt.Errorf("create AES-GCM cipher: %w", err)
	}
	return Cipher{aead: aead}, nil
}

func (c Cipher) Encrypt(name string, plaintext []byte) ([]byte, error) {
	if c.aead == nil {
		return nil, fmt.Errorf("secret cipher is not configured")
	}
	nonce := make([]byte, c.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate secret nonce: %w", err)
	}
	payload := make([]byte, 1, 1+len(nonce)+len(plaintext)+c.aead.Overhead())
	payload[0] = ciphertextVersion
	payload = append(payload, nonce...)
	payload = c.aead.Seal(payload, nonce, plaintext, []byte(name))
	return payload, nil
}

func (c Cipher) Decrypt(name string, payload []byte) ([]byte, error) {
	if c.aead == nil {
		return nil, fmt.Errorf("secret cipher is not configured")
	}
	minimumLength := 1 + c.aead.NonceSize() + c.aead.Overhead()
	if len(payload) < minimumLength || payload[0] != ciphertextVersion {
		return nil, fmt.Errorf("invalid encrypted secret payload")
	}
	nonceEnd := 1 + c.aead.NonceSize()
	plaintext, err := c.aead.Open(nil, payload[1:nonceEnd], payload[nonceEnd:], []byte(name))
	if err != nil {
		return nil, fmt.Errorf("decrypt secret: %w", err)
	}
	return plaintext, nil
}
