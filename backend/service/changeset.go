package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type ChangesetService struct {
	currentUserAccessor     *auth.CurrentUserAccessor
	variationContextService *VariationContextService
	unitOfWorkRunner        db.UnitOfWorkRunner
	queries                 *db.Queries
	validator               *Validator
}

func NewChangesetService(
	queries *db.Queries,
	variationContextService *VariationContextService,
	unitOfWorkRunner db.UnitOfWorkRunner,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
) *ChangesetService {
	return &ChangesetService{
		queries:                 queries,
		variationContextService: variationContextService,
		unitOfWorkRunner:        unitOfWorkRunner,
		currentUserAccessor:     currentUserAccessor,
		validator:               validator,
	}
}

// use userID here, we dont have the user in the context yet, since this is called as part of creating it
func (s *ChangesetService) GetOpenChangesetIDForUser(ctx context.Context, userID uint) (uint, error) {
	id, err := s.queries.GetOpenChangesetIDForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, NewDbError(err, "Changeset")
	}

	return id, nil
}

func (s *ChangesetService) EnsureChangesetForUser(ctx context.Context) (uint, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	if user.ChangesetID != 0 {
		return user.ChangesetID, nil
	}

	id, err := s.queries.CreateChangeset(ctx, user.ID)

	if err != nil {
		return 0, err
	}

	user.ChangesetID = id

	return id, nil
}

type ChangesetChange struct {
	ID                             uint                   `json:"id" validate:"required"`
	Type                           db.ChangesetChangeType `json:"type" validate:"required"`
	ServiceVersionID               uint                   `json:"serviceVersionId" validate:"required"`
	ServiceName                    string                 `json:"serviceName" validate:"required"`
	ServiceVersion                 int                    `json:"serviceVersion" validate:"required"`
	PreviousServiceVersionID       *uint                  `json:"previousServiceVersionId"`
	FeatureVersionID               *uint                  `json:"featureVersionId"`
	FeatureName                    *string                `json:"featureName"`
	FeatureVersion                 *int                   `json:"featureVersion"`
	PreviousFeatureVersionID       *uint                  `json:"previousFeatureVersionId"`
	FeatureVersionServiceVersionID *uint                  `json:"featureVersionServiceVersionId"`
	KeyID                          *uint                  `json:"keyId"`
	KeyName                        *string                `json:"keyName"`
	NewVariationValueID            *uint                  `json:"newVariationValueId"`
	NewVariationValueData          *string                `json:"newVariationValueData"`
	OldVariationValueID            *uint                  `json:"oldVariationValueId"`
	OldVariationValueData          *string                `json:"oldVariationValueData"`
	Variation                      map[uint]string        `json:"variation"`
}

type Changeset struct {
	ID       uint
	UserID   uint
	UserName string
	State    db.ChangesetState
}

type ChangesetWithChanges struct {
	Changeset
	ChangesetChanges []ChangesetChange
}

func (c ChangesetWithChanges) CanBeAppliedBy(user *auth.User) bool {
	if !c.IsOpen() && !c.IsCommitted() {
		return false
	}

	if c.IsOpen() && !c.BelongsTo(user.ID) {
		return false
	}

	if user.IsGlobalAdmin {
		return true
	}

	for _, change := range c.ChangesetChanges {
		if user.GetPermissionForService(change.ServiceVersionID) != constants.PermissionAdmin {
			return false
		}
	}

	return true
}

func (c Changeset) BelongsTo(userID uint) bool {
	return c.UserID == userID
}

func (c Changeset) IsOpen() bool {
	return c.State == db.ChangesetStateOpen
}

func (c Changeset) IsCommitted() bool {
	return c.State == db.ChangesetStateCommitted
}

func (c Changeset) IsDiscarded() bool {
	return c.State == db.ChangesetStateDiscarded
}

func (c Changeset) IsStashed() bool {
	return c.State == db.ChangesetStateStashed
}

func (c ChangesetWithChanges) IsEmpty() bool {
	return len(c.ChangesetChanges) == 0
}

