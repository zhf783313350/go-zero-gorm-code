package util

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"golang.org/x/crypto/scrypt"
)

const (
	saltSize = 16
	keyLen   = 32
	n        = 1 << 14
	r        = 8
	p        = 1
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash, err := scrypt.Key([]byte(password), salt, n, r, p, keyLen)
	if err != nil {
		return "", err
	}

	result := make([]byte, saltSize+keyLen)
	copy(result[:saltSize], salt)
	copy(result[saltSize:], hash)

	return base64.StdEncoding.EncodeToString(result), nil
}

func VerifyPassword(hashedPassword, password string) bool {
	data, err := base64.StdEncoding.DecodeString(hashedPassword)
	if err != nil {
		return false
	}

	if len(data) < saltSize+keyLen {
		return false
	}

	salt := data[:saltSize]
	hash := data[saltSize : saltSize+keyLen]

	computedHash, err := scrypt.Key([]byte(password), salt, n, r, p, keyLen)
	if err != nil {
		return false
	}

	return subtle.ConstantTimeCompare(hash, computedHash) == 1
}
