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

	log.Println("Adding migrations table")
	_, err = db.Exec(addMigrationsTable)
	if err != nil {
		log.Fatalf("Failed to add migration table. %v", err)
	}
}
