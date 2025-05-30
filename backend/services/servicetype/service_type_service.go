package servicetype

import (
	"context"
	"slices"

	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/validator"
)

type Service struct {
	unitOfWorkRunner          db.UnitOfWorkRunner
	queries                   *db.Queries
	validator                 *validator.Validator
	validationService         *validation.Service
	currentUserAccessor       *auth.CurrentUserAccessor
	variationHierarchyService *variation.HierarchyService
}

func NewService(
	unitOfWorkRunner db.UnitOfWorkRunner,
	queries *db.Queries,
	validator *validator.Validator,
	validationService *validation.Service,
	currentUserAccessor *auth.CurrentUserAccessor,
	variationHierarchyService *variation.HierarchyService,
) *Service {
	return &Service{
		unitOfWorkRunner:          unitOfWorkRunner,
		queries:                   queries,
		validator:                 validator,
		validationService:         validationService,
		currentUserAccessor:       currentUserAccessor,
		variationHierarchyService: variationHierarchyService,
	}
}

type ServiceTypeItemDto struct {
	ID   uint   `json:"id" validate:"required"`
	Name string `json:"name" validate:"required"`
}

func (s *Service) GetServiceTypes(ctx context.Context) ([]ServiceTypeItemDto, error) {
	serviceTypes, err := s.queries.GetServiceTypes(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ServiceTypeItemDto, 0, len(serviceTypes))
	for _, serviceType := range serviceTypes {
		result = append(result, ServiceTypeItemDto{
			ID:   serviceType.ID,
			Name: serviceType.Name,
		})
	}

	return result, nil
}

type ServiceTypeVariationPropertyLinkDto struct {
	ID          uint   `json:"id" validate:"required"`
	PropertyID  uint   `json:"propertyId" validate:"required"`
	Name        string `json:"name" validate:"required"`
	DisplayName string `json:"displayName" validate:"required"`
	Priority    int    `json:"priority" validate:"required"`
	UsageCount  int    `json:"usageCount" validate:"required"`
}

type ServiceTypeDto struct {
	ServiceTypeItemDto
	UsageCount          int                                   `json:"usageCount" validate:"required"`
	VariationProperties []ServiceTypeVariationPropertyLinkDto `json:"properties" validate:"required"`
}

func (s *Service) GetServiceType(ctx context.Context, id uint) (ServiceTypeDto, error) {
	serviceType, err := s.queries.GetServiceType(ctx, id)
	if err != nil {
		return ServiceTypeDto{}, core.NewDbError(err, "ServiceType")
	}

	variationProperties, err := s.queries.GetServiceTypeVariationPropertyLinks(ctx, id)
	if err != nil {
		return ServiceTypeDto{}, err
	}

	dto := ServiceTypeDto{
		ServiceTypeItemDto: ServiceTypeItemDto{
			ID:   serviceType.ID,
			Name: serviceType.Name,
		},
		UsageCount:          serviceType.UsageCount,
		VariationProperties: make([]ServiceTypeVariationPropertyLinkDto, 0, len(variationProperties)),
	}

	for _, variationProperty := range variationProperties {
		dto.VariationProperties = append(dto.VariationProperties, ServiceTypeVariationPropertyLinkDto{
			ID:          variationProperty.ID,
			PropertyID:  variationProperty.PropertyID,
			Name:        variationProperty.Name,
			DisplayName: variationProperty.DisplayName,
			Priority:    variationProperty.Priority,
			UsageCount:  variationProperty.UsageCount,
		})
	}

	return dto, nil
}

type CreateServiceTypeParams struct {
	Name string `json:"name" validate:"required"`
}

func (s *Service) validateCreateServiceType(ctx context.Context, params CreateServiceTypeParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to create service types")
	}

	err := s.validator.Validate(params.Name, "Name").Required().MaxLength(50).Regex(`^[\w\-_\. ]+$`).Error(ctx)
	if err != nil {
		return err
	}

	taken, err := s.validationService.IsServiceTypeNameTaken(ctx, params.Name)
	if err != nil {
		return err
	}

	if taken {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Service type name is already taken")
	}

	return nil
}

func (s *Service) CreateServiceType(ctx context.Context, params CreateServiceTypeParams) (uint, error) {
	if err := s.validateCreateServiceType(ctx, params); err != nil {
		return 0, err
	}

	serviceTypeID, err := s.queries.CreateServiceType(ctx, params.Name)
	if err != nil {
		return 0, core.NewDbError(err, "ServiceType")
	}

	s.variationHierarchyService.ClearCache(ctx)

	return serviceTypeID, nil
}

type LinkVariationPropertyToServiceTypeParams struct {
	ServiceTypeID       uint
	VariationPropertyID uint
}

func (s *Service) validateLinkVariationPropertyToServiceType(ctx context.Context, params LinkVariationPropertyToServiceTypeParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to link variation properties to service types")
	}

	_, err := s.queries.GetServiceType(ctx, params.ServiceTypeID)
	if err != nil {
		return core.NewDbError(err, "ServiceType")
	}

	_, err = s.queries.GetVariationProperty(ctx, params.VariationPropertyID)
	if err != nil {
		return core.NewDbError(err, "VariationProperty")
	}

	linked, err := s.queries.IsVariationPropertyLinkedToServiceType(ctx, db.IsVariationPropertyLinkedToServiceTypeParams{
		ServiceTypeID:       params.ServiceTypeID,
		VariationPropertyID: params.VariationPropertyID,
	})
	if err != nil {
		return err
	}

	if linked {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property is already linked to this service type")
	}

	return nil
}

