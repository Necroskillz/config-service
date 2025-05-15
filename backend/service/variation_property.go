package service

import (
	"context"
	"slices"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/util/ptr"
)

type VariationPropertyService struct {
	queries                   *db.Queries
	variationHierarchyService *VariationHierarchyService
	validator                 *Validator
	validationService         *ValidationService
	currentUserAccessor       *auth.CurrentUserAccessor
	unitOfWorkRunner          db.UnitOfWorkRunner
}

func NewVariationPropertyService(queries *db.Queries, variationHierarchyService *VariationHierarchyService, validator *Validator, validationService *ValidationService, currentUserAccessor *auth.CurrentUserAccessor, unitOfWorkRunner db.UnitOfWorkRunner) *VariationPropertyService {
	return &VariationPropertyService{
		queries:                   queries,
		variationHierarchyService: variationHierarchyService,
		validator:                 validator,
		validationService:         validationService,
		currentUserAccessor:       currentUserAccessor,
		unitOfWorkRunner:          unitOfWorkRunner,
	}
}

type VariationPropertyDto struct {
	ID          uint   `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName" validate:"required"`
}

func NewVariationPropertyDto(property *VariationHierarchyProperty) VariationPropertyDto {
	return VariationPropertyDto{
		ID:          property.ID,
		Name:        property.Name,
		DisplayName: property.DisplayName,
	}
}

type VariationPropertyValueDto struct {
	ID         uint                        `json:"id" validate:"required"`
	Value      string                      `json:"value" validate:"required"`
	Children   []VariationPropertyValueDto `json:"children" validate:"required"`
	UsageCount int                         `json:"usageCount" validate:"required"`
	Archived   bool                        `json:"archived" validate:"required"`
}

type VariationPropertyWithValuesDto struct {
	VariationPropertyDto
	Values []VariationPropertyValueDto `json:"values" validate:"required"`
}

func (s *VariationPropertyService) GetVariationProperties(ctx context.Context) ([]VariationPropertyDto, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())

	if err != nil {
		return nil, err
	}

	variationProperties := variationHierarchy.GetAllProperties()

	result := make([]VariationPropertyDto, len(variationProperties))

	for i, property := range variationProperties {
		result[i] = NewVariationPropertyDto(property)
	}

	return result, nil
}

func (s *VariationPropertyService) makeVariationPropertyValueDto(value *VariationHierarchyValue, usageMap map[uint]int) VariationPropertyValueDto {
	dto := VariationPropertyValueDto{
		ID:         value.ID,
		Value:      value.Value,
		Archived:   value.Archived,
		UsageCount: usageMap[value.ID],
		Children:   make([]VariationPropertyValueDto, len(value.Children)),
	}

	for i, child := range value.Children {
		dto.Children[i] = s.makeVariationPropertyValueDto(child, usageMap)
	}

	return dto
}

func (s *VariationPropertyService) GetVariationProperty(ctx context.Context, id uint) (VariationPropertyWithValuesDto, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())

	if err != nil {
		return VariationPropertyWithValuesDto{}, err
	}

	property, err := variationHierarchy.GetProperty(id)

	if err != nil {
		return VariationPropertyWithValuesDto{}, err
	}

	propertyValuesUsage, err := s.queries.GetVariationPropertyValuesUsage(ctx, id)
	if err != nil {
		return VariationPropertyWithValuesDto{}, err
	}

	usageMap := make(map[uint]int)
	for _, usage := range propertyValuesUsage {
		usageMap[usage.ID] = usage.UsageCount
	}

	dto := VariationPropertyWithValuesDto{
		VariationPropertyDto: NewVariationPropertyDto(property),
		Values:               make([]VariationPropertyValueDto, len(property.Values)),
	}

	for i, value := range property.Values {
		dto.Values[i] = s.makeVariationPropertyValueDto(value, usageMap)
	}

	return dto, nil
}

type CreateVariationPropertyParams struct {
	Name        string
	DisplayName string
}

func (s *VariationPropertyService) validateVariationProperty(ctx context.Context, data CreateVariationPropertyParams) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to create variation properties")
	}

	err := s.validator.Validate(data.Name, "Name").Required().MaxLength(20).Regex(`^[a-z_\-]+$`).
		Validate(data.DisplayName, "Display Name").MaxLength(20).Regex(`^[a-zA-Z\-]+$`).
		Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsVariationPropertyNameTaken(ctx, data.Name); err != nil {
		return err
	} else if taken {
		return NewServiceError(ErrorCodeInvalidOperation, "Variation property name is already taken")
	}

	return nil
}

