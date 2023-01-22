package test

import (
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

type LocalTestFixture struct {
	dockerComposePath string
	compose           testcontainers.DockerCompose
}

func NewLocalTestFixture(dockerComposePath string) (LocalTestFixture, error) {
	compose := testcontainers.NewLocalDockerCompose(
		[]string{dockerComposePath},
		uuid.New().String(),
	)

	f := LocalTestFixture{
		dockerComposePath: dockerComposePath,
		compose:           compose.WithCommand([]string{"up", "--build", "-d"}),
	}

	return f, nil
}

func (f *LocalTestFixture) Start() error {
	execErr := f.compose.Invoke()
	return execErr.Error
}

func (f *LocalTestFixture) Stop() error {
	execErr := f.compose.Down()
	return execErr.Error
}
