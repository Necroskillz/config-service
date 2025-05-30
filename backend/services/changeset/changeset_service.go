package changeset

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	currentUserAccessor     *auth.CurrentUserAccessor
	variationContextService *variation.ContextService
	unitOfWorkRunner        db.UnitOfWorkRunner
	queries                 *db.Queries
	validator               *validator.Validator
}

func NewService(
	queries *db.Queries,
	variationContextService *variation.ContextService,
	unitOfWorkRunner db.UnitOfWorkRunner,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *validator.Validator,
) *Service {
	return &Service{
		queries:                 queries,
		variationContextService: variationContextService,
		unitOfWorkRunner:        unitOfWorkRunner,
		currentUserAccessor:     currentUserAccessor,
		validator:               validator,
	}
}

// use userID here, we dont have the user in the context yet, since this is called as part of creating it
func (s *Service) GetOpenChangesetIDForUser(ctx context.Context, userID uint) (uint, error) {
	id, err := s.queries.GetOpenChangesetIDForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}

		return 0, core.NewDbError(err, "Changeset")
	}

	return id, nil
}

func (s *Service) EnsureChangesetForUser(ctx context.Context) (uint, error) {
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

type Filter struct {
	Page       int
	PageSize   int
	AuthorID   *uint
	Approvable bool
}

type ChangesetItemDto struct {
	ID           uint              `json:"id" validate:"required"`
	CreatedAt    time.Time         `json:"createdAt" validate:"required"`
	State        db.ChangesetState `json:"state" validate:"required"`
	LastActionAt time.Time         `json:"lastActionAt" validate:"required"`
	ActionCount  int               `json:"actionCount" validate:"required"`
	UserName     string            `json:"userName" validate:"required"`
	UserID       uint              `json:"userId" validate:"required"`
}

func (s *Service) GetChangesets(ctx context.Context, filter Filter) (core.PaginatedResult[ChangesetItemDto], error) {
	if filter.Page < 1 {
		return core.PaginatedResult[ChangesetItemDto]{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Page must be 1 or greater")
	}

	if filter.PageSize < 1 || filter.PageSize > 100 {
		return core.PaginatedResult[ChangesetItemDto]{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Page size must be between 1 and 100")
	}

	var approverID *uint
	if filter.Approvable {
		user := s.currentUserAccessor.GetUser(ctx)
		approverID = &user.ID
	}

	changesets, err := s.queries.GetChangesets(ctx, db.GetChangesetsParams{
		Limit:      filter.PageSize,
		Offset:     (filter.Page - 1) * filter.PageSize,
		UserID:     filter.AuthorID,
		ApproverID: approverID,
	})
	if err != nil {
		return core.PaginatedResult[ChangesetItemDto]{}, core.NewDbError(err, "Changesets")
	}

	changesetItems := make([]ChangesetItemDto, len(changesets))
	for i, changeset := range changesets {
		changesetItems[i] = ChangesetItemDto{
			ID:           changeset.ID,
			CreatedAt:    changeset.CreatedAt,
			State:        changeset.State,
			LastActionAt: changeset.LastActionAt,
			ActionCount:  changeset.ActionCount,
			UserName:     changeset.UserName,
			UserID:       changeset.UserID,
		}
	}

	var total int
	if len(changesets) > 0 {
		total = changesets[0].TotalCount
	}

	return core.PaginatedResult[ChangesetItemDto]{
		Items:      changesetItems,
		TotalCount: total,
	}, nil
}

func (s *Service) GetApprovableChangesetCount(ctx context.Context) (int, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	count, err := s.queries.GetApprovableChangesetCount(ctx, user.ID)
	if err != nil {
		return 0, core.NewDbError(err, "ApprovableChangesetCount")
	}

	return count, nil
}

func (s *Service) getChangesetWithoutChanges(ctx context.Context, changesetID uint) (Changeset, error) {
	changeset, err := s.queries.GetChangeset(ctx, changesetID)
	if err != nil {
		return Changeset{}, core.NewDbError(err, "Changeset")
	}

	return Changeset{
		ID:       changeset.ID,
		UserID:   changeset.UserID,
		UserName: changeset.UserName,
		State:    changeset.State,
	}, nil
}

func (s *Service) getChangeset(ctx context.Context, changesetID uint) (ChangesetWithChanges, error) {
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
			ServiceID:                      change.ServiceID,
			PreviousServiceVersionID:       change.PreviousServiceVersionID,
			FeatureVersionID:               change.FeatureVersionID,
			FeatureName:                    change.FeatureName,
			FeatureVersion:                 change.FeatureVersion,
			FeatureID:                      change.FeatureID,
			PreviousFeatureVersionID:       change.PreviousFeatureVersionID,
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

func (s *Service) GetChangeset(ctx context.Context, changesetID uint) (ChangesetDto, error) {
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

func (s *Service) ApplyChangeset(ctx context.Context, changesetID uint, comment *string) error {
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.CanBeAppliedBy(user) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "To apply changeset, it needs to be in an open or committed state and the user needs to have admin permissions for all changes")
	}

	startTime := time.Now()
	endTime := startTime.Add(time.Microsecond * -1)

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		for _, change := range changeset.ChangesetChanges {
			if change.NewVariationValueID != nil || change.OldVariationValueID != nil {
				if change.OldVariationValueID != nil {
					if err := tx.EndValueValidity(ctx, db.EndValueValidityParams{
						VariationValueID: *change.OldVariationValueID,
						ValidTo:          &endTime,
					}); err != nil {
						return err
					}
				}

				if change.NewVariationValueID != nil {
					if err := tx.StartValueValidity(ctx, db.StartValueValidityParams{
						VariationValueID: *change.NewVariationValueID,
						ValidFrom:        &startTime,
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
			} else {
				if change.Type == db.ChangesetChangeTypeCreate {
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
			AppliedAt:   &startTime,
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

func (s *Service) CommitChangeset(ctx context.Context, changesetID uint, comment *string) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.IsOpen() {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Cannot commit changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not allowed to commit this changeset")
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

func (s *Service) ReopenChangeset(ctx context.Context, changesetID uint) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.ChangesetID != 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You already have an open changeset")
	}

	if !changeset.IsCommitted() && !changeset.IsStashed() {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Cannot reopen changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not allowed to reopen this changeset")
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

func (s *Service) discardChangeEntity(ctx context.Context, tx *db.Queries, change ChangesetChange) error {
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
			if err := tx.DeleteFeature(ctx, *change.FeatureID); err != nil {
				return err
			}
		}
	} else if change.Type == db.ChangesetChangeTypeCreate {
		if err := tx.DeleteServiceVersion(ctx, change.ServiceVersionID); err != nil {
			return err
		}

		if change.ServiceVersion == 1 {
			if err := tx.DeleteService(ctx, change.ServiceID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *Service) DiscardChangeset(ctx context.Context, changesetID uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() && !changeset.IsCommitted() {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Cannot discard changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not allowed to discard this changeset")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		for _, change := range slices.Backward(changeset.ChangesetChanges) {
			if err := s.discardChangeEntity(ctx, tx, change); err != nil {
				return err
			}
		}

		if err := tx.DeleteChangesForChangeset(ctx, changeset.ID); err != nil {
			return err
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

func (s *Service) DiscardChange(ctx context.Context, changesetID uint, changeID uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	changeset, err := s.getChangeset(ctx, changesetID)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Cannot discard changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not allowed to discard this changeset")
	}

	changeIndex := slices.IndexFunc(changeset.ChangesetChanges, func(c ChangesetChange) bool {
		return c.ID == changeID
	})

	if changeIndex == -1 {
		return core.NewServiceError(core.ErrorCodeRecordNotFound, "Change not found")
	}

	change := changeset.ChangesetChanges[changeIndex]

	if change.Type == db.ChangesetChangeTypeCreate && change.Variation != nil && len(change.Variation) == 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot discard the default value of a key")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err := tx.DeleteChange(ctx, change.ID); err != nil {
			return err
		}

		if err := s.discardChangeEntity(ctx, tx, change); err != nil {
			return err
		}

		return nil
	})
}

func (s *Service) StashChangeset(ctx context.Context, changesetID uint) error {
	changeset, err := s.getChangesetWithoutChanges(ctx, changesetID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if !changeset.IsOpen() {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Cannot stash changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(user.ID) {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not allowed to stash this changeset")
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

func (s *Service) validateAddComment(ctx context.Context, comment string) error {
	return s.validator.Validate(comment, "comment").Required().MaxLength(1000).Error(ctx)
}

func (s *Service) AddComment(ctx context.Context, changesetID uint, comment string) error {
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

func (s *Service) GetChangesetChangesCount(ctx context.Context) (int, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	if user.ChangesetID == 0 {
		return 0, nil
	}

	count, err := s.queries.GetChangesetChangesCount(ctx, user.ChangesetID)
	if err != nil {
		return 0, core.NewDbError(err, "ChangesetChangesCount")
	}

	return count, nil
}
