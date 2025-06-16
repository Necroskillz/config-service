package configuration

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/jsonmerge"
)

type Service struct {
	queries                   *db.Queries
	variationContextService   *variation.ContextService
	variationHierarchyService *variation.HierarchyService
}

type ServiceVersions []db.GetServiceVersionByNameAndVersionRow

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

func (s ServiceVersions) GetServiceTypeIds() []uint {
	visited := make(map[uint]bool)
	ids := make([]uint, 0, len(s))

	for _, serviceVersion := range s {
		if !visited[serviceVersion.ServiceTypeID] {
			visited[serviceVersion.ServiceTypeID] = true
			ids = append(ids, serviceVersion.ServiceTypeID)
		}
	}

	return ids
}

func NewService(queries *db.Queries, variationContextService *variation.ContextService, variationHierarchyService *variation.HierarchyService) *Service {
	return &Service{queries: queries, variationContextService: variationContextService, variationHierarchyService: variationHierarchyService}
}

func (s *Service) getServiceVersions(ctx context.Context, serviceVersionSpecifiers []core.ServiceVersionSpecifier) (ServiceVersions, error) {
	result := make(ServiceVersions, len(serviceVersionSpecifiers))

	for i, svs := range serviceVersionSpecifiers {
		serviceVersion, err := s.queries.GetServiceVersionByNameAndVersion(ctx, db.GetServiceVersionByNameAndVersionParams{
			Name:    svs.Name,
			Version: svs.Version,
		})
		if err != nil {
			return nil, core.NewDbError(err, fmt.Sprintf("ServiceVersion %s v%d", svs.Name, svs.Version))
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

type VariationHierarchyPropertyValueDto struct {
	Value    string                               `json:"value" validate:"required"`
	Children []VariationHierarchyPropertyValueDto `json:"children,omitempty"`
}

type VariationHierarchyPropertyDto struct {
	Name   string                               `json:"name" validate:"required"`
	Values []VariationHierarchyPropertyValueDto `json:"values" validate:"required"`
}

type VariationHierarchyDto struct {
	Properties []VariationHierarchyPropertyDto `json:"properties" validate:"required"`
}

func makeValueDtos(values []*variation.HierarchyValue) []VariationHierarchyPropertyValueDto {
	dtos := make([]VariationHierarchyPropertyValueDto, 0, len(values))
	for _, value := range values {
		dtos = append(dtos, VariationHierarchyPropertyValueDto{
			Value:    value.Value,
			Children: makeValueDtos(value.Children),
		})
	}

	return dtos
}

func (s *Service) GetVariationHierarchy(ctx context.Context, serviceVersionSpecifiers []core.ServiceVersionSpecifier) (*VariationHierarchyDto, error) {
	serviceVersions, err := s.getServiceVersions(ctx, serviceVersionSpecifiers)
	if err != nil {
		return nil, err
	}

	serviceTypeIds := serviceVersions.GetServiceTypeIds()

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return nil, err
	}

	visited := make(map[string]bool)
	result := make([]VariationHierarchyPropertyDto, 0, len(serviceTypeIds))

	for _, serviceTypeID := range serviceTypeIds {
		properties, err := variationHierarchy.GetProperties(serviceTypeID)
		if err != nil {
			return nil, err
		}

		for _, property := range properties {
			if !visited[property.Name] {
				visited[property.Name] = true
				result = append(result, VariationHierarchyPropertyDto{
					Name:   property.Name,
					Values: makeValueDtos(property.Values),
				})
			}
		}
	}

	return &VariationHierarchyDto{
		Properties: result,
	}, nil
}

type ConfigurationDto struct {
	ChangesetID uint                      `json:"changesetId" validate:"required"`
	Features    []FeatureConfigurationDto `json:"features" validate:"required"`
}

type FeatureConfigurationDto struct {
	Name string                `json:"name" validate:"required"`
	Keys []KeyConfigurationDto `json:"keys" validate:"required"`
}

type KeyConfigurationDto struct {
	Name     string                  `json:"name" validate:"required"`
	DataType string                  `json:"dataType" validate:"required"`
	Values   []ValueConfigurationDto `json:"values" validate:"required"`
}

type ValueConfigurationDto struct {
	Data      string            `json:"data" validate:"required"`
	Variation map[string]string `json:"variation,omitempty"`
	Rank      int               `json:"rank" validate:"required"`
}

type GetConfigurationParams struct {
	ServiceVersionSpecifiers []core.ServiceVersionSpecifier
	ChangesetID              *uint
	Mode                     string
	Variation                map[uint]string
}

func (s *Service) GetConfiguration(ctx context.Context, params GetConfigurationParams) (ConfigurationDto, error) {
	var timestamp time.Time
	var changesetID uint
	isProduction := params.Mode == "production"
	isChangesetApplied := false

	variationHierarchy, err := s.variationHierarchyService.GetVariationHierarchy(ctx)
	if err != nil {
		return ConfigurationDto{}, err
	}

	if params.ChangesetID != nil {
		changeset, err := s.queries.GetChangeset(ctx, *params.ChangesetID)
		if err != nil {
			return ConfigurationDto{}, err
		}

		if isProduction && changeset.AppliedAt == nil {
			return ConfigurationDto{}, core.NewServiceError(core.ErrorCodeInvalidOperation, "Getting configuration for production mode requires the changeset to be applied")
		}

		if changeset.AppliedAt != nil {
			isChangesetApplied = true
			timestamp = *changeset.AppliedAt
		}

		changesetID = changeset.ID
	} else {
		lastAppliedChangeset, err := s.queries.GetLastAppliedChangeset(ctx)
		if err != nil {
			return ConfigurationDto{}, err
		}

		changesetID = lastAppliedChangeset.ID
		timestamp = *lastAppliedChangeset.AppliedAt
		isChangesetApplied = true
	}

	serviceVersions, err := s.getServiceVersions(ctx, params.ServiceVersionSpecifiers)
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
		ChangesetID:       changesetID,
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

		match, unresolved, err := variationHierarchy.Filter(valueVariation, params.Variation)
		if err != nil {
			return ConfigurationDto{}, err
		}

		if !match {
			continue
		}

		rank, err := variationHierarchy.GetRank(value.ServiceTypeID, valueVariation)
		if err != nil {
			return ConfigurationDto{}, err
		}

		variationMap, err := variationHierarchy.GetVariationStringMap(unresolved)
		if err != nil {
			return ConfigurationDto{}, err
		}

		valueDto := ValueConfigurationDto{
			Data:      value.Data,
			Variation: variationMap,
			Rank:      rank,
		}

		features[fi].Keys[ki].Values = append(features[fi].Keys[ki].Values, valueDto)
	}

	for _, feature := range features {
		for ki, key := range feature.Keys {
			slices.SortFunc(key.Values, func(a, b ValueConfigurationDto) int {
				return a.Rank - b.Rank
			})

			values := make([]ValueConfigurationDto, 1, len(key.Values))

			if key.DataType == "json" {
				var defaultValue any

				for _, value := range key.Values {
					if len(value.Variation) == 0 {
						var jsonData any
						err := json.Unmarshal([]byte(value.Data), &jsonData)
						if err != nil {
							return ConfigurationDto{}, err
						}

						if defaultValue == nil {
							defaultValue = jsonData
						} else {
							defaultValue = jsonmerge.Merge(defaultValue, jsonData)
						}
					} else {
						values = append(values, value)
					}
				}

				jsonString, err := json.Marshal(defaultValue)
				if err != nil {
					return ConfigurationDto{}, err
				}

				values[0] = ValueConfigurationDto{
					Data: string(jsonString),
				}
			} else {
				for _, value := range key.Values {
					if len(value.Variation) == 0 {
						values[0] = value
					} else {
						values = append(values, value)
					}
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
