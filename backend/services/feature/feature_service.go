package feature

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	unitOfWorkRunner    db.UnitOfWorkRunner
	queries             *db.Queries
	changesetService    *changeset.Service
	currentUserAccessor *auth.CurrentUserAccessor
	validator           *validator.Validator
	coreService         *core.Service
	validationService   *validation.Service
}

func NewService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	queries *db.Queries,
	changesetService *changeset.Service,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *validator.Validator,
	coreService *core.Service,
	validationService *validation.Service,
) *Service {
	return &Service{
		unitOfWorkRunner:    unitOfWorkRunner,
		queries:             queries,
		changesetService:    changesetService,
		currentUserAccessor: currentUserAccessor,
		validator:           validator,
		coreService:         coreService,
		validationService:   validationService,
	}
}

type FeatureVersionDto struct {
	ID            uint   `json:"id" validate:"required"`
	Version       int    `json:"version" validate:"required"`
	Description   string `json:"description" validate:"required"`
	Name          string `json:"name" validate:"required"`
	CanEdit       bool   `json:"canEdit" validate:"required"`
	IsLastVersion bool   `json:"isLastVersion" validate:"required"`
}

func (s *Service) GetFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) (FeatureVersionDto, error) {
	serviceVersion, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return FeatureVersionDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return FeatureVersionDto{
		ID:            featureVersion.ID,
		Version:       featureVersion.Version,
		Description:   featureVersion.FeatureDescription,
		Name:          featureVersion.FeatureName,
		CanEdit:       user.GetPermissionForFeature(serviceVersion.ServiceID, featureVersion.FeatureID) >= constants.PermissionAdmin,
		IsLastVersion: featureVersion.LastVersion == featureVersion.Version,
	}, nil
}

type FeatureVersionItemDto struct {
	ID          uint   `json:"id" validate:"required"`
	Version     int    `json:"version" validate:"required"`
	Description string `json:"description" validate:"required"`
	Name        string `json:"name" validate:"required"`
	CanUnlink   bool   `json:"canUnlink" validate:"required"`
}

func (s *Service) GetServiceFeatures(ctx context.Context, serviceVersionID uint) ([]FeatureVersionItemDto, error) {
	user := s.currentUserAccessor.GetUser(ctx)
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return nil, err
	}

	featureVersions, err := s.queries.GetFeatureVersionsForServiceVersion(ctx, db.GetFeatureVersionsForServiceVersionParams{
		ServiceVersionID: serviceVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]FeatureVersionItemDto, len(featureVersions))
	for i, featureVersion := range featureVersions {
		result[i] = FeatureVersionItemDto{
			ID:          featureVersion.ID,
			Version:     featureVersion.Version,
			Description: featureVersion.FeatureDescription,
			Name:        featureVersion.FeatureName,
			CanUnlink:   !serviceVersion.Published || featureVersion.LinkedInChangesetID == user.ChangesetID,
		}
	}

	return result, nil
}

type FeatureVersionLinkDto struct {
	ServiceVersionID uint `json:"serviceVersionId" validate:"required"`
	FeatureVersionID uint `json:"featureVersionId" validate:"required"`
	Version          int  `json:"version" validate:"required"`
}

func (s *Service) GetVersionsOfFeatureForServiceVersion(ctx context.Context, featureVersionID uint, serviceVersionID uint) ([]FeatureVersionLinkDto, error) {
	_, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	featureVersions, err := s.queries.GetVersionsOfFeatureForServiceVersion(ctx, db.GetVersionsOfFeatureForServiceVersionParams{
		FeatureID:   featureVersion.FeatureID,
		ChangesetID: user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]FeatureVersionLinkDto, len(featureVersions))
	for i, featureVersion := range featureVersions {
		result[i] = FeatureVersionLinkDto{
			ServiceVersionID: featureVersion.ServiceVersionID,
			FeatureVersionID: featureVersion.ID,
			Version:          featureVersion.Version,
		}
	}

	return result, nil
}

func (s *Service) GetFeatureVersionsLinkableToServiceVersion(ctx context.Context, serviceVersionID uint) ([]FeatureVersionDto, error) {
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

func (s *Service) validateCreateFeature(ctx context.Context, data CreateFeatureParams, serviceVersion db.GetServiceVersionRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create features for this service")
	}

	err := s.validator.
		Validate(data.Name, "Name").Required().MaxLength(100).Regex(`^[\w\-_\.]+$`).
		Validate(data.Description, "Description").Required().MaxLength(core.DefaultDescriptionMaxLength).
		Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsFeatureNameTaken(ctx, data.Name); err != nil {
		return err
	} else if taken {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Feature name is already taken")
	}

	return nil
}

func (s *Service) CreateFeature(ctx context.Context, params CreateFeatureParams) (uint, error) {
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

func (s *Service) validateUpdateFeature(ctx context.Context, data UpdateFeatureParams, serviceVersion db.GetServiceVersionRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create features for this service")
	}

	return s.validator.
		Validate(data.Description, "Description").Required().MaxLength(core.DefaultDescriptionMaxLength).
		Error(ctx)
}

