package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/necroskillz/config-service/db"
	"golang.org/x/crypto/bcrypt"
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

	password, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("failed to generate password: %v", err)
	}

	MustCreate(queries.CreateUser(ctx, db.CreateUserParams{
		Name:                "admin",
		Password:            string(password),
		GlobalAdministrator: true,
	}))

	serviceTypeId := MustCreate(queries.CreateServiceType(ctx, "TestServiceType"))
	MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name:   "String",
		Editor: "text",
	}))
	MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name:   "Boolean",
		Editor: "boolean",
	}))
	MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name:   "Number",
		Editor: "number",
	}))
	MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name:   "JSON",
		Editor: "json",
	}))
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

	serviceId := MustCreate(queries.CreateService(ctx, db.CreateServiceParams{
		Name:          "TestService",
		Description:   "Test Service Description",
		ServiceTypeID: serviceTypeId,
	}))

	validFrom := time.Now()

	serviceVersionId := MustCreate(queries.CreateServiceVersion(ctx, db.CreateServiceVersionParams{
		ServiceID: serviceId,
		Version:   1,
	}))

	MustExec(queries.StartServiceVersionValidity(ctx, db.StartServiceVersionValidityParams{
		ServiceVersionID: serviceVersionId,
		ValidFrom:        &validFrom,
	}))

	MustExec(queries.PublishServiceVersion(ctx, serviceVersionId))

	MustCreate(queries.CreatePermission(ctx, db.CreatePermissionParams{
		UserID:     1,
		ServiceID:  serviceId,
		FeatureID:  nil,
		KeyID:      nil,
		Permission: "admin",
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
