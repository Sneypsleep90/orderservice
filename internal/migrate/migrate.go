package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

type Migrator struct {
	migrate *migrate.Migrate
}

func NewMigrator(db *sql.DB, migrationsPath string) (*Migrator, error) {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"postgres",
		driver,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}

	return &Migrator{migrate: m}, nil
}

func (m *Migrator) Up() error {
	if err := m.migrate.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to run")
			return nil
		}
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Migrations completed successfully")
	return nil
}

func (m *Migrator) Down() error {
	if err := m.migrate.Steps(-1); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("No migrations to rollback")
			return nil
		}
		return fmt.Errorf("failed to rollback migration: %w", err)
	}
	log.Println("Migration rollback completed successfully")
	return nil
}

func (m *Migrator) Version() (uint, bool, error) {
	return m.migrate.Version()
}

func (m *Migrator) Force(version int) error {
	if err := m.migrate.Force(version); err != nil {
		return fmt.Errorf("failed to force migration version: %w", err)
	}
	log.Printf("Migration version forced to %d", version)
	return nil
}

func (m *Migrator) Close() error {
	sourceErr, dbErr := m.migrate.Close()
	if sourceErr != nil {
		return fmt.Errorf("failed to close source: %w", sourceErr)
	}
	if dbErr != nil {
		return fmt.Errorf("failed to close database: %w", dbErr)
	}
	return nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	migrator, err := NewMigrator(db, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	args := os.Args[1:]
	if len(args) == 0 {
		return migrator.Up()
	}

	switch args[0] {
	case "up":
		return migrator.Up()
	case "down":
		return migrator.Down()
	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			return fmt.Errorf("failed to get version: %w", err)
		}
		log.Printf("Current migration version: %d (dirty: %t)", version, dirty)
		return nil
	case "force":
		if len(args) < 2 {
			return fmt.Errorf("force command requires a version number")
		}
		var version int
		if _, err := fmt.Sscanf(args[1], "%d", &version); err != nil {
			return fmt.Errorf("invalid version number: %w", err)
		}
		return migrator.Force(version)
	default:
		return fmt.Errorf("unknown command: %s. Available commands: up, down, version, force", args[0])
	}
}
