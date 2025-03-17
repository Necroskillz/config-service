package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type ValueService struct {
	unitOfWorkCreator        repository.UnitOfWorkCreator
	variationValueRepository *repository.VariationValueRepository
}

func NewValueService(unitOfWorkCreator repository.UnitOfWorkCreator, variationValueRepository *repository.VariationValueRepository) *ValueService {
	return &ValueService{unitOfWorkCreator: unitOfWorkCreator, variationValueRepository: variationValueRepository}
}

func (s *ValueService) GetKeyValues(ctx context.Context, keyID uint) ([]model.VariationValue, error) {
	return s.variationValueRepository.GetActive(ctx, keyID)
}
