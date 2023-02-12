package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Password_Matches_Hash(t *testing.T) {
	// Arrange
	password := uuid.NewString()

	hasher := NewSHA256PasswordHasher()

	passwordHash, err := hasher.HashPassword(password)

	require.NoError(t, err)
	require.NotEmpty(t, passwordHash)

	// Act
	err = hasher.Verify(passwordHash, password)

	// Assert
	require.NoError(t, err)
}
