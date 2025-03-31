package service

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type ChangesetService struct {
	variationContextService *VariationContextService
	unitOfWorkRunner        db.UnitOfWorkRunner
	queries                 *db.Queries
}

func NewChangesetService(queries *db.Queries, variationContextService *VariationContextService, unitOfWorkRunner db.UnitOfWorkRunner) *ChangesetService {
	return &ChangesetService{queries: queries, variationContextService: variationContextService, unitOfWorkRunner: unitOfWorkRunner}
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
	ServiceVersion                 *int
	FeatureVersionID               *uint
	FeatureName                    *string
	FeatureVersion                 *int
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

func (c *Changeset) CanBeAppliedBy(checker PermissionChecker) bool {
	userID := checker.GetID()

	if !c.IsOpen() && !c.IsCommitted() {
		return false
	}

	if c.IsOpen() && !c.BelongsTo(userID) {
		return false
	}

	if checker.IsGlobalAdministrator() {
		return true
	}

	for _, change := range c.ChangesetChanges {
		if change.ServiceVersionID == nil {
			panic("service version id is nil")
		}

		if checker.GetPermissionForService(*change.ServiceVersionID) != constants.PermissionAdmin {
			return false
		}
	}

	return true
}

func (c *Changeset) BelongsTo(userID uint) bool {
	return c.UserID == userID
}

func (c *Changeset) IsOpen() bool {
	return c.State == db.ChangesetStateOpen
}

func (c *Changeset) IsCommitted() bool {
	return c.State == db.ChangesetStateCommitted
}

func (c *Changeset) IsDiscarded() bool {
	return c.State == db.ChangesetStateDiscarded
}

func (c *Changeset) IsStashed() bool {
	return c.State == db.ChangesetStateStashed
}

func (c *Changeset) IsEmpty() bool {
	return len(c.ChangesetChanges) == 0
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
			ServiceVersion:                 change.ServiceVersion,
			FeatureVersionID:               change.FeatureVersionID,
			FeatureName:                    change.FeatureName,
			FeatureVersion:                 change.FeatureVersion,
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

func (s *ChangesetService) ApplyChangeset(ctx context.Context, changeset *Changeset) error {
	startTime := time.Now()
	endTime := startTime.Add(time.Microsecond * -1)

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		for _, change := range changeset.ChangesetChanges {
			if change.NewVariationValueID != nil || change.OldVariationValueID != nil {
				if change.NewVariationValueID != nil {
					if err := tx.StartValueValidity(ctx, db.StartValueValidityParams{
						VariationValueID: *change.NewVariationValueID,
						ValidFrom:        &startTime,
					}); err != nil {
						return err
					}
				}

				if change.OldVariationValueID != nil {
					if err := tx.EndValueValidity(ctx, db.EndValueValidityParams{
						VariationValueID: *change.OldVariationValueID,
						ValidTo:          &endTime,
					}); err != nil {
						return err
					}
				}
			} else if change.KeyID != nil {
				if change.Type == db.ChangesetChangeTypeCreate {
					if err := tx.StartKeyValidity(ctx, db.StartKeyValidityParams{
						KeyID:     *change.KeyID,
						ValidFrom: &startTime,
					}); err != nil {
						return err
					}
				} else if change.Type == db.ChangesetChangeTypeDelete {
					if err := tx.EndKeyValidity(ctx, db.EndKeyValidityParams{
						KeyID:   *change.KeyID,
						ValidTo: &endTime,
					}); err != nil {
						return err
					}
				}
			} else if change.FeatureVersionServiceVersionID != nil {
				if change.Type == db.ChangesetChangeTypeCreate {
					if err := tx.StartFeatureVersionServiceVersionValidity(ctx, db.StartFeatureVersionServiceVersionValidityParams{
						FeatureVersionServiceVersionID: *change.FeatureVersionServiceVersionID,
						ValidFrom:                      &startTime,
					}); err != nil {
						return err
					}
				} else if change.Type == db.ChangesetChangeTypeDelete {
					if err := tx.EndFeatureVersionServiceVersionValidity(ctx, db.EndFeatureVersionServiceVersionValidityParams{
						FeatureVersionServiceVersionID: *change.FeatureVersionServiceVersionID,
						ValidTo:                        &endTime,
					}); err != nil {
						return err
					}
				}
			} else if change.FeatureVersionID != nil {
				if change.Type == db.ChangesetChangeTypeCreate {
					if err := tx.StartFeatureVersionValidity(ctx, db.StartFeatureVersionValidityParams{
						FeatureVersionID: *change.FeatureVersionID,
						ValidFrom:        &startTime,
					}); err != nil {
						return err
					}
				} else if change.Type == db.ChangesetChangeTypeDelete {
					if err := tx.EndFeatureVersionValidity(ctx, db.EndFeatureVersionValidityParams{
						FeatureVersionID: *change.FeatureVersionID,
						ValidTo:          &endTime,
					}); err != nil {
						return err
					}
				}
			} else if change.ServiceVersionID != nil {
				if change.Type == db.ChangesetChangeTypeCreate {
					if err := tx.StartServiceVersionValidity(ctx, db.StartServiceVersionValidityParams{
						ServiceVersionID: *change.ServiceVersionID,
						ValidFrom:        &startTime,
					}); err != nil {
						return err
					}
				} else if change.Type == db.ChangesetChangeTypeDelete {
					if err := tx.EndServiceVersionValidity(ctx, db.EndServiceVersionValidityParams{
						ServiceVersionID: *change.ServiceVersionID,
						ValidTo:          &endTime,
					}); err != nil {
						return err
					}
				}
			}
		}

		if err := tx.SetChangesetState(ctx, db.SetChangesetStateParams{
			ChangesetID: changeset.ID,
			State:       db.ChangesetStateApplied,
		}); err != nil {
			return err
		}

		changeset.State = db.ChangesetStateApplied

		return nil
	})
}

