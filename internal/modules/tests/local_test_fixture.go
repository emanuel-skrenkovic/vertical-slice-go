package tests

import (
	"context"
	"os"

	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LocalTestFixture struct {
	compose tc.ComposeStack
}

func NewLocalTestFixture(dockerComposePath string, strategies map[string]wait.Strategy) (LocalTestFixture, error) {
	compose, err := tc.NewDockerCompose(dockerComposePath)
	if err != nil {
		return LocalTestFixture{}, err
	}

	fixture := LocalTestFixture{compose}

	for serviceName, strategy := range strategies {
		fixture.compose = compose.WaitForService(serviceName, strategy)
	}

	return fixture, nil
}

func (f *LocalTestFixture) Start(ctx context.Context) error {
	if skip := os.Getenv("SKIP_INFRASTRUCTURE"); skip == "true" {
		return nil
	}

	return f.compose.Up(ctx)
}

func (f *LocalTestFixture) Stop(ctx context.Context) error {
	if skip := os.Getenv("SKIP_INFRASTRUCTURE"); skip == "true" {
		return nil
	}

	return f.compose.Down(ctx)
}
