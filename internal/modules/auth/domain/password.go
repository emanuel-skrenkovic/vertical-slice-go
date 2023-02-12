package domain

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"hash"
	"strings"
)

const (
	SaltBytes = 32

	separator = ":"
)

var ErrInvalidPassword error = fmt.Errorf("given password does not match")

type HashFactory func() hash.Hash

type PasswordHasher struct{
	createHash HashFactory
}

func NewSHA256PasswordHasher(hashFactory HashFactory) *PasswordHasher {
	return &PasswordHasher{createHash: hashFactory}
}

func (h *PasswordHasher) HashPassword(password string) (string, error) {
	salt := make([]byte, SaltBytes, SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	passwordBytes := []byte(password)

	hashedBytes, err := hashPassword(h.createHash(), salt, passwordBytes)
	if err != nil {
		return "", err
	}

	base64Hash := base64.StdEncoding.EncodeToString(hashedBytes)
	base64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("%s%s%s", base64Salt, separator, base64Hash), nil
}

func (h *PasswordHasher) Verify(passwordHash, givenPassword string) error {
	parts := strings.Split(passwordHash, separator)

	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return err
	}

	givenPasswordHash, err := hashPassword(h.createHash(), salt, []byte(givenPassword))
	if err != nil {
		return err
	}

	password, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	if !bytes.Equal(givenPasswordHash, password) {
		return ErrInvalidPassword
	}

	return nil
}

func hashPassword(h hash.Hash, salt, password []byte) ([]byte, error) {
	inputLen := SaltBytes + len(password)

	inputBytes := make([]byte, 0, inputLen)
	inputBytes = append(inputBytes, salt...)
	inputBytes = append(inputBytes, password...)

	if _, err := h.Write(inputBytes); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}
