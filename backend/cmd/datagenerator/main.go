package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/hashicorp/go-metrics"
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
	iterations := flag.Int("iterations", 1000, "number of iterations to run")
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

	inm := metrics.NewInmemSink(time.Hour*24, time.Hour*24)
	metrics, err := metrics.New(metrics.DefaultConfig("testdata"), inm)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}

	m := testdata.NewManager(dbpool, cache, metrics)
	if err := m.Run(ctx, testdata.WithSeed(*seed), testdata.WithIterationCount(*iterations)); err != nil {
		log.Fatalf("failed to run data generator: %v", err)
	}

	metricData := inm.Data()
	for _, data := range metricData {
		fmt.Printf("| Name%s | Count  | Mean       | Min        | Max        | Stddev\n", strings.Repeat(" ", 26))
		for _, agg := range data.Samples {
			if agg.Name == "testdata.runtime.gc_pause_ns" {
				continue
			}

			count := agg.AggregateSample.Count
			mean := agg.AggregateSample.Mean()
			min := agg.AggregateSample.Min
			max := agg.AggregateSample.Max
			stddev := agg.AggregateSample.Stddev()

			fmt.Printf("| %s | %s%d | %s%.4f | %s%.4f | %s%.4f | %s%.4f\n",
				agg.Name+strings.Repeat(" ", 30-len(agg.Name)),
				strings.Repeat(" ", 6-len(strconv.Itoa(count))),
				count,
				strings.Repeat(" ", 10-len(strconv.FormatFloat(mean, 'f', 4, 64))),
				mean,
				strings.Repeat(" ", 10-len(strconv.FormatFloat(min, 'f', 4, 64))),
				min,
				strings.Repeat(" ", 10-len(strconv.FormatFloat(max, 'f', 4, 64))),
				max,
				strings.Repeat(" ", 10-len(strconv.FormatFloat(stddev, 'f', 4, 64))),
				stddev,
			)
		}
	}
}
