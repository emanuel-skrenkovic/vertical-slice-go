package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	gamesession "github.com/eskrenkovic/vertical-slice-go/internal/modules/game-session"

	"github.com/google/uuid"
)

func Test_CreateSessionCommand_Creates_New_Session_For_User(t *testing.T) {
	// Arrange
	createGameSessionCommand := gamesession.CreateSessionCommand{
		OwnerID: uuid.New(),
		Name:    uuid.New().String(),
	}

	payload, err := json.Marshal(createGameSessionCommand)

	// Act
	resp, err := fixture.client.Post(
		fmt.Sprintf("%s%s", fixture.baseURL, "/game-sessions"),
		"application/json",
		bytes.NewReader(payload),
	)

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code: %d received: %d", http.StatusCreated, resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		t.Errorf("empty 'Location' header")
	}
}

func Test_GetOwnedSessions_Returns_Empty_List_If_No_Active_Owned_Sessions(t *testing.T) {
	// Act
	resp, err := fixture.client.Get(
		fmt.Sprintf("%s%s?ownerId=%s", fixture.baseURL, "/game-sessions", uuid.New().String()),
	)

	// Assert
	if err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status code: %d received: %d", http.StatusCreated, resp.StatusCode)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	var response []gamesession.Session
	if err := json.Unmarshal(bytes, &response); err != nil {
		t.Errorf("unexpected error occurred: %s", err.Error())
	}

	if len(response) != 0 {
		t.Errorf("expected length: %d found: %d", 0, len(response))
	}
}
