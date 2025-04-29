package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type FeatureService struct {
	unitOfWorkRunner    db.UnitOfWorkRunner
	queries             *db.Queries
	changesetService    *ChangesetService
	currentUserAccessor *auth.CurrentUserAccessor
	validator           *Validator
	coreService         *CoreService
}

func NewFeatureService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	queries *db.Queries,
	changesetService *ChangesetService,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
	coreService *CoreService,
) *FeatureService {
	return &FeatureService{
		unitOfWorkRunner:    unitOfWorkRunner,
		queries:             queries,
		changesetService:    changesetService,
		currentUserAccessor: currentUserAccessor,
		validator:           validator,
		coreService:         coreService,
	}
}

type FeatureVersionDto struct {
	ID          uint   `json:"id" validate:"required"`
	Version     int    `json:"version" validate:"required"`
	Description string `json:"description" validate:"required"`
	Name        string `json:"name" validate:"required"`
}

type FeatureVersionWithPermissionDto struct {
	FeatureVersionDto
	CanEdit bool `json:"canEdit" validate:"required"`
}

func (s *FeatureService) GetFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) (FeatureVersionWithPermissionDto, error) {
	_, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return FeatureVersionWithPermissionDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return FeatureVersionWithPermissionDto{
		FeatureVersionDto: FeatureVersionDto{
			ID:          featureVersion.ID,
			Version:     featureVersion.Version,
			Description: featureVersion.FeatureDescription,
			Name:        featureVersion.FeatureName,
		},
		CanEdit: user.GetPermissionForService(featureVersion.FeatureID) >= constants.PermissionAdmin,
	}, nil
}

func (s *FeatureService) GetServiceFeatures(ctx context.Context, serviceVersionID uint) ([]FeatureVersionDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)
	_, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return nil, err
	}

	featureVersions, err := s.queries.GetActiveFeatureVersionsForServiceVersion(ctx, db.GetActiveFeatureVersionsForServiceVersionParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]FeatureVersionDto, len(featureVersions))
	for i, featureVersion := range featureVersions {
		result[i] = FeatureVersionDto{
			ID:          featureVersion.ID,
			Version:     featureVersion.Version,
			Description: featureVersion.FeatureDescription,
			Name:        featureVersion.FeatureName,
		}
	}

	return result, nil
}

func (s *FeatureService) GetFeatureVersionsLinkedToServiceVersion(ctx context.Context, featureVersionID uint, serviceVersionID uint) ([]VersionLinkDto, error) {
	_, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	featureVersions, err := s.queries.GetFeatureVersionsLinkedToServiceVersion(ctx, db.GetFeatureVersionsLinkedToServiceVersionParams{
		FeatureID:        featureVersion.FeatureID,
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]VersionLinkDto, len(featureVersions))
	for i, featureVersion := range featureVersions {
		result[i] = VersionLinkDto{
			ID:      featureVersion.ID,
			Version: featureVersion.Version,
		}
	}

	return result, nil
}

func (s *FeatureService) GetFeatureVersionsLinkableToServiceVersion(ctx context.Context, serviceVersionID uint) ([]FeatureVersionDto, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	featureVersions, err := s.queries.GetFeatureVersionsLinkableToServiceVersion(ctx, db.GetFeatureVersionsLinkableToServiceVersionParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
		ServiceID:        serviceVersion.ServiceID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]FeatureVersionDto, len(featureVersions))
	for i, featureVersion := range featureVersions {
		result[i] = FeatureVersionDto{
			ID:          featureVersion.ID,
			Version:     featureVersion.Version,
			Description: featureVersion.FeatureDescription,
			Name:        featureVersion.FeatureName,
		}
	}

	return result, nil
}

type CreateFeatureParams struct {
	ServiceVersionID uint
	Name             string
	Description      string
}

func descriptionValidatorFunc(v *ValidatorContext) *ValidatorContext {
	return v.Required().MaxLength(500)
}

func (s *FeatureService) validateCreateFeature(ctx context.Context, data CreateFeatureParams, serviceVersion db.GetServiceVersionRow) error {
	v := s.validator.
		Validate(data.Name, "Name").Required().FeatureNameNotTaken().MaxLength(100).
		Validate(data.Description, "Description").Func(descriptionValidatorFunc)

	if err := v.Error(ctx); err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create features for this service")
	}

	return nil
}

