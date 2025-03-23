package service

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/necroskillz/config-service/db"
)

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
	propertyLookup   map[string]uint
	serviceTypeOrder map[uint][]uint
}

func NewVariationHierarchy(variationPropertyValues []db.GetVariationPropertyValuesRow, serviceTypesProperties []db.GetServiceTypeVariationPropertiesRow) *VariationHierarchy {
	variationHierarchy := &VariationHierarchy{
		properties:       make(map[uint]*VariationHierarchyProperty),
		lookup:           make(map[uint]map[string]*VariationHierarchyValue),
		serviceTypeOrder: make(map[uint][]uint),
		propertyLookup:   make(map[string]uint),
	}

	values := make(map[uint]*VariationHierarchyValue)
	propertyValues := make(map[uint][]*VariationHierarchyValue)

	for _, variationPropertyValue := range variationPropertyValues {
		propertyID := variationPropertyValue.PropertyID

		if _, exists := variationHierarchy.properties[propertyID]; !exists {
			variationHierarchy.properties[propertyID] = &VariationHierarchyProperty{
				ID:          propertyID,
				Name:        variationPropertyValue.PropertyName,
				DisplayName: variationPropertyValue.PropertyDisplayName,
			}
			variationHierarchy.lookup[propertyID] = make(map[string]*VariationHierarchyValue)
			variationHierarchy.propertyLookup[variationPropertyValue.PropertyName] = propertyID
			propertyValues[propertyID] = []*VariationHierarchyValue{}
		}

		value := VariationHierarchyValue{
			ID:       variationPropertyValue.ID,
			Value:    variationPropertyValue.Value,
			Children: []*VariationHierarchyValue{},
		}

		if variationPropertyValue.ParentID != nil {
			parentValue := values[*variationPropertyValue.ParentID]
			value.Parent = parentValue
			value.Depth = parentValue.Depth + 1
			variationHierarchy.properties[propertyID].MaxDepth = max(variationHierarchy.properties[propertyID].MaxDepth, value.Depth)
			parentValue.Children = append(parentValue.Children, &value)
		} else {
			propertyValues[propertyID] = append(propertyValues[propertyID], &value)
		}

		values[variationPropertyValue.ID] = &value
		variationHierarchy.lookup[propertyID][value.Value] = &value
	}

	for propertyID, values := range propertyValues {
		order := 1
		assignOrderToPropertyValues(values, &order)
		variationHierarchy.properties[propertyID].Values = values
	}

	for _, serviceTypeProperty := range serviceTypesProperties {
		variationHierarchy.serviceTypeOrder[serviceTypeProperty.ServiceTypeID] = append(variationHierarchy.serviceTypeOrder[serviceTypeProperty.ServiceTypeID], serviceTypeProperty.VariationPropertyID)
	}

	return variationHierarchy
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

func (v *VariationHierarchy) GetPropertyId(property string) (uint, error) {
	propertyID, ok := v.propertyLookup[property]

	if !ok {
		return 0, fmt.Errorf("property %s not found", property)
	}

	return propertyID, nil
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
	Variation map[uint]string
}

type ValueSortOrder int

const (
	ValueSortOrderTree ValueSortOrder = iota
	ValueSortOrderSpecificity
)

type VariationValueFilter struct {
	Filter          map[uint]string
	IncludeChildren bool
	ValueSortOrder  ValueSortOrder
}

func (v *VariationHierarchy) SortAndFilterValues(serviceTypeID uint, values []VariationValue, filter VariationValueFilter) []EvaluatedVariationValue {
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

		for propertyID, variationPropertyValue := range value.Variation {
			filterValue, ok := filter.Filter[propertyID]

			if ok {
				if filterValue != variationPropertyValue && (filter.IncludeChildren && !slices.Contains(v.GetParents(propertyID, variationPropertyValue), filterValue)) {
					rank = -1
					break
				}
			}

			variationHierarchyValue := v.lookup[propertyID][variationPropertyValue]
			order[orderMap[propertyID]] = variationHierarchyValue.Order

			rank += 1<<rankMap[propertyID] + variationHierarchyValue.Depth
		}

		if rank != -1 {
			evaluatedValues = append(evaluatedValues, EvaluatedVariationValue{
				ID:        value.ID,
				Value:     value.Data,
				Rank:      rank,
				Variation: value.Variation,
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

type VariationHierarchyService struct {
	queries *db.Queries
	cache   *ristretto.Cache[string, any]
}

func NewVariationHierarchyService(queries *db.Queries, cache *ristretto.Cache[string, any]) *VariationHierarchyService {
	return &VariationHierarchyService{queries: queries, cache: cache}
}

func (s *VariationHierarchyService) GetVariationHierarchy(ctx context.Context) (*VariationHierarchy, error) {
	cacheKey := "variation_hierarchy"
	cachedVariationHierarchy, exists := s.cache.Get(cacheKey)

	if exists {
		return cachedVariationHierarchy.(*VariationHierarchy), nil
	}

	variationPropertyValues, err := s.queries.GetVariationPropertyValues(ctx)
	if err != nil {
		return nil, err
	}

	serviceTypesProperties, err := s.queries.GetServiceTypeVariationProperties(ctx)
	if err != nil {
		return nil, err
	}

	variationHierarchy := NewVariationHierarchy(variationPropertyValues, serviceTypesProperties)

	s.cache.SetWithTTL(cacheKey, variationHierarchy, int64(len(variationPropertyValues)*10+len(serviceTypesProperties)), time.Minute*10)

	return variationHierarchy, nil
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
