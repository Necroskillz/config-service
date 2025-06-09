package testdata

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/hashicorp/go-metrics"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services"
	"github.com/necroskillz/config-service/services/changeset"
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
	"github.com/schollz/progressbar/v3"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Manager struct {
	DbPool  *pgxpool.Pool
	Cache   *ristretto.Cache[string, any]
	DB      *db.Queries
	Metrics *metrics.Metrics

	Registry     *Registry
	Rng          *Rng
	ChangeScopes []*ChangeScope
	ValueTypes   []valuetype.ValueTypeDto

	ServiceService            *service.Service
	MembershipService         *membership.Service
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
	AuthService               *membership.AuthService
}

func NewManager(dbpool *pgxpool.Pool, cache *ristretto.Cache[string, any], metrics *metrics.Metrics) *Manager {
	svc := services.InitializeServices(dbpool, cache)

	changeScopes := make([]*ChangeScope, 3)
	changeScopes[0] = &ChangeScope{
		Weight:      95,
		CreateValue: NewChangeAmount(1, 2),
		UpdateValue: NewChangeAmount(1, 3),
		DeleteValue: NewChangeAmount(0, 1),
	}

	changeScopes[1] = &ChangeScope{
		Weight:               4,
		CreateValue:          NewChangeAmount(10, 20),
		UpdateValue:          NewChangeAmount(5, 10),
		DeleteValue:          NewChangeAmount(0, 1),
		CreateKey:            NewChangeAmount(1, 2),
		CreateFeatureVersion: NewChangeAmount(0, 1),
	}

	changeScopes[2] = &ChangeScope{
		Weight:               1,
		CreateValue:          NewChangeAmount(30, 50),
		UpdateValue:          NewChangeAmount(10, 20),
		DeleteValue:          NewChangeAmount(1, 3),
		CreateKey:            NewChangeAmount(1, 2),
		DeleteKey:            NewChangeAmount(1, 2),
		CreateFeatureVersion: NewChangeAmount(1, 2),
		CreateServiceVersion: NewChangeAmount(0, 1),
		CreateFeature:        NewChangeAmount(0, 1),
		LinkFeature:          NewChangeAmount(0, 1),
		UnlinkFeature:        NewChangeAmount(0, 1),
	}

	return &Manager{
		DbPool:                    dbpool,
		Cache:                     cache,
		Metrics:                   metrics,
		DB:                        db.New(dbpool),
		ChangeScopes:              changeScopes,
		ServiceService:            svc.ServiceService,
		MembershipService:         svc.MembershipService,
		FeatureService:            svc.FeatureService,
		KeyService:                svc.KeyService,
		ChangesetService:          svc.ChangesetService,
		ValidationService:         svc.ValidationService,
		ValueService:              svc.ValueService,
		VariationHierarchyService: svc.VariationHierarchyService,
		CurrentUserAccessor:       svc.CurrentUserAccessor,
		ValueTypeService:          svc.ValueTypeService,
		VariationPropertyService:  svc.VariationPropertyService,
		ServiceTypeService:        svc.ServiceTypeService,
		AuthService:               svc.AuthService,
	}
}

type RunOptions struct {
	Seed       int64
	Iterations int
}

type RunOptionFunc func(*RunOptions)

func WithSeed(seed int64) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Seed = seed
	}
}

func WithIterationCount(count int) RunOptionFunc {
	return func(opts *RunOptions) {
		opts.Iterations = count
	}
}

