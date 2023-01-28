package commands

import (
	"context"
	"fmt"

	"github.com/eskrenkovic/vertical-slice-go/internal/modules/core"
	"github.com/jmoiron/sqlx"
)

type LoginCommand struct {
	Email    string
	Password string
}

func (c LoginCommand) Validate() error {
	if c.Email == "" {
		return fmt.Errorf("invalid email: '%s'", c.Email)
	}

	if c.Password == "" {
		return fmt.Errorf("invalid password")
	}

	return nil
}

type LoginCommandHandler struct {
	db *sqlx.DB
}

func (h *LoginCommandHandler) Handle(ctx context.Context, request LoginCommand) (core.Unit, error) {
	panic("TODO: implement LoginCommandHandler.Handle")
}
