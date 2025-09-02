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
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("exec migration %s: %w", migration.Filename, err)
	}

	// Insert into schema
	_, err = tx.Exec(insertMigration, migration.Version)
	if err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Filename, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("commit tx: %w", err)
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
			parts := strings.SplitN(filename, "_", 2)
			if len(parts) != 2 {
				return fmt.Errorf("invalid migration filename (expected <version>_<name>.sql): %s", filename)
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read migration file %s: %w",
					path, err)
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
		return nil, fmt.Errorf("fetch applied migrations: %w", err)
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
		return fmt.Errorf("try advisory lock %d: %w", lockID, err)
	}

	if !acquired {
		return fmt.Errorf("advisory lock %d is already held", lockID)
	}

	return nil
}

func ReleaseAdvisoryLock(db *sql.DB, lockID int64) error {
	var released bool
	if err := db.QueryRow("SELECT pg_advisory_unlock($1)", lockID).Scan(&released); err != nil {
		return fmt.Errorf("release advisory lock %d: %w", lockID, err)
	}
	if !released {
		return fmt.Errorf("advisory lock %d was not held", lockID)
	}
	return nil
}

//go:embed insert_migration.sql
var insertMigration string

//go:embed get_migrations.sql
var getMigrations string

func Run(db *sql.DB) error {
	// Lock
	lockID := int64(42069)
	if err := AcquireAdvisoryLock(db, lockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer ReleaseAdvisoryLock(db, lockID)

	// Load migration files
	migrations, err := LoadMigrations()
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := GetAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
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
			return fmt.Errorf("apply migration %s: %w",
				migration.Filename, err)
		}
		log.Printf("Successfully applied migration %s (%s)...",
			migration.Version, migration.Name)
	}

	return nil
}

func main() {
	_ = godotenv.Load()

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	err = Run(db)
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
