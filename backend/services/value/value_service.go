package value

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	unitOfWorkRunner          db.UnitOfWorkRunner
	variationContextService   *variation.ContextService
	variationHierarchyService *variation.HierarchyService
	queries                   *db.Queries
	changesetService          *changeset.Service
	currentUserAccessor       *auth.CurrentUserAccessor
	validator                 *validator.Validator
	coreService               *core.Service
	validationService         *validation.Service
	valueValidatorService     *validation.ValueValidatorService
}

func NewService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	variationContextService *variation.ContextService,
	variationHierarchyService *variation.HierarchyService,
	queries *db.Queries,
	changesetService *changeset.Service,
	currentUserAccessor *auth.CurrentUserAccessor,
	validator *validator.Validator,
	coreService *core.Service,
	validationService *validation.Service,
	valueValidatorService *validation.ValueValidatorService,
) *Service {
	return &Service{
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

type VariationValueDto struct {
	ID        uint            `json:"id" validate:"required"`
	Data      string          `json:"data" validate:"required"`
	Variation map[uint]string `json:"variation" validate:"required"`
	CanEdit   bool            `json:"canEdit" validate:"required"`
	Rank      int             `json:"rank" validate:"required"`
	Order     []int           `json:"order" validate:"required"`
}

func (s *Service) GetKeyValues(ctx context.Context, serviceVersionID uint, featureVersionID uint, keyID uint) ([]VariationValueDto, error) {
	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return nil, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	values, err := s.queries.GetVariationValuesForKey(ctx, db.GetVariationValuesForKeyParams{
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

	variationValues := make([]VariationValueDto, len(values))
	for i, value := range values {
		variation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
		if err != nil {
			return nil, err
		}

		order, err := variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, variation)
		if err != nil {
			return nil, err
		}

		rank, err := variationHierarchy.GetRank(serviceVersion.ServiceTypeID, variation)
		if err != nil {
			return nil, err
		}

		variationValues[i] = VariationValueDto{
			ID:        value.ID,
			Data:      value.Data,
			Variation: variation,
			CanEdit:   user.GetPermissionForValue(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID, variation) >= constants.PermissionEditor,
			Rank:      rank,
			Order:     order,
		}
	}

	slices.SortFunc(variationValues, func(a, b VariationValueDto) int {
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

func (s *Service) valueDataValidator(ctx context.Context, valueTypeID uint, keyID uint) (validator.ValidatorFunc, error) {
	valueValidators, err := s.valueValidatorService.GetValueValidators(ctx, &keyID, &valueTypeID)
	if err != nil {
		return nil, err
	}

	validatorFunc, err := s.valueValidatorService.CreateValueValidatorFunc(valueValidators)
	if err != nil {
		return nil, err
	}

	return validatorFunc, nil
}

func (s *Service) validateCreateValue(ctx context.Context, data CreateValueParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow) error {
	err := s.validationService.CanAddValueInternal(ctx, serviceVersion, featureVersion, key, data.Variation)
	if err != nil {
		return err
	}

	validatorFunc, err := s.valueDataValidator(ctx, key.ValueTypeID, key.ID)
	if err != nil {
		return err
	}

	return s.validator.
		Validate(data.Data, "Data").Func(validatorFunc).
		Error(ctx)
}

func (s *Service) CreateValue(ctx context.Context, data CreateValueParams) (NewValueInfo, error) {
	user := s.currentUserAccessor.GetUser(ctx)

	serviceVersion, featureVersion, key, err := s.coreService.GetKey(ctx, data.ServiceVersionID, data.FeatureVersionID, data.KeyID)
	if err != nil {
		return NewValueInfo{}, err
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return NewValueInfo{}, err
	}

	if err := s.validateCreateValue(ctx, data, serviceVersion, featureVersion, key); err != nil {
		return NewValueInfo{}, err
	}

	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, data.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	existingDeleteChange, err := s.queries.GetDeleteChangeForVariationContextID(ctx, db.GetDeleteChangeForVariationContextIDParams{
		ChangesetID:        user.ChangesetID,
		VariationContextID: variationContextID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return NewValueInfo{}, err
	}

	var variationValueID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		if existingDeleteChange.ID != 0 {
			if err = tx.DeleteChange(ctx, existingDeleteChange.ID); err != nil {
				return err
			}

			if existingDeleteChange.VariationValueData != data.Data {
				variationValueID, err = tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
					KeyID:              data.KeyID,
					VariationContextID: variationContextID,
					Data:               data.Data,
				})
				if err != nil {
					return err
				}

				if err = tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
					ChangesetID:         changesetID,
					ServiceVersionID:    data.ServiceVersionID,
					FeatureVersionID:    data.FeatureVersionID,
					KeyID:               data.KeyID,
					OldVariationValueID: existingDeleteChange.VariationValueID,
					NewVariationValueID: variationValueID,
				}); err != nil {
					return err
				}
			}
		} else {
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
		}

		return nil
	})

	if err != nil {
		return NewValueInfo{}, err
	}

	order, err := variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, data.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	return NewValueInfo{ID: variationValueID, Order: order}, nil
}

type DeleteValueParams struct {
	ServiceVersionID uint
	FeatureVersionID uint
	KeyID            uint
	ValueID          uint
}

func (s *Service) DeleteValue(ctx context.Context, params DeleteValueParams) error {
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
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to delete this value")
	}

	if len(variation) == 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot delete default value")
	}

	return s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
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
			if err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
				ChangesetID:         changesetID,
				FeatureVersionID:    params.FeatureVersionID,
				KeyID:               params.KeyID,
				ServiceVersionID:    params.ServiceVersionID,
				OldVariationValueID: params.ValueID,
			}); err != nil {
				return err
			}
		} else {
			if err = tx.DeleteVariationValue(ctx, *variationValueChange.NewVariationValueID); err != nil {
				return err
			}

			if variationValueChange.Type == db.ChangesetChangeTypeUpdate {
				if err = tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
					ChangesetID:         changesetID,
					FeatureVersionID:    params.FeatureVersionID,
					KeyID:               params.KeyID,
					ServiceVersionID:    params.ServiceVersionID,
					OldVariationValueID: *variationValueChange.OldVariationValueID,
				}); err != nil {
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

func (s *Service) validateUpdateValue(ctx context.Context, data UpdateValueParams, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.GetKeyRow, value db.VariationValue) error {
	err := s.validationService.CanEditValueInternal(ctx, serviceVersion, featureVersion, key, value, data.Variation)
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

type UpdateValueState struct {
	ServiceVersion       db.GetServiceVersionRow
	FeatureVersion       db.GetFeatureVersionRow
	Key                  db.GetKeyRow
	Value                db.VariationValue
	VariationContextID   uint
	Data                 string
	ExistingChange       *db.GetChangeForVariationValueRow
	ExistingDeleteChange *db.GetDeleteChangeForVariationContextIDRow
	ChangesetID          uint
}

type UpdateStrategy interface {
	CanHandle(state *UpdateValueState) bool
	Execute(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error)
}

type NoExistingChangeStrategy struct{}

func (s *NoExistingChangeStrategy) CanHandle(state *UpdateValueState) bool {
	return state.ExistingChange == nil
}

func (s *NoExistingChangeStrategy) Execute(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if state.ExistingDeleteChange != nil {
		return s.handleWithSameVariationAlreadyHavingDeletedChange(ctx, tx, state)
	}

	return s.handleDefault(ctx, tx, state)
}

func (s *NoExistingChangeStrategy) handleWithSameVariationAlreadyHavingDeletedChange(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if err := tx.DeleteChange(ctx, state.ExistingDeleteChange.ID); err != nil {
		return 0, err
	}

	if err := tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
		ChangesetID:         state.ChangesetID,
		ServiceVersionID:    state.ServiceVersion.ID,
		FeatureVersionID:    state.FeatureVersion.ID,
		KeyID:               state.Key.ID,
		OldVariationValueID: state.Value.ID,
	}); err != nil {
		return 0, err
	}

	if state.ExistingDeleteChange.VariationValueData != state.Data {
		variationValueID, err := tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
			KeyID:              state.Key.ID,
			VariationContextID: state.VariationContextID,
			Data:               state.Data,
		})
		if err != nil {
			return 0, err
		}

		if err := tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			OldVariationValueID: state.ExistingDeleteChange.VariationValueID,
			NewVariationValueID: variationValueID,
		}); err != nil {
			return 0, err
		}

		return variationValueID, nil
	}

	return state.Value.ID, nil
}