func (m *Manager) Run(ctx context.Context, opts ...RunOptionFunc) error {
	runOpts := RunOptions{}
	for _, opt := range opts {
		opt(&runOpts)
	}
	m.Rng = NewRng(runOpts.Seed)
	m.Registry = NewRegistry(m.Rng)

	log.Printf("Manager starting with seed %d and %d iterations\n", runOpts.Seed, runOpts.Iterations)

	user, err := m.DB.GetUserByName(ctx, "admin")
	if err != nil {
		return err
	}

	adminCtx, err := m.createUserContext(ctx, user.ID)
	if err != nil {
		return err
	}

	err = m.createUsers(adminCtx)
	if err != nil {
		return err
	}

	err = m.createVariationHierarchy(adminCtx)
	if err != nil {
		return err
	}

	err = m.createServiceTypes(adminCtx)
	if err != nil {
		return err
	}

	valueTypes, err := m.ValueTypeService.GetValueTypes(ctx)
	if err != nil {
		return err
	}
	m.ValueTypes = valueTypes

	bar := progressbar.Default(20, "Creating base services")
	for i := 0; i < 20; i++ {
		err = m.createService(adminCtx)
		if err != nil {
			return err
		}
		bar.Add(1)
	}

	bar = progressbar.Default(50, "Creating base features")
	for i := 0; i < 50; i++ {
		service := m.Registry.GetRandomService()
		err = m.createFeature(adminCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
		bar.Add(1)
	}

	bar = progressbar.Default(500, "Creating base keys")
	for i := 0; i < 500; i++ {
		service := m.Registry.GetRandomService()
		err = m.createKey(adminCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
		bar.Add(1)
	}

	changesetID, err := m.getUserOpenChangesetID(adminCtx)
	if err != nil {
		return err
	}

	err = m.applyChangeset(adminCtx, changesetID)
	if err != nil {
		return err
	}

	bar = progressbar.Default(int64(runOpts.Iterations), "Making changes")
	for i := 0; i < runOpts.Iterations; i++ {
		err = m.makeChanges(ctx)
		if err != nil {
			return err
		}

		bar.Add(1)
	}

	return nil
}

func (m *Manager) createUserContext(ctx context.Context, userId uint) (context.Context, error) {
	user, err := m.AuthService.GetUser(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error creating user context for user %d: %w", userId, err)
	}

	variationHierarchy, err := m.VariationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating user context for user %d: %w", userId, err)
	}

	userBuilder := auth.NewUserBuilder(variationHierarchy)
	userBuilder.WithBasicInfo(user.ID, user.Username, user.GlobalAdministrator)
	for _, permission := range user.Permissions {
		userBuilder.WithPermission(permission.ServiceID, permission.FeatureID, permission.KeyID, permission.Variation, permission.Permission)
	}

	return context.WithValue(ctx, constants.UserContextKey, userBuilder.User()), nil
}

func (m *Manager) createUsers(ctx context.Context) error {
	count := 8000 + m.Rng.Intn(4000)

	log.Printf("Creating %d users...", count)

	userParams := make([]db.CreateUsersParams, count)
	createdAt := time.Now()
	passwordBytes, err := bcrypt.GenerateFromPassword([]byte("test"), 10)
	if err != nil {
		return err
	}
	password := string(passwordBytes)

	for i := 0; i < count; i++ {
		userParams[i] = db.CreateUsersParams{
			Name:                fmt.Sprintf("%s%s_%d", m.Rng.CapitalizedAdjective(), m.Rng.CapitalizedNoun(), i),
			Password:            password,
			GlobalAdministrator: false,
			CreatedAt:           createdAt,
		}
	}

	if _, err = m.DB.CreateUsers(ctx, userParams); err != nil {
		return err
	}

	users, err := m.DB.GetUsersAndGroups(ctx, db.GetUsersAndGroupsParams{
		Offset: 0,
		Limit:  15000,
	})
	if err != nil {
		return err
	}

	for _, user := range users {
		m.Registry.RegisterUser(user.ID)
	}

	return err
}

func (m *Manager) createVariationHierarchy(ctx context.Context) error {
	properties := []struct {
		Name         string
		DisplayName  string
		Values       []string
		ValueParents map[string]string
	}{
		{
			Name:         "env",
			DisplayName:  "Environment",
			Values:       []string{"dev", "qa1", "qa2", "test", "prod"},
			ValueParents: map[string]string{"qa2": "qa", "qa1": "qa"},
		},
		{
			Name:         "product",
			Values:       []string{"blog", "gaming", "shop"},
			ValueParents: map[string]string{},
		},
		{
			Name:         "domain",
			Values:       []string{"necronet.com", "necronet.org", "necronet.net", "necroskillz.io", "necroskillz.dev"},
			ValueParents: map[string]string{},
		},
	}

	for _, property := range properties {
		log.Printf("Creating variation property %s...\n", property.Name)
		propertyID, err := m.VariationPropertyService.CreateVariationProperty(ctx, variationproperty.CreateVariationPropertyParams{
			Name:        property.Name,
			DisplayName: property.DisplayName,
		})
		if err != nil {
			return err
		}

		parentMap := make(map[string]uint)
		for _, value := range property.Values {
			fmt.Printf(" - Creating value %s...\n", value)
			parent, ok := property.ValueParents[value]
			parentID := uint(0)
			if ok {
				parentID, ok = parentMap[parent]
				if !ok {
					fmt.Printf(" - Creating parent value %s...\n", parent)
					parentID, err = m.VariationPropertyService.CreateVariationPropertyValue(ctx, variationproperty.CreateVariationPropertyValueParams{
						Value:      parent,
						PropertyID: propertyID,
					})
					if err != nil {
						return err
					}

					parentMap[parent] = parentID
				}
			}
			_, err = m.VariationPropertyService.CreateVariationPropertyValue(ctx, variationproperty.CreateVariationPropertyValueParams{
				Value:      value,
				PropertyID: propertyID,
				ParentID:   parentID,
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (m *Manager) createServiceTypes(ctx context.Context) error {
	variationHierarchy, err := m.VariationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return err
	}

	serviceTypes := []struct {
		Name       string
		Properties []string
	}{
		{
			Name:       "Simple",
			Properties: []string{"env"},
		},
		{
			Name:       "Complex",
			Properties: []string{"product", "domain", "env"},
		},
	}

	for _, serviceType := range serviceTypes {
		log.Printf("Creating service type %s...\n", serviceType.Name)
		serviceTypeID, err := m.ServiceTypeService.CreateServiceType(ctx, servicetype.CreateServiceTypeParams{
			Name: serviceType.Name,
		})
		if err != nil {
			return err
		}

		for _, property := range serviceType.Properties {
			fmt.Printf(" - Linking property %s...\n", property)
			propertyID, err := variationHierarchy.GetPropertyID(property)
			if err != nil {
				return err
			}

			err = m.ServiceTypeService.LinkVariationPropertyToServiceType(ctx, servicetype.LinkVariationPropertyToServiceTypeParams{
				ServiceTypeID:       serviceTypeID,
				VariationPropertyID: propertyID,
			})
			if err != nil {
				return err
			}
		}

		m.Registry.RegisterServiceType(serviceTypeID)
	}

	return nil
}

func (m *Manager) createService(ctx context.Context) error {
	defer m.Metrics.MeasureSince([]string{"createService"}, time.Now())

	name := fmt.Sprintf("%sService", m.Rng.CapitalizedAdjective())

	serviceVersionID, err := m.ServiceService.CreateService(ctx, service.CreateServiceParams{
		Name:          name,
		Description:   fmt.Sprintf("Description text for service `%s`", name),
		ServiceTypeID: m.Registry.GetRandomServiceType(),
	})
	if err != nil {
		return err
	}

	serviceVersion, err := m.ServiceService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	admins := make([]uint, 3)
	editors := make([]uint, 3)
	for i := 0; i < 3; i++ {
		admin := m.Registry.GetRandomUser()
		editor := m.Registry.GetRandomUser()

		if _, err = m.DB.CreatePermission(ctx, db.CreatePermissionParams{
			UserID:     &admin,
			Kind:       db.PermissionKindService,
			ServiceID:  serviceVersion.ServiceID,
			Permission: db.PermissionLevelAdmin,
		}); err != nil {
			return err
		}

		if _, err = m.DB.CreatePermission(ctx, db.CreatePermissionParams{
			UserID:     &editor,
			Kind:       db.PermissionKindService,
			ServiceID:  serviceVersion.ServiceID,
			Permission: db.PermissionLevelEditor,
		}); err != nil {
			return err
		}

		editors[i] = editor
		admins[i] = admin
	}

	m.Registry.RegisterService(serviceVersionID, editors, admins)

	return nil
}

func (m *Manager) createServiceVersion(ctx context.Context, service *TestDataService) error {
	defer m.Metrics.MeasureSince([]string{"createServiceVersion"}, time.Now())

	newId, err := m.ServiceService.CreateServiceVersion(ctx, service.ServiceVersionID)
	if err != nil {
		return fmt.Errorf("error creating service version from %d: %w", service.ServiceVersionID, err)
	}

	m.Registry.RegisterService(newId, service.Editors, service.Admins)

	return nil
}

func (m *Manager) createFeature(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"createFeature"}, time.Now())

	serviceVersion, err := m.ServiceService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	name := fmt.Sprintf("%s.%s%sFeature", serviceVersion.Name, m.Rng.CapitalizedAdjective(), m.Rng.CapitalizedNoun())

	_, err = m.FeatureService.CreateFeature(ctx, feature.CreateFeatureParams{
		Name:             name,
		ServiceVersionID: serviceVersion.ID,
		Description:      fmt.Sprintf("Description text for feature `%s`", name),
	})

	return err
}

func (m *Manager) createFeatureVersion(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"createFeatureVersion"}, time.Now())

	featureVersion, err := m.getRandomFeature(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if featureVersion == nil {
		return nil
	}

	info, err := m.FeatureService.GetFeatureVersion(ctx, serviceVersionID, featureVersion.ID)
	if err != nil {
		return err
	}

	if !info.IsLastVersion {
		return nil
	}

	_, err = m.FeatureService.CreateFeatureVersion(ctx, feature.CreateFeatureVersionParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersion.ID,
	})

	if err != nil {
		return fmt.Errorf("error creating feature version for /services/%d/features/%d: %w", serviceVersionID, featureVersion.ID, err)
	}

	return nil
}

func (m *Manager) createData(ctx context.Context, valueTypeKind db.ValueTypeKind) string {
	switch valueTypeKind {
	case db.ValueTypeKindString:
		return m.Rng.Noun()
	case db.ValueTypeKindInteger:
		return strconv.Itoa(m.Rng.Intn(1000))
	case db.ValueTypeKindDecimal:
		return strconv.FormatFloat(m.Rng.Float64()*1000, 'f', -1, 64)
	case db.ValueTypeKindBoolean:
		return cases.Upper(language.English).String(strconv.FormatBool(m.Rng.Intn(2) == 1))
	case db.ValueTypeKindJson:
		if m.Rng.Intn(2) == 1 {
			return fmt.Sprintf(`{"%s": "%s"}`, m.Rng.Noun(), m.Rng.Adjective())
		}

		return fmt.Sprintf(`["%s", "%s"]`, m.Rng.Noun(), m.Rng.Noun())
	}

	panic("invalid value type kind")
}

func (m *Manager) getRandomFeature(ctx context.Context, serviceVersionID uint) (*feature.FeatureVersionItemDto, error) {
	features, err := m.FeatureService.GetServiceFeatures(ctx, serviceVersionID)
	if err != nil {
		return nil, err
	}

	if len(features) == 0 {
		return nil, nil
	}

	return &features[m.Rng.Intn(len(features))], nil
}

func (m *Manager) getRandomKey(ctx context.Context, serviceVersionID uint) (*feature.FeatureVersionItemDto, *key.KeyItemDto, error) {
	featureVersion, err := m.getRandomFeature(ctx, serviceVersionID)
	if err != nil {
		return nil, nil, err
	}

	if featureVersion == nil {
		return nil, nil, nil
	}

	keys, err := m.KeyService.GetFeatureKeys(ctx, serviceVersionID, featureVersion.ID)
	if err != nil {
		return nil, nil, err
	}

	if len(keys) == 0 {
		return nil, nil, nil
	}

	return featureVersion, &keys[m.Rng.Intn(len(keys))], nil
}

func (m *Manager) getRandomValue(ctx context.Context, serviceVersionID uint) (*feature.FeatureVersionItemDto, *key.KeyItemDto, *value.VariationValueDto, error) {
	feature, key, err := m.getRandomKey(ctx, serviceVersionID)
	if err != nil {
		return nil, nil, nil, err
	}

	if key == nil {
		return nil, nil, nil, nil
	}

	values, err := m.ValueService.GetKeyValues(ctx, serviceVersionID, feature.ID, key.ID)
	if err != nil {
		return nil, nil, nil, err
	}

	if len(values) == 0 {
		return nil, nil, nil, nil
	}

	return feature, key, &values[m.Rng.Intn(len(values))], nil
}

func (m *Manager) unlinkFeature(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"unlinkFeature"}, time.Now())

	featureVersion, err := m.getRandomFeature(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if featureVersion == nil {
		return nil
	}

	err = m.FeatureService.UnlinkFeatureVersion(ctx, serviceVersionID, featureVersion.ID)
	if err != nil {
		return fmt.Errorf("error unlinking feature version %d for service version %d: %w", featureVersion.ID, serviceVersionID, err)
	}

	return nil
}

