package validation

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
)

type Service struct {
	queries                   *db.Queries
	variationContextService   *variation.ContextService
	variationHierarchyService *variation.HierarchyService
	currentUserAccessor       *auth.CurrentUserAccessor
	coreService               *core.Service
}

func NewService(queries *db.Queries, variationContextService *variation.ContextService, variationHierarchyService *variation.HierarchyService, currentUserAccessor *auth.CurrentUserAccessor, coreService *core.Service) *Service {
	return &Service{
		queries:                   queries,
		variationContextService:   variationContextService,
		variationHierarchyService: variationHierarchyService,
		currentUserAccessor:       currentUserAccessor,
		coreService:               coreService,
	}
}

func (s *Service) IsServiceTypeNameTaken(ctx context.Context, name string) (bool, error) {
	_, err := s.queries.GetServiceTypeIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
	}

	return true, nil
}

func (s *Service) IsServiceNameTaken(ctx context.Context, name string) (bool, error) {
	_, err := s.queries.GetServiceIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *Service) IsFeatureNameTaken(ctx context.Context, name string) (bool, error) {
	_, err := s.queries.GetFeatureIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *Service) IsKeyNameTaken(ctx context.Context, featureVersionID uint, keyName string) (bool, error) {
	_, err := s.queries.GetKeyIDByName(ctx, db.GetKeyIDByNameParams{
		Name:             keyName,
		FeatureVersionID: featureVersionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *Service) IsVariationPropertyNameTaken(ctx context.Context, name string) (bool, error) {
	_, err := s.queries.GetVariationPropertyIDByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *Service) IsVariationPropertyValueTaken(ctx context.Context, variationPropertyID uint, value string) (bool, error) {
	_, err := s.queries.GetVariationPropertyValueIDByValue(ctx, db.GetVariationPropertyValueIDByValueParams{
		VariationPropertyID: variationPropertyID,
		Value:               value,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (s *Service) DoesVariationExist(ctx context.Context, keyID uint, variation map[uint]string) (uint, error) {
	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, variation)
	if err != nil {
		return 0, err
	}

	valueID, err := s.queries.GetVariationValueIDByVariationContextID(ctx, db.GetVariationValueIDByVariationContextIDParams{
		VariationContextID: variationContextID,
		KeyID:              keyID,
		ChangesetID:        s.currentUserAccessor.GetUser(ctx).ChangesetID,
	})

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	return valueID, nil
}

func (s *Service) CanAddValueInternal(ctx context.Context, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow, variation map[uint]string) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, variation) < constants.PermissionEditor {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to add a value to this key")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return err
	}

	err = variationHierarchy.ValidateIDVariation(serviceVersion.ServiceTypeID, variation)
	if err != nil {
		return err
	}

	valueID, err := s.DoesVariationExist(ctx, key.ID, variation)
	if err != nil {
		return err
	}

	if valueID != 0 {
		return core.NewServiceError(core.ErrorCodeDuplicateVariation, "Value with this variation already exists")
	}

	return nil
}

func (s *Service) CanAddValue(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint, variation map[uint]string) error {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return err
	}

	return s.CanAddValueInternal(ctx, serviceVersion, featureVersion, key, variation)
}

func (s *Service) CanEditValueInternal(ctx context.Context, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow, value db.VariationValue, variation map[uint]string) error {
	previousVariation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
	if err != nil {
		return err
	}

	if len(previousVariation) == 0 && len(variation) != 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot change default value variation")
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, previousVariation) < constants.PermissionEditor {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to edit this value")
	}

	if user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, variation) < constants.PermissionEditor {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to save value with this variation")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return err
	}

	err = variationHierarchy.ValidateIDVariation(serviceVersion.ServiceTypeID, variation)
	if err != nil {
		return err
	}

	existingValueWithVariation, err := s.DoesVariationExist(ctx, key.ID, variation)
	if err != nil {
		return err
	}

	if existingValueWithVariation != 0 && existingValueWithVariation != value.ID {
		return core.NewServiceError(core.ErrorCodeDuplicateVariation, "Value with this variation already exists")
	}

	return nil
}

func (s *Service) CanEditValue(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint, valueID uint, variation map[uint]string) error {
	serviceVersion, featureVersion, key, value, err := s.coreService.GetVariationValue(ctx, serviceVersionID, featureVersionID, keyID, valueID)
	if err != nil {
		return err
	}

	return s.CanEditValueInternal(ctx, serviceVersion, featureVersion, key, value, variation)
}

func (s *Service) IsUsernameTaken(ctx context.Context, username string) (bool, error) {
	_, err := s.queries.GetUserByName(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