func (s *NoExistingChangeStrategy) handleDefault(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	variationValueID, err := tx.CreateVariationValue(ctx, db.CreateVariationValueParams{
		KeyID:              state.Key.ID,
		VariationContextID: state.VariationContextID,
		Data:               state.Data,
	})
	if err != nil {
		return 0, err
	}

	if state.Value.VariationContextID == state.VariationContextID {
		if err := tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			OldVariationValueID: state.Value.ID,
			NewVariationValueID: variationValueID,
		}); err != nil {
			return 0, err
		}
	} else {
		if err := tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			OldVariationValueID: state.Value.ID,
		}); err != nil {
			return 0, err
		}

		if err := tx.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			NewVariationValueID: variationValueID,
		}); err != nil {
			return 0, err
		}
	}

	return variationValueID, nil
}

type ExistingUpdateChangeStrategy struct{}

func (s *ExistingUpdateChangeStrategy) CanHandle(state *UpdateValueState) bool {
	return state.ExistingChange != nil && state.ExistingChange.Type == db.ChangesetChangeTypeUpdate
}

func (s *ExistingUpdateChangeStrategy) Execute(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if state.ExistingDeleteChange != nil {
		return s.handleWithSameVariationAlreadyHavingDeletedChange(ctx, tx, state)
	}

	return s.handleDefault(ctx, tx, state)
}

