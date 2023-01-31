package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Register_Registers_New_User_When_Email_Unique(t *testing.T) {
	// Arrange
	command := commands.RegisterCommand{
		Email:    "test@test.com",
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(command)
	require.NoError(t, err)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var user domain.User
	err = fixture.db.Get(&user, "SELECT * FROM auth.user WHERE email = $1;", command.Email)
	require.NoError(t, err)

	require.Equal(t, command.Username, user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.NotEqual(t, uuid.Nil, user.SecurityStamp)
	require.False(t, user.Locked)
	require.False(t, user.EmailConfirmed)
	require.Zero(t, user.UnsuccessfulLoginAttempts)

	var code domain.ActivationCode
	err = fixture.db.Get(&code, "SELECT * FROM auth.activation_code WHERE user_id = $1;", user.ID)
	require.NoError(t, err)

	require.Equal(t, user.SecurityStamp, code.SecurityStamp)
	require.NotEmpty(t, code.Token)
	require.False(t, code.Used)
	require.Less(t, time.Now().UTC(), code.ExpiresAt)
}

func Test_Register_Does_Not_Create_Another_User_When_Username_Exists(t *testing.T) {
	// Arrange
	username := uuid.New().String()
	command1 := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.New().String()),
		Username: username,
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(command1)
	require.NoError(t, err)

	_, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	command := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.New().String()),
		Username: username,
		Password: uuid.New().String(),
	}

	payload, err = json.Marshal(command)
	require.NoError(t, err)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var count int
	err = fixture.db.Get(&count, "SELECT COUNT(id) FROM auth.user WHERE username = $1;", username)
	require.NoError(t, err)

	expectedUsersCount := 1
	require.Equal(t, expectedUsersCount, count)

	var user domain.User
	err = fixture.db.Get(&user, "SELECT * FROM auth.user WHERE username = $1;", username)
	require.NoError(t, err)

	require.Equal(t, username, user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.NotEqual(t, uuid.Nil, user.SecurityStamp)
	require.False(t, user.Locked)
	require.False(t, user.EmailConfirmed)
	require.Zero(t, user.UnsuccessfulLoginAttempts)

	var code domain.ActivationCode
	err = fixture.db.Get(&code, "SELECT * FROM auth.activation_code WHERE user_id = $1;", user.ID)
	require.NoError(t, err)

	require.Equal(t, user.SecurityStamp, code.SecurityStamp)
	require.NotEmpty(t, code.Token)
	require.False(t, code.Used)
	require.Less(t, time.Now().UTC(), code.ExpiresAt)
}

func Test_Register_Does_Not_Create_Another_User_When_Email_Exists(t *testing.T) {
	// Arrange
	email := fmt.Sprintf("%s@test.com", uuid.New().String())
	command1 := commands.RegisterCommand{
		Email:    email,
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(command1)
	require.NoError(t, err)

	_, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	command := commands.RegisterCommand{
		Email:    email,
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err = json.Marshal(command)
	require.NoError(t, err)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var count int
	err = fixture.db.Get(&count, "SELECT COUNT(id) FROM auth.user WHERE email = $1;", command.Email)
	require.NoError(t, err)

	expectedUsersCount := 1
	require.Equal(t, expectedUsersCount, count)

	var user domain.User
	err = fixture.db.Get(&user, "SELECT * FROM auth.user WHERE email = $1;", command.Email)
	require.NoError(t, err)

	require.Equal(t, command1.Username, user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.NotEqual(t, uuid.Nil, user.SecurityStamp)
	require.False(t, user.Locked)
	require.False(t, user.EmailConfirmed)
	require.Zero(t, user.UnsuccessfulLoginAttempts)
}
