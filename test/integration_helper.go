package test

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SpinUpPostgresContainer(pathToMigration string) (testcontainers.Container, func(ctx context.Context), error) {
	ctx := context.Background()

	net, err := network.New(ctx)
	if err != nil {
		log.Printf("Could not create a new docker network: %s", err)
		return nil, nil, err
	}

	postgresReqEnv := map[string]string{
		"POSTGRES_DB":       "postgres",
		"POSTGRES_USER":     "postgres",
		"POSTGRES_PASSWORD": "postgres",
	}
	postgresReq := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Networks:     []string{net.Name},
		Env:          postgresReqEnv,
		WaitingFor: wait.ForSQL("5432/tcp", "pgx", func(host string, port nat.Port) string {
			return fmt.Sprintf("postgres://%v:%v@localhost:%s/%v", postgresReqEnv["POSTGRES_USER"], postgresReqEnv["POSTGRES_PASSWORD"], port.Port(), postgresReqEnv["POSTGRES_DB"])
		}),
	}
	postgres, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: postgresReq,
		Started:          true,
	})
	if err != nil {
		log.Printf("Could not start postgres: %s", err)
		return nil, nil, err
	}

	postgresIp, err := postgres.ContainerIP(ctx)
	if err != nil {
		log.Fatal(err)
	}

	migrationsReqEnv := map[string]string{
		"GOOSE_DRIVER": "postgres",
		"GOOSE_DBSTRING": fmt.Sprintf(
			"host=%v port=5432 user=%v password=%v dbname=%v",
			postgresIp,
			postgresReqEnv["POSTGRES_USER"],
			postgresReqEnv["POSTGRES_PASSWORD"],
			postgresReqEnv["POSTGRES_DB"],
		),
	}

	migrationsReq := testcontainers.ContainerRequest{
		Image:    "ghcr.io/kukymbr/goose-docker:3.21.1",
		Env:      migrationsReqEnv,
		Networks: []string{net.Name},
		Mounts: testcontainers.ContainerMounts{
			{
				Source: testcontainers.GenericBindMountSource{
					HostPath: pathToMigration,
				},
				Target: testcontainers.ContainerMountTarget("/migrations"),
			},
		},
		WaitingFor: wait.ForExit(),
	}
	migrations, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: migrationsReq,
		Started:          true,
	})
	if err != nil {
		log.Printf("Could not start migrations: %s", err)
		postgres.Terminate(ctx)
		return nil, nil, err
	}

	closeHandler := func(ctx context.Context) {
		migrations.Terminate(ctx)
		postgres.Terminate(ctx)
		net.Remove(ctx)
	}

	return postgres, closeHandler, nil
}

func RunInTransaction(t *testing.T, db *pgxpool.Pool, testFunc func(*testing.T, pgx.Tx)) {
	ctx := context.Background()

	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer tx.Rollback(ctx)
	testFunc(t, tx)
}

func SpinUpPostgresContainerAndGetPgxpool(pathToMigration string) (*pgxpool.Pool, testcontainers.Container, func(ctx context.Context)) {
	ctx := context.Background()

	postgres, close, err := SpinUpPostgresContainer(pathToMigration)
	if err != nil {
		close(ctx)
		log.Fatal(err)
	}

	dbUser := "postgres"
	dbPass := dbUser
	dbName := dbUser
	dbHost := "localhost"
	dbPort, err := postgres.MappedPort(ctx, "5432/tcp")
	if err != nil {
		log.Fatal(err)
	}

	dbUrl := fmt.Sprintf("postgresql://%v:%v@%v:%v/%v", dbUser, dbPass, dbHost, dbPort.Port(), dbName)
	connPool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		log.Fatal(err)
	}

	return connPool, postgres, close
}

func GetValueFromCnOrLogFatalWithTimeout[T any](cn <-chan T, timeout time.Duration, errorMsg string) T {
	var val T

	select {
	case val = <-cn:
		return val
	case <-time.After(timeout):
		log.Fatal(errorMsg)
	}

	return val
}