func (s *ExistingUpdateChangeStrategy) handleWithSameVariationAlreadyHavingDeletedChange(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if err := tx.DeleteChange(ctx, state.ExistingDeleteChange.ID); err != nil {
		return 0, err
	}

	if err := tx.DeleteChange(ctx, state.ExistingChange.ID); err != nil {
		return 0, err
	}

	if state.ExistingDeleteChange.VariationValueData != state.Data {
		if err := tx.UpdateVariationValue(ctx, db.UpdateVariationValueParams{
			VariationValueID:   state.Value.ID,
			VariationContextID: state.VariationContextID,
			Data:               state.Data,
		}); err != nil {
			return 0, err
		}

		if err := tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			NewVariationValueID: state.Value.ID,
			OldVariationValueID: state.ExistingDeleteChange.VariationValueID,
		}); err != nil {
			return 0, err
		}
	}

	return state.Value.ID, nil
}

func (s *ExistingUpdateChangeStrategy) handleDefault(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if err := tx.UpdateVariationValue(ctx, db.UpdateVariationValueParams{
		VariationValueID:   state.Value.ID,
		VariationContextID: state.VariationContextID,
		Data:               state.Data,
	}); err != nil {
		return 0, err
	}

	if state.Value.VariationContextID != state.VariationContextID {
		if err := tx.DeleteChange(ctx, state.ExistingChange.ID); err != nil {
			return 0, err
		}

		if err := tx.AddDeleteVariationValueChange(ctx, db.AddDeleteVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			OldVariationValueID: *state.ExistingChange.OldVariationValueID,
		}); err != nil {
			return 0, err
		}

		if err := tx.AddCreateVariationValueChange(ctx, db.AddCreateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			NewVariationValueID: state.Value.ID,
		}); err != nil {
			return 0, err
		}
	}

	return state.Value.ID, nil
}

type ExistingCreateChangeStrategy struct{}

func (s *ExistingCreateChangeStrategy) CanHandle(state *UpdateValueState) bool {
	return state.ExistingChange != nil && state.ExistingChange.Type == db.ChangesetChangeTypeCreate
}

func (s *ExistingCreateChangeStrategy) Execute(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if state.ExistingDeleteChange != nil {
		return s.handleWithSameVariationAlreadyHavingDeletedChange(ctx, tx, state)
	}
	return s.handleDefault(ctx, tx, state)
}

func (s *ExistingCreateChangeStrategy) handleWithSameVariationAlreadyHavingDeletedChange(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if err := tx.DeleteChange(ctx, state.ExistingDeleteChange.ID); err != nil {
		return 0, err
	}

	if err := tx.DeleteChange(ctx, state.ExistingChange.ID); err != nil {
		return 0, err
	}

	if state.ExistingDeleteChange.VariationValueData != state.Data {
		if err := tx.UpdateVariationValue(ctx, db.UpdateVariationValueParams{
			VariationValueID:   state.Value.ID,
			VariationContextID: state.VariationContextID,
			Data:               state.Data,
		}); err != nil {
			return 0, err
		}

		if err := tx.AddUpdateVariationValueChange(ctx, db.AddUpdateVariationValueChangeParams{
			ChangesetID:         state.ChangesetID,
			ServiceVersionID:    state.ServiceVersion.ID,
			FeatureVersionID:    state.FeatureVersion.ID,
			KeyID:               state.Key.ID,
			OldVariationValueID: state.ExistingDeleteChange.VariationValueID,
			NewVariationValueID: state.Value.ID,
		}); err != nil {
			return 0, err
		}
	} else {
		if err := tx.DeleteVariationValue(ctx, state.Value.ID); err != nil {
			return 0, err
		}

		return state.ExistingDeleteChange.VariationValueID, nil
	}

	return state.Value.ID, nil
}

func (s *ExistingCreateChangeStrategy) handleDefault(ctx context.Context, tx *db.Queries, state *UpdateValueState) (uint, error) {
	if err := tx.UpdateVariationValue(ctx, db.UpdateVariationValueParams{
		VariationValueID:   state.Value.ID,
		VariationContextID: state.VariationContextID,
		Data:               state.Data,
	}); err != nil {
		return 0, err
	}

	return state.Value.ID, nil
}

