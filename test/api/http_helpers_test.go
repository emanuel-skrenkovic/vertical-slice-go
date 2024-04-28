package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/commands"

	"github.com/eskrenkovic/tql"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type responseAssertion func(*http.Response)

func sendRequest[TReq any, TResp any](
	c *http.Client,
	url string,
	method string,
	req TReq,
	opts ...responseAssertion,
) (TResp, error) {
	var resp TResp

	payload, err := json.Marshal(req)
	if err != nil {
		return resp, err
	}

	httpReq, err := http.NewRequest(method, url, bytes.NewReader(payload))
	if err != nil {
		return resp, err
	}

	httpResp, err := c.Do(httpReq)
	if err != nil {
		return resp, err
	}

	for _, opt := range opts {
		opt(httpResp)
	}

	if httpResp.ContentLength > 0 {
		defer func() {
			_ = httpResp.Body.Close()
		}()

		responsePayload, err := io.ReadAll(httpResp.Body)
		if err != nil {
			return resp, err
		}

		if err := json.Unmarshal(responsePayload, &resp); err != nil {
			return resp, err
		}
	}

	return resp, err
}

func login(t *testing.T) string {
	// Arrange
	registerUserCommand := commands.RegisterCommand{
		Email:    fmt.Sprintf("%s@tests.com", uuid.NewString()),
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

	var cookie string

	_, err = sendRequest[commands.LoginCommand, any](
		fixture.client,
		fmt.Sprintf("%s%s", fixture.baseURL, "/auth/login"),
		http.MethodPost,
		loginCommand,
		func(resp *http.Response) {
			require.Equal(t, http.StatusOK, resp.StatusCode)
			require.Greater(t, len(resp.Cookies()), 0)

			for _, c := range resp.Cookies() {
				if c.Name != "chess-session" {
					continue
				}

				cookie = c.Value
				break
			}
		},
	)
	require.NoError(t, err)

	if cookie == "" {
		t.Error("found no cookie 'chess-session'")
		t.Fail()
	}

	return cookie
}
