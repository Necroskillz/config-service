package variationproperty

import (
	"context"
	"slices"
	"strings"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/ptr"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	queries                   *db.Queries
	variationHierarchyService *variation.HierarchyService
	validator                 *validator.Validator
	validationService         *validation.Service
	currentUserAccessor       *auth.CurrentUserAccessor
	unitOfWorkRunner          db.UnitOfWorkRunner
}

func NewService(queries *db.Queries, variationHierarchyService *variation.HierarchyService, validator *validator.Validator, validationService *validation.Service, currentUserAccessor *auth.CurrentUserAccessor, unitOfWorkRunner db.UnitOfWorkRunner) *Service {
	return &Service{
		queries:                   queries,
		variationHierarchyService: variationHierarchyService,
		validator:                 validator,
		validationService:         validationService,
		currentUserAccessor:       currentUserAccessor,
		unitOfWorkRunner:          unitOfWorkRunner,
	}
}

type VariationPropertyItemDto struct {
	ID          uint   `json:"id" validate:"required"`
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName" validate:"required"`
}

func NewVariationPropertyItemDto(property *variation.HierarchyProperty) VariationPropertyItemDto {
	return VariationPropertyItemDto{
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

type VariationPropertyDto struct {
	VariationPropertyItemDto
	UsageCount int                         `json:"usageCount" validate:"required"`
	Values     []VariationPropertyValueDto `json:"values" validate:"required"`
}

func (s *Service) GetVariationProperties(ctx context.Context) ([]VariationPropertyItemDto, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())

	if err != nil {
		return nil, err
	}

	variationProperties := variationHierarchy.GetAllProperties()

	result := make([]VariationPropertyItemDto, len(variationProperties))

	for i, property := range variationProperties {
		result[i] = NewVariationPropertyItemDto(property)
	}

	slices.SortFunc(result, func(a, b VariationPropertyItemDto) int {
		return strings.Compare(a.Name, b.Name)
	})

	return result, nil
}

type FlatVariationPropertyValueDto struct {
	ID    uint   `json:"id" validate:"required"`
	Value string `json:"value" validate:"required"`
	Depth int    `json:"depth" validate:"required"`
}

type ServiceTypeVariationPropertyDto struct {
	VariationPropertyItemDto
	Values []FlatVariationPropertyValueDto `json:"values" validate:"required"`
}

func (s *Service) makeFlatVariationPropertyValues(indent int, values []*variation.HierarchyValue) []FlatVariationPropertyValueDto {
	flatValues := []FlatVariationPropertyValueDto{}

	for _, value := range values {
		if value.Archived {
			continue
		}

		flatValues = append(flatValues, FlatVariationPropertyValueDto{
			ID:    value.ID,
			Value: value.Value,
			Depth: value.Depth,
		})

		if len(value.Children) > 0 {
			flatValues = append(flatValues, s.makeFlatVariationPropertyValues(indent+1, value.Children)...)
		}
	}

	return flatValues
}

func (s *Service) getFlatValues(property *variation.HierarchyProperty) []FlatVariationPropertyValueDto {
	values := []FlatVariationPropertyValueDto{
		{
			Value: "any",
			Depth: 0,
		},
	}

	values = append(values, s.makeFlatVariationPropertyValues(0, property.Values)...)

	return values
}

func (s *Service) GetVariationPropertiesForServiceType(ctx context.Context, serviceTypeID uint) ([]ServiceTypeVariationPropertyDto, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	properties, err := variationHierarchy.GetProperties(serviceTypeID)
	if err != nil {
		return nil, err
	}

	response := []ServiceTypeVariationPropertyDto{}

	for _, property := range properties {
		response = append(response, ServiceTypeVariationPropertyDto{
			VariationPropertyItemDto: NewVariationPropertyItemDto(property),
			Values:                   s.getFlatValues(property),
		})
	}

	return response, nil
}

func (s *Service) makeVariationPropertyValueDto(value *variation.HierarchyValue, usageMap map[uint]int) VariationPropertyValueDto {
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

func (s *Service) GetVariationProperty(ctx context.Context, id uint) (VariationPropertyDto, error) {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())

	if err != nil {
		return VariationPropertyDto{}, err
	}

	property, err := variationHierarchy.GetProperty(id)

	if err != nil {
		return VariationPropertyDto{}, err
	}

	propertyValuesUsage, err := s.queries.GetVariationPropertyValuesUsage(ctx, id)
	if err != nil {
		return VariationPropertyDto{}, err
	}

	valueUsageMap := make(map[uint]int)
	for _, usage := range propertyValuesUsage {
		valueUsageMap[usage.ID] = usage.UsageCount
	}

	propertyUsage, err := s.queries.GetVariationPropertyUsage(ctx, id)
	if err != nil {
		return VariationPropertyDto{}, err
	}

	dto := VariationPropertyDto{
		VariationPropertyItemDto: NewVariationPropertyItemDto(property),
		UsageCount:               propertyUsage,
		Values:                   make([]VariationPropertyValueDto, len(property.Values)),
	}

	for i, value := range property.Values {
		dto.Values[i] = s.makeVariationPropertyValueDto(value, valueUsageMap)
	}

	return dto, nil
}

type CreateVariationPropertyParams struct {
	Name        string
	DisplayName string
}

func (s *Service) validateVariationProperty(ctx context.Context, data CreateVariationPropertyParams) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to create variation properties")
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
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property name is already taken")
	}

	return nil
}

