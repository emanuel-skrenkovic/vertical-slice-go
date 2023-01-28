package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ConfirmationToken struct {
	UserID        uuid.UUID
	SecurityStamp uuid.UUID
	ExpirationUTC time.Time
}

func CreateRegistrationConfirmationToken(user User, expiration time.Duration) (string, error) {
	token := ConfirmationToken{
		UserID:        user.ID,
		SecurityStamp: user.SecurityStamp,
		ExpirationUTC: time.Now().UTC().Add(expiration),
	}

	serialized, err := json.Marshal(&token)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(serialized), nil
}

func ParseConfirmationToken(token string) (ConfirmationToken, error) {
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return ConfirmationToken{}, err
	}

	var parsedToken ConfirmationToken
	return parsedToken, json.Unmarshal(tokenBytes, &parsedToken)
}

func ValidateUserConfirmationToken(token ConfirmationToken, user User) error {
	if token.ExpirationUTC.After(time.Now().UTC()) {
		return fmt.Errorf("confirmation token expired")
	}

	if token.SecurityStamp != user.SecurityStamp {
		return fmt.Errorf("token security stamp does not match the user security stamp")
	}

	// TODO: what about user id?

	return nil
}
