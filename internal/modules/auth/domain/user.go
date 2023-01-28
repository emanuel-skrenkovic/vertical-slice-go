package domain

import "github.com/google/uuid"

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

// TODO: move to register.go ?
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
