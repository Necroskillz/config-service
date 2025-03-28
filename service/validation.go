package service

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/db"
)

type ValidationService struct {
	queries                 *db.Queries
	variationContextService *VariationContextService
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationService(queries *db.Queries, variationContextService *VariationContextService) *ValidationService {
	return &ValidationService{
		queries:                 queries,
		variationContextService: variationContextService,
	}
}

func (s *ValidationService) ValidateServiceNameUniqueness(ctx context.Context, name string) error {
	_, err := s.queries.GetServiceIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return &ValidationError{
		Field:   "Name",
		Message: "Service with this name already exists",
	}
}

func (s *ValidationService) ValidateFeatureNameUniqueness(ctx context.Context, name string) error {
	_, err := s.queries.GetFeatureIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return &ValidationError{
		Field:   "Name",
		Message: "Feature with this name already exists",
	}
}

func (s *ValidationService) ValidateKeyNameUniqueness(ctx context.Context, featureVersionID uint, keyName string) error {
	_, err := s.queries.GetKeyIDByName(ctx, db.GetKeyIDByNameParams{
		Name:             keyName,
		FeatureVersionID: featureVersionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	return &ValidationError{
		Field:   "Name",
		Message: "Key with this name already exists in this feature",
	}
}

func (s *ValidationService) ValidateVariationUniqueness(ctx context.Context, keyID uint, variationIDs []uint, changesetID uint) error {
	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, variationIDs)
	if err != nil {
		return err
	}

	id, err := s.queries.GetActiveVariationValueIDByVariationContextID(ctx, db.GetActiveVariationValueIDByVariationContextIDParams{
		VariationContextID: variationContextID,
		ChangesetID:        changesetID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}

		return err
	}

	log.Println(id)

	return &ValidationError{
		Field:   "Value",
		Message: "Variation with these property values already exists",
	}
}
