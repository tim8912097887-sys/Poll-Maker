package main

import (
	"context"
	"embed"
	"log"
	"os"

	// Import the pgx driver explicitly for goose
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func main() {
	dbString := os.Getenv("VOTE_SERVICE_DB_URL")
	if dbString == "" {
		log.Fatal("VOTE_SERVICE_DB_URL environment variable is required")
	}

	// Set Goose to use the embedded filesystem
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Open the DB connection using Goose's driver wrapper
	// "pgx" matches the github.com/jackc/pgx/v5/stdlib driver
	db, err := goose.OpenDBWithDriver("pgx", dbString)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	log.Println("Running database migrations...")
	
	// Now 'db' is a standard *sql.DB which Goose accepts perfectly
	if err := goose.UpContext(context.Background(), db, "migrations"); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migrations completed successfully!")
}