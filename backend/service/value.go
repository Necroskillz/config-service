package service

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
)

type ValueService struct {
	unitOfWorkRunner          db.UnitOfWorkRunner
	variationContextService   *VariationContextService
	variationHierarchyService *VariationHierarchyService
	queries                   *db.Queries
	changesetService          *ChangesetService
	currentUserAccessor       *auth.CurrentUserAccessor
	validator                 *Validator
	coreService               *CoreService
	validationService         *ValidationService
	valueValidatorService     *ValueValidatorService
}

func NewValueService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	variationContextService *VariationContextService,
	variationHierarchyService *VariationHierarchyService,
	queries *db.Queries,
	changesetService *ChangesetService,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *Validator,
	coreService *CoreService,
	validationService *ValidationService,
	valueValidatorService *ValueValidatorService,
) *ValueService {
	return &ValueService{
		unitOfWorkRunner:          unitOfWorkRunner,
		variationContextService:   variationContextService,
		variationHierarchyService: variationHierarchyService,
		queries:                   queries,
		changesetService:          changesetService,
		currentUserAccessor:       currentUserAccessor,
		validator:                 validator,
		coreService:               coreService,
		validationService:         validationService,
		valueValidatorService:     valueValidatorService,
	}
}

type VariationValue struct {
	ID        uint            `json:"id" validate:"required"`
	Data      string          `json:"data" validate:"required"`
	Variation map[uint]string `json:"variation" validate:"required"`
	CanEdit   bool            `json:"canEdit" validate:"required"`
	Rank      int             `json:"rank" validate:"required"`
	Order     []int           `json:"order" validate:"required"`
}

func (s *ValueService) GetKeyValues(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) ([]VariationValue, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	values, err := s.queries.GetActiveVariationValuesForKey(ctx, db.GetActiveVariationValuesForKeyParams{
		KeyID:       keyID,
		ChangesetID: user.ChangesetID,
	})
	if err != nil {
		return nil, err
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	variationValues := make([]VariationValue, len(values))
	for i, value := range values {
		variation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
		if err != nil {
			return nil, err
		}

		variationValues[i] = VariationValue{
			ID:        value.ID,
			Data:      value.Data,
			Variation: variation,
			CanEdit:   user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, variation) >= constants.PermissionEditor,
			Rank:      variationHierarchy.GetRank(serviceVersion.ServiceTypeID, variation),
			Order:     variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, variation),
		}
	}

	slices.SortFunc(variationValues, func(a, b VariationValue) int {
		for i := range a.Order {
			if a.Order[i] != b.Order[i] {
				return a.Order[i] - b.Order[i]
			}
		}

		return 0
	})

	return variationValues, nil
}

type NewValueInfo struct {
	ID    uint  `json:"id" validate:"required"`
	Order []int `json:"order" validate:"required"`
}

type CreateValueParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	KeyID            uint
	Data             string
	Variation        map[uint]string
}

func (s *ValueService) valueDataValidator(ctx context.Context, valueTypeID uint, keyID uint) (ValidatorFunc, error) {
	valueValidators, err := s.valueValidatorService.GetValueValidators(ctx, keyID, valueTypeID)
	if err != nil {
		return nil, err
	}

	validatorFunc, err := s.valueValidatorService.CreateValueValidatorFunc(valueValidators)
	if err != nil {
		return nil, err
	}

	return validatorFunc, nil
}

func (s *ValueService) validateCreateValue(ctx context.Context, data CreateValueParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow) error {
	err := s.validationService.canAddValueInternal(ctx, serviceVersion, featureVersion, key, data.Variation)
	if err != nil {
		return err
	}

	validatorFunc, err := s.valueDataValidator(ctx, key.ValueTypeID, key.ID)
	if err != nil {
		return err
	}

	return s.validator.
		Validate(data.Data, "Data").Func(validatorFunc).
		Validate(data.Variation, "Variation").Required().
		Error(ctx)
}

func (s *ValueService) CreateValue(ctx context.Context, data CreateValueParams) (NewValueInfo, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, data.ServiceVersionID, data.FeatureVersionID, data.KeyID)
	if err != nil {
		return NewValueInfo{}, err
	}

	if err := s.validateCreateValue(ctx, data, serviceVersion, featureVersion, key); err != nil {
		return NewValueInfo{}, err
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return NewValueInfo{}, err
	}

	variationIds, err := variationHierarchy.VariationMapToIds(serviceVersion.ServiceTypeID, data.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, variationIds)
	if err != nil {
		return NewValueInfo{}, err
	}

	var variationValueID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		variationValueID, err = tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
			KeyID:              data.KeyID,
			VariationContextID: variationContextID,
			Data:               data.Data,
		})
		if err != nil {
			return err
		}

		err = tx.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         changesetID,
			NewVariationValueID: variationValueID,
			FeatureVersionID:    data.FeatureVersionID,
			KeyID:               data.KeyID,
			ServiceVersionID:    data.ServiceVersionID,
		})
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return NewValueInfo{}, err
	}

	return NewValueInfo{ID: variationValueID, Order: variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, data.Variation)}, nil
}

type DeleteValueParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	KeyID            uint
	ValueID          uint
}

func (s *ValueService) DeleteValue(ctx context.Context, params DeleteValueParams) error {
	serviceVersion, featureVersion, key, value, err := s.coreService.GetVariationValue(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID, params.ValueID)
	if err != nil {
		return err
	}

	variation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
	if err != nil {
		return err
	}

	user := s.currentUserAccessor.GetUser(ctx)
	if user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, variation) < constants.PermissionEditor {
		return NewServiceError(ErrorCodePermissionDenied, "You are not authorized to delete this value")
	}

	if len(variation) == 0 {
		return NewServiceError(ErrorCodeInvalidOperation, "Cannot delete default value")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		// existing change
		variationValueChange, err := s.queries.GetChangeForVariationValue(ctx, db.GetChangeForVariationValueParams{
			ChangesetID:      user.ChangesetID,
			VariationValueID: params.ValueID,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		if variationValueChange.ID == 0 {
			err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
				ChangesetID:         changesetID,
				FeatureVersionID:    params.FeatureVersionID,
				KeyID:               params.KeyID,
				ServiceVersionID:    params.ServiceVersionID,
				OldVariationValueID: params.ValueID,
			})
			if err != nil {
				return err
			}
		} else {
			if variationValueChange.Type == db.ChangesetChangeTypeDelete {
				return NewServiceError(ErrorCodeInvalidOperation, "Cannot delete a already deleted value")
			}

			err = tx.DeleteChange(ctx, variationValueChange.ID)
			if err != nil {
				return err
			}

			if variationValueChange.Type == db.ChangesetChangeTypeCreate {
				err = tx.DeleteVariationValue(ctx, *variationValueChange.NewVariationValueID)
				if err != nil {
					return err
				}
			} else if variationValueChange.Type == db.ChangesetChangeTypeUpdate {
				err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
					ChangesetID:         changesetID,
					FeatureVersionID:    params.FeatureVersionID,
					KeyID:               params.KeyID,
					ServiceVersionID:    params.ServiceVersionID,
					OldVariationValueID: *variationValueChange.OldVariationValueID,
				})
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

type UpdateValueParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	KeyID            uint
	ValueID          uint
	Data             string
	Variation        map[uint]string
}

func (s *ValueService) validateUpdateValue(ctx context.Context, data UpdateValueParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow, value db.VariationValue) error {
	err := s.validationService.canEditValueInternal(ctx, serviceVersion, featureVersion, key, value, data.Variation)
	if err != nil {
		return err
	}

	validatorFunc, err := s.valueDataValidator(ctx, key.ValueTypeID, key.ID)
	if err != nil {
		return err
	}

	v := s.validator.
		Validate(data.Data, "Data").Func(validatorFunc).
		Validate(data.Variation, "Variation").Required()

	return v.Error(ctx)
}

func (s *ValueService) UpdateValue(ctx context.Context, params UpdateValueParams) (NewValueInfo, error) {
	serviceVersion, featureVersion, key, value, err := s.coreService.GetVariationValue(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID, params.ValueID)
	if err != nil {
		return NewValueInfo{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	if err := s.validateUpdateValue(ctx, params, serviceVersion, featureVersion, key, value); err != nil {
		return NewValueInfo{}, err
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return NewValueInfo{}, err
	}

	variationIds, err := variationHierarchy.VariationMapToIds(serviceVersion.ServiceTypeID, params.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, variationIds)
	if err != nil {
		return NewValueInfo{}, err
	}

	var variationValueID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		// existing change
		variationValueChange, err := s.queries.GetChangeForVariationValue(ctx, db.GetChangeForVariationValueParams{
			ChangesetID:      user.ChangesetID,
			VariationValueID: params.ValueID,
		})
		if err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}

		if variationValueChange.ID == 0 {
			variationValueID, err = tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
				KeyID:              params.KeyID,
				VariationContextID: variationContextID,
				Data:               params.Data,
			})
			if err != nil {
				return err
			}

			err = tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
				ChangesetID:         changesetID,
				FeatureVersionID:    params.FeatureVersionID,
				KeyID:               params.KeyID,
				ServiceVersionID:    params.ServiceVersionID,
				OldVariationValueID: params.ValueID,
				NewVariationValueID: variationValueID,
			})
			if err != nil {
				return err
			}
		} else {
			err = tx.UpdateVariationValue(ctx, db.UpdateVariationValueParams{
				VariationValueID:   params.ValueID,
				VariationContextID: variationContextID,
				Data:               params.Data,
			})
			if err != nil {
				return err
			}

			variationValueID = params.ValueID
		}

		return nil
	})

	if err != nil {
		return NewValueInfo{}, err
	}

	return NewValueInfo{ID: variationValueID, Order: variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, params.Variation)}, nil
}
