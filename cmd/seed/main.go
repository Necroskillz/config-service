package main

import (
	"context"
	"log"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/necroskillz/config-service/db"
)

func main() {
	cmd := exec.Command("dbmate", "down")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to run dbmate down: %v", err)
	}

	cmd = exec.Command("dbmate", "up")
	err = cmd.Run()
	if err != nil {
		log.Fatalf("failed to run dbmate up: %v", err)
	}

	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	defer conn.Close(ctx)

	queries := db.New(conn)

	MustCreate(queries.CreateUser(ctx, db.CreateUserParams{
		Name:                "admin",
		Password:            "admin",
		GlobalAdministrator: true,
	}))

	serviceTypeId := MustCreate(queries.CreateServiceType(ctx, "TestServiceType"))
	MustCreate(queries.CreateValueType(ctx, "String"))
	MustCreate(queries.CreateValueType(ctx, "Boolean"))
	MustCreate(queries.CreateValueType(ctx, "Number"))
	MustCreate(queries.CreateValueType(ctx, "JSON"))
	envPropertyId := MustCreate(queries.CreateVariationProperty(ctx, db.CreateVariationPropertyParams{
		Name:        "env",
		DisplayName: "Environment",
	}))

	MustCreate(queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		VariationPropertyID: envPropertyId,
		Value:               "prod",
	}))

	qaId := MustCreate(queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		VariationPropertyID: envPropertyId,
		Value:               "qa",
	}))

	MustCreate(queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		VariationPropertyID: envPropertyId,
		Value:               "qa1",
		ParentID:            &qaId,
	}))

	MustCreate(queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		VariationPropertyID: envPropertyId,
		Value:               "qa2",
		ParentID:            &qaId,
	}))

	MustCreate(queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		VariationPropertyID: envPropertyId,
		Value:               "dev",
	}))

	MustExec(queries.AddPropertyToServiceType(ctx, db.AddPropertyToServiceTypeParams{
		ServiceTypeID:       serviceTypeId,
		VariationPropertyID: envPropertyId,
		Priority:            1,
	}))
}

var counter int

func MustCreate(id uint, err error) uint {
	if err != nil {
		log.Fatalf("failed to run query %d: %v", counter, err)
	}

	counter++

	return id
}

func MustExec(err error) {
	if err != nil {
		log.Fatalf("failed to execute query %d: %v", counter, err)
	}

	counter++
}
