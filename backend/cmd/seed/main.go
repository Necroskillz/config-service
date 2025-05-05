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

func Ptr[T any](v T) *T {
	return &v
}

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
		Name: "String",
		Kind: "string",
	}))
	booleanTypeId := MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name: "Boolean",
		Kind: "boolean",
	}))
	integerTypeId := MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name: "Integer",
		Kind: "integer",
	}))
	decimalTypeId := MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name: "Decimal",
		Kind: "decimal",
	}))
	jsonTypeId := MustCreate(queries.CreateValueType(ctx, db.CreateValueTypeParams{
		Name: "JSON",
		Kind: "json",
	}))

	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &booleanTypeId,
		ValidatorType: "required",
	}))
	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &booleanTypeId,
		ValidatorType: "regex",
		Parameter:     Ptr("^TRUE|FALSE$"),
		ErrorText:     Ptr("Value must be TRUE or FALSE"),
	}))

	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &integerTypeId,
		ValidatorType: "required",
	}))
	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &integerTypeId,
		ValidatorType: "valid_integer",
		ErrorText:     Ptr("Value must be an integer"),
	}))

	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &decimalTypeId,
		ValidatorType: "required",
	}))
	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &decimalTypeId,
		ValidatorType: "valid_float",
		ErrorText:     Ptr("Value must be a number with optional decimal part"),
	}))

	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &jsonTypeId,
		ValidatorType: "required",
	}))
	MustCreate(queries.CreateValueValidatorForValueType(ctx, db.CreateValueValidatorForValueTypeParams{
		ValueTypeID:   &jsonTypeId,
		ValidatorType: "valid_json",
		ErrorText:     Ptr("Value must be valid JSON: {0}"),
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
		Kind:       "service",
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
