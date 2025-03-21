package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type ValueService struct {
	unitOfWorkCreator        repository.UnitOfWorkCreator
	variationValueRepository *repository.VariationValueRepository
	changesetService         *ChangesetService
}

func NewValueService(unitOfWorkCreator repository.UnitOfWorkCreator, variationValueRepository *repository.VariationValueRepository, changesetService *ChangesetService) *ValueService {
	return &ValueService{unitOfWorkCreator: unitOfWorkCreator, variationValueRepository: variationValueRepository, changesetService: changesetService}
}

func (s *ValueService) GetKeyValues(ctx context.Context, keyID uint) ([]model.VariationValue, error) {
	return s.variationValueRepository.GetActive(ctx, keyID)
}

type CreateValueParams struct {
	FeatureVersionID uint
	KeyID            uint
	ChangesetID      uint
	Value            string
	Variation        []uint
}

func (s *ValueService) CreateValue(ctx context.Context, params CreateValueParams) error {
	variationValue := &model.VariationValue{
		KeyID: params.KeyID,
		Data:  &params.Value,
	}

	return s.unitOfWorkCreator.Run(ctx, func(ctx context.Context) error {
		err := s.variationValueRepository.Create(ctx, variationValue)
		if err != nil {
			return err
		}

		for _, variationPropertyValueID := range params.Variation {
			err = s.variationValueRepository.AddVariationPropertyValue(ctx, variationValue.ID, variationPropertyValueID)
			if err != nil {
				return err
			}
		}

		err = s.changesetService.AddVariationValueChange(ctx, params.ChangesetID, params.FeatureVersionID, params.KeyID, variationValue.ID, model.ChangesetChangeTypeCreate, &params.Value, nil)
		if err != nil {
			return err
		}

		return nil
	})
}
