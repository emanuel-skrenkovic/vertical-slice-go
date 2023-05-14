package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/tql"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_Register_Registers_New_User_When_Email_Unique(t *testing.T) {
	// Arrange
	command := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.New().String()),
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	// Act
	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		http.MethodPost,
		command,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)

	// Assert
	require.NoError(t, err)

	user, err := tql.QueryFirst[domain.User](
		context.Background(),
		fixture.db,
		"SELECT * FROM auth.user WHERE email = $1;",
		command.Email,
	)
	require.NoError(t, err)

	require.Equal(t, command.Username, user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.NotEqual(t, uuid.Nil, user.SecurityStamp)
	require.False(t, user.Locked)
	require.False(t, user.EmailConfirmed)
	require.Zero(t, user.UnsuccessfulLoginAttempts)

	code, err := tql.QueryFirst[domain.ActivationCode](
		context.Background(),
		fixture.db,
		"SELECT * FROM auth.activation_code WHERE user_id = $1;",
		user.ID,
	)
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

	registerPath := fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations")
	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		registerPath,
		http.MethodPost,
		command1,
	)
	require.NoError(t, err)

	command := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.New().String()),
		Username: username,
		Password: uuid.New().String(),
	}

	// Act
	_, err = sendRequest[commands.RegisterCommand, any](
		fixture.client,
		registerPath,
		http.MethodPost,
		command,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)

	// Assert
	require.NoError(t, err)

	count, err := tql.QueryFirst[int](
		context.Background(),
		fixture.db,
		"SELECT COUNT(id) FROM auth.user WHERE username = $1;",
		username,
	)
	require.NoError(t, err)

	expectedUsersCount := 1
	require.Equal(t, expectedUsersCount, count)

	user, err := tql.QueryFirst[domain.User](
		context.Background(),
		fixture.db,
		"SELECT * FROM auth.user WHERE username = $1;",
		username,
	)
	require.NoError(t, err)

	require.Equal(t, username, user.Username)
	require.NotEmpty(t, user.PasswordHash)
	require.NotEqual(t, uuid.Nil, user.SecurityStamp)
	require.False(t, user.Locked)
	require.False(t, user.EmailConfirmed)
	require.Zero(t, user.UnsuccessfulLoginAttempts)

	code, err := tql.QueryFirst[domain.ActivationCode](
		context.Background(),
		fixture.db,
		"SELECT * FROM auth.activation_code WHERE user_id = $1;",
		user.ID,
	)
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

	registerPath := fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations")
	_, err := sendRequest[commands.RegisterCommand, any](fixture.client, registerPath, http.MethodPost, command1)
	require.NoError(t, err)

	command := commands.RegisterCommand{
		Email:    email,
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	// Act
	_, err = sendRequest[commands.RegisterCommand, any](
		fixture.client,
		registerPath,
		http.MethodPost,
		command,
		func(resp *http.Response) {
			require.Equal(t, http.StatusOK, resp.StatusCode)
		},
	)
	// Assert
	require.NoError(t, err)

	count, err := tql.QueryFirst[int](context.Background(), fixture.db, "SELECT COUNT(id) FROM auth.user WHERE email = $1;", command.Email)
	require.NoError(t, err)

	expectedUsersCount := 1
	require.Equal(t, expectedUsersCount, count)

	user, err := tql.QueryFirst[domain.User](context.Background(), fixture.db, "SELECT * FROM auth.user WHERE email = $1;", command.Email)
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

	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		http.MethodPost,
		registerUserCommand,
	)
	require.NoError(t, err)

	// Act
	_, err = sendRequest[any, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/publish-confirmation-emails"),
		http.MethodPost,
		nil,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)

	// Assert
	require.NoError(t, err)

	const q = `
		SELECT ac.*
		FROM auth.activation_code ac
		INNER JOIN auth.user u ON ac.user_id = u.id
		WHERE u.email = $1;`
	activationCode, err := tql.QueryFirst[domain.ActivationCode](context.Background(), fixture.db, q, registerUserCommand.Email)
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

	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		http.MethodPost,
		registerUserCommand,
	)
	require.NoError(t, err)

	_, err = sendRequest[any, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/publish-confirmation-emails"),
		http.MethodPost,
		nil,
	)
	require.NoError(t, err)

	userID, err := tql.QueryFirst[uuid.UUID](
		context.Background(),
		fixture.db,
		"SELECT id FROM auth.user WHERE email = $1;",
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	// Act
	reSendConfirmationCommand := commands.ReSendActivationEmailCommand{UserID: userID}
	_, err = sendRequest[commands.ReSendActivationEmailCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations/actions/send-activation-code"),
		http.MethodPost,
		reSendConfirmationCommand,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)

	// Assert
	require.NoError(t, err)

	const q = `
		SELECT ac.*
		FROM auth.activation_code ac
		INNER JOIN auth.user u ON ac.user_id = u.id
		WHERE u.email = $1;`
	activationCodes, err := tql.Query[domain.ActivationCode](context.Background(), fixture.db, q, registerUserCommand.Email)
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

	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		http.MethodPost,
		registerUserCommand,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)
	require.NoError(t, err)

	token, err := tql.QueryFirst[string](
		context.Background(),
		fixture.db,
		`SELECT
			token
		FROM
			auth.activation_code ac
		INNER JOIN auth.user u ON u.id = ac.user_id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	fakeToken := uuid.New()

	// Act
	_, err = sendRequest[any, any](
		fixture.client,
		fmt.Sprintf("%s%s?token=%s", fixture.baseURL, "/auth/registrations/actions/confirm", fakeToken),
		http.MethodPost,
		nil,
		func(resp *http.Response) { require.Equal(t, http.StatusBadRequest, resp.StatusCode) },
	)

	// Assert
	require.NoError(t, err)

	confirmed, err := tql.QueryFirst[bool](
		context.Background(),
		fixture.db,
		"SELECT email_confirmed FROM auth.user WHERE email = $1;",
		registerUserCommand.Email,
	)
	require.NoError(t, err)
	require.False(t, confirmed)

	activationCodeUsed, err := tql.QueryFirst[bool](
		context.Background(),
		fixture.db,
		"SELECT used FROM auth.activation_code WHERE token = $1;",
		token,
	)
	require.NoError(t, err)
	require.False(t, activationCodeUsed)
}

func Test_Login_With_Valid_Password_Succeeds(t *testing.T) {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@test.com", uuid.NewString()),
		Username: uuid.New().String(),
		Password: uuid.New().String(),
	}

	_, err := sendRequest[commands.RegisterCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/registrations"),
		http.MethodPost,
		registerUserCommand,
		func(resp *http.Response) { require.Equal(t, http.StatusOK, resp.StatusCode) },
	)
	require.NoError(t, err)

	token, err := tql.QueryFirst[string](
		context.Background(),
		fixture.db,
		`SELECT
			token
		FROM
			auth.activation_code ac
		INNER JOIN auth.user u ON u.id = ac.user_id
		WHERE u.email = $1;`,
		registerUserCommand.Email,
	)
	require.NoError(t, err)

	_, err = sendRequest[commands.VerifyRegistrationCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s?token=%s", fixture.baseURL, "/auth/registrations/actions/confirm", token),
		http.MethodPost,
		commands.VerifyRegistrationCommand{Token: token},
	)
	require.NoError(t, err)

	// Act
	loginCommand := commands.LoginCommand{
		Email:    registerUserCommand.Email,
		Password: registerUserCommand.Password,
	}

	_, err = sendRequest[commands.LoginCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/login"),
		http.MethodPost,
		loginCommand,
		func(resp *http.Response) {
			require.Equal(t, http.StatusOK, resp.StatusCode)

			require.Greater(t, len(resp.Cookies()), 0)
		},
	)
	require.NoError(t, err)
}
