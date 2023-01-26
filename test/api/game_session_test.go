package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	if resp == nil {
		t.Errorf("response is nil")
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected status code: %d received: %d", http.StatusCreated, resp.StatusCode)
	}
}

func Test_GetOwnedSessions_Returns_Empty_List_If_No_Active_Owned_Sessions(t *testing.T) {
	// TODO
}
