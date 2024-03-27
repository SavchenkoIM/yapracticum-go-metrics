package testhelpers

import (
	"context"
	"fmt"
	"github.com/docker/docker/pkg/ioutils"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestPostgres struct {
	instance testcontainers.Container
}

func NewTestPostgres() (*TestPostgres, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	testcontainers.Logger = log.New(&ioutils.NopWriter{}, "", 0)
	req := testcontainers.ContainerRequest{
		Image:        "postgres:14",
		ExposedPorts: []string{"5432/tcp"},
		AutoRemove:   true,
		Env: map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       "postgres",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp"),
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		return nil, err
	}
	return &TestPostgres{
		instance: postgres,
	}, nil
}

func (db *TestPostgres) Port() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := db.instance.MappedPort(ctx, "5432")
	if err != nil {
		return 0, err
	}
	return p.Int(), nil
}

func (db *TestPostgres) ConnectionString() (string, error) {
	port, err := db.Port()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("postgres://postgres:postgres@127.0.0.1:%d/postgres", port), nil
}

func (db *TestPostgres) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	return db.instance.Terminate(ctx)
}

func (db *TestPostgres) Host() string {
	return "localhost"
}