func (s *ChangesetService) getChangesetWithoutChanges(ctx context.Context, changesetID uint) (Changeset, error) {
	changeset, err := s.queries.GetChangeset(ctx, changesetID)
	if err != nil {
		return Changeset{}, NewDbError(err, "Changeset")
	}

	return Changeset{
		ID:       changeset.ID,
		UserID:   changeset.UserID,
		UserName: changeset.UserName,
		State:    changeset.State,
	}, nil
}

func (s *ChangesetService) getChangeset(ctx context.Context, changesetID uint) (ChangesetWithChanges, error) {
	changesetWithChanges := ChangesetWithChanges{}

	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return changesetWithChanges, err
	}

	changesetWithChanges.Changeset = changeset

	changes, err := s.queries.GetChangesetChanges(ctx, changesetID)
	if err != nil {
		return changesetWithChanges, err
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
				return changesetWithChanges, err
			}

			changesetChanges[i].Variation = variation
		}
	}

	changesetWithChanges.ChangesetChanges = changesetChanges

	return changesetWithChanges, nil
}

type ChangesetAction struct {
	ID        uint                   `json:"id" validate:"required"`
	Type      db.ChangesetActionType `json:"type" validate:"required"`
	Comment   *string                `json:"comment"`
	CreatedAt time.Time              `json:"createdAt" validate:"required"`
	UserID    uint                   `json:"userId" validate:"required"`
	UserName  string                 `json:"userName" validate:"required"`
}

type ChangesetDto struct {
	ID               uint              `json:"id" validate:"required"`
	UserID           uint              `json:"userId" validate:"required"`
	UserName         string            `json:"userName" validate:"required"`
	State            db.ChangesetState `json:"state" validate:"required"`
	CanApply         bool              `json:"canApply" validate:"required"`
	VariationContext map[uint]string   `json:"variationContext" validate:"required"`
	Changes          []ChangesetChange `json:"changes" validate:"required"`
	Actions          []ChangesetAction `json:"actions" validate:"required"`
}

func (s *ChangesetService) GetChangeset(ctx context.Context, changesetID uint) (ChangesetDto, error) {
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return ChangesetDto{}, err
	}

	actions, err := s.queries.GetChangesetActions(ctx, changesetID)
	if err != nil {
		return ChangesetDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	dto := ChangesetDto{
		ID:       changeset.ID,
		UserID:   changeset.UserID,
		UserName: changeset.UserName,
		State:    changeset.State,
		Changes:  changeset.ChangesetChanges,
		CanApply: changeset.CanBeAppliedBy(user),
	}

	dto.Actions = make([]ChangesetAction, 0, len(actions))
	for _, action := range actions {
		dto.Actions = append(dto.Actions, ChangesetAction{
			ID:        action.ID,
			Type:      action.Type,
			Comment:   action.Comment,
			CreatedAt: action.CreatedAt,
			UserID:    action.UserID,
			UserName:  action.UserName,
		})
	}

	return dto, nil
}