func (s *ChangesetService) CommitChangeset(ctx context.Context, changeset *Changeset) error {
	err := s.queries.SetChangesetState(ctx, db.SetChangesetStateParams{
		ChangesetID: changeset.ID,
		State:       db.ChangesetStateCommitted,
	})

	if err != nil {
		return err
	}

	changeset.State = db.ChangesetStateCommitted

	return nil
}

func (s *ChangesetService) ReopenChangeset(ctx context.Context, changeset *Changeset) error {
	err := s.queries.SetChangesetState(ctx, db.SetChangesetStateParams{
		ChangesetID: changeset.ID,
		State:       db.ChangesetStateOpen,
	})

	if err != nil {
		return err
	}

	changeset.State = db.ChangesetStateOpen

	return nil
}

func (s *ChangesetService) DiscardChangeset(ctx context.Context, changeset *Changeset) error {
	err := s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		for _, change := range slices.Backward(changeset.ChangesetChanges) {
			if err := tx.DeleteChange(ctx, change.ID); err != nil {
				return err
			}

			if change.NewVariationValueID != nil || change.OldVariationValueID != nil {
				if change.NewVariationValueID != nil {
					if err := tx.DeleteVariationValue(ctx, *change.NewVariationValueID); err != nil {
						return err
					}
				}
			} else if change.KeyID != nil && change.Type == db.ChangesetChangeTypeCreate {
				if err := tx.DeleteKey(ctx, *change.KeyID); err != nil {
					return err
				}
			} else if change.FeatureVersionServiceVersionID != nil && change.Type == db.ChangesetChangeTypeCreate {
				if err := tx.DeleteFeatureVersionServiceVersion(ctx, *change.FeatureVersionServiceVersionID); err != nil {
					return err
				}
			} else if change.FeatureVersionID != nil && change.Type == db.ChangesetChangeTypeCreate {
				if err := tx.DeleteFeatureVersion(ctx, *change.FeatureVersionID); err != nil {
					return err
				}

				if *change.FeatureVersion == 1 {
					if err := tx.DeleteFeature(ctx, *change.FeatureVersionID); err != nil {
						return err
					}
				}
			} else if change.ServiceVersionID != nil && change.Type == db.ChangesetChangeTypeCreate {
				if err := tx.DeleteServiceVersion(ctx, *change.ServiceVersionID); err != nil {
					return err
				}

				if *change.ServiceVersion == 1 {
					if err := tx.DeleteService(ctx, *change.ServiceVersionID); err != nil {
						return err
					}
				}
			}
		}

		if err := tx.SetChangesetState(ctx, db.SetChangesetStateParams{
			ChangesetID: changeset.ID,
			State:       db.ChangesetStateDiscarded,
		}); err != nil {
			return err
		}

		changeset.State = db.ChangesetStateDiscarded
		changeset.ChangesetChanges = nil

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (s *ChangesetService) StashChangeset(ctx context.Context, changeset *Changeset) error {
	err := s.queries.SetChangesetState(ctx, db.SetChangesetStateParams{
		ChangesetID: changeset.ID,
		State:       db.ChangesetStateStashed,
	})

	if err != nil {
		return err
	}

	changeset.State = db.ChangesetStateStashed

	return nil
}
