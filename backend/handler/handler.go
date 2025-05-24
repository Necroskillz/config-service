package handler

import (
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/configuration"
	"github.com/necroskillz/config-service/services/feature"
	"github.com/necroskillz/config-service/services/key"
	"github.com/necroskillz/config-service/services/membership"
	"github.com/necroskillz/config-service/services/service"
	"github.com/necroskillz/config-service/services/servicetype"
	"github.com/necroskillz/config-service/services/validation"
	"github.com/necroskillz/config-service/services/value"
	"github.com/necroskillz/config-service/services/valuetype"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/services/variationproperty"
)

type Handler struct {
	ServiceService            *service.Service
	UserService               *membership.UserService
	FeatureService            *feature.Service
	KeyService                *key.Service
	ChangesetService          *changeset.Service
	ValidationService         *validation.Service
	ValueService              *value.Service
	VariationHierarchyService *variation.HierarchyService
	CurrentUserAccessor       *auth.CurrentUserAccessor
	ValueTypeService          *valuetype.Service
	VariationPropertyService  *variationproperty.Service
	ServiceTypeService        *servicetype.Service
	ConfigurationService      *configuration.Service
}

func NewHandler(
	serviceService *service.Service,
	userService *membership.UserService,
	featureService *feature.Service,
	keyService *key.Service,
	changesetService *changeset.Service,
	validationService *validation.Service,
	valueService *value.Service,
	variationHierarchyService *variation.HierarchyService,
	currentUserAccessor *auth.CurrentUserAccessor,
	valueTypeService *valuetype.Service,
	variationPropertyService *variationproperty.Service,
	serviceTypeService *servicetype.Service,
	configurationService *configuration.Service,
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
		ConfigurationService:      configurationService,
	}
}
