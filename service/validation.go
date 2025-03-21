package service

import (
	"context"

	"github.com/necroskillz/config-service/repository"
)

type ValidationService struct {
	keyRepository            *repository.KeyRepository
	variationValueRepository *repository.VariationValueRepository
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationService(keyRepository *repository.KeyRepository, variationValueRepository *repository.VariationValueRepository) *ValidationService {
	return &ValidationService{keyRepository: keyRepository, variationValueRepository: variationValueRepository}
}

func (s *ValidationService) ValidateKeyNameUniqueness(ctx context.Context, featureVersionID uint, keyName string) error {
	// Check if key with this name already exists for this feature
	key, err := s.keyRepository.GetActiveKeyByName(ctx, featureVersionID, keyName)
	if err != nil {
		return err
	}

	if key != nil {
		return &ValidationError{
			Field:   "Name",
			Message: "Key with this name already exists in this feature",
		}
	}

	return nil
}

func (s *ValidationService) ValidateVariationUniqueness(ctx context.Context, keyID uint, variationIDs []uint) error {
	id, err := s.variationValueRepository.GetIDByVariation(ctx, keyID, variationIDs)
	if err != nil {
		return err
	}

	if id != 0 {
		return &ValidationError{
			Field:   "Value",
			Message: "Variation with these property values already exists",
		}
	}

	return nil
}