func (s *Service) CreateVariationProperty(ctx context.Context, params CreateVariationPropertyParams) (uint, error) {
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

func (s *Service) validateUpdateVariationProperty(ctx context.Context, data UpdateVariationPropertyParams) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to update variation properties")
	}

	err := s.validator.Validate(data.DisplayName, "Display Name").MaxLength(20).Regex(`^[a-zA-Z\-]+$`).
		Error(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) UpdateVariationProperty(ctx context.Context, id uint, params UpdateVariationPropertyParams) error {
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

func (s *Service) validateDeleteVariationProperty(ctx context.Context, id uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to delete variation properties")
	}

	_, err := s.queries.GetVariationProperty(ctx, id)
	if err != nil {
		return core.NewDbError(err, "VariationProperty")
	}

	usage, err := s.queries.GetVariationPropertyUsage(ctx, id)
	if err != nil {
		return err
	}

	if usage > 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot delete variation property with values that are in use")
	}

	return nil
}

func (s *Service) DeleteVariationProperty(ctx context.Context, id uint) error {
	err := s.validateDeleteVariationProperty(ctx, id)
	if err != nil {
		return err
	}

	err = s.queries.DeleteVariationProperty(ctx, id)
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

func (s *Service) validateCreateVariationPropertyValue(ctx context.Context, data CreateVariationPropertyValueParams, variationHierarchy *variation.Hierarchy) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to create variation property values")
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
			return core.NewServiceError(core.ErrorCodeInvalidOperation, "Parent value belongs to a different property")
		}

		if parent.Archived {
			return core.NewServiceError(core.ErrorCodeInvalidOperation, "Parent value is archived")
		}
	}

	err := s.validator.Validate(data.Value, "Value").Required().MaxLength(20).Regex(`^[\w\-_\.]+$`).
		Error(ctx)

	if err != nil {
		return err
	}

	if taken, err := s.validationService.IsVariationPropertyValueTaken(ctx, data.PropertyID, data.Value); err != nil {
		return err
	} else if taken {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property value is already taken")
	}

	return nil
}

func (s *Service) CreateVariationPropertyValue(ctx context.Context, params CreateVariationPropertyValueParams) (uint, error) {
	variationHierachy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())
	if err != nil {
		return 0, err
	}

	err = s.validateCreateVariationPropertyValue(ctx, params, variationHierachy)
	if err != nil {
		return 0, err
	}

	values, err := variationHierachy.GetValuesWithSameParent(params.PropertyID, variation.ByParentID(params.ParentID))
	if err != nil {
		return 0, err
	}

	index := slices.IndexFunc(values, func(value *variation.HierarchyValue) bool {
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

func (s *Service) validateUpdateVariationPropertyValueOrder(ctx context.Context, params UpdateVariationPropertyValueOrderParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to update variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())
	if err != nil {
		return err
	}

	if _, err = variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	valueGroup, err := variationHierarchy.GetValuesWithSameParent(params.PropertyID, variation.ByValueID(params.ValueID))
	if err != nil {
		return err
	}

	return s.validator.Validate(params.Order, "Order").Min(1).Max(len(valueGroup)).Error(ctx)
}

func (s *Service) UpdateVariationPropertyValueOrder(ctx context.Context, params UpdateVariationPropertyValueOrderParams) error {
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

func (s *Service) validateDeleteVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams, variationHierarchy *variation.Hierarchy) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to archive variation property values")
	}

	if _, err := variationHierarchy.GetProperty(params.PropertyID); err != nil {
		return err
	}

	value, err := variationHierarchy.GetValue(params.ValueID)
	if err != nil {
		return err
	}

	if value.Archived {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property value is already archived")
	}

	usageCounts, err := s.queries.GetVariationPropertyValuesUsage(ctx, params.PropertyID)
	if err != nil {
		return err
	}

	index := slices.IndexFunc(usageCounts, func(u db.GetVariationPropertyValuesUsageRow) bool {
		return u.ID == params.ValueID && u.UsageCount > 0
	})

	if index != -1 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot delete variation property value in use")
	}

	if len(value.Children) > 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot delete variation property value with children")
	}

	return nil
}

func (s *Service) DeleteVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())
	if err != nil {
		return err
	}

	err = s.validateDeleteVariationPropertyValue(ctx, params, variationHierarchy)
	if err != nil {
		return err
	}

	values, err := variationHierarchy.GetValuesWithSameParent(params.PropertyID, variation.ByValueID(params.ValueID))
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

func (s *Service) validateArchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to archive variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())
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
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property value is already archived")
	}

	for _, child := range value.Children {
		if !child.Archived {
			return core.NewServiceError(core.ErrorCodeInvalidOperation, "Cannot archive variation property value with unarchived children")
		}
	}

	return nil
}

func (s *Service) ArchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
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

func (s *Service) validateUnarchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to unarchive variation property values")
	}

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx, variation.WithForceRefresh())
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
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property value is not archived")
	}

	return nil
}

func (s *Service) UnarchiveVariationPropertyValue(ctx context.Context, params VariationPropertyValueParams) error {
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
