package main

import (
	"time"

	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=postgres dbname=config_service port=5432 sslmode=disable TimeZone=UTC"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.Exec("DROP TABLE IF EXISTS changeset_changes;")
	db.Exec("DROP TABLE IF EXISTS variation_value_variation_property_values;")
	db.Exec("DROP TABLE IF EXISTS variation_values;")
	db.Exec("DROP TABLE IF EXISTS keys;")
	db.Exec("DROP TABLE IF EXISTS value_types;")
	db.Exec("DROP TABLE IF EXISTS feature_version_service_versions;")
	db.Exec("DROP TABLE IF EXISTS feature_versions;")
	db.Exec("DROP TABLE IF EXISTS service_versions;")
	db.Exec("DROP TABLE IF EXISTS features;")
	db.Exec("DROP TABLE IF EXISTS services;")
	db.Exec("DROP TABLE IF EXISTS service_type_variation_properties;")
	db.Exec("DROP TABLE IF EXISTS service_types;")
	db.Exec("DROP TABLE IF EXISTS changesets;")
	db.Exec("DROP TABLE IF EXISTS user_permission_variation_property_values;")
	db.Exec("DROP TABLE IF EXISTS variation_property_values;")
	db.Exec("DROP TABLE IF EXISTS variation_properties;")
	db.Exec("DROP TABLE IF EXISTS user_permissions;")
	db.Exec("DROP TABLE IF EXISTS users;")

	db.AutoMigrate(
		&model.Service{},
		&model.ServiceVersion{},
		&model.Changeset{},
		&model.Feature{},
		&model.FeatureVersion{},
		&model.ValueType{},
		&model.Key{},
		&model.VariationValue{},
		&model.VariationProperty{},
		&model.VariationPropertyValue{},
		&model.FeatureVersionServiceVersion{},
		&model.User{},
		&model.ChangesetChange{},
		&model.UserPermission{},
		&model.ServiceType{},
		&model.ServiceTypeVariationProperty{},
	)

	password, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)

	admin := model.User{
		Name:                "admin",
		Password:            string(password),
		GlobalAdministrator: true,
	}

	db.Create(&admin)

	strValueType := model.ValueType{
		Name: "String",
	}

	db.Create(&strValueType)

	db.Create(&model.ValueType{
		Name: "Number",
	})

	db.Create(&model.ValueType{
		Name: "Boolean",
	})

	db.Create(&model.ValueType{
		Name: "JSON",
	})

	envVariationProperty := model.VariationProperty{
		Name: "Environment",
	}

	db.Create(&envVariationProperty)

	prodVariationPropertyValue := model.VariationPropertyValue{
		VariationPropertyID: envVariationProperty.ID,
		Value:               "Production",
	}

	db.Create(&prodVariationPropertyValue)

	qaVariationPropertyValue := model.VariationPropertyValue{
		VariationPropertyID: envVariationProperty.ID,
		Value:               "QA",
	}

	db.Create(&qaVariationPropertyValue)

	qa1VariationPropertyValue := model.VariationPropertyValue{
		VariationPropertyID: envVariationProperty.ID,
		Value:               "QA1",
		ParentID:            &qaVariationPropertyValue.ID,
	}

	db.Create(&qa1VariationPropertyValue)

	serviceType := model.ServiceType{
		Name: "TestServiceType",
		VariationProperties: []model.ServiceTypeVariationProperty{
			{
				VariationProperty: envVariationProperty,
				Priority:          1,
			},
		},
	}

	db.Create(&serviceType)

	now := time.Now()
	past := now.Add(time.Hour * -24)
	nowEnd := now.Add(time.Millisecond * -1)

	service := model.Service{
		Name:          "TestService",
		Description:   "TestService Description",
		ServiceTypeID: serviceType.ID,
	}

	db.Create(&service)

	service2 := model.Service{
		Name:          "TestService2",
		Description:   "TestService2 Description",
		ServiceTypeID: serviceType.ID,
	}

	db.Create(&service2)

	userPermission := model.UserPermission{
		UserID:     admin.ID,
		ServiceID:  service.ID,
		Permission: constants.PermissionAdmin,
	}

	db.Create(&userPermission)

	changeset := model.Changeset{
		UserID: admin.ID,
		State:  model.ChangesetStateApplied,
	}

	db.Create(&changeset)

	serviceVersion := model.ServiceVersion{
		ServiceID: service.ID,
		Version:   1,
		ValidFrom: &past,
		ValidTo:   &nowEnd,
	}

	db.Create(&serviceVersion)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		ServiceVersionID: &serviceVersion.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	serviceVersion12 := model.ServiceVersion{
		ServiceID: service.ID,
		Version:   2,
		ValidFrom: &now,
	}

	db.Create(&serviceVersion12)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		ServiceVersionID: &serviceVersion12.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	serviceVersion2 := model.ServiceVersion{
		ServiceID: service2.ID,
		Version:   1,
		ValidFrom: &now,
	}

	db.Create(&serviceVersion2)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		ServiceVersionID: &serviceVersion2.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	feature := model.Feature{
		Name:        "TestService.TestFeature",
		Description: "TestService.TestFeature Description",
		ServiceID:   service.ID,
	}

	db.Create(&feature)

	userPermission2 := model.UserPermission{
		UserID:     admin.ID,
		ServiceID:  service.ID,
		FeatureID:  &feature.ID,
		Permission: constants.PermissionAdmin,
	}

	db.Create(&userPermission2)

	featureVersion := model.FeatureVersion{
		FeatureID: feature.ID,
		Version:   1,
		ValidFrom: &now,
	}

	db.Create(&featureVersion)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		FeatureVersionID: &featureVersion.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	featureVersionServiceVersion := model.FeatureVersionServiceVersion{
		FeatureVersionID: featureVersion.ID,
		ServiceVersionID: serviceVersion.ID,
		ValidFrom:        &now,
	}

	db.Create(&featureVersionServiceVersion)
	db.Create(&model.ChangesetChange{
		ChangesetID:                    changeset.ID,
		FeatureVersionServiceVersionID: &featureVersionServiceVersion.ID,
		ServiceVersionID:               &serviceVersion.ID,
		FeatureVersionID:               &featureVersion.ID,
		Type:                           model.ChangesetChangeTypeCreate,
	})

	featureVersionServiceVersion2 := model.FeatureVersionServiceVersion{
		FeatureVersionID: featureVersion.ID,
		ServiceVersionID: serviceVersion12.ID,
		ValidFrom:        &now,
	}

	db.Create(&featureVersionServiceVersion2)
	db.Create(&model.ChangesetChange{
		ChangesetID:                    changeset.ID,
		FeatureVersionServiceVersionID: &featureVersionServiceVersion2.ID,
		ServiceVersionID:               &serviceVersion12.ID,
		FeatureVersionID:               &featureVersion.ID,
		Type:                           model.ChangesetChangeTypeCreate,
	})

	key := model.Key{
		Name:             "TestKey",
		ValueTypeID:      strValueType.ID,
		FeatureVersionID: featureVersion.ID,
		ValidFrom:        &now,
	}

	db.Create(&key)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		FeatureVersionID: &featureVersion.ID,
		KeyID:            &key.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	defaultValue := "Default"

	variationValue := model.VariationValue{
		KeyID:     key.ID,
		Data:      &defaultValue,
		ValidFrom: &now,
	}

	db.Create(&variationValue)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		FeatureVersionID: &featureVersion.ID,
		KeyID:            &key.ID,
		VariationValueID: &variationValue.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})

	prodValue := "Prod"

	prodVariationValue := model.VariationValue{
		KeyID: key.ID,
		Data:  &prodValue,
		VariationPropertyValues: []model.VariationPropertyValue{
			prodVariationPropertyValue,
		},
		ValidFrom: &now,
	}

	db.Create(&prodVariationValue)
	db.Create(&model.ChangesetChange{
		ChangesetID:      changeset.ID,
		FeatureVersionID: &featureVersion.ID,
		KeyID:            &key.ID,
		VariationValueID: &prodVariationValue.ID,
		Type:             model.ChangesetChangeTypeCreate,
	})
}