type UpdateValueStrategyResolver struct {
	strategies []UpdateStrategy
}

func NewUpdateValueStrategyResolver() *UpdateValueStrategyResolver {
	return &UpdateValueStrategyResolver{
		strategies: []UpdateStrategy{
			&ExistingUpdateChangeStrategy{},
			&ExistingCreateChangeStrategy{},
			&NoExistingChangeStrategy{},
		},
	}
}

func (r *UpdateValueStrategyResolver) ResolveStrategy(state *UpdateValueState) UpdateStrategy {
	for _, strategy := range r.strategies {
		if strategy.CanHandle(state) {
			return strategy
		}
	}
	return nil
}

func (s *Service) UpdateValue(ctx context.Context, params UpdateValueParams) (NewValueInfo, error) {
	serviceVersion, featureVersion, key, value, err := s.coreService.GetVariationValue(ctx, params.ServiceVersionID, params.FeatureVersionID, params.KeyID, params.ValueID)
	if err != nil {
		return NewValueInfo{}, err
	}

	user := s.currentUserAccessor.GetUser(ctx)

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return NewValueInfo{}, err
	}

	if err := s.validateUpdateValue(ctx, params, serviceVersion, featureVersion, key, value); err != nil {
		return NewValueInfo{}, err
	}

	variationContextID, err := s.variationContextService.GetVariationContextID(ctx, params.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	if value.VariationContextID == variationContextID && value.Data == params.Data {
		order, err := variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, params.Variation)
		if err != nil {
			return NewValueInfo{}, err
		}

		return NewValueInfo{ID: value.ID, Order: order}, nil
	}

	existingChange, err := s.queries.GetChangeForVariationValue(ctx, db.GetChangeForVariationValueParams{
		ChangesetID:      user.ChangesetID,
		VariationValueID: params.ValueID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return NewValueInfo{}, err
	}

	existingDeleteChange, err := s.queries.GetDeleteChangeForVariationContextID(ctx, db.GetDeleteChangeForVariationContextIDParams{
		ChangesetID:        user.ChangesetID,
		KeyID:              key.ID,
		VariationContextID: variationContextID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return NewValueInfo{}, err
	}

	state := &UpdateValueState{
		ServiceVersion:     serviceVersion,
		FeatureVersion:     featureVersion,
		Key:                key,
		Value:              value,
		VariationContextID: variationContextID,
		Data:               params.Data,
	}

	if existingChange.ID != 0 {
		state.ExistingChange = &existingChange
	}

	if existingDeleteChange.ID != 0 {
		state.ExistingDeleteChange = &existingDeleteChange
	}

	var variationValueID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		changesetID, err := s.changesetService.EnsureChangesetForUser(ctx)
		if err != nil {
			return err
		}

		state.ChangesetID = changesetID

		// Possible scenarios:
		//
		// 1. NO EXISTING CHANGE for this value:
		//    a) Target variation has existing delete change:
		//       - Remove the delete change, add delete change for current value
		//       - If new data differs from deleted value: create new value + add update change
		//       - If same data: just restore the deleted value
		//    b) No conflicts (default):
		//       - Create new value with new data/variation
		//       - Same variation: add update change (old -> new value)
		//       - Different variation: add delete change (old value) + create change (new value)
		//
		// 2. EXISTING UPDATE CHANGE for this value:
		//    a) Target variation has existing delete change:
		//       - Remove both changes, update current value with new data/variation
		//       - If new data differs from deleted value: add update change (deleted -> current)
		//       - If same data: deleted value is restored
		//    b) No conflicts (default):
		//       - Update current value with new data/variation
		//       - Same variation: keep existing update change
		//       - Different variation: replace update with delete (original) + create (current)
		//
		// 3. EXISTING CREATE CHANGE for this value:
		//    a) Target variation has existing delete change:
		//       - Remove both changes
		//       - If new data differs from deleted value: update current + add update change (deleted -> current)
		//       - If same data: delete current value and restore original deleted value
		//    b) No conflicts (default):
		//       - Just update the value (no change record needed, it's already a create)
		resolver := NewUpdateValueStrategyResolver()
		strategy := resolver.ResolveStrategy(state)
		if strategy == nil {
			return errors.New("no strategy found for update value state")
		}

		variationValueID, err = strategy.Execute(ctx, tx, state)
		return err
	})

	if err != nil {
		return NewValueInfo{}, err
	}

	order, err := variationHierarchy.GetOrder(serviceVersion.ServiceTypeID, params.Variation)
	if err != nil {
		return NewValueInfo{}, err
	}

	return NewValueInfo{ID: variationValueID, Order: order}, nil
}
