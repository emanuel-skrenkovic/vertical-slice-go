package domain

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                        uuid.UUID `db:"id"`
	SecurityStamp             uuid.UUID `db:"security_stamp"`
	Username                  string    `db:"username"`
	Email                     string    `db:"email"`
	PasswordHash              string    `db:"password_hash"`
	EmailConfirmed            bool      `db:"email_confirmed"`
	Locked                    bool      `db:"locked"`
	UnsuccessfulLoginAttempts int       `db:"unsuccessful_login_attempts"`
}

func RegisterUser(
	username string,
	email string,
	password string,
	passwordHasher PasswordHasher,
) (User, error) {
	passwordHash, err := passwordHasher.HashPassword(password)
	if err != nil {
		return User{}, err
	}

	return User{
		ID:            uuid.New(),
		SecurityStamp: uuid.New(),
		Username:      username,
		Email:         email,
		PasswordHash:  passwordHash,
	}, nil
}

var ErrSessionExpired = errors.New("session expired")

type Session struct {
	ID           uuid.UUID `db:"id"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
	UserID       uuid.UUID `db:"user_id"`
	ExpiresAtUTC time.Time `db:"expires_at"`
}

func (s Session) Validate() error {
	if time.Now().UTC().After(s.ExpiresAtUTC) {
		return ErrSessionExpired
	}

	return nil
}

func (u *User) Authenticate(password string, passwordHasher PasswordHasher) (Session, error) {
	err := passwordHasher.Verify(u.PasswordHash, password)
	if err == nil {
		u.UnsuccessfulLoginAttempts = 0

		now := time.Now().UTC()
		return Session{
			ID:           uuid.New(),
			CreatedAt:    now,
			UpdatedAt:    now,
			UserID:       u.ID,
			ExpiresAtUTC: now.Add(15 * time.Minute), // TODO: from configuration?

		}, nil
	}

	reason := err.Error()

	u.UnsuccessfulLoginAttempts++

	if u.UnsuccessfulLoginAttempts >= 3 {
		u.Locked = true
		u.SecurityStamp = uuid.New()
		reason = "account locked"
	}

	return Session{}, fmt.Errorf("authentication failed: %s", reason)
}
