package key

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/ptr"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	unitOfWorkRunner          db.UnitOfWorkRunner
	variationContextService   *variation.ContextService
	queries                   *db.Queries
	changesetService          *changeset.Service
	currentUserAccessor       *auth.CurrentUserAccessor
	validator                 *validator.Validator
	coreService               *core.Service
	valueValidatorService     *validation.ValueValidatorService
	variationHierarchyService *variation.HierarchyService
	validationService         *validation.Service
}

func NewService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	variationContextService *variation.ContextService,
	queries *db.Queries,
	changesetService *changeset.Service,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *validator.Validator,
	coreService *core.Service,
	valueValidatorService *validation.ValueValidatorService,
	variationHierarchyService *variation.HierarchyService,
	validationService *validation.Service,
) *Service {
	return &Service{
		unitOfWorkRunner:          unitOfWorkRunner,
		variationContextService:   variationContextService,
		queries:                   queries,
		changesetService:          changesetService,
		currentUserAccessor:       currentUserAccessor,
		validator:                 validator,
		coreService:               coreService,
		valueValidatorService:     valueValidatorService,
		variationHierarchyService: variationHierarchyService,
		validationService:         validationService,
	}
}

type ValueTypeName = string

type KeyItemDto struct {
	ID            uint             `json:"id" validate:"required"`
	Name          string           `json:"name" validate:"required"`
	Description   string           `json:"description" validate:"required"`
	ValueTypeName string           `json:"valueTypeName" validate:"required"`
	ValueType     db.ValueTypeKind `json:"valueType" validate:"required"`
	ValueTypeID   uint             `json:"valueTypeId" validate:"required"`
}

type KeyDto struct {
	KeyItemDto
	CanEdit    bool                      `json:"canEdit" validate:"required"`
	Validators []validation.ValidatorDto `json:"validators" validate:"required"`
}

func (s *Service) GetKey(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) (KeyDto, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return KeyDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	validators, err := s.valueValidatorService.GetValueValidators(ctx, &key.ID, &key.ValueTypeID)
	if err != nil {
		return KeyDto{}, err
	}

	return KeyDto{
		KeyItemDto: KeyItemDto{
			ID:            key.ID,
			Name:          key.Name,
			Description:   ptr.From(key.Description),
			ValueTypeName: key.ValueTypeName,
			ValueType:     key.ValueTypeKind,
			ValueTypeID:   key.ValueTypeID,
		},
		CanEdit:    user.GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) >= constants.PermissionAdmin,
		Validators: validators,
	}, nil
}

func (s *Service) GetFeatureKeys(ctx context.Context, serviceVersionID uint, featureVersionID uint) ([]KeyItemDto, error) {
	_, _, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	keys, err := s.queries.GetKeysForFeatureVersion(ctx, db.GetKeysForFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]KeyItemDto, len(keys))
	for i, key := range keys {
		result[i] = KeyItemDto{
			ID:            key.ID,
			Name:          key.Name,
			Description:   ptr.From(key.Description),
			ValueTypeName: key.ValueTypeName,
			ValueType:     key.ValueTypeKind,
		}
	}

	return result, nil
}

type AppliedKeyDto struct {
	Name string `json:"name" validate:"required"`
}

func (s *Service) GetAppliedKeys(ctx context.Context, featureVersionID *uint, featureID *uint) ([]AppliedKeyDto, error) {
	if featureVersionID == nil && featureID == nil {
		return nil, core.NewServiceError(core.ErrorCodeInvalidOperation, "Either feature version ID or feature ID must be provided")
	}

	if featureID != nil {
		_, err := s.coreService.GetFeature(ctx, *featureID)
		if err != nil {
			return nil, err
		}
	}

	if featureVersionID != nil {
		_, err := s.coreService.GetFeatureVersionWithoutLink(ctx, *featureVersionID)
		if err != nil {
			return nil, err
		}
	}

	keys, err := s.queries.GetAppliedKeys(ctx, db.GetAppliedKeysParams{
		FeatureVersionID: featureVersionID,
		FeatureID:        featureID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]AppliedKeyDto, len(keys))
	for i, key := range keys {
		result[i] = AppliedKeyDto{
			Name: key,
		}
	}

	return result, nil
}

type CreateKeyParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	Name             string
	Description      string
	DefaultValue     string
	ValueTypeID      uint
	Validators       []validation.ValidatorDto
}

