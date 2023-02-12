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

func Test_ProcessActivationCodes_Updates_ActivationCode_On_Send(t *testing.T) {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    "test@test.com",
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(registerUserCommand)
	require.NoError(t, err)

	// Act
	_, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/publish-confirmation-emails"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var activationCode domain.ActivationCode
	err = fixture.db.Get(
		&activationCode,
		`SELECT ac.*
		FROM auth.activation_code ac
		INNER JOIN auth.user u ON ac.user_id = u.id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	require.NotNil(t, activationCode.SentAt)
	require.Equal(t, time.Now().UTC().Day(), activationCode.SentAt.Day())
	require.Equal(t, time.Now().UTC().Month(), activationCode.SentAt.Month())
	require.Equal(t, time.Now().UTC().Year(), activationCode.SentAt.Year())
}

func Test_SendActivationCode_Creates_New_ActivationCode(t *testing.T) {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.NewString()),
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(registerUserCommand)
	require.NoError(t, err)

	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/publish-confirmation-emails"),
		"application/json",
		nil,
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var userID uuid.UUID
	err = fixture.db.Get(&userID, "SELECT id FROM auth.user WHERE email = $1;", registerUserCommand.Email)
	require.NoError(t, err)


	// Act
	reSendConfirmationCommand := commands.ReSendActivationEmailCommand{UserID: userID}
	payload, err = json.Marshal(reSendConfirmationCommand)
	require.NoError(t, err)

	resp, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/send-activation-code"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var activationCodes []domain.ActivationCode
	err = fixture.db.Select(
		&activationCodes,
		`SELECT ac.*
		FROM auth.activation_code ac
		INNER JOIN auth.user u ON ac.user_id = u.id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	require.Len(t, activationCodes, 2)

	for _, activationCode := range activationCodes {
		require.NotNil(t, activationCode.SentAt)
		require.Equal(t, time.Now().UTC().Day(), activationCode.SentAt.Day())
		require.Equal(t, time.Now().UTC().Month(), activationCode.SentAt.Month())
		require.Equal(t, time.Now().UTC().Year(), activationCode.SentAt.Year())
	}
}

func Test_VerifyRegistration_Returns_Error_When_ActivationCode_Is_Invalid(t *testing.T) {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.NewString()),
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(registerUserCommand)
	require.NoError(t, err)

	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var token string
	err = fixture.db.Get(
		&token,
		`SELECT
			token
		FROM
			auth.activation_code ac
		INNER JOIN auth.user u ON u.id = ac.user_id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	verifyRegistrationCommand := commands.VerifyRegistrationCommand{Token: token}

	payload, err = json.Marshal(verifyRegistrationCommand)
	require.NoError(t, err)

	// Act
	resp, err = fixture.client.Post(
		fmt.Sprintf("%s%s?token=%s", fixture.baseURL, "/auth/registrations/actions/confirm", token),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var confirmed bool
	err = fixture.db.Get(
		&confirmed,
		"SELECT email_confirmed FROM auth.user WHERE email = $1;",
		registerUserCommand.Email,
	)
	require.NoError(t, err)
	require.True(t, confirmed)

	var activationCodeUsed bool
	err = fixture.db.Get(
		&activationCodeUsed,
		"SELECT used FROM auth.activation_code WHERE token = $1;",
		token,
	)
	require.NoError(t, err)
	require.True(t, activationCodeUsed)
}

func Test_Login_With_Valid_Password_Succeeds(t *testing.T) {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.NewString()),
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	payload, err := json.Marshal(registerUserCommand)
	require.NoError(t, err)

	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var token string
	err = fixture.db.Get(
		&token,
		`SELECT
			token
		FROM
			auth.activation_code ac
		INNER JOIN auth.user u ON u.id = ac.user_id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	verifyRegistrationCommand := commands.VerifyRegistrationCommand{Token: token}

	payload, err = json.Marshal(verifyRegistrationCommand)
	require.NoError(t, err)

	resp, err = fixture.client.Post(
		fmt.Sprintf("%s%s?token=%s", fixture.baseURL, "/auth/registrations/actions/confirm", token),
		"application/json",
		bytes.NewReader(payload),
	)

	// Act
	loginCommand := commands.LoginCommand{
		Email: registerUserCommand.Email,
		Password: registerUserCommand.Password,
	}

	payload, err = json.Marshal(loginCommand)
	require.NoError(t, err)

	resp, err = fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/login"),
		"application/json",
		bytes.NewReader(payload),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	require.Greater(t, len(resp.Cookies()), 0)
}