func (m *Manager) linkFeature(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"linkFeature"}, time.Now())

	featureVersions, err := m.FeatureService.GetFeatureVersionsLinkableToServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if len(featureVersions) == 0 {
		return nil
	}

	featureVersion := featureVersions[m.Rng.Intn(len(featureVersions))]

	err = m.FeatureService.LinkFeatureVersion(ctx, serviceVersionID, featureVersion.ID)
	if err != nil {
		return fmt.Errorf("error linking feature version %d for service version %d: %w", featureVersion.ID, serviceVersionID, err)
	}

	return nil
}

func (m *Manager) createKey(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"createKey"}, time.Now())

	featureVersion, err := m.getRandomFeature(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if featureVersion == nil {
		return nil
	}

	valueType := m.ValueTypes[m.Rng.Intn(len(m.ValueTypes))]
	name := fmt.Sprintf("%s%s%s", m.Rng.CapitalizedAdjective(), m.Rng.CapitalizedNoun(), valueType.Name)

	_, err = m.KeyService.CreateKey(ctx, key.CreateKeyParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersion.ID,
		Description:      fmt.Sprintf("Description text for key `%s`", name),
		Name:             name,
		Validators:       make([]validation.ValidatorDto, 0),
		ValueTypeID:      valueType.ID,
		DefaultValue:     m.createData(ctx, valueType.Kind),
	})

	if err != nil {
		return fmt.Errorf("error creating key for feature /services/%d/features/%d: %w", serviceVersionID, featureVersion.ID, err)
	}

	return nil
}