func (s *Service) validateValidators(vc *validator.Context, validators []validation.ValidatorDto) *validator.Context {
	vc.
		Validate(validators, "Validators").Required()

	for i, validator := range validators {
		prefix := fmt.Sprintf("Validators[%d]", i)
		vc.
			Validate(validator.ValidatorType, fmt.Sprintf("%s.ValidatorType", prefix)).Required().
			Validate(validator.ErrorText, fmt.Sprintf("%s.ErrorText", prefix)).MaxLength(100)

		pvc := vc.Validate(validator.Parameter, fmt.Sprintf("%s.Parameter", prefix))

		parameterType := s.valueValidatorService.GetValidatorParameterType(validator.ValidatorType)
		switch parameterType {
		case "none":
			pvc.MaxLength(0)
		case "integer":
			pvc.Required().MaxLength(10).ValidInteger()
		case "float":
			pvc.Required().MaxLength(50).ValidFloat()
		case "regex":
			pvc.Required().MaxLength(500).ValidRegex()
		case "json_schema":
			pvc.Required().MaxLength(10000).ValidJsonSchema()
		}
	}

	return vc
}

func (s *Service) validateCreateKey(ctx context.Context, data CreateKeyParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForFeature(serviceVersion.ServiceID, featureVersion.FeatureID) < constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create keys for this feature")
	}

	vc := s.validator.
		Validate(data.Name, "Name").Required().MaxLength(100).Regex(`^[a-zA-Z_][\w_]*$`).
		Validate(data.ValueTypeID, "Value Type ID").Min(1).
		Validate(data.Description, "Description").MaxLength(core.DefaultDescriptionMaxLength)

	s.validateValidators(vc, data.Validators)

	valueTypeValidators, err := s.valueValidatorService.GetValueValidators(ctx, nil, &data.ValueTypeID)
	if err != nil {
		return err
	}

	validatorFunc, err := s.valueValidatorService.CreateValueValidatorFunc(slices.Concat(valueTypeValidators, data.Validators))
	if err != nil {
		return err
	}

	vc.Validate(data.DefaultValue, "Default Value").Func(validatorFunc)

	err = vc.Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsKeyNameTaken(ctx, featureVersion.ID, data.Name); err != nil {
		return err
	} else if taken {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Key name is already taken")
	}

	return nil
}

func (s *Service) CreateKey(ctx context.Context, params CreateKeyParams) (uint, error) {
	serviceVersion, featureVersion, err := s.coreService.GetFeatureVersion(ctx, params.ServiceVersionID, params.FeatureVersionID)
	if err != nil {
		return 0, err
	}

	err = s.validateCreateKey(ctx, params, serviceVersion, featureVersion)
	if err != nil {
		return 0, err
	}

	var keyID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		keyID, err = tx.CreateKey(ctx, db.CreateKeyParams{
			Name:             params.Name,
			Description:      ptr.To(params.Description, ptr.NilIfZero()),
			ValueTypeID:      params.ValueTypeID,
			FeatureVersionID: params.FeatureVersionID,
		})
		if err != nil {
			return err
		}

		for _, validator := range params.Validators {
			_, err = tx.CreateValueValidatorForKey(ctx, db.CreateValueValidatorForKeyParams{
				KeyID:         &keyID,
				ValidatorType: validator.ValidatorType,
				Parameter:     ptr.To(validator.Parameter, ptr.NilIfZero()),
				ErrorText:     ptr.To(validator.ErrorText, ptr.NilIfZero()),
			})
			if err != nil {
				return err
			}
		}

		err = tx.AddCreateKeyChange(ctx, db.AddCreateKeyChangeParams{
			ChangesetID:      changesetID,
			KeyID:            keyID,
			FeatureVersionID: params.FeatureVersionID,
			ServiceVersionID: params.ServiceVersionID,
		})
		if err != nil {
			return err
		}

		defaultVariationContextID, err := s.variationContextService.GetVariationContextID(ctx, map[uint]string{})
		if err != nil {
			return err
		}

		variationValueID, err := tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
			KeyID:              keyID,
			Data:               params.DefaultValue,
			VariationContextID: defaultVariationContextID,
		})
		if err != nil {
			return err
		}

		err = tx.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         changesetID,
			NewVariationValueID: variationValueID,
			FeatureVersionID:    params.FeatureVersionID,
			KeyID:               keyID,
			ServiceVersionID:    params.ServiceVersionID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	return keyID, err
}

type UpdateKeyParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	KeyID            uint
	Description      string
	Validators       []validation.ValidatorDto
}

