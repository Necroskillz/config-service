package service

import (
	"context"
	"fmt"
	"slices"
	"sort"

	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
)

type VariationHierarchyService struct {
	variationPropertyRepository *repository.VariationPropertyRepository
}

func NewVariationHierarchyService(variationPropertyRepository *repository.VariationPropertyRepository) *VariationHierarchyService {
	return &VariationHierarchyService{variationPropertyRepository: variationPropertyRepository}
}

type VariationHierarchyProperty struct {
	ID          uint
	Name        string
	DisplayName string
	MaxDepth    int
	Values      []*VariationHierarchyValue
}

type VariationHierarchyValue struct {
	ID       uint
	Value    string
	Order    int
	Parent   *VariationHierarchyValue
	Children []*VariationHierarchyValue
	Depth    int
}

type VariationHierarchy struct {
	properties       map[uint]*VariationHierarchyProperty
	lookup           map[uint]map[string]*VariationHierarchyValue
	serviceTypeOrder map[uint][]uint
}

func (v *VariationHierarchy) GetParents(propertyId uint, value string) []string {
	variationHierarchyValue, ok := v.lookup[propertyId][value]

	if !ok {
		return nil
	}

	values := []string{}

	for variationHierarchyValue.Parent != nil {
		values = append(values, variationHierarchyValue.Parent.Value)
		variationHierarchyValue = variationHierarchyValue.Parent
	}

	return values
}

func (v *VariationHierarchy) GetProperties(serviceTypeID uint) []*VariationHierarchyProperty {
	properties := []*VariationHierarchyProperty{}

	for _, propertyID := range v.serviceTypeOrder[serviceTypeID] {
		properties = append(properties, v.properties[propertyID])
	}

	return properties
}

func (v *VariationHierarchy) VariationMapToIds(serviceTypeID uint, variation map[string]string) ([]uint, error) {
	ids := []uint{}

	properties := v.GetProperties(serviceTypeID)

	for _, property := range properties {
		value, ok := variation[property.Name]

		if !ok || value == "any" {
			continue
		}

		hierarchyValue, ok := v.lookup[property.ID][value]
		if !ok {
			return nil, fmt.Errorf("value %s not found for property %s", variation[property.Name], property.Name)
		}

		ids = append(ids, hierarchyValue.ID)
	}

	return ids, nil
}

type EvaluatedVariationValue struct {
	ID        uint
	Value     *string
	Rank      int
	Order     []int
	Variation map[string]string
}

type ValueSortOrder int

const (
	ValueSortOrderTree ValueSortOrder = iota
	ValueSortOrderSpecificity
)

type VariationValueFilter struct {
	Filter          map[string]string
	IncludeChildren bool
	ValueSortOrder  ValueSortOrder
}

func (v *VariationHierarchy) SortAndFilterValues(serviceTypeID uint, values []model.VariationValue, filter VariationValueFilter) []EvaluatedVariationValue {
	evaluatedValues := []EvaluatedVariationValue{}
	rankMap := map[uint]int{}
	orderMap := map[uint]int{}

	accumulatedDepth := 0
	for i, variationPropertyID := range v.serviceTypeOrder[serviceTypeID] {
		rankMap[variationPropertyID] = accumulatedDepth
		accumulatedDepth += v.properties[variationPropertyID].MaxDepth + 1

		orderMap[variationPropertyID] = i
	}

	for _, value := range values {
		rank := 0
		order := make([]int, len(v.serviceTypeOrder[serviceTypeID]))
		variation := map[string]string{}

		for _, variationPropertyValue := range value.VariationPropertyValues {
			property := v.properties[variationPropertyValue.VariationPropertyID]
			filterValue, ok := filter.Filter[property.Name]

			if ok {
				if filterValue != variationPropertyValue.Value && (filter.IncludeChildren && !slices.Contains(v.GetParents(variationPropertyValue.VariationPropertyID, variationPropertyValue.Value), filterValue)) {
					rank = -1
					break
				}
			}

			variation[property.Name] = variationPropertyValue.Value
			variationHierarchyValue := v.lookup[variationPropertyValue.VariationPropertyID][variationPropertyValue.Value]
			order[orderMap[property.ID]] = variationHierarchyValue.Order

			rank += 1<<rankMap[variationPropertyValue.VariationPropertyID] + variationHierarchyValue.Depth
		}

		if rank != -1 {
			evaluatedValues = append(evaluatedValues, EvaluatedVariationValue{
				ID:        value.ID,
				Value:     value.Data,
				Rank:      rank,
				Variation: variation,
				Order:     order,
			})
		}
	}

	slices.SortFunc(evaluatedValues, func(a, b EvaluatedVariationValue) int {
		if filter.ValueSortOrder == ValueSortOrderTree {
			for i := range a.Order {
				if a.Order[i] != b.Order[i] {
					return a.Order[i] - b.Order[i]
				}
			}
		} else if filter.ValueSortOrder == ValueSortOrderSpecificity {
			return a.Rank - b.Rank
		} else {
			panic("invalid value sort order")
		}

		return 0
	})

	return evaluatedValues
}

