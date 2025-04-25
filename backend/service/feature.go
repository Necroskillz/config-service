package service

import (
	"context"

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
	FeatureID   uint   `json:"featureId" validate:"required"`
	Version     int    `json:"version" validate:"required"`
	Description string `json:"description" validate:"required"`
	Name        string `json:"name" validate:"required"`
	CanEdit     bool   `json:"canEdit" validate:"required"`
}

func (s *FeatureService) GetFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) (FeatureVersionDto, error) {
	_, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return FeatureVersionDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return FeatureVersionDto{
		ID:          featureVersion.ID,
		FeatureID:   featureVersion.FeatureID,
		Version:     featureVersion.Version,
		Description: featureVersion.FeatureDescription,
		Name:        featureVersion.FeatureName,
		CanEdit:     user.GetPermissionForService(featureVersion.FeatureID) >= constants.PermissionAdmin,
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
			FeatureID:   featureVersion.FeatureID,
			Version:     featureVersion.Version,
			Description: featureVersion.FeatureDescription,
			Name:        featureVersion.FeatureName,
			CanEdit:     user.GetPermissionForService(featureVersion.FeatureID) >= constants.PermissionAdmin,
		}
	}

	return result, nil
}

func (s *FeatureService) GetFeatureVersionsLinkedToServiceVersion(ctx context.Context, featureID uint, serviceVersionID uint) ([]VersionLinkDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	featureVersions, err := s.queries.GetFeatureVersionsLinkedToServiceVersion(ctx, db.GetFeatureVersionsLinkedToServiceVersionParams{
		FeatureID:        featureID,
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

type CreateFeatureParams struct {
	ServiceVersionID uint
	Name             string
	Description      string
}

func (s *FeatureService) validateCreateFeature(ctx context.Context, data CreateFeatureParams) error {
	return s.validator.
		Validate(data.Name, "Name").Required().FeatureNameNotTaken().MaxLength(100).
		Validate(data.Description, "Description").Required().MaxLength(500).
		Error(ctx)
}

func (s *FeatureService) CreateFeature(ctx context.Context, params CreateFeatureParams) (uint, error) {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, params.ServiceVersionID)
	if err != nil {
		return 0, err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return 0, NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create features for this service")
	}

	if err := s.validateCreateFeature(ctx, params); err != nil {
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
