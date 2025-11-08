package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptionService handles encryption and decryption of sensitive data
type EncryptionService struct {
	key []byte
}

// NewEncryptionService creates a new encryption service
// key should be 32 bytes for AES-256
func NewEncryptionService(encryptionKey string) (*EncryptionService, error) {
	key := []byte(encryptionKey)

	// Ensure key is 32 bytes for AES-256
	if len(key) != 32 {
		return nil, errors.New("encryption key must be exactly 32 bytes for AES-256")
	}

	return &EncryptionService{
		key: key,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (s *EncryptionService) Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and encrypted data
	nonce, encryptedData := data[:nonceSize], data[nonceSize:]

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, encryptedData, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// EncryptIfNotEmpty encrypts the value if it's not empty
func (s *EncryptionService) EncryptIfNotEmpty(value *string) (*string, error) {
	if value == nil || *value == "" {
		return value, nil
	}

	encrypted, err := s.Encrypt(*value)
	if err != nil {
		return nil, err
	}

	return &encrypted, nil
}

// DecryptIfNotEmpty decrypts the value if it's not empty
func (s *EncryptionService) DecryptIfNotEmpty(value *string) (*string, error) {
	if value == nil || *value == "" {
		return value, nil
	}

	decrypted, err := s.Decrypt(*value)
	if err != nil {
		return nil, err
	}

	return &decrypted, nil
}
