package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type ValueService struct {
	unitOfWorkCreator         repository.UnitOfWorkCreator
	variationValueRepository  *repository.VariationValueRepository
	changesetService          *ChangesetService
	changesetChangeRepository *repository.ChangesetChangeRepository
}

func NewValueService(unitOfWorkCreator repository.UnitOfWorkCreator, variationValueRepository *repository.VariationValueRepository, changesetService *ChangesetService, changesetChangeRepository *repository.ChangesetChangeRepository) *ValueService {
	return &ValueService{unitOfWorkCreator: unitOfWorkCreator, variationValueRepository: variationValueRepository, changesetService: changesetService, changesetChangeRepository: changesetChangeRepository}
}

func (s *ValueService) GetKeyValues(ctx context.Context, keyID uint, changesetID uint) ([]model.VariationValue, error) {
	return s.variationValueRepository.GetActive(ctx, keyID, changesetID)
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

		err = s.changesetService.AddCreateVariationValueChange(ctx, params.ChangesetID, params.FeatureVersionID, params.KeyID, variationValue.ID)
		if err != nil {
			return err
		}

		return nil
	})
}

type DeleteValueParams struct {
	ChangesetID      uint
	FeatureVersionID uint
	KeyID            uint
	ValueID          uint
}

func (s *ValueService) DeleteValue(ctx context.Context, params DeleteValueParams) error {
	return s.unitOfWorkCreator.Run(ctx, func(ctx context.Context) error {
		variationValue, err := s.variationValueRepository.GetById(ctx, params.ValueID)
		if err != nil {
			return err
		}

		if variationValue.ValidFrom != nil {
			err = s.changesetService.AddDeleteVariationValueChange(ctx, params.ChangesetID, params.FeatureVersionID, params.KeyID, variationValue.ID)
			if err != nil {
				return err
			}
		} else {
			err = s.changesetChangeRepository.DeleteByVariationValueID(ctx, variationValue.ID)
			if err != nil {
				return err
			}

			err = s.variationValueRepository.Delete(ctx, variationValue.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