func (s *Service) LinkVariationPropertyToServiceType(ctx context.Context, params LinkVariationPropertyToServiceTypeParams) error {
	if err := s.validateLinkVariationPropertyToServiceType(ctx, params); err != nil {
		return err
	}

	if _, err := s.queries.CreateServiceTypeVariationPropertyLink(ctx, db.CreateServiceTypeVariationPropertyLinkParams{
		ServiceTypeID:       params.ServiceTypeID,
		VariationPropertyID: params.VariationPropertyID,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

func (s *Service) validateUnlinkVariationPropertyToServiceType(ctx context.Context, params LinkVariationPropertyToServiceTypeParams) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to unlink variation properties from service types")
	}

	if _, err := s.queries.GetServiceType(ctx, params.ServiceTypeID); err != nil {
		return core.NewDbError(err, "ServiceType")
	}

	if _, err := s.queries.GetVariationProperty(ctx, params.VariationPropertyID); err != nil {
		return core.NewDbError(err, "VariationProperty")
	}

	links, err := s.queries.GetServiceTypeVariationPropertyLinks(ctx, params.ServiceTypeID)
	if err != nil {
		return err
	}

	idx := slices.IndexFunc(links, func(link db.GetServiceTypeVariationPropertyLinksRow) bool {
		return link.PropertyID == params.VariationPropertyID
	})

	if idx == -1 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property is not linked to this service type")
	}

	if links[idx].UsageCount > 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property is already used in some services of this service type")
	}

	return nil
}

func (s *Service) UnlinkVariationPropertyToServiceType(ctx context.Context, params LinkVariationPropertyToServiceTypeParams) error {
	if err := s.validateUnlinkVariationPropertyToServiceType(ctx, params); err != nil {
		return err
	}

	if err := s.queries.DeleteServiceTypeVariationPropertyLink(ctx, db.DeleteServiceTypeVariationPropertyLinkParams{
		ServiceTypeID:       params.ServiceTypeID,
		VariationPropertyID: params.VariationPropertyID,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

type UpdateServiceTypeVariationPropertyPriorityParams struct {
	ServiceTypeID       uint
	VariationPropertyID uint
	Priority            int
}

func (s *Service) validateUpdateServiceTypeVariationPropertyPriority(ctx context.Context, params UpdateServiceTypeVariationPropertyPriorityParams) (uint, error) {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return 0, core.NewServiceError(core.ErrorCodeInvalidOperation, "You are not authorized to update variation property values")
	}

	if _, err := s.queries.GetServiceType(ctx, params.ServiceTypeID); err != nil {
		return 0, core.NewDbError(err, "ServiceType")
	}

	if _, err := s.queries.GetVariationProperty(ctx, params.VariationPropertyID); err != nil {
		return 0, core.NewDbError(err, "VariationProperty")
	}

	links, err := s.queries.GetServiceTypeVariationPropertyLinks(ctx, params.ServiceTypeID)
	if err != nil {
		return 0, err
	}

	if err := s.validator.Validate(params.Priority, "Priority").Min(1).Max(len(links)).Error(ctx); err != nil {
		return 0, err
	}

	idx := slices.IndexFunc(links, func(link db.GetServiceTypeVariationPropertyLinksRow) bool {
		return link.PropertyID == params.VariationPropertyID
	})

	if idx == -1 {
		return 0, core.NewServiceError(core.ErrorCodeInvalidOperation, "Variation property is not linked to this service type")
	}

	return links[idx].ID, nil
}

func (s *Service) UpdateServiceTypeVariationPropertyPriority(ctx context.Context, params UpdateServiceTypeVariationPropertyPriorityParams) error {
	linkID, err := s.validateUpdateServiceTypeVariationPropertyPriority(ctx, params)
	if err != nil {
		return err
	}

	if err = s.queries.UpdateServiceTypeVariationPropertyPriority(ctx, db.UpdateServiceTypeVariationPropertyPriorityParams{
		ID:             linkID,
		TargetPriority: params.Priority,
	}); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}

func (s *Service) validateDeleteServiceType(ctx context.Context, id uint) error {
	user := s.currentUserAccessor.GetUser(ctx)
	if !user.IsGlobalAdmin {
		return core.NewServiceError(core.ErrorCodePermissionDenied, "You are not authorized to delete service types")
	}

	serviceType, err := s.queries.GetServiceType(ctx, id)
	if err != nil {
		return core.NewDbError(err, "ServiceType")
	}

	if serviceType.UsageCount > 0 {
		return core.NewServiceError(core.ErrorCodeInvalidOperation, "Service type is already used in some services")
	}

	return nil
}

func (s *Service) DeleteServiceType(ctx context.Context, id uint) error {
	if err := s.validateDeleteServiceType(ctx, id); err != nil {
		return err
	}

	if err := s.queries.DeleteServiceType(ctx, id); err != nil {
		return err
	}

	s.variationHierarchyService.ClearCache(ctx)

	return nil
}