func (s *ChangesetService) ApplyChangeset(ctx context.Context, changesetID uint, comment *string) error {
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.CanBeAppliedBy(user) {
		return NewServiceError(ErrorCodePermissionDenied, "user does not have permission to apply changeset")
	}

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
					if change.PreviousFeatureVersionID != nil {
						if err := tx.EndFeatureVersionValidity(ctx, db.EndFeatureVersionValidityParams{
							FeatureVersionID: *change.PreviousFeatureVersionID,
							ValidTo:          &endTime,
						}); err != nil {
							return err
						}
					}
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
			} else {
				if change.Type == db.ChangesetChangeTypeCreate {
					if change.PreviousServiceVersionID != nil {
						if err := tx.EndServiceVersionValidity(ctx, db.EndServiceVersionValidityParams{
							ServiceVersionID: *change.PreviousServiceVersionID,
							ValidTo:          &endTime,
						}); err != nil {
							return err
						}
					}

					if err := tx.StartServiceVersionValidity(ctx, db.StartServiceVersionValidityParams{
						ServiceVersionID: change.ServiceVersionID,
						ValidFrom:        &startTime,
					}); err != nil {
						return err
					}
				} else if change.Type == db.ChangesetChangeTypeDelete {
					if err := tx.EndServiceVersionValidity(ctx, db.EndServiceVersionValidityParams{
						ServiceVersionID: change.ServiceVersionID,
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

		if err := tx.AddChangesetAction(ctx, db.AddChangesetActionParams{
			ChangesetID: changeset.ID,
			UserID:      user.ID,
			Type:        db.ChangesetActionTypeApply,
			Comment:     comment,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *ChangesetService) CommitChangeset(ctx context.Context, changesetID uint, comment *string) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.IsOpen() {
		return NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Cannot commit changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return NewServiceError(ErrorCodePermissionDenied, "You are not allowed to commit this changeset")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err := tx.SetChangesetState(ctx, db.SetChangesetStateParams{
			ChangesetID: changeset.ID,
			State:       db.ChangesetStateCommitted,
		}); err != nil {
			return err
		}

		if err := tx.AddChangesetAction(ctx, db.AddChangesetActionParams{
			ChangesetID: changeset.ID,
			UserID:      user.ID,
			Type:        db.ChangesetActionTypeCommit,
			Comment:     comment,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *ChangesetService) ReopenChangeset(ctx context.Context, changesetID uint) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.ChangesetID != 0 {
		return NewServiceError(ErrorCodeInvalidOperation, "You already have an open changeset")
	}

	if !changeset.IsCommitted() && !changeset.IsStashed() {
		return NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Cannot reopen changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return NewServiceError(ErrorCodePermissionDenied, "You are not allowed to reopen this changeset")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err := tx.SetChangesetState(ctx, db.SetChangesetStateParams{
			ChangesetID: changeset.ID,
			State:       db.ChangesetStateOpen,
		}); err != nil {
			return err
		}

		if err := tx.AddChangesetAction(ctx, db.AddChangesetActionParams{
			ChangesetID: changeset.ID,
			UserID:      user.ID,
			Type:        db.ChangesetActionTypeReopen,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *ChangesetService) DiscardChangeset(ctx context.Context, changesetID uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() && !changeset.IsCommitted() {
		return NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Cannot discard changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return NewServiceError(ErrorCodePermissionDenied, "You are not allowed to discard this changeset")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
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
			} else if change.Type == db.ChangesetChangeTypeCreate {
				if err := tx.DeleteServiceVersion(ctx, change.ServiceVersionID); err != nil {
					return err
				}

				if change.ServiceVersion == 1 {
					if err := tx.DeleteService(ctx, change.ServiceVersionID); err != nil {
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

		if err := tx.AddChangesetAction(ctx, db.AddChangesetActionParams{
			ChangesetID: changeset.ID,
			UserID:      user.ID,
			Type:        db.ChangesetActionTypeDiscard,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *ChangesetService) StashChangeset(ctx context.Context, changesetID uint) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.IsOpen() {
		return NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Cannot stash changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return NewServiceError(ErrorCodePermissionDenied, "You are not allowed to stash this changeset")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err := tx.SetChangesetState(ctx, db.SetChangesetStateParams{
			ChangesetID: changeset.ID,
			State:       db.ChangesetStateStashed,
		}); err != nil {
			return err
		}

		if err := tx.AddChangesetAction(ctx, db.AddChangesetActionParams{
			ChangesetID: changeset.ID,
			UserID:      user.ID,
			Type:        db.ChangesetActionTypeStash,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (s *ChangesetService) validateAddComment(ctx context.Context, comment string) error {
	return s.validator.Validate(comment, "comment").Required().MaxLength(1000).Error(ctx)
}

func (s *ChangesetService) AddComment(ctx context.Context, changesetID uint, comment string) error {
	if err := s.validateAddComment(ctx, comment); err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if err := s.queries.AddChangesetAction(ctx, db.AddChangesetActionParams{
		ChangesetID: changesetID,
		UserID:      user.ID,
		Type:        db.ChangesetActionTypeComment,
		Comment:     &comment,
	}); err != nil {
		return err
	}

	return nil
}