func (s *VariationPropertyService) CreateVariationProperty(ctx context.Context, params CreateVariationPropertyParams) (uint, error) {
	err := s.validateVariationProperty(ctx, params)

	if err != nil {
		return 0, err
	}

	displayName := params.DisplayName
	if displayName == "" {
		displayName = params.Name
	}

	variationPropertyID, err := s.queries.CreateVariationProperty(ctx, db.CreateVariationPropertyParams{
		Name:        params.Name,
		DisplayName: displayName,
	})

	if err != nil {
		return 0, err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return variationPropertyID, nil
}

type UpdateVariationPropertyParams struct {
	DisplayName string
}

func (s *VariationPropertyService) validateUpdateVariationProperty(ctx context.Context, data UpdateVariationPropertyParams) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to update variation properties")
	}

	err := s.validator.Validate(data.DisplayName, "Display Name").MaxLength(20).Regex(`^[a-zA-Z\-]+$`).
		Error(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *VariationPropertyService) UpdateVariationProperty(ctx context.Context, id uint, params UpdateVariationPropertyParams) error {
	err := s.validateUpdateVariationProperty(ctx, params)

	if err != nil {
		return err
	}

	err = s.queries.UpdateVariationProperty(ctx, db.UpdateVariationPropertyParams{
		ID:          id,
		DisplayName: params.DisplayName,
	})

	if err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

type CreateVariationPropertyValueParams struct {
	Value      string
	ParentID   uint
	PropertyID uint
}

func (s *VariationPropertyService) validateCreateVariationPropertyValue(ctx context.Context, data CreateVariationPropertyValueParams, variationHierarchy *VariationHierarchy) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to create variation property values")
	}

	if _, err := variationHierarchy.GetProperty(data.PropertyID); err != nil {
		return err
	}

	if data.ParentID != 0 {
		parent, err := variationHierarchy.GetValue(data.ParentID)
		if err != nil {
			return err
		}

		if parent.PropertyID != data.PropertyID {
			return NewServiceError(ErrorCodeInvalidOperation, "Parent value belongs to a different property")
		}

		if parent.Archived {
			return NewServiceError(ErrorCodeInvalidOperation, "Parent value is archived")
		}
	}

	err := s.validator.Validate(data.Value, "Value").Required().MaxLength(20).Regex(`^[\w\-_]+$`).
		Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsVariationPropertyValueTaken(ctx, data.PropertyID, data.Value); err != nil {
		return err
	} else if taken {
		return NewServiceError(ErrorCodeInvalidOperation, "Variation property value is already taken")
	}

	return nil
}

func (s *VariationPropertyService) CreateVariationPropertyValue(ctx context.Context, params CreateVariationPropertyValueParams) (uint, error) {
	variationHierachy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())
	if err != nil {
		return 0, err
	}

	err = s.validateCreateVariationPropertyValue(ctx, params, variationHierachy)
	if err != nil {
		return 0, err
	}

	values, err := variationHierachy.GetValuesWithSameParent(params.PropertyID, ByParentID(params.ParentID))
	if err != nil {
		return 0, err
	}

	index := slices.IndexFunc(values, func(value *VariationHierarchyValue) bool {
		return params.Value < value.Value
	})

	if index == -1 {
		index = len(values) + 1
	} else {
		index++
	}

	var valueID uint

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		valueID, err = tx.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
			Value:               params.Value,
			ParentID:            ptr.To(params.ParentID, ptr.NilIfZero()),
			VariationPropertyID: params.PropertyID,
		})
		if err != nil {
			return err
		}

		if err := tx.UpdateVariationPropertyValueOrder(ctx, db.UpdateVariationPropertyValueOrderParams{
			ID:          valueID,
			TargetIndex: index,
		}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return valueID, nil
}

type UpdateVariationPropertyValueOrderParams struct {
	PropertyID uint
	ValueID    uint
	Order      int
}

