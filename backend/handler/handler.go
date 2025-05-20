package handler

import (
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/service"
)

type Handler struct {
	ServiceService            *service.ServiceService
	UserService               *service.UserService
	FeatureService            *service.FeatureService
	KeyService                *service.KeyService
	ChangesetService          *service.ChangesetService
	ValidationService         *service.ValidationService
	ValueService              *service.ValueService
	VariationHierarchyService *service.VariationHierarchyService
	CurrentUserAccessor       *auth.CurrentUserAccessor
	ValueTypeService          *service.ValueTypeService
	VariationPropertyService  *service.VariationPropertyService
	ServiceTypeService        *service.ServiceTypeService
}

func NewHandler(
	serviceService *service.ServiceService,
	userService *service.UserService,
	featureService *service.FeatureService,
	keyService *service.KeyService,
	changesetService *service.ChangesetService,
	validationService *service.ValidationService,
	valueService *service.ValueService,
	variationHierarchyService *service.VariationHierarchyService,
	currentUserAccessor *auth.CurrentUserAccessor,
	valueTypeService *service.ValueTypeService,
	variationPropertyService *service.VariationPropertyService,
	serviceTypeService *service.ServiceTypeService,
) *Handler {
	return &Handler{
		ServiceService:            serviceService,
		UserService:               userService,
		FeatureService:            featureService,
		KeyService:                keyService,
		ChangesetService:          changesetService,
		ValidationService:         validationService,
		ValueService:              valueService,
		VariationHierarchyService: variationHierarchyService,
		CurrentUserAccessor:       currentUserAccessor,
		ValueTypeService:          valueTypeService,
		VariationPropertyService:  variationPropertyService,
		ServiceTypeService:        serviceTypeService,
	}
}
