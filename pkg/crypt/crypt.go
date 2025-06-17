package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64" 
	"fmt"
	"io"

	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/pbkdf2" 
	"crypto/sha256" 
)

const (
	// KeySize is the size of the AES key in bytes (e.g., 32 for AES-256)
	KeySize = 32
	// SaltSize is the recommended size for the salt (e.g., 16 bytes)
	SaltSize = 16
	// PBKDF2Iterations is the number of iterations for PBKDF2 (should be high)
	PBKDF2Iterations = 100000 // recommended minimum
)

// this generates a bcrypt hash of the provided password.
func GenerateHash(password string, cost int) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to generate bcrypt hash: %w", err)
	}
	return string(hashedPassword), nil
}

// compares a bcrypt hashed password with its plaintext equivalent.
func CompareHashAndPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// generates a cryptographically secure random salt.
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}
	return salt, nil
}

// derives a cryptographic key from a password and salt using PBKDF2.
func DeriveKey(password []byte, salt []byte) []byte {
	return pbkdf2.Key(password, salt, PBKDF2Iterations, KeySize, sha256.New)
}

// encrypts plaintext using AES-256 GCM.
// it returns the Base64-encoded ciphertext and the nonce (IV).
func Encrypt(plaintext []byte, key []byte) (base64Ciphertext string, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	nonce = make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nonce, nil 
}

// decrypts Base64-encoded ciphertext using AES-256 GCM.
func Decrypt(base64Ciphertext string, nonce, key []byte) (plaintext []byte, err error) {
	ciphertext, err := base64.StdEncoding.DecodeString(base64Ciphertext)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode ciphertext: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM cipher: %w", err)
	}

	if len(nonce) != aesGCM.NonceSize() {
		return nil, fmt.Errorf("invalid nonce size")
	}

	plaintext, err = aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt ciphertext (likely wrong key/password): %w", err)
	}
	return plaintext, nil
}