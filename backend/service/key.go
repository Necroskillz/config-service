package service

import (
	"context"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type KeyService struct {
	unitOfWorkRunner        db.UnitOfWorkRunner
	variationContextService *VariationContextService
	queries                 *db.Queries
	changesetService        *ChangesetService
	currentUserAccessor     *auth.CurrentUserAccessor
	validator               *Validator
	coreService             *CoreService
}

func NewKeyService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	variationContextService *VariationContextService,
	queries *db.Queries,
	changesetService *ChangesetService,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
	coreService *CoreService,
) *KeyService {
	return &KeyService{
		unitOfWorkRunner:        unitOfWorkRunner,
		variationContextService: variationContextService,
		queries:                 queries,
		changesetService:        changesetService,
		currentUserAccessor:     currentUserAccessor,
		validator:               validator,
		coreService:             coreService,
	}
}

type EditorTypes = string

const (
	EditorTypeText    EditorTypes = "text"
	EditorTypeBoolean EditorTypes = "boolean"
	EditorTypeInteger EditorTypes = "integer"
	EditorTypeDecimal EditorTypes = "decimal"
	EditorTypeJSON    EditorTypes = "json"
)

type KeyDto struct {
	ID          uint        `json:"id" validate:"required"`
	Name        string      `json:"name" validate:"required"`
	Description *string     `json:"description"`
	CanEdit     bool        `json:"canEdit" validate:"required"`
	ValueType   string      `json:"valueType" validate:"required"`
	Editor      EditorTypes `json:"editor" validate:"required"`
}

func (s *KeyService) GetKey(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) (KeyDto, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return KeyDto{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	return KeyDto{
		ID:          key.ID,
		Name:        key.Name,
		Description: key.Description,
		ValueType:   key.ValueTypeName,
		Editor:      key.ValueTypeEditor,
		CanEdit:     user.GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) >= constants.PermissionAdmin,
	}, nil
}

func (s *KeyService) GetFeatureKeys(ctx context.Context, serviceVersionID uint, featureVersionID uint) ([]KeyDto, error) {
	serviceVersion, featureVersion, err := s.coreService.GetFeatureVersion(ctx, serviceVersionID, featureVersionID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	keys, err := s.queries.GetActiveKeysForFeatureVersion(ctx, db.GetActiveKeysForFeatureVersionParams{
		FeatureVersionID: featureVersionID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]KeyDto, len(keys))
	for i, key := range keys {
		result[i] = KeyDto{
			ID:          key.ID,
			Name:        key.Name,
			Description: key.Description,
			ValueType:   key.ValueTypeName,
			Editor:      key.ValueTypeEditor,
			CanEdit:     user.GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) >= constants.PermissionAdmin,
		}
	}

	return result, nil
}

func (s *KeyService) GetValueTypes(ctx context.Context) ([]db.ValueType, error) {
	return s.queries.GetValueTypes(ctx)
}

type CreateKeyParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	Name             string
	Description      string
	DefaultValue     string
	ValueTypeID      uint
}

func (s *KeyService) validateCreateKey(ctx context.Context, data CreateKeyParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow) error {
	err := s.validator.
		Validate(data.Name, "Name").Required().KeyNameNotTaken(data.FeatureVersionID).
		Validate(data.ValueTypeID, "Value Type ID").Min(1).
		Validate(data.Description, "Description").Func(optionalDescriptionValidatorFunc).
		Error(ctx)

	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForFeature(serviceVersion.ServiceID, featureVersion.FeatureID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to create keys for this feature")
	}

	return nil
}

func (s *KeyService) CreateKey(ctx context.Context, params CreateKeyParams) (uint, error) {
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
			Description:      &params.Description,
			ValueTypeID:      params.ValueTypeID,
			FeatureVersionID: params.FeatureVersionID,
		})
		if err != nil {
			return err
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

		defaultVariationContextID, err := s.variationContextService.GetVariationContextID(ctx, []uint{})
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
}

func (s *KeyService) validateUpdateKey(ctx context.Context, data UpdateKeyParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow) error {
	err := s.validator.
		Validate(data.Description, "Description").Func(optionalDescriptionValidatorFunc).
		Error(ctx)

	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) < constants.PermissionAdmin {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to update keys for this feature")
	}

	return nil
}

func (s *KeyService) UpdateKey(ctx context.Context, params UpdateKeyParams) error {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID)
	if err != nil {
		return err
	}

	err = s.validateUpdateKey(ctx, params, serviceVersion, featureVersion, key)
	if err != nil {
		return err
	}

	err = s.queries.UpdateKey(ctx, db.UpdateKeyParams{
		KeyID:       params.KeyID,
		Description: &params.Description,
	})
	if err != nil {
		return err
	}

	return nil
}
