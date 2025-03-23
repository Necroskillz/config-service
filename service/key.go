package service

import (
	"context"

	"github.com/necroskillz/config-service/db"
)

type KeyService struct {
	unitOfWorkRunner        db.UnitOfWorkRunner
	variationContextService *VariationContextService
	queries                 *db.Queries
}

func NewKeyService(unitOfWorkRunner db.UnitOfWorkRunner, variationContextService *VariationContextService, queries *db.Queries) *KeyService {
	return &KeyService{unitOfWorkRunner: unitOfWorkRunner, variationContextService: variationContextService, queries: queries}
}

func (s *KeyService) GetKey(ctx context.Context, keyID uint) (db.Key, error) {
	return s.queries.GetKey(ctx, keyID)
}

func (s *KeyService) GetFeatureKeys(ctx context.Context, featureVersionID uint, changesetID uint) ([]db.Key, error) {
	keys, err := s.queries.GetActiveKeysForFeatureVersion(ctx, db.GetActiveKeysForFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      changesetID,
	})

	return keys, err
}

func (s *KeyService) GetValueTypes(ctx context.Context) ([]db.ValueType, error) {
	return s.queries.GetValueTypes(ctx)
}

type CreateKeyParams struct {
	ChangesetID      uint
	FeatureVersionID uint
	Name             string
	Description      string
	DefaultValue     string
	ValueTypeID      uint
	ServiceVersionID uint
}

func (s *KeyService) CreateKey(ctx context.Context, params CreateKeyParams) (uint, error) {
	var keyID uint

	err := s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		var err error
		keyID, err = tx.CreateKey(ctx, db.CreateKeyParams{
			Name:             params.Name,
			Description:      &params.Description,
			ValueTypeID:      params.ValueTypeID,
			FeatureVersionID: params.FeatureVersionID,
		})
		if err != nil {
			return err
		}

		err = s.queries.AddCreateKeyChange(ctx, db.AddCreateKeyChangeParams{
			ChangesetID:      params.ChangesetID,
			KeyID:            keyID,
			FeatureVersionID: params.FeatureVersionID,
			ServiceVersionID: params.ServiceVersionID,
		})
		if err != nil {
			return err
		}

		defaultVariationContextID, err := s.variationContextService.GetVariationContextID(ctx, []uint{})
		if err != nil {
			return err
		}

		variationValueID, err := s.queries.CreateVariationValue(ctx, db.CreateVariationValueParams{
			KeyID:              keyID,
			Data:               &params.DefaultValue,
			VariationContextID: defaultVariationContextID,
		})
		if err != nil {
			return err
		}

		err = s.queries.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         params.ChangesetID,
			NewVariationValueID: variationValueID,
			FeatureVersionID:    params.FeatureVersionID,
			KeyID:               keyID,
			ServiceVersionID:    params.ServiceVersionID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return keyID, err
}
