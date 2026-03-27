package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
)

const (
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024
	Argon2Threads = 4
	Argon2KeyLen  = 32
)

type Crypto struct{}

func New() *Crypto {
	return &Crypto{}
}

func (c *Crypto) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt hash failed: %w", err)
	}
	return string(hash), nil
}

func (c *Crypto) VerifyPassword(password, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (c *Crypto) HashAPIKey(key string) string {
	hash := argon2.IDKey([]byte(key), []byte("api-key"), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
	return base64.RawURLEncoding.EncodeToString(hash)
}

func (c *Crypto) VerifyAPIKey(key, hash string) bool {
	expected := argon2.IDKey([]byte(key), []byte("api-key"), Argon2Time, Argon2Memory, Argon2Threads, Argon2KeyLen)
	decoded, err := base64.RawURLEncoding.DecodeString(hash)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(expected, decoded) == 1
}

func GenerateAPIKey(prefix string) (key string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return fmt.Sprintf("%s_%s", prefix, base64.RawURLEncoding.EncodeToString(bytes)), nil
}

func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
