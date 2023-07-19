package tests

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LocalTestFixture struct {
	dockerComposePath string
	compose           testcontainers.DockerCompose
}

func NewLocalTestFixture(dockerComposePath string, dbURL string) (LocalTestFixture, error) {
	compose := testcontainers.NewLocalDockerCompose(
		[]string{dockerComposePath},
		uuid.New().String(),
	)

	for serviceName := range compose.Services {
		if strings.Contains(serviceName, "postgres") {
			port := 5432 // TODO: actually pull the port from the service definition

			compose.WithExposedService(
				serviceName,
				port,
				wait.ForSQL(nat.Port(fmt.Sprintf("%d", port)), "postgres", func(nat.Port) string {
					return dbURL
				}),
			)
		}
	}

	return LocalTestFixture{
		dockerComposePath: dockerComposePath,
		compose:           compose.WithCommand([]string{"up", "--build", "-d"}),
	}, nil
}

func (f *LocalTestFixture) Start() error {
	log.Println(os.Getenv("SKIP_INFRASTRUCTURE"))
	if skip := os.Getenv("SKIP_INFRASTRUCTURE"); skip == "true" {
		return nil
	}

	execErr := f.compose.Invoke()
	return execErr.Error
}

func (f *LocalTestFixture) Stop() error {
	if skip := os.Getenv("SKIP_INFRASTRUCTURE"); skip == "true" {
		return nil
	}

	execErr := f.compose.Down()
	return execErr.Error
}