func (m *Manager) deleteKey(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"deleteKey"}, time.Now())

	featureVersion, key, err := m.getRandomKey(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if key == nil {
		return nil
	}

	user := m.CurrentUserAccessor.GetUser(ctx)

	fv, err := m.DB.GetFeatureVersion(ctx, db.GetFeatureVersionParams{
		FeatureVersionID: featureVersion.ID,
		ChangesetID:      user.ChangesetID,
	})
	if err != nil {
		return err
	}

	if fv.LinkedToPublishedServiceVersion {
		return nil
	}

	err = m.KeyService.DeleteKey(ctx, serviceVersionID, featureVersion.ID, key.ID)
	if err != nil {
		return fmt.Errorf("error deleting key %d for feature %d: %w", key.ID, featureVersion.ID, err)
	}

	return nil
}

func (m *Manager) deleteValue(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"deleteValue"}, time.Now())

	feature, key, val, err := m.getRandomValue(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if val == nil {
		return nil
	}

	if len(val.Variation) == 0 {
		return nil
	}

	err = m.ValueService.DeleteValue(ctx, value.DeleteValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: feature.ID,
		KeyID:            key.ID,
		ValueID:          val.ID,
	})

	if err != nil {
		return fmt.Errorf("error deleting value %d for at /services/%d/features/%d/keys/%d/values: %w", val.ID, serviceVersionID, feature.ID, key.ID, err)
	}

	return nil
}

