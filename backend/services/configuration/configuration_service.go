package configuration

import (
	"context"
	"slices"
	"time"

	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
)

type Service struct {
	queries                   *db.Queries
	variationContextService   *variation.ContextService
	variationHierarchyService *variation.HierarchyService
}

type ServiceVersions []db.ServiceVersion

func (s ServiceVersions) GetIds() []uint {
	result := make([]uint, len(s))
	for i, serviceVersion := range s {
		result[i] = serviceVersion.ID
	}
	return result
}

func (s ServiceVersions) ArePublished() bool {
	for _, serviceVersion := range s {
		if !serviceVersion.Published {
			return false
		}
	}

	return true
}

func NewService(queries *db.Queries, variationContextService *variation.ContextService, variationHierarchyService *variation.HierarchyService) *Service {
	return &Service{queries: queries, variationContextService: variationContextService, variationHierarchyService: variationHierarchyService}
}

func (s *Service) getServiceVersions(ctx context.Context, serviceVersionSpecifiers []core.ServiceVersionSpecifier) (ServiceVersions, error) {
	result := make(ServiceVersions, len(serviceVersionSpecifiers))

	for i, serviceVersion := range serviceVersionSpecifiers {
		serviceVersion, err := s.queries.GetServiceVersionByNameAndVersion(ctx, db.GetServiceVersionByNameAndVersionParams{
			Name:    serviceVersion.Name,
			Version: serviceVersion.Version,
		})
		if err != nil {
			return nil, err
		}
		result[i] = serviceVersion
	}

	return result, nil
}

func (s *Service) GetNextChangesets(ctx context.Context, serviceVersionSpecifiers []core.ServiceVersionSpecifier, afterChangesetID uint) ([]uint, error) {
	serviceVersions, err := s.getServiceVersions(ctx, serviceVersionSpecifiers)
	if err != nil {
		return nil, err
	}

	changeset, err := s.queries.GetChangeset(ctx, afterChangesetID)
	if err != nil {
		return nil, err
	}

	if changeset.AppliedAt == nil {
		return nil, core.NewServiceError(core.ErrorCodeInvalidOperation, "The changeset after which to get the next changesets must be applied")
	}

	changesetIds, err := s.queries.GetNextChangesetsRelatedToServiceVersions(ctx, db.GetNextChangesetsRelatedToServiceVersionsParams{
		ServiceVersionIds: serviceVersions.GetIds(),
		AppliedAfter:      changeset.AppliedAt,
	})
	if err != nil {
		return nil, err
	}

	if changesetIds == nil {
		return []uint{}, nil
	}

	return changesetIds, nil
}

type ConfigurationDto struct {
	ChangesetID uint                      `json:"changesetId"`
	Features    []FeatureConfigurationDto `json:"features"`
}

type FeatureConfigurationDto struct {
	Name string                `json:"name"`
	Keys []KeyConfigurationDto `json:"keys"`
}

type KeyConfigurationDto struct {
	Name     string                  `json:"name"`
	DataType string                  `json:"dataType"`
	Values   []ValueConfigurationDto `json:"values"`
}

type ValueConfigurationDto struct {
	Data      string            `json:"data"`
	Variation map[string]string `json:"variation,omitempty"`
	Rank      int               `json:"rank"`
}

func (s *Service) GetConfiguration(ctx context.Context, serviceVersionSpecifiers []core.ServiceVersionSpecifier, changesetID uint, mode string, variation map[string]string) (ConfigurationDto, error) {
	timestamp := time.Now()
	isProduction := mode == "production"
	isChangesetApplied := false

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return ConfigurationDto{}, err
	}

	filterVariation := variationHierarchy.GetVariationIDMap(variation)

	if changesetID != 0 {
		changeset, err := s.queries.GetChangeset(ctx, changesetID)
		if err != nil {
			return ConfigurationDto{}, err
		}

		if isProduction && changeset.AppliedAt == nil {
			return ConfigurationDto{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Getting configuration for production mode requires the changeset to be applied")
		}

		if changeset.AppliedAt != nil {
			isChangesetApplied = true
		}

		timestamp = *changeset.AppliedAt
	}

	serviceVersions, err := s.getServiceVersions(ctx, serviceVersionSpecifiers)
	if err != nil {
		return ConfigurationDto{}, err
	}

	if isProduction && !serviceVersions.ArePublished() {
		return ConfigurationDto{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Getting configuration for production mode requires all service versions to be published")
	}

	configuration, err := s.queries.GetConfiguration(ctx, db.GetConfigurationParams{
		ServiceVersionIds: serviceVersions.GetIds(),
		Timestamp:         timestamp,
		IsApplied:         isChangesetApplied,
	})
	if err != nil {
		return ConfigurationDto{}, err
	}

	featureIndex := make(map[uint]int)
	keyIndex := make(map[uint]int)
	features := []FeatureConfigurationDto{}

	for _, value := range configuration {
		fi, ok := featureIndex[value.FeatureID]
		if !ok {
			fi = len(features)
			featureIndex[value.FeatureID] = fi
			features = append(features, FeatureConfigurationDto{
				Name: value.FeatureName,
				Keys: []KeyConfigurationDto{},
			})
		}

		ki, ok := keyIndex[value.KeyID]
		if !ok {
			ki = len(features[fi].Keys)
			keyIndex[value.KeyID] = ki
			features[fi].Keys = append(features[fi].Keys, KeyConfigurationDto{
				Name:     value.KeyName,
				DataType: string(value.ValueType),
				Values:   []ValueConfigurationDto{},
			})
		}

		valueVariation, err := s.variationContextService.GetVariationContextValues(ctx, value.VariationContextID)
		if err != nil {
			return ConfigurationDto{}, err
		}

		match, unresolved, err := variationHierarchy.Filter(valueVariation, filterVariation)
		if err != nil {
			return ConfigurationDto{}, err
		}

		if !match {
			continue
		}

		valueDto := ValueConfigurationDto{
			Data:      value.Data,
			Variation: variationHierarchy.GetVariationStringMap(unresolved),
			Rank:      variationHierarchy.GetRank(value.ServiceTypeID, valueVariation),
		}

		features[fi].Keys[ki].Values = append(features[fi].Keys[ki].Values, valueDto)
	}

	for _, feature := range features {
		for ki, key := range feature.Keys {
			slices.SortFunc(key.Values, func(a, b ValueConfigurationDto) int {
				return a.Rank - b.Rank
			})

			values := make([]ValueConfigurationDto, 1, len(key.Values))
			for _, value := range key.Values {
				if len(value.Variation) == 0 {
					// TODO: merge json
					values[0] = value
				} else {
					values = append(values, value)
				}
			}

			key.Values = values
			feature.Keys[ki] = key
		}
	}

	return ConfigurationDto{
		ChangesetID: changesetID,
		Features:    features,
	}, nil
}
