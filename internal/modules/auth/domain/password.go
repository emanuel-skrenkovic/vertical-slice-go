package domain

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"hash"
	"strings"
)

const (
	SaltBytes = 32

	separator = ":"
)

var (
	_ PasswordHasher = (*SHA256PasswordHasher)(nil)

	ErrInvalidPassword error = fmt.Errorf("given password does not match")
)

// TODO: remove interface. The hash.Hash interface serves the intended purpose of having
// plug-inable hash algorithms.
// TODO: think about saving the hash algorithm alongside password. Is that secure?
type PasswordHasher interface {
	HashPassword(password string) (string, error)
	Verify(passwordHash, givenPassword string) error
}

type SHA256PasswordHasher struct {
	hasher hash.Hash
}

func NewSHA256PasswordHasher() *SHA256PasswordHasher {
	return &SHA256PasswordHasher{hasher: sha256.New()}
}

func (h *SHA256PasswordHasher) HashPassword(password string) (string, error) {
	salt := make([]byte, SaltBytes, SaltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	passwordBytes := []byte(password)

	hashedBytes := h.hashPassword(salt, passwordBytes)

	base64Hash := base64.StdEncoding.EncodeToString(hashedBytes)
	base64Salt := base64.StdEncoding.EncodeToString(salt)

	return fmt.Sprintf("%s%s%s", base64Salt, separator, base64Hash), nil
}

func (h *SHA256PasswordHasher) Verify(passwordHash, givenPassword string) error {
	saltPart := strings.Split(passwordHash, separator)[0]
	salt, err := base64.StdEncoding.DecodeString(saltPart)
	if err != nil {
		return err
	}

	givenPasswordHash := h.hashPassword(salt, []byte(givenPassword))

	if !bytes.Equal(givenPasswordHash, []byte(passwordHash)) {
		return ErrInvalidPassword
	}

	return nil
}

func (h *SHA256PasswordHasher) hashPassword(salt, password []byte) []byte {
	inputLen := SaltBytes + len(password)

	inputBytes := make([]byte, 0, inputLen)
	inputBytes = append(inputBytes, salt...)
	inputBytes = append(inputBytes, password...)

	return h.hasher.Sum(inputBytes)
}
