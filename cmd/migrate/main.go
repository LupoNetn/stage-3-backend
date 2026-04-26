package main

import (
	"context"
	"flag"
	"log"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/luponetn/hng-stage-1/internals/config"
	_ "github.com/luponetn/hng-stage-1/internals/db/schema"
	"github.com/pressly/goose/v3"
)

func main() {
	_ = godotenv.Load()
	cfg := config.LoadConfig()

	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		args = append(args, "up")
	}

	command := args[0]

	db, err := goose.OpenDBWithDriver("pgx", cfg.DBURL)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	// This assumes you want goose to read SQL migrations from the schema folder.
	// Since the go migration files are self-registering via the blank import above,
	// supplying the directory here ensures goose finds both SQL and GO migrations.
	if err := goose.RunContext(context.Background(), command, db, "internals/db/schema", args[1:]...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
