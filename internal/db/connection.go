package db

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/kense/home-task/migrations"
)

func ConnectAndMigrate(ctx context.Context, connString string) (*Queries, *pgx.Conn, error) {
	slog.Info("Connecting to database", "connString", connString)
	conn, err := pgx.Connect(ctx, connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	slog.Info("Running migrations")
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create iofs: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, connString)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return nil, nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	slog.Info("Migrations applied successfully or already up to date")

	return New(conn), conn, nil
}
