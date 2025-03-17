package service

import (
	"context"
	"sort"

	"github.com/necroskillz/config-service/repository"
)

type VariationHierarchyService struct {
	variationPropertyRepository *repository.VariationPropertyRepository
}

func NewVariationHierarchyService(variationPropertyRepository *repository.VariationPropertyRepository) *VariationHierarchyService {
	return &VariationHierarchyService{variationPropertyRepository: variationPropertyRepository}
}

type VariationHierarchyProperty struct {
	ID       uint
	Name     string
	MaxDepth int
	Values   []VariationHierarchyValue
}

type VariationHierarchyValue struct {
	ID       uint
	Value    string
	Parent   *VariationHierarchyValue
	Children []VariationHierarchyValue
	Depth    int
}

type VariationHierarchy struct {
	tree             map[uint]VariationHierarchyProperty
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
		tree:             make(map[uint]VariationHierarchyProperty),
		lookup:           make(map[uint]map[string]*VariationHierarchyValue),
		serviceTypeOrder: make(map[uint][]uint),
	}

	values := make(map[uint]*VariationHierarchyValue)

	serviceTypePropertyPriority := map[uint][]ServiceTypePropertyPriority{}

	for _, variationProperty := range variationProperties {
		propertyValues := []VariationHierarchyValue{}
		variationHierarchy.lookup[variationProperty.ID] = make(map[string]*VariationHierarchyValue)
		maxDepth := 0

		for _, vartiationPropertyValue := range variationProperty.Values {
			value := VariationHierarchyValue{
				ID:       vartiationPropertyValue.ID,
				Value:    vartiationPropertyValue.Value,
				Children: []VariationHierarchyValue{},
			}

			if vartiationPropertyValue.ParentID != nil {
				parentValue := values[*vartiationPropertyValue.ParentID]
				value.Parent = parentValue
				value.Depth = parentValue.Depth + 1
				maxDepth = max(maxDepth, value.Depth)
				parentValue.Children = append(parentValue.Children, value)
			}

			values[vartiationPropertyValue.ID] = &value
			variationHierarchy.lookup[variationProperty.ID][value.Value] = &value

			propertyValues = append(propertyValues, value)
		}

		variationHierarchy.tree[variationProperty.ID] = VariationHierarchyProperty{
			ID:       variationProperty.ID,
			Name:     variationProperty.Name,
			Values:   propertyValues,
			MaxDepth: maxDepth,
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