func (s *Service) validateUpdateKey(ctx context.Context, data UpdateKeyParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow, hasValidatorsChanges bool) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if user.GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) < constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to update keys for this feature")
	}

	vc := s.validator.
		Validate(data.Description, "Description").MaxLength(core.DefaultDescriptionMaxLength)

	if hasValidatorsChanges {
		s.validateValidators(vc, data.Validators)

		if key.CreatedInChangesetID != user.ChangesetID {
			changesCount, err := s.queries.GetRelatedKeyChangesCount(ctx, db.GetRelatedKeyChangesCountParams{
				KeyID:       data.KeyID,
				ChangesetID: user.ChangesetID,
			})
			if err != nil {
				return err
			}

			if changesCount > 0 {
				return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Your current changeset contains %d changes related to this key. Please apply or discard them before updating validators.", changesCount))
			}
		}

		variationValues, err := s.queries.GetVariationValuesForKey(ctx, db.GetVariationValuesForKeyParams{
			KeyID:       data.KeyID,
			ChangesetID: user.ChangesetID,
		})
		if err != nil {
			return err
		}

		variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
		if err != nil {
			return err
		}

		valueTypeValidators, err := s.valueValidatorService.GetValueValidators(ctx, nil, &key.ValueTypeID)
		if err != nil {
			return err
		}

		validatorFunc, err := s.valueValidatorService.CreateValueValidatorFunc(slices.Concat(valueTypeValidators, data.Validators))
		if err != nil {
			return err
		}

		for _, variationValue := range variationValues {
			variationContextValues, err := s.variationContextService.GetVariationContextValues(ctx, variationValue.VariationContextID)
			if err != nil {
				return err
			}

			valueNameBuilder := strings.Builder{}
			valueNameBuilder.WriteString("Value for variation: ")
			if len(variationContextValues) > 0 {
				for propertyID, propertyValue := range variationContextValues {
					property, err := variationHierarchy.GetProperty(propertyID)
					if err != nil {
						return err
					}

					valueNameBuilder.WriteString(fmt.Sprintf("%s: %s, ", property.Name, propertyValue))
				}
			} else {
				valueNameBuilder.WriteString("Default")
			}

			vc.Validate(variationValue.Data, valueNameBuilder.String()).Func(validatorFunc)
		}
	}

	err := vc.Error(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateKey(ctx context.Context, params UpdateKeyParams) error {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID)
	if err != nil {
		return err
	}

	hasValidatorsChanges := false

	if params.Validators != nil {
		validators, err := s.queries.GetValueValidators(ctx, db.GetValueValidatorsParams{
			KeyID: &params.KeyID,
		})
		if err != nil {
			return err
		}

		hasValidatorsChanges = slices.CompareFunc(validators, params.Validators, func(a db.ValueValidator, b validation.ValidatorDto) int {
			if a.ValidatorType != b.ValidatorType || ptr.From(a.Parameter) != b.Parameter || ptr.From(a.ErrorText) != b.ErrorText {
				return 1
			}

			return 0
		}) != 0
	}

	err = s.validateUpdateKey(ctx, params, serviceVersion, featureVersion, key, hasValidatorsChanges)
	if err != nil {
		return err
	}

	validatorsUpdatedAt := key.ValidatorsUpdatedAt
	if hasValidatorsChanges {
		validatorsUpdatedAt = time.Now()
	}

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err = tx.UpdateKey(ctx, db.UpdateKeyParams{
			KeyID:               params.KeyID,
			Description:         ptr.To(params.Description, ptr.NilIfZero()),
			ValidatorsUpdatedAt: validatorsUpdatedAt,
		}); err != nil {
			return err
		}

		if hasValidatorsChanges {
			err = tx.DeleteValueValidatorsForKey(ctx, key.ID)
			if err != nil {
				return err
			}

			for _, validator := range params.Validators {
				_, err = tx.CreateValueValidatorForKey(ctx, db.CreateValueValidatorForKeyParams{
					KeyID:         &params.KeyID,
					ValidatorType: validator.ValidatorType,
					Parameter:     ptr.To(validator.Parameter, ptr.NilIfZero()),
					ErrorText:     ptr.To(validator.ErrorText, ptr.NilIfZero()),
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return err
}

func (s *Service) validateDeleteKey(ctx context.Context, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForService(serviceVersion.ServiceID) < constants.PermissionAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to delete keys for this service")
	}

	if key.ValidFrom != nil {
		changesCount, err := s.queries.GetRelatedKeyChangesCount(ctx, db.GetRelatedKeyChangesCountParams{
			KeyID:       key.ID,
			ChangesetID: user.ChangesetID,
		})
		if err != nil {
			return err
		}

		if changesCount > 0 {
			return core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Your current changeset contains %d changes related to this key. Please apply or discard them before deleting.", changesCount))
		}

		if featureVersion.LinkedToPublishedServiceVersion {
			return core.NewServiceError(core.ErrorCodePermissionDenied, "You cannot delete a key for a feature version that is linked to a published service version")
		}
	}

	return nil
}

func (s *Service) DeleteKey(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) error {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return err
	}

	err = s.validateDeleteKey(ctx, serviceVersion, featureVersion, key)
	if err != nil {
		return err
	}

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		change, err := tx.GetChangeForKey(ctx, db.GetChangeForKeyParams{
			ChangesetID: changesetID,
			KeyID:       key.ID,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		if change.ID == 0 {
			if err = tx.AddDeleteKeyChange(ctx, db.AddDeleteKeyChangeParams{
				ChangesetID:      changesetID,
				KeyID:            key.ID,
				FeatureVersionID: featureVersionID,
				ServiceVersionID: serviceVersionID,
			}); err != nil {
				return err
			}
		} else {
			if err = tx.DeleteKey(ctx, key.ID); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}
