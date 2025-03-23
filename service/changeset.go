package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/db"
)

type ChangesetService struct {
	variationContextService *VariationContextService
	queries                 *db.Queries
}

func NewChangesetService(queries *db.Queries, variationContextService *VariationContextService) *ChangesetService {
	return &ChangesetService{queries: queries, variationContextService: variationContextService}
}

func (s *ChangesetService) GetOpenChangesetForUser(ctx context.Context, userID uint) (uint, error) {
	id, err := s.queries.GetOpenChangesetIDForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return id, nil
}

func (s *ChangesetService) CreateChangesetForUser(ctx context.Context, userID uint) (uint, error) {
	id, err := s.queries.CreateChangeset(ctx, userID)

	if err != nil {
		return 0, err
	}

	return id, nil
}

type ChangesetChange struct {
	ID                             uint
	Type                           db.ChangesetChangeType
	ServiceVersionID               *uint
	ServiceName                    *string
	FeatureVersionID               *uint
	FeatureName                    *string
	FeatureVersionServiceVersionID *uint
	KeyID                          *uint
	KeyName                        *string
	NewVariationValueID            *uint
	NewVariationValueData          *string
	OldVariationValueID            *uint
	OldVariationValueData          *string
	Variation                      map[uint]string
}

type Changeset struct {
	ID               uint
	UserID           uint
	UserName         string
	State            db.ChangesetState
	ChangesetChanges []ChangesetChange
}

func (s *ChangesetService) GetChangeset(ctx context.Context, changesetID uint) (Changeset, error) {
	changeset, err := s.queries.GetChangeset(ctx, changesetID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Changeset{}, ErrRecordNotFound
		}

		return Changeset{}, err
	}

	changes, err := s.queries.GetChangesetChanges(ctx, changesetID)
	if err != nil {
		return Changeset{}, err
	}

	changesetChanges := make([]ChangesetChange, len(changes))
	for i, change := range changes {
		changesetChanges[i] = ChangesetChange{
			ID:                             change.ID,
			Type:                           change.Type,
			ServiceVersionID:               change.ServiceVersionID,
			ServiceName:                    change.ServiceName,
			FeatureVersionID:               change.FeatureVersionID,
			FeatureName:                    change.FeatureName,
			KeyID:                          change.KeyID,
			KeyName:                        change.KeyName,
			NewVariationValueID:            change.NewVariationValueID,
			NewVariationValueData:          change.NewVariationValueData,
			OldVariationValueID:            change.OldVariationValueID,
			OldVariationValueData:          change.OldVariationValueData,
			FeatureVersionServiceVersionID: change.FeatureVersionServiceVersionID,
		}

		if change.VariationContextID != nil {
			variation, err := s.variationContextService.GetVariationContextValues(ctx, *change.VariationContextID)
			if err != nil {
				return Changeset{}, err
			}

			changesetChanges[i].Variation = variation
		}
	}

	return Changeset{
		ID:               changeset.ID,
		UserID:           changeset.UserID,
		UserName:         changeset.UserName,
		State:            changeset.State,
		ChangesetChanges: changesetChanges,
	}, nil
}
