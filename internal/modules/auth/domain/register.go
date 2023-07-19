package domain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"time"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/google/uuid"
)

type ActivationCode struct {
	ID            int64      `db:"id"`
	UserID        uuid.UUID  `db:"user_id"`
	SecurityStamp uuid.UUID  `db:"security_stamp"`
	ExpiresAt     time.Time  `db:"expires_at"`
	SentAt        *time.Time `db:"sent_at"`
	Token         string     `db:"token"`
	// Is this needed if we have sent_at?
	Used bool `db:"used"`
}

func CreateRegistrationActivationCode(user User, expiration time.Duration, h hash.Hash) (ActivationCode, error) {
	code := ActivationCode{
		UserID:        user.ID,
		SecurityStamp: user.SecurityStamp,
		ExpiresAt:     time.Now().UTC().Add(expiration),
	}

	serialized, err := json.Marshal(code)
	if err != nil {
		return ActivationCode{}, err
	}

	securityBytes, err := user.SecurityStamp.MarshalBinary()
	if err != nil {
		return ActivationCode{}, err
	}

	inputLen := len(securityBytes) + len(serialized)

	inputBytes := make([]byte, 0, inputLen)
	inputBytes = append(inputBytes, securityBytes...)
	inputBytes = append(inputBytes, serialized...)

	if _, err := h.Write(inputBytes); err != nil {
		return ActivationCode{}, err
	}

	hashed := h.Sum(nil)
	code.Token = base64.StdEncoding.EncodeToString(hashed)

	return code, nil
}

func ValidateUserActivationCode(code ActivationCode, user User) error {
	if time.Now().UTC().After(code.ExpiresAt) {
		return fmt.Errorf("confirmation token expired")
	}

	if code.SecurityStamp != user.SecurityStamp {
		return fmt.Errorf("token security stamp does not match the user security stamp")
	}

	// TODO: what about user id?

	return nil
}

func RegistrationActivationEmail(user User, sender string) core.MailMessage {
	return core.MailMessage{
		Subject:    "Chess account verification",
		From:       sender,
		To:         []string{user.Email},
		IsHTML:     true,
		BodyString: "This is an email verification mail",
	}
}
