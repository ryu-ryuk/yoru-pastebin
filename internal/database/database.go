package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ryu-ryuk/yoru-pastebin/internal/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// DB represents our database connection pool.
type DB struct {
	Pool *pgxpool.Pool
}

// this creates a new database connection pool and runs migrations.
func NewDB(cfg *config.Config) (*DB, error) {
	connStr := cfg.Database.ConnectionString
	if connStr == "" {
		return nil, fmt.Errorf("database connection string is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	fmt.Println("Successfully connected to PostgreSQL!")

	// run migrations after successful connection
	if err := RunMigrations(connStr); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to run database migrations: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// applying database migrations.
func RunMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file://db/migrations", // path to migration files
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}
	defer m.Close()

	// apply migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		// log the error for debugging
		log.Printf("Migration error: %v", err)
		// check for specific migration errors like dirty database
		version, dirty, migErr := m.Version()
		if migErr == nil && dirty {
			log.Printf("Database is in a dirty state at version %d. You might need to force a fix (e.g., 'migrate force %d')", version, version)
		}
		return fmt.Errorf("could not run up migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		fmt.Println("No new database migrations to apply.")
	} else {
		fmt.Println("Database migrations applied successfully!")
	}
	return nil
}

// closes the database connection pool.
func (d *DB) Close() {
	if d.Pool != nil {
		d.Pool.Close()
		fmt.Println("Database connection pool closed.")
	}
}