func (m *Manager) getRandomVariation(ctx context.Context, serviceTypeID uint) (map[uint]string, error) {
	variationHierarchy, err := m.VariationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	properties, err := variationHierarchy.GetProperties(serviceTypeID)
	if err != nil {
		return nil, err
	}

	variation := map[uint]string{}

	for _, property := range properties {
		if m.Rng.Float64()*float64(len(properties)) >= (1-1/float64(len(properties)))*float64(len(properties)) {
			values := property.GetAllValues()
			variation[property.ID] = values[m.Rng.Intn(len(values))].Value
		}
	}

	return variation, nil
}

func (m *Manager) createValue(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"createValue"}, time.Now())

	serviceVersion, err := m.ServiceService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	featureVersion, key, err := m.getRandomKey(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if key == nil {
		return nil
	}

	variation, err := m.getRandomVariation(ctx, serviceVersion.ServiceTypeID)
	if err != nil {
		return err
	}

	valueID, err := m.ValidationService.DoesVariationExist(ctx, key.ID, variation)
	if err != nil {
		return err
	}

	if valueID != 0 {
		return nil
	}

	data := m.createData(ctx, key.ValueType)

	_, err = m.ValueService.CreateValue(ctx, value.CreateValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersion.ID,
		KeyID:            key.ID,
		Data:             data,
		Variation:        variation,
	})

	if err != nil {
		return fmt.Errorf("error creating value with data %s and variation %v for at /services/%d/features/%d/keys/%d/values: %w", data, variation, serviceVersionID, featureVersion.ID, key.ID, err)
	}

	return nil
}