func (s *FeatureService) CreateFeature(ctx context.Context, params CreateFeatureParams) (uint, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, params.ServiceVersionID)
	if err != nil {
		return 0, err
	}

	if err := s.validateCreateFeature(ctx, params, serviceVersion); err != nil {
		return 0, err
	}

	var featureVersionID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		featureID, err := tx.CreateFeature(ctx, db.CreateFeatureParams{
			Name:        params.Name,
			Description: params.Description,
			ServiceID:   serviceVersion.ServiceID,
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
			ChangesetID:      changesetID,
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
			ChangesetID:                    changesetID,
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

type UpdateFeatureParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	Description      string
}

func (s *FeatureService) validateUpdateFeature(ctx context.Context, data UpdateFeatureParams, serviceVersion db.GetServiceVersionRow) error {
	err := s.validator.
		Validate(data.Description, "Description").Func(descriptionValidatorFunc).
		Error(ctx)

	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create features for this service")
	}

	return nil
}

func (s *FeatureService) UpdateFeature(ctx context.Context, params UpdateFeatureParams) error {
	serviceVersion, featureVersion, err := s.coreService.GetFeatureVersion(ctx, params.ServiceVersionID, params.FeatureVersionID)
	if err != nil {
		return err
	}

	if err := s.validateUpdateFeature(ctx, params, serviceVersion); err != nil {
		return err
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		err = tx.UpdateFeature(ctx, db.UpdateFeatureParams{
			FeatureID:   featureVersion.FeatureID,
			Description: params.Description,
		})
		if err != nil {
			return err
		}

		return nil
	})
}

func (s *FeatureService) UnlinkFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) error {
	serviceVersion, _, linkID, err := s.coreService.GetFeatureVersionWithLinkID(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to unlink features for this service")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		change, err := tx.GetChangeForFeatureVersionServiceVersion(ctx, db.GetChangeForFeatureVersionServiceVersionParams{
			ChangesetID:      changesetID,
			ServiceVersionID: serviceVersionID,
			FeatureVersionID: featureVersionID,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		if change.ID == 0 {
			if err = tx.AddDeleteFeatureVersionServiceVersionChange(ctx, db.AddDeleteFeatureVersionServiceVersionChangeParams{
				ChangesetID:                    changesetID,
				FeatureVersionServiceVersionID: linkID,
				ServiceVersionID:               serviceVersionID,
				FeatureVersionID:               featureVersionID,
			}); err != nil {
				return err
			}
		} else {
			if change.Type == "delete" {
				return NewServiceError(ErrorCodeInvalidOperation, "Feature version already has an unlink change in the current changeset")
			}

			if err = tx.DeleteChange(ctx, change.ID); err != nil {
				return err
			}

			if err = tx.DeleteFeatureVersionServiceVersion(ctx, linkID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *FeatureService) LinkFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) error {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to link features for this service")
	}

	featureVersion, err := s.queries.GetFeatureVersion(ctx, featureVersionID)
	if err != nil {
		return err
	}

	if featureVersion.ServiceID != serviceVersion.ServiceID {
		return NewServiceError(ErrorCodeInvalidOperation, "Unable to link feature version to a different service version")
	}

	linked, err := s.queries.IsFeatureLinkedToServiceVersion(ctx, db.IsFeatureLinkedToServiceVersionParams{
		FeatureID:        featureVersion.FeatureID,
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return err
	}

	if linked {
		return NewServiceError(ErrorCodeInvalidOperation, "Feature is already linked to this service version")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		change, err := tx.GetChangeForFeatureVersionServiceVersion(ctx, db.GetChangeForFeatureVersionServiceVersionParams{
			ChangesetID:      changesetID,
			ServiceVersionID: serviceVersionID,
			FeatureVersionID: featureVersionID,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		if change.ID == 0 {
			linkID, err := tx.CreateFeatureVersionServiceVersion(ctx, db.CreateFeatureVersionServiceVersionParams{
				ServiceVersionID: serviceVersionID,
				FeatureVersionID: featureVersionID,
			})
			if err != nil {
				return err
			}

			if err = tx.AddCreateFeatureVersionServiceVersionChange(ctx, db.AddCreateFeatureVersionServiceVersionChangeParams{
				ChangesetID:                    changesetID,
				FeatureVersionServiceVersionID: linkID,
				ServiceVersionID:               serviceVersionID,
				FeatureVersionID:               featureVersionID,
			}); err != nil {
				return err
			}
		} else {
			// has to be a delete change
			if err = tx.DeleteChange(ctx, change.ID); err != nil {
				return err
			}
		}

		return nil
	})
}
