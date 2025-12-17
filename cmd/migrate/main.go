package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// 1. Read DATABASE_URL from env
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// 2. Connect to Postgres/Neon using pgx stdlib
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("error opening database: %v", err)
	}
	defer db.Close()

	// Optional: small ping to be sure
	if err := db.Ping(); err != nil {
		log.Fatalf("error pinging database: %v", err)
	}

	// 3. Read the SQL file
	sqlBytes, err := os.ReadFile("migrations/migrations.sql")
	if err != nil {
		log.Fatalf("error reading migrations file: %v", err)
	}
	sqlText := string(sqlBytes)

	// 4. Apply migrations with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Applying migrations...")
	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		log.Fatalf("error applying migrations: %v", err)
	}

	log.Println("Migrations applied successfully ✅")
}