func (m *Manager) updateValue(ctx context.Context, serviceVersionID uint) error {
	defer m.Metrics.MeasureSince([]string{"updateValue"}, time.Now())

	serviceVersion, err := m.ServiceService.GetServiceVersion(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	featureVersion, key, val, err := m.getRandomValue(ctx, serviceVersionID)
	if err != nil {
		return err
	}

	if val == nil {
		return nil
	}

	variation := val.Variation

	if m.Rng.Intn(10) == 0 && len(val.Variation) > 0 {
		changedVariation, err := m.getRandomVariation(ctx, serviceVersion.ServiceTypeID)
		if err != nil {
			return err
		}

		valueID, err := m.ValidationService.DoesVariationExist(ctx, key.ID, changedVariation)
		if err != nil {
			return err
		}

		if valueID != 0 {
			return nil
		}

		variation = changedVariation
	}

	data := m.createData(ctx, key.ValueType)
	_, err = m.ValueService.UpdateValue(ctx, value.UpdateValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersion.ID,
		KeyID:            key.ID,
		ValueID:          val.ID,
		Data:             data,
		Variation:        variation,
	})

	if err != nil {
		return fmt.Errorf("error updating value %d with data %s -> %s and variation %v -> %v for at /services/%d/features/%d/keys/%d/values: %w", val.ID, val.Data, data, val.Variation, variation, serviceVersionID, featureVersion.ID, key.ID, err)
	}

	return nil
}

func (m *Manager) getUserOpenChangesetID(ctx context.Context) (uint, error) {
	changesetID, err := m.ChangesetService.GetOpenChangesetIDForUser(ctx, m.CurrentUserAccessor.GetUser(ctx).ID)
	if err != nil {
		return 0, err
	}

	return changesetID, nil
}

func (m *Manager) applyChangeset(ctx context.Context, changesetID uint) error {
	defer m.Metrics.MeasureSince([]string{"applyChangeset"}, time.Now())

	return m.ChangesetService.ApplyChangeset(ctx, changesetID, nil)
}

func (m *Manager) commitChangeset(ctx context.Context, changesetID uint) error {
	return m.ChangesetService.CommitChangeset(ctx, changesetID, nil)
}

