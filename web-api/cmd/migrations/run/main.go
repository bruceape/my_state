package main

import (
	"database/sql"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

type Migration struct {
	Filename   string
	Name       string
	Version    string
	SQL        string
	Applied_at *time.Time
}

func ApplyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("Failed to begin transction")
	}
	defer tx.Rollback()

	_, err = tx.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("Failed to execute migration")
	}

	// Insert into schema
	_, err = tx.Exec(insertMigration, migration.Version)
	if err != nil {
		return fmt.Errorf("Failed to record migration")
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("Failed to commit transaction")
	}

	return nil
}

func LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	err := filepath.WalkDir("db/migrations",
		func(path string, d fs.DirEntry, err error) error {
			// Guards
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(path, ".sql") {
				return nil
			}

			filename := d.Name()
			parts := strings.Split(filename, "_")
			if len(parts) != 2 {
				return fmt.Errorf("Invalid name: %s", filename)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w", path, err)
			}

			version := parts[0]
			name := strings.TrimSuffix(parts[1], ".sql")

			migrations = append(migrations, Migration{
				Version:  version,
				Name:     name,
				Filename: filename,
				SQL:      string(content),
			})

			return nil
		})

	if err != nil {
		return nil, err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func GetAppliedMigrations(db *sql.DB) (map[string]time.Time, error) {
	rows, err := db.Query(getMigrations)
	if err != nil {
		return nil, fmt.Errorf("Failed to fetch migrations from DB, %w", err)
	}
	defer rows.Close()

	appliedMigrations := make(map[string]time.Time)
	for rows.Next() {
		var version string
		var appliedAt time.Time
		if err := rows.Scan(&version, &appliedAt); err != nil {
			return nil, fmt.Errorf("scanning migration row: %w", err)
		}
		appliedMigrations[version] = appliedAt
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating migration rows: %w", err)
	}

	return appliedMigrations, nil
}
func AcquireAdvisoryLock(db *sql.DB, lockID int64) error {
	var acquired bool
	err := db.QueryRow("SELECT pg_try_advisory_lock($1)", lockID).
		Scan(&acquired)
	if err != nil {
		return fmt.Errorf("Failed to acquire lock: %w", err)
	}

	if !acquired {
		return fmt.Errorf("Failed to acquire lock, it may be in use.")
	}

	return nil
}

func ReleaseAdvisoryLock(db *sql.DB, lockID int64) error {
	_, err := db.Exec("SELECT pg_advisory_unlock($1)", lockID)
	if err != nil {
		return fmt.Errorf("Failed to release lock: %w", err)
	}

	return nil
}

//go:embed insert_migration.sql
var insertMigration string

//go:embed get_migrations.sql
var getMigrations string

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	databaseUrl := os.Getenv("DATABASE_URL")

	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatal("Failed to connect to DB")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Lock
	lockID := int64(42069)
	if err := AcquireAdvisoryLock(db, lockID); err != nil {
		log.Fatalf("failed to acquire migration lock: %v", err)
	}
	defer ReleaseAdvisoryLock(db, lockID)

	// Load migration files
	migrations, err := LoadMigrations()
	if err != nil {
		log.Fatal("Failed to load migration files.")
	}

	// Get applied migrations
	appliedMigrations, err := GetAppliedMigrations(db)
	if err != nil {
		log.Fatalf("Failed to get applied migrations: %v", err)
	}

	// Apply pending migrations
	for _, migration := range migrations {
		_, exists := appliedMigrations[migration.Version]
		if exists {
			log.Printf("Migration %s already applied, skipping",
				migration.Version)
			continue
		}

		log.Printf("Applying migration %s (%s)...",
			migration.Version, migration.Name)
		if err := ApplyMigration(db, migration); err != nil {
			log.Fatalf("Failed to apply migration: %s, %v",
				migration.Filename, err)
		}
		log.Printf("Successfully applied migration %s (%s)...",
			migration.Version, migration.Name)
	}
}
