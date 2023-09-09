## What is this?

This is a project I use to play around with and explore different concept in programming using the 
Go programming language.
It does not have a purpose other than for me to play around when learning.

## Make commands
* **build**: builds the project
* **lint**: runs golangci-lint on the project
* **run**: runs the API on port 8080. Make sure to have infrastructure running before executing this command.
* **test**: runs the integration tests. All the supporting infrastructure will start automatically.
* **infra-up**: starts the supporting infrastructure for the integration tests. Useful for avoiding waiting for Docker
containers to start up before running the tests.
* **infra-down**: stops the supporting infrastructure for integration tests.