func (s *Service) UpdateFeature(ctx context.Context, params UpdateFeatureParams) error {
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

func (s *Service) validateUnlinkFeatureVersion(ctx context.Context, serviceVersion db.GetServiceVersionRow, link db.GetFeatureVersionServiceVersionLinkRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to unlink features for this service")
	}

	if serviceVersion.Published && link.CreatedInChangesetID != user.ChangesetID {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Features cannot be unlinked from a published service version")
	}

	return nil
}

func (s *Service) UnlinkFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) error {
	serviceVersion, _, link, err := s.coreService.GetFeatureVersionWithLink(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return err
	}

	if err := s.validateUnlinkFeatureVersion(ctx, serviceVersion, link); err != nil {
		return err
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
				FeatureVersionServiceVersionID: link.ID,
				ServiceVersionID:               serviceVersionID,
				FeatureVersionID:               featureVersionID,
			}); err != nil {
				return err
			}
		} else {
			if change.Type == "delete" {
				return core.NewServiceError(core.ErrorCodeInvalidOperation, "Feature version already has an unlink change in the current changeset")
			}

			if err = tx.DeleteFeatureVersionServiceVersion(ctx, link.ID); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *Service) validateLinkFeatureVersion(ctx context.Context, serviceVersion db.GetServiceVersionRow, featureVersionID uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to link features for this service")
	}

	featureVersion, err := s.queries.GetFeatureVersion(ctx, db.GetFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return err
	}

	if featureVersion.ServiceID != serviceVersion.ServiceID {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Unable to link feature version to a different service version")
	}

	linked, err := s.queries.IsFeatureLinkedToServiceVersion(ctx, db.IsFeatureLinkedToServiceVersionParams{
		FeatureID:        featureVersion.FeatureID,
		ServiceVersionID: serviceVersion.ID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return err
	}

	if linked {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Feature is already linked to this service version")
	}

	return nil
}

func (s *Service) LinkFeatureVersion(ctx context.Context, serviceVersionID uint, featureVersionID uint) error {
	serviceVersion, err := s.coreService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if err := s.validateLinkFeatureVersion(ctx, serviceVersion, featureVersionID); err != nil {
		return err
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

type FeatureVersionKeyDataValue struct {
	Data               string
	VariationContextID uint
}

type FeatureVersionKeyDataValidator struct {
	ValidatorType db.ValueValidatorType
	Parameter     *string
	ErrorText     *string
}

type FeatureVersionKeyData struct {
	Name        string
	Description *string
	ValueTypeID uint
	Validators  []FeatureVersionKeyDataValidator
	Values      []FeatureVersionKeyDataValue
}

func (s *Service) getFeatureVersionKeyData(ctx context.Context, featureVersionID uint) (map[uint]FeatureVersionKeyData, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	valuesData, err := s.queries.GetFeatureVersionValuesData(ctx, db.GetFeatureVersionValuesDataParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	validatorsData, err := s.queries.GetFeatureVersionValidatorData(ctx, db.GetFeatureVersionValidatorDataParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	keyMap := make(map[uint]FeatureVersionKeyData)
	for _, key := range valuesData {
		existingKey, ok := keyMap[key.KeyID]
		if !ok {
			keyMap[key.KeyID] = FeatureVersionKeyData{
				Name:        key.KeyName,
				Description: key.KeyDescription,
				ValueTypeID: key.KeyValueTypeID,
				Values: []FeatureVersionKeyDataValue{
					{
						Data:               key.Data,
						VariationContextID: key.VariationContextID,
					},
				},
				Validators: []FeatureVersionKeyDataValidator{},
			}
		} else {
			existingKey.Values = append(existingKey.Values, FeatureVersionKeyDataValue{
				Data:               key.Data,
				VariationContextID: key.VariationContextID,
			})

			keyMap[key.KeyID] = existingKey
		}
	}

	for _, validator := range validatorsData {
		existingKey := keyMap[validator.KeyID]

		existingKey.Validators = append(existingKey.Validators, FeatureVersionKeyDataValidator{
			ValidatorType: validator.ValidatorType,
			Parameter:     validator.Parameter,
			ErrorText:     validator.ErrorText,
		})

		keyMap[validator.KeyID] = existingKey
	}

	return keyMap, nil
}

type CreateFeatureVersionParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
}

func (s *Service) validateCreateFeatureVersion(ctx context.Context, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) != constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create feature versions for this service")
	}

	if serviceVersion.Published {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot create a feature version for a published service version")
	}

	if featureVersion.LastVersion != featureVersion.Version {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "New feature version can only be created from the latest version")
	}

	return nil
}

