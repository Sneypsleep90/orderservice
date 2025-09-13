package database

import (
	"database/sql"
	"fmt"
	"log"
	"myapp/internal/config"
	"myapp/internal/migrate"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func Connect(cfg config.Config) (*sql.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Successfully connected to database")
	return db, nil
}

func RunMigrations(cfg config.Config, migrationsPath string) error {
	if os.Getenv("SKIP_MIGRATIONS") == "true" {
		log.Println("Skipping migrations (SKIP_MIGRATIONS=true)")
		return nil
	}

	migrationDB, err := Connect(cfg)
	if err != nil {
		return fmt.Errorf("failed to connect to database for migrations: %w", err)
	}
	defer migrationDB.Close()

	migrator, err := migrate.NewMigrator(migrationDB, migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	return migrator.Up()
}
