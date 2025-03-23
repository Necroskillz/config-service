package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/db"
)

type ValueService struct {
	unitOfWorkRunner        db.UnitOfWorkRunner
	variationContextService *VariationContextService
	queries                 *db.Queries
}

func NewValueService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	variationContextService *VariationContextService,
	queries *db.Queries,
) *ValueService {
	return &ValueService{
		unitOfWorkRunner:        unitOfWorkRunner,
		variationContextService: variationContextService,
		queries:                 queries,
	}
}

type VariationValue struct {
	ID        uint
	Data      *string
	Variation map[uint]string
}

func (s *ValueService) GetKeyValues(ctx context.Context, keyID uint, changesetID uint) ([]VariationValue, error) {
	values, err := s.queries.GetActiveVariationValuesForKey(ctx, db.GetActiveVariationValuesForKeyParams{
		KeyID:       keyID,
		ChangesetID: changesetID,
	})
	if err != nil {
		return nil, err
	}

	variationValues := make([]VariationValue, len(values))
	for i, value := range values {
		variation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
		if err != nil {
			return nil, err
		}

		variationValues[i] = VariationValue{
			ID:        value.ID,
			Data:      value.Data,
			Variation: variation,
		}
	}

	return variationValues, nil
}

type CreateValueParams struct {
	FeatureVersionID uint
	KeyID            uint
	ChangesetID      uint
	Value            string
	Variation        []uint
	ServiceVersionID uint
}

func (s *ValueService) CreateValue(ctx context.Context, params CreateValueParams) error {
	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, params.Variation)
	if err != nil {
		return err
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		variationValueID, err := tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
			KeyID:              params.KeyID,
			VariationContextID: variationContextID,
			Data:               &params.Value,
		})
		if err != nil {
			return err
		}

		err = tx.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         params.ChangesetID,
			NewVariationValueID: variationValueID,
			FeatureVersionID:    params.FeatureVersionID,
			KeyID:               params.KeyID,
			ServiceVersionID:    params.ServiceVersionID,
		})
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
	ServiceVersionID uint
}

func (s *ValueService) DeleteValue(ctx context.Context, params DeleteValueParams) error {
	variationValueChange, err := s.queries.GetChangeForVariationValue(ctx, db.GetChangeForVariationValueParams{
		ChangesetID:      params.ChangesetID,
		VariationValueID: params.ValueID,
	})
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if variationValueChange.ID == 0 {
			err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
				ChangesetID:         params.ChangesetID,
				FeatureVersionID:    params.FeatureVersionID,
				KeyID:               params.KeyID,
				ServiceVersionID:    params.ServiceVersionID,
				OldVariationValueID: params.ValueID,
			})
			if err != nil {
				return err
			}
		} else {
			if variationValueChange.Type == "delete" {
				panic("attempt to delete a already deleted value")
			}

			err = tx.DeleteChange(ctx, variationValueChange.ID)
			if err != nil {
				return err
			}

			if variationValueChange.Type == "create" {
				err = tx.DeleteVariationValue(ctx, *variationValueChange.NewVariationValueID)
				if err != nil {
					return err
				}
			} else if variationValueChange.Type == "update" {
				err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
					ChangesetID:         params.ChangesetID,
					FeatureVersionID:    params.FeatureVersionID,
					KeyID:               params.KeyID,
					ServiceVersionID:    params.ServiceVersionID,
					OldVariationValueID: *variationValueChange.OldVariationValueID,
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}