type ServiceTypePropertyPriority struct {
	PropertyID uint
	Priority   int
}

func (s *VariationHierarchyService) GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error) {
	variationProperties, err := s.variationPropertyRepository.GetAll(ctx)

	if err != nil {
		return nil, err
	}

	variationHierarchy := VariationHierarchy{
		properties:       make(map[uint]*VariationHierarchyProperty),
		lookup:           make(map[uint]map[string]*VariationHierarchyValue),
		serviceTypeOrder: make(map[uint][]uint),
	}

	values := make(map[uint]*VariationHierarchyValue)

	serviceTypePropertyPriority := map[uint][]ServiceTypePropertyPriority{}

	for _, variationProperty := range variationProperties {
		propertyValues := []*VariationHierarchyValue{}
		variationHierarchy.lookup[variationProperty.ID] = make(map[string]*VariationHierarchyValue)
		maxDepth := 0

		for _, vartiationPropertyValue := range variationProperty.Values {
			value := VariationHierarchyValue{
				ID:       vartiationPropertyValue.ID,
				Value:    vartiationPropertyValue.Value,
				Children: []*VariationHierarchyValue{},
			}

			if vartiationPropertyValue.ParentID != nil {
				parentValue := values[*vartiationPropertyValue.ParentID]
				value.Parent = parentValue
				value.Depth = parentValue.Depth + 1
				maxDepth = max(maxDepth, value.Depth)
				parentValue.Children = append(parentValue.Children, &value)
			} else {
				propertyValues = append(propertyValues, &value)
			}

			values[vartiationPropertyValue.ID] = &value
			variationHierarchy.lookup[variationProperty.ID][value.Value] = &value
		}

		order := 1
		assignOrderToPropertyValues(propertyValues, &order)

		variationHierarchy.properties[variationProperty.ID] = &VariationHierarchyProperty{
			ID:          variationProperty.ID,
			Name:        variationProperty.Name,
			DisplayName: variationProperty.DisplayName,
			Values:      propertyValues,
			MaxDepth:    maxDepth,
		}

		for _, serviceType := range variationProperty.ServiceTypes {
			serviceTypePropertyPriority[serviceType.ServiceTypeID] = append(serviceTypePropertyPriority[serviceType.ServiceTypeID], ServiceTypePropertyPriority{
				PropertyID: variationProperty.ID,
				Priority:   serviceType.Priority,
			})
		}
	}

	for serviceTypeID, propertyWithPriority := range serviceTypePropertyPriority {
		sort.Slice(propertyWithPriority, func(i, j int) bool {
			return propertyWithPriority[i].Priority < propertyWithPriority[j].Priority
		})

		for _, prop := range propertyWithPriority {
			variationHierarchy.serviceTypeOrder[serviceTypeID] = append(variationHierarchy.serviceTypeOrder[serviceTypeID], prop.PropertyID)
		}
	}

	return &variationHierarchy, nil
}

func assignOrderToPropertyValues(propertyValues []*VariationHierarchyValue, order *int) {
	for _, propertyValue := range propertyValues {
		propertyValue.Order = *order
		*order++

		if len(propertyValue.Children) > 0 {
			assignOrderToPropertyValues(propertyValue.Children, order)
		}
	}
}
