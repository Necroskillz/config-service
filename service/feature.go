package service

import (
	"context"

	"github.com/necroskillz/config-service/db"
)

type FeatureService struct {
	unitOfWorkRunner db.UnitOfWorkRunner
	queries          *db.Queries
}

func NewFeatureService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	queries *db.Queries,
) *FeatureService {
	return &FeatureService{
		unitOfWorkRunner: unitOfWorkRunner,
		queries:          queries,
	}
}

func (s *FeatureService) GetFeatureVersion(ctx context.Context, featureVersionID uint) (db.GetFeatureVersionRow, error) {
	return s.queries.GetFeatureVersion(ctx, featureVersionID)
}

func (s *FeatureService) GetServiceFeatures(ctx context.Context, serviceVersionID uint, changesetID uint) ([]db.GetActiveFeatureVersionsForServiceVersionRow, error) {
	return s.queries.GetActiveFeatureVersionsForServiceVersion(ctx, db.GetActiveFeatureVersionsForServiceVersionParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      changesetID,
	})
}

func (s *FeatureService) GetFeatureVersionsLinkedToServiceVersion(ctx context.Context, featureID uint, serviceVersionID uint, changesetID uint) ([]db.GetFeatureVersionsLinkedToServiceVersionRow, error) {
	return s.queries.GetFeatureVersionsLinkedToServiceVersion(ctx, db.GetFeatureVersionsLinkedToServiceVersionParams{
		FeatureID:        featureID,
		ServiceVersionID: serviceVersionID,
		ChangesetID:      changesetID,
	})
}

type CreateFeatureParams struct {
	ChangesetID      uint
	ServiceVersionID uint
	Name             string
	Description      string
	ServiceID        uint
}

func (s *FeatureService) CreateFeature(ctx context.Context, params CreateFeatureParams) (uint, error) {
	var featureVersionID uint

	err := s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		featureID, err := tx.CreateFeature(ctx, db.CreateFeatureParams{
			Name:        params.Name,
			Description: params.Description,
			ServiceID:   params.ServiceID,
		})
		if err != nil {
			return err
		}

		featureVersionID, err = tx.CreateFeatureVersion(ctx, db.CreateFeatureVersionParams{
			FeatureID: featureID,
			Version:   1,
		})
		if err != nil {
			return err
		}

		if err = tx.AddCreateFeatureVersionChange(ctx, db.AddCreateFeatureVersionChangeParams{
			ChangesetID:      params.ChangesetID,
			FeatureVersionID: featureVersionID,
			ServiceVersionID: params.ServiceVersionID,
		}); err != nil {
			return err
		}

		linkID, err := tx.CreateFeatureVersionServiceVersion(ctx, db.CreateFeatureVersionServiceVersionParams{
			ServiceVersionID: params.ServiceVersionID,
			FeatureVersionID: featureVersionID,
		})
		if err != nil {
			return err
		}

		if err = tx.AddCreateFeatureVersionServiceVersionChange(ctx, db.AddCreateFeatureVersionServiceVersionChangeParams{
			ChangesetID:                    params.ChangesetID,
			FeatureVersionServiceVersionID: linkID,
			ServiceVersionID:               params.ServiceVersionID,
			FeatureVersionID:               featureVersionID,
		}); err != nil {
			return err
		}

		return nil
	})

	return featureVersionID, err
}