func (s *Service) CreateFeatureVersion(ctx context.Context, params CreateFeatureVersionParams) (uint, error) {
	serviceVersion, featureVersion, link, err := s.coreService.GetFeatureVersionWithLink(ctx, params.ServiceVersionID, params.FeatureVersionID)
	if err != nil {
		return 0, err
	}

	if err := s.validateCreateFeatureVersion(ctx, serviceVersion, featureVersion); err != nil {
		return 0, err
	}

	keyData, err := s.getFeatureVersionKeyData(ctx, featureVersion.ID)
	if err != nil {
		return 0, err
	}

	var newFeatureVersionID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		newFeatureVersionID, err = tx.CreateFeatureVersion(ctx, db.CreateFeatureVersionParams{
			FeatureID: featureVersion.FeatureID,
			Version:   featureVersion.LastVersion + 1,
		})
		if err != nil {
			return err
		}

		if err = tx.AddCreateFeatureVersionChange(ctx, db.AddCreateFeatureVersionChangeParams{
			ChangesetID:              changesetID,
			FeatureVersionID:         newFeatureVersionID,
			ServiceVersionID:         serviceVersion.ID,
			PreviousFeatureVersionID: &featureVersion.ID,
		}); err != nil {
			return err
		}

		if err = tx.AddDeleteFeatureVersionServiceVersionChange(ctx, db.AddDeleteFeatureVersionServiceVersionChangeParams{
			ChangesetID:                    changesetID,
			FeatureVersionServiceVersionID: link.ID,
			ServiceVersionID:               serviceVersion.ID,
			FeatureVersionID:               featureVersion.ID,
		}); err != nil {
			return err
		}

		linkID, err := tx.CreateFeatureVersionServiceVersion(ctx, db.CreateFeatureVersionServiceVersionParams{
			ServiceVersionID: serviceVersion.ID,
			FeatureVersionID: newFeatureVersionID,
		})
		if err != nil {
			return err
		}

		if err = tx.AddCreateFeatureVersionServiceVersionChange(ctx, db.AddCreateFeatureVersionServiceVersionChangeParams{
			ChangesetID:                    changesetID,
			FeatureVersionServiceVersionID: linkID,
			ServiceVersionID:               serviceVersion.ID,
			FeatureVersionID:               newFeatureVersionID,
		}); err != nil {
			return err
		}

		newKeys := []db.CreateKeysParams{}
		newVariationValues := []db.CreateVariationValuesParams{}
		newChanges := []db.AddChangesParams{}
		newValueValidators := []db.CreateValueValidatorsParams{}

		for _, key := range keyData {
			newKeys = append(newKeys, db.CreateKeysParams{
				FeatureVersionID: newFeatureVersionID,
				Name:             key.Name,
				Description:      key.Description,
				ValueTypeID:      key.ValueTypeID,
			})
		}

		if _, err := tx.CreateKeys(ctx, newKeys); err != nil {
			return err
		}

		createdKeys, err := tx.GetKeysForWipFeatureVersion(ctx, newFeatureVersionID)
		if err != nil {
			return err
		}

		keyMap := make(map[string]uint)
		for _, key := range createdKeys {
			keyMap[key.Name] = key.ID
		}

		for _, key := range keyData {
			keyID, ok := keyMap[key.Name]
			if !ok {
				return core.NewServiceError(core.ErrorCodeUnexpectedError, "Key that should have been created was not found")
			}

			for _, validator := range key.Validators {
				newValueValidators = append(newValueValidators, db.CreateValueValidatorsParams{
					KeyID:         &keyID,
					ValidatorType: validator.ValidatorType,
					Parameter:     validator.Parameter,
					ErrorText:     validator.ErrorText,
				})
			}

			newChanges = append(newChanges, db.AddChangesParams{
				ChangesetID:      changesetID,
				ServiceVersionID: serviceVersion.ID,
				FeatureVersionID: &newFeatureVersionID,
				KeyID:            &keyID,
				Type:             db.ChangesetChangeTypeCreate,
				Kind:             db.ChangesetChangeKindKey,
			})

			for _, value := range key.Values {
				newVariationValues = append(newVariationValues, db.CreateVariationValuesParams{
					KeyID:              keyID,
					VariationContextID: value.VariationContextID,
					Data:               value.Data,
				})
			}
		}

		if _, err := tx.CreateVariationValues(ctx, newVariationValues); err != nil {
			return err
		}

		if _, err := tx.CreateValueValidators(ctx, newValueValidators); err != nil {
			return err
		}

		createdVariationValues, err := tx.GetVariationValuesForWipFeatureVersion(ctx, newFeatureVersionID)
		if err != nil {
			return err
		}

		for _, variationValue := range createdVariationValues {
			newChanges = append(newChanges, db.AddChangesParams{
				ChangesetID:         changesetID,
				ServiceVersionID:    serviceVersion.ID,
				FeatureVersionID:    &newFeatureVersionID,
				KeyID:               &variationValue.KeyID,
				NewVariationValueID: &variationValue.ID,
				Type:                db.ChangesetChangeTypeCreate,
				Kind:                db.ChangesetChangeKindVariationValue,
			})
		}

		if _, err := tx.AddChanges(ctx, newChanges); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return newFeatureVersionID, nil
}
