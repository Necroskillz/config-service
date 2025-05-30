package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/testdata"
)

func main() {
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	seed := flag.Int64("seed", time.Now().UnixNano(), "seed for the random number generator")
	flag.Parse()

	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}

	defer cache.Close()

	connectionString := os.Getenv("DATABASE_URL")

	log.Println("Dropping database...")
	if err := db.Drop(connectionString); err != nil {
		log.Fatalf("failed to drop database: %v", err)
	}

	log.Println("Creating database...")
	if err := db.Migrate(connectionString); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	dbpool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer dbpool.Close()

	m := testdata.NewManager(dbpool, cache)
	if err := m.Run(ctx, testdata.WithSeed(*seed)); err != nil {
		log.Fatalf("failed to run data generator: %v", err)
	}
}
