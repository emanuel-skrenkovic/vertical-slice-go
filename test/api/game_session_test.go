package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/auth/domain"
	"github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session/commands"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func Test_CreateSessionCommand_Creates_New_Session_For_User(t *testing.T) {
	// Arrange
	sessionCookie := login(t)

	createGameSessionCommand := commands.CreateSessionCommand{
		OwnerID: uuid.New(),
		Name:    uuid.New().String(),
	}

	payload, err := json.Marshal(createGameSessionCommand)
	require.NoError(t, err)

	r, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s", fixture.baseURL, "/game-sessions"),
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	// "application/json",

	r.AddCookie(&http.Cookie{
		Name:  "chess-session",
		Value: sessionCookie,
	})

	// Act
	resp, err := fixture.client.Do(r)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	location := resp.Header.Get("Location")
	require.NotEmpty(t, location)
}

func Test_CreateSessionCommand_Creates_Returns_400_When_OwnerID_Invalid(t *testing.T) {
	// Arrange
	sessionCookie := login(t)

	createGameSessionCommand := commands.CreateSessionCommand{
		OwnerID: uuid.Nil,
		Name:    uuid.New().String(),
	}

	payload, err := json.Marshal(createGameSessionCommand)
	require.NoError(t, err)

	r, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s", fixture.baseURL, "/game-sessions"),
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	r.AddCookie(&http.Cookie{
		Name:  "chess-session",
		Value: sessionCookie,
	})

	// Act
	resp, err := fixture.client.Do(r)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	location := resp.Header.Get("Location")
	require.Empty(t, location)
}

func Test_CreateSessionCommand_Creates_Returns_400_When_Name_Empty(t *testing.T) {
	// Arrange
	sessionCookie := login(t)

	createGameSessionCommand := commands.CreateSessionCommand{
		OwnerID: uuid.New(),
		Name:    "",
	}

	payload, err := json.Marshal(createGameSessionCommand)
	require.NoError(t, err)

	r, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s%s", fixture.baseURL, "/game-sessions"),
		bytes.NewReader(payload),
	)
	require.NoError(t, err)

	r.AddCookie(&http.Cookie{
		Name:  "chess-session",
		Value: sessionCookie,
	})

	// Act
	resp, err := fixture.client.Do(r)

	// Assert
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	location := resp.Header.Get("Location")
	require.Empty(t, location)
}

func Test_GetOwnedSessions_Returns_Empty_List_If_No_Active_Owned_Sessions(t *testing.T) {
	// Arrange
	sessionCookie := login(t)

	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s%s?ownerId=%s", fixture.baseURL, "/game-sessions", uuid.New().String()),
		nil,
	)
	require.NoError(t, err)

	r.AddCookie(&http.Cookie{
		Name:  "chess-session",
		Value: sessionCookie,
	})

	// Act
	resp, err := fixture.client.Do(r)

	// Assert
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response []domain.Session
	require.NoError(t, json.Unmarshal(bytes, &response))

	require.Equal(t, 0, len(response))
}

func Test_GetOwnedSessions_Returns_Sessions_Owned_By_User(t *testing.T) {
	// Arrange
	sessionCookie := login(t)

	count := 5
	ownerID := uuid.New()

	for i := 0; i < count; i++ {
		// Arrange

		createGameSessionCommand := commands.CreateSessionCommand{
			OwnerID: ownerID,
			Name:    uuid.New().String(),
		}

		payload, err := json.Marshal(createGameSessionCommand)
		require.NoError(t, err)

		r, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("%s%s", fixture.baseURL, "/game-sessions"),
			bytes.NewReader(payload),
		)
		require.NoError(t, err)

		r.AddCookie(&http.Cookie{
			Name:  "chess-session",
			Value: sessionCookie,
		})

		// Act
		_, err = fixture.client.Do(r)
		require.NoError(t, err)
	}

	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s%s?ownerId=%s", fixture.baseURL, "/game-sessions", ownerID.String()),
		nil,
	)
	require.NoError(t, err)

	r.AddCookie(&http.Cookie{
		Name:  "chess-session",
		Value: sessionCookie,
	})

	// Act
	resp, err := fixture.client.Do(r)

	// Assert
	require.NoError(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	bytes, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response []domain.Session
	require.NoError(t, json.Unmarshal(bytes, &response))

	require.Equal(t, count, len(response))
}