func (s *VariationPropertyService) validateUpdateVariationPropertyValueOrder(ctx context.Context, params UpdateVariationPropertyValueOrderParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to update variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())
	if err != nil {
		return err
	}

	if _, err = variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	valueGroup, err := variationHierarchy.GetValuesWithSameParent(params.PropertyID, ByValueID(params.ValueID))
	if err != nil {
		return err
	}

	return s.validator.Validate(params.Order, "Order").Min(1).Max(len(valueGroup)).Error(ctx)
}

func (s *VariationPropertyService) UpdateVariationPropertyValueOrder(ctx context.Context, params UpdateVariationPropertyValueOrderParams) error {
	err := s.validateUpdateVariationPropertyValueOrder(ctx, params)

	if err != nil {
		return err
	}

	if err = s.queries.UpdateVariationPropertyValueOrder(ctx, db.UpdateVariationPropertyValueOrderParams{
		ID:          params.ValueID,
		TargetIndex: params.Order,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

type VariationPropertyValueParams struct {
	PropertyID uint
	ValueID    uint
}

func (s *VariationPropertyService) validateDeleteVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams, variationHierarchy *VariationHierarchy) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to archive variation property values")
	}

	if _, err := variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	value, err := variationHierarchy.GetValue(params.ValueID)
	if err != nil {
		return err
	}

	if value.Archived {
		return NewServiceError(ErrorCodeInvalidOperation, "Variation property value is already archived")
	}

	usageCounts, err := s.queries.GetVariationPropertyValuesUsage(ctx, params.PropertyID)
	if err != nil {
		return err
	}

	index := slices.IndexFunc(usageCounts, func(u db.GetVariationPropertyValuesUsageRow) bool {
		return u.ID == params.ValueID && u.UsageCount > 0
	})

	if index != -1 {
		return NewServiceError(ErrorCodeInvalidOperation, "Cannot delete variation property value in use")
	}

	if len(value.Children) > 0 {
		return NewServiceError(ErrorCodeInvalidOperation, "Cannot delete variation property value with children")
	}

	return nil
}

func (s *VariationPropertyService) DeleteVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())
	if err != nil {
		return err
	}

	err = s.validateDeleteVariationPropertyValue(ctx, params, variationHierarchy)
	if err != nil {
		return err
	}

	values, err := variationHierarchy.GetValuesWithSameParent(params.PropertyID, ByValueID(params.ValueID))
	if err != nil {
		return err
	}

	err = s.unitOfWorkRunner.Run(ctx, func(tx *db.Queries) error {
		if err := tx.UpdateVariationPropertyValueOrder(ctx, db.UpdateVariationPropertyValueOrderParams{
			ID:          params.ValueID,
			TargetIndex: len(values),
		}); err != nil {
			return err
		}

		if err := tx.DeleteVariationPropertyValue(ctx, params.ValueID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

func (s *VariationPropertyService) validateArchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to archive variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())
	if err != nil {
		return err
	}

	if _, err := variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	value, err := variationHierarchy.GetValue(params.ValueID)
	if err != nil {
		return err
	}

	if value.Archived {
		return NewServiceError(ErrorCodeInvalidOperation, "Variation property value is already archived")
	}

	for _, child := range value.Children {
		if !child.Archived {
			return NewServiceError(ErrorCodeInvalidOperation, "Cannot archive variation property value with unarchived children")
		}
	}

	return nil
}

func (s *VariationPropertyService) ArchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	err := s.validateArchiveVariationPropertyValue(ctx, params)
	if err != nil {
		return err
	}

	if err = s.queries.SetVariationPropertyValueArchived(ctx, db.SetVariationPropertyValueArchivedParams{
		ID:       params.ValueID,
		Archived: true,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

func (s *VariationPropertyService) validateUnarchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to unarchive variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, WithForceRefresh())
	if err != nil {
		return err
	}

	if _, err := variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	value, err := variationHierarchy.GetValue(params.ValueID)
	if err != nil {
		return err
	}

	if !value.Archived {
		return NewServiceError(ErrorCodeInvalidOperation, "Variation property value is not archived")
	}

	return nil
}

func (s *VariationPropertyService) UnarchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	err := s.validateUnarchiveVariationPropertyValue(ctx, params)
	if err != nil {
		return err
	}

	if err = s.queries.SetVariationPropertyValueArchived(ctx, db.SetVariationPropertyValueArchivedParams{
		ID:       params.ValueID,
		Archived: false,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}
