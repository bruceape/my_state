package main

import (
	"database/sql"
	_ "embed"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

//go:embed init.sql
var addMigrationsTable string

func main() {
	_ = godotenv.Load()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer func() {
		err = db.Close()
	}()
	if err != nil {
		log.Fatalf("Failed to close DB: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Adding migrations table")
	_, err = db.Exec(addMigrationsTable)
	if err != nil {
		log.Fatalf("Failed to create schema_migrations table: %v", err)
	}
}
