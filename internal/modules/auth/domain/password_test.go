package domain

import (
	"crypto/sha256"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Password_Matches_Hash(t *testing.T) {
	// Arrange
	password := uuid.NewString()

	hasher := NewPasswordHasher(sha256.New)

	passwordHash, err := hasher.HashPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, passwordHash)

	// Act
	err = hasher.Verify(passwordHash, password)

	// Assert
	require.NoError(t, err)
}
