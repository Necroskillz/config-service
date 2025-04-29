package service

import (
	"context"
	"fmt"
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

type VariationHierarchyServiceType struct {
	Order    []uint
	RankMap  map[uint]int
	OrderMap map[uint]int
}

type VariationHierarchy struct {
	properties     map[uint]*VariationHierarchyProperty
	lookup         map[uint]map[string]*VariationHierarchyValue
	propertyLookup map[string]uint
	serviceTypes   map[uint]*VariationHierarchyServiceType
}

func NewVariationHierarchy(variationPropertyValues []db.GetVariationPropertyValuesRow, serviceTypesProperties []db.GetServiceTypeVariationPropertiesRow) *VariationHierarchy {
	variationHierarchy := &VariationHierarchy{
		properties:     make(map[uint]*VariationHierarchyProperty),
		lookup:         make(map[uint]map[string]*VariationHierarchyValue),
		serviceTypes:   make(map[uint]*VariationHierarchyServiceType),
		propertyLookup: make(map[string]uint),
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

	accumulatedDepth := 0
	for i, serviceTypeProperty := range serviceTypesProperties {
		serviceType, ok := variationHierarchy.serviceTypes[serviceTypeProperty.ServiceTypeID]
		if !ok {
			serviceType = &VariationHierarchyServiceType{
				Order:    []uint{},
				RankMap:  make(map[uint]int),
				OrderMap: make(map[uint]int),
			}

			variationHierarchy.serviceTypes[serviceTypeProperty.ServiceTypeID] = serviceType
		}

		serviceType.Order = append(serviceType.Order, serviceTypeProperty.VariationPropertyID)
		serviceType.RankMap[serviceTypeProperty.VariationPropertyID] = accumulatedDepth
		serviceType.OrderMap[serviceTypeProperty.VariationPropertyID] = i

		accumulatedDepth += variationHierarchy.properties[serviceTypeProperty.VariationPropertyID].MaxDepth + 1
	}

	return variationHierarchy
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

	for _, propertyID := range v.serviceTypes[serviceTypeID].Order {
		properties = append(properties, v.properties[propertyID])
	}

	return properties
}

func (v *VariationHierarchy) GetPropertyId(property string) (uint, error) {
	propertyID, ok := v.propertyLookup[property]

	if !ok {
		return 0, NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Property %s not found", property))
	}

	return propertyID, nil
}

func (v *VariationHierarchy) VariationMapToIds(serviceTypeID uint, variation map[uint]string) ([]uint, error) {
	ids := []uint{}

	properties := v.GetProperties(serviceTypeID)

	for _, property := range properties {
		value, ok := variation[property.ID]

		if !ok || value == "any" {
			continue
		}

		hierarchyValue, ok := v.lookup[property.ID][value]
		if !ok {
			return nil, NewServiceError(ErrorCodeInvalidOperation, fmt.Sprintf("Value %s not found for property %s", variation[property.ID], property.Name))
		}

		ids = append(ids, hierarchyValue.ID)
	}

	return ids, nil
}

func (v *VariationHierarchy) GetRank(serviceTypeID uint, variation map[uint]string) int {
	if _, ok := v.serviceTypes[serviceTypeID]; !ok {
		panic(fmt.Sprintf("Service type %d not found", serviceTypeID))
	}

	rank := 0

	for propertyID, value := range variation {
		baseRank := v.serviceTypes[serviceTypeID].RankMap[propertyID]
		depth := v.lookup[propertyID][value].Depth

		rank += 1<<baseRank + depth
	}

	return rank
}

func (v *VariationHierarchy) GetOrder(serviceTypeID uint, variation map[uint]string) []int {
	if _, ok := v.serviceTypes[serviceTypeID]; !ok {
		panic(fmt.Sprintf("Service type %d not found", serviceTypeID))
	}

	order := make([]int, len(v.serviceTypes[serviceTypeID].Order))

	for propertyID, value := range variation {
		order[v.serviceTypes[serviceTypeID].OrderMap[propertyID]] = v.lookup[propertyID][value].Order
	}

	return order
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
