package service

import (
	"context"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type KeyService struct {
	unitOfWorkCreator        repository.UnitOfWorkCreator
	keyRepository            *repository.KeyRepository
	valueTypeRepository      *repository.ValueTypeRepository
	changesetService         *ChangesetService
	variationValueRepository *repository.VariationValueRepository
}

func NewKeyService(unitOfWorkCreator repository.UnitOfWorkCreator, keyRepository *repository.KeyRepository, valueTypeRepository *repository.ValueTypeRepository, changesetService *ChangesetService, variationValueRepository *repository.VariationValueRepository) *KeyService {
	return &KeyService{unitOfWorkCreator: unitOfWorkCreator, keyRepository: keyRepository, valueTypeRepository: valueTypeRepository, changesetService: changesetService, variationValueRepository: variationValueRepository}
}

func (s *KeyService) GetFeatureKeys(ctx context.Context, featureVersionID uint) ([]model.Key, error) {
	return s.keyRepository.GetActive(ctx, featureVersionID)
}

func (s *KeyService) GetValueTypes(ctx context.Context) ([]model.ValueType, error) {
	return s.valueTypeRepository.GetAll(ctx)
}

type CreateKeyParams struct {
	ChangesetID      uint
	ServiceVersionID uint
	FeatureVersionID uint
	Name             string
	Description      string
	DefaultValue     string
	ValueTypeID      uint
}

func (s *KeyService) CreateKey(ctx context.Context, params CreateKeyParams) error {
	return s.unitOfWorkCreator.Run(ctx, func(ctx context.Context) error {
		key := model.Key{
			Name:             params.Name,
			Description:      &params.Description,
			ValueTypeID:      params.ValueTypeID,
			FeatureVersionID: params.FeatureVersionID,
		}

		err := s.keyRepository.Create(ctx, &key)
		if err != nil {
			return err
		}

		err = s.changesetService.AddKeyChange(ctx, params.ChangesetID, params.FeatureVersionID, key.ID, model.ChangesetChangeTypeCreate)
		if err != nil {
			return err
		}

		value := model.VariationValue{
			KeyID: key.ID,
			Data:  &params.DefaultValue,
		}

		err = s.variationValueRepository.Create(ctx, &value)
		if err != nil {
			return err
		}

		err = s.changesetService.AddVariationValueChange(ctx, params.ChangesetID, params.FeatureVersionID, key.ID, value.ID, model.ChangesetChangeTypeCreate, &params.DefaultValue, nil)
		if err != nil {
			return err
		}

		return nil
	})
}
