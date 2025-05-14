package service

import (
	"context"

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
}

func NewVariationPropertyService(queries *db.Queries, variationHierarchyService *VariationHierarchyService, validator *Validator, validationService *ValidationService, currentUserAccessor *auth.CurrentUserAccessor) *VariationPropertyService {
	return &VariationPropertyService{
		queries:                   queries,
		variationHierarchyService: variationHierarchyService,
		validator:                 validator,
		validationService:         validationService,
		currentUserAccessor:       currentUserAccessor,
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
	ID       uint                        `json:"id" validate:"required"`
	Value    string                      `json:"value" validate:"required"`
	Children []VariationPropertyValueDto `json:"children" validate:"required"`
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

func (s *VariationPropertyService) makeVariationPropertyValueDto(value *VariationHierarchyValue) VariationPropertyValueDto {
	dto := VariationPropertyValueDto{
		ID:       value.ID,
		Value:    value.Value,
		Children: make([]VariationPropertyValueDto, len(value.Children)),
	}

	for i, child := range value.Children {
		dto.Children[i] = s.makeVariationPropertyValueDto(child)
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

	dto := VariationPropertyWithValuesDto{
		VariationPropertyDto: NewVariationPropertyDto(property),
		Values:               make([]VariationPropertyValueDto, len(property.Values)),
	}

	for i, value := range property.Values {
		dto.Values[i] = s.makeVariationPropertyValueDto(value)
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

func (s *VariationPropertyService) validateCreateVariationPropertyValue(ctx context.Context, data CreateVariationPropertyValueParams) error {
	user := s.currentUserAccessor.GetUser(ctx)

	if !user.IsGlobalAdmin {
		return NewServiceError(ErrorCodeInvalidOperation, "You are not authorized to create variation property values")
	}

	_, err := s.queries.GetVariationProperty(ctx, data.PropertyID)

	if err != nil {
		return NewDbError(err, "VariationProperty")
	}

	err = s.validator.Validate(data.Value, "Value").Required().MaxLength(20).Regex(`^[a-zA-Z0-9\-_]+$`).
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
	err := s.validateCreateVariationPropertyValue(ctx, params)

	if err != nil {
		return 0, err
	}

	variationPropertyValueID, err := s.queries.CreateVariationPropertyValue(ctx, db.CreateVariationPropertyValueParams{
		Value:               params.Value,
		ParentID:            ptr.To(params.ParentID, ptr.WithNilIfZero()),
		VariationPropertyID: params.PropertyID,
	})

	if err != nil {
		return 0, err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return variationPropertyValueID, nil
}