func (m *Manager) getRandomChangeScope() *ChangeScope {
	n := m.Rng.Intn(100)
	weight := 0
	for _, scope := range m.ChangeScopes {
		weight += scope.Weight
		if n < weight {
			return scope
		}
	}
	panic("should not happen")
}

func (m *Manager) makeChanges(ctx context.Context) error {
	scope := m.getRandomChangeScope()

	service := m.Registry.GetRandomService()
	admin := service.Admins[m.Rng.Intn(len(service.Admins))]
	editor := service.Editors[m.Rng.Intn(len(service.Editors))]
	approver := service.Admins[m.Rng.Intn(len(service.Admins))]

	adminCtx, err := m.createUserContext(ctx, admin)
	if err != nil {
		return err
	}

	editorCtx, err := m.createUserContext(ctx, editor)
	if err != nil {
		return err
	}

	approverCtx, err := m.createUserContext(ctx, approver)
	if err != nil {
		return err
	}

	serviceVersion, err := m.ServiceService.GetServiceVersion(adminCtx, service.ServiceVersionID)
	if err != nil {
		return fmt.Errorf("error making changes: %w", err)
	}

	if serviceVersion.IsLastVersion {
		for range scope.CreateServiceVersion.Next(m.Rng) {
			err := m.createServiceVersion(adminCtx, service)
			if err != nil {
				return err
			}
		}
	}

	for range scope.CreateFeature.Next(m.Rng) {
		err := m.createFeature(adminCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
	}

	if !serviceVersion.Published {
		for range scope.UnlinkFeature.Next(m.Rng) {
			err := m.unlinkFeature(adminCtx, service.ServiceVersionID)
			if err != nil {
				return err
			}
		}

		for range scope.LinkFeature.Next(m.Rng) {
			err := m.linkFeature(adminCtx, service.ServiceVersionID)
			if err != nil {
				return err
			}
		}

		for range scope.CreateFeatureVersion.Next(m.Rng) {
			err := m.createFeatureVersion(adminCtx, service.ServiceVersionID)
			if err != nil {
				return err
			}
		}
	}

	for range scope.CreateKey.Next(m.Rng) {
		err := m.createKey(adminCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
	}

	if !serviceVersion.Published {
		for range scope.DeleteKey.Next(m.Rng) {
			err := m.deleteKey(adminCtx, service.ServiceVersionID)
			if err != nil {
				return err
			}
		}
	}

	adminChangeset, err := m.getUserOpenChangesetID(adminCtx)
	if err != nil {
		return err
	}

	if adminChangeset != 0 {
		if admin != approver {
			err = m.commitChangeset(adminCtx, adminChangeset)
			if err != nil {
				return err
			}
		}

		err = m.applyChangeset(approverCtx, adminChangeset)
		if err != nil {
			return err
		}
	}

	for range scope.DeleteValue.Next(m.Rng) {
		err := m.deleteValue(editorCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
	}

	for range scope.CreateValue.Next(m.Rng) {
		err := m.createValue(editorCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
	}

	for range scope.UpdateValue.Next(m.Rng) {
		err := m.updateValue(editorCtx, service.ServiceVersionID)
		if err != nil {
			return err
		}
	}

	editorChangeset, err := m.getUserOpenChangesetID(editorCtx)
	if err != nil {
		return err
	}

	if editorChangeset != 0 {
		err = m.commitChangeset(editorCtx, editorChangeset)
		if err != nil {
			return err
		}

		err = m.applyChangeset(approverCtx, editorChangeset)
		if err != nil {
			return err
		}
	}

	if !serviceVersion.Published {
		if m.Rng.Intn(100) == 0 {
			adminCtx, err = m.createUserContext(ctx, admin)
			if err != nil {
				return err
			}

			err = m.ServiceService.PublishServiceVersion(adminCtx, service.ServiceVersionID)
			if err != nil {
				return fmt.Errorf("error publishing service version %d: %w", service.ServiceVersionID, err)
			}
		}
	}

	return nil
}
