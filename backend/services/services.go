package services

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/configuration"
	"github.com/necroskillz/config-service/services/core"
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
	"github.com/necroskillz/config-service/util/validator"
)

type Services struct {
	ValueService              *value.Service
	ValueTypeService          *valuetype.Service
	ServiceTypeService        *servicetype.Service
	ServiceService            *service.Service
	AuthService               *membership.AuthService
	FeatureService            *feature.Service
	KeyService                *key.Service
	VariationPropertyService  *variationproperty.Service
	ConfigurationService      *configuration.Service
	UnitOfWorkRunner          db.UnitOfWorkRunner
	CurrentUserAccessor       *auth.CurrentUserAccessor
	Validator                 *validator.Validator
	ValueValidatorService     *validation.ValueValidatorService
	VariationHierarchyService *variation.HierarchyService
	VariationContextService   *variation.ContextService
	ValidationService         *validation.Service
	ChangesetService          *changeset.Service
	MembershipService         *membership.Service
}

func InitializeServices(dbpool *pgxpool.Pool, cache *ristretto.Cache[string, any]) *Services {
	queries := db.New(dbpool)
	currentUserAccessor := auth.NewCurrentUserAccessor()

	unitOfWorkRunner := db.NewPgxUnitOfWorkRunner(dbpool, queries)
	valueValidatorService := validation.NewValueValidatorService(queries)
	validator := validator.New()
	valueTypeService := valuetype.NewService(queries, valueValidatorService)
	coreService := core.NewService(queries, currentUserAccessor)
	variationHierarchyService := variation.NewHierarchyService(queries, cache)
	variationContextService := variation.NewContextService(queries, variationHierarchyService, unitOfWorkRunner, cache)
	validationService := validation.NewService(queries, variationContextService, variationHierarchyService, currentUserAccessor, coreService)
	serviceTypeService := servicetype.NewService(unitOfWorkRunner, queries, validator, validationService, currentUserAccessor, variationHierarchyService)
	changesetService := changeset.NewService(queries, variationContextService, unitOfWorkRunner, currentUserAccessor, validator)
	serviceService := service.NewService(queries, unitOfWorkRunner, changesetService, currentUserAccessor, validator, coreService, validationService)
	authService := membership.NewAuthService(queries, variationContextService, validationService, validator)
	featureService := feature.NewService(unitOfWorkRunner, queries, changesetService, currentUserAccessor, validator, coreService, validationService)
	keyService := key.NewService(unitOfWorkRunner, variationContextService, queries, changesetService, currentUserAccessor, validator, coreService, valueValidatorService, variationHierarchyService, validationService)
	valueService := value.NewService(unitOfWorkRunner, variationContextService, variationHierarchyService, queries, changesetService, currentUserAccessor, validator, coreService, validationService, valueValidatorService)
	variationPropertyService := variationproperty.NewService(queries, variationHierarchyService, validator, validationService, currentUserAccessor, unitOfWorkRunner)
	configurationService := configuration.NewService(queries, variationContextService, variationHierarchyService)
	membershipService := membership.NewService(queries, variationContextService, validationService, variationHierarchyService, validator, coreService)

	return &Services{
		ValueService:              valueService,
		ValueTypeService:          valueTypeService,
		ServiceTypeService:        serviceTypeService,
		ServiceService:            serviceService,
		AuthService:               authService,
		FeatureService:            featureService,
		KeyService:                keyService,
		VariationPropertyService:  variationPropertyService,
		ConfigurationService:      configurationService,
		UnitOfWorkRunner:          unitOfWorkRunner,
		CurrentUserAccessor:       currentUserAccessor,
		Validator:                 validator,
		ValueValidatorService:     valueValidatorService,
		VariationHierarchyService: variationHierarchyService,
		VariationContextService:   variationContextService,
		ValidationService:         validationService,
		ChangesetService:          changesetService,
		MembershipService:         membershipService,
	}
}
