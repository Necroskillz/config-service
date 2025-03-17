package service

import (
	"context"

	"github.com/necroskillz/config-service/repository"
)

type ValidationService struct {
	keyRepository *repository.KeyRepository
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidationService(keyRepository *repository.KeyRepository) *ValidationService {
	return &ValidationService{keyRepository: keyRepository}
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
