package main

import (
	"context"
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
	Filename string
	Name     string
	Version  string
	SQL      string
}

// A helper to apply a specific migration.
// We do the entire migration in a transaction, so we don't apply it without also recording it in the applied table.
func ApplyMigration(ctx context.Context, conn *sql.Conn, migration Migration) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, migration.SQL); err != nil {
		return fmt.Errorf("exec migration %s: %w", migration.Filename, err)
	}

	if _, err := tx.ExecContext(ctx, insertMigration, migration.Version); err != nil {
		return fmt.Errorf("record migration %s: %w", migration.Filename, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// A helper to load the migration files from disk.
func LoadMigrations() ([]Migration, error) {
	var migrations []Migration

	// We take no risks here...if the directory doesn't exist, we bail and just return empty.
	if _, err := os.Stat("db/migrations"); err != nil {
		if os.IsNotExist(err) {
			return []Migration{}, nil
		}
		return nil, fmt.Errorf("stat migrations dir: %w", err)
	}

	err := filepath.WalkDir("db/migrations",
		func(path string, d fs.DirEntry, err error) error {
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

	// We sort by version number (which should be date). Doing this hopefully applies the migrations in an ideal order (though it is possible, due to the relative nature of time, for these things to get out sync)
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// A helped to fetch the applied migrations.
// We fetch what has been applied in the DB, and store it as a map.
func GetAppliedMigrations(ctx context.Context, conn *sql.Conn) (map[string]time.Time, error) {
	rows, err := conn.QueryContext(ctx, getMigrations)
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

// A helper to create an advisory lock.
func AcquireAdvisoryLock(ctx context.Context, conn *sql.Conn, lockID int64) error {
	var acquired bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock($1)", lockID).Scan(&acquired); err != nil {
		return fmt.Errorf("try advisory lock %d: %w", lockID, err)
	}
	if !acquired {
		return fmt.Errorf("advisory lock %d is already held", lockID)
	}
	return nil
}

// A helper to release an advisory lock.
func ReleaseAdvisoryLock(ctx context.Context, conn *sql.Conn, lockID int64) error {
	var released bool
	if err := conn.QueryRowContext(ctx, "SELECT pg_advisory_unlock($1)", lockID).Scan(&released); err != nil {
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

func Run(ctx context.Context, conn *sql.Conn) error {
	// Lock
	// We lock the DB here, JUST IN CASE we have two migrations running at the same time
	lockID := int64(42069)
	if err := AcquireAdvisoryLock(ctx, conn, lockID); err != nil {
		return fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		if err := ReleaseAdvisoryLock(ctx, conn, lockID); err != nil {
			log.Printf("failed to release advisory lock: %v", err)
		}
	}()

	// Load migration files
	// Next, we load the migration files from `db/migrations`
	migrations, err := LoadMigrations()
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	// Get applied migrations
	// We also make a trip to the DB, to see what we've already applied.
	// We want migrations to be idempotent when possible, but it's not a guarantee.
	appliedMigrations, err := GetAppliedMigrations(ctx, conn)
	if err != nil {
		return fmt.Errorf("get applied migrations: %w", err)
	}

	// Apply pending migrations
	// If a migration hasn't been applied yet, we go for it.
	for _, migration := range migrations {
		if _, exists := appliedMigrations[migration.Version]; exists {
			log.Printf("Migration %s already applied, skipping", migration.Version)
			continue
		}

		log.Printf("Applying migration %s (%s)...", migration.Version, migration.Name)
		if err := ApplyMigration(ctx, conn, migration); err != nil {
			return fmt.Errorf("apply migration %s: %w", migration.Filename, err)
		}
		log.Printf("Successfully applied migration %s (%s)...", migration.Version, migration.Name)
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
	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatalf("Failed to close DB: %v", err)
		}
	}()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	conn, err := db.Conn(ctx)
	if err != nil {
		log.Fatalf("Failed to acquire DB connection: %v", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatalf("Failed to close connection to DB: %v", err)
		}
	}()

	if err := Run(ctx, conn); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
