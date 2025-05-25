package variation

import (
	"fmt"

	"slices"

	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/services/core"
)

type HierarchyProperty struct {
	ID          uint
	Name        string
	DisplayName string
	MaxDepth    int
	Values      []*HierarchyValue
}

type HierarchyValue struct {
	ID         uint
	PropertyID uint
	Value      string
	Archived   bool
	Parent     *HierarchyValue
	Children   []*HierarchyValue
	Depth      int
	Order      int
}

type HierarchyServiceType struct {
	Order    []uint
	RankMap  map[uint]int
	OrderMap map[uint]int
}

type Hierarchy struct {
	properties     map[uint]*HierarchyProperty
	values         map[uint]*HierarchyValue
	lookup         map[uint]map[string]*HierarchyValue
	propertyLookup map[string]uint
	serviceTypes   map[uint]*HierarchyServiceType
}

func NewHierarchy(variationPropertyValues []db.GetVariationPropertyValuesRow, serviceTypesProperties []db.GetServiceTypeVariationPropertiesRow) *Hierarchy {
	variationHierarchy := &Hierarchy{
		properties:     make(map[uint]*HierarchyProperty),
		values:         make(map[uint]*HierarchyValue),
		lookup:         make(map[uint]map[string]*HierarchyValue),
		serviceTypes:   make(map[uint]*HierarchyServiceType),
		propertyLookup: make(map[string]uint),
	}

	propertyValues := make(map[uint][]*HierarchyValue)

	for _, variationPropertyValue := range variationPropertyValues {
		propertyID := variationPropertyValue.PropertyID

		if _, exists := variationHierarchy.properties[propertyID]; !exists {
			variationHierarchy.properties[propertyID] = &HierarchyProperty{
				ID:          propertyID,
				Name:        variationPropertyValue.PropertyName,
				DisplayName: variationPropertyValue.PropertyDisplayName,
			}
			variationHierarchy.lookup[propertyID] = make(map[string]*HierarchyValue)
			variationHierarchy.propertyLookup[variationPropertyValue.PropertyName] = propertyID
			propertyValues[propertyID] = []*HierarchyValue{}
		}

		if variationPropertyValue.ID == nil {
			continue
		}

		value := HierarchyValue{
			ID:         *variationPropertyValue.ID,
			Value:      *variationPropertyValue.Value,
			Archived:   *variationPropertyValue.Archived,
			PropertyID: propertyID,
			Children:   []*HierarchyValue{},
		}

		if variationPropertyValue.ParentID != 0 {
			parentValue := variationHierarchy.values[variationPropertyValue.ParentID]
			value.Parent = parentValue
			value.Depth = parentValue.Depth + 1
			variationHierarchy.properties[propertyID].MaxDepth = max(variationHierarchy.properties[propertyID].MaxDepth, value.Depth)
			parentValue.Children = append(parentValue.Children, &value)
		} else {
			propertyValues[propertyID] = append(propertyValues[propertyID], &value)
		}

		variationHierarchy.values[*variationPropertyValue.ID] = &value
		variationHierarchy.lookup[propertyID][value.Value] = &value
	}

	for propertyID, values := range propertyValues {
		order := 1
		assignOrderToPropertyValues(values, &order)
		variationHierarchy.properties[propertyID].Values = values
	}

	accumulatedDepth := 0
	for _, serviceTypeProperty := range serviceTypesProperties {
		serviceType, ok := variationHierarchy.serviceTypes[serviceTypeProperty.ServiceTypeID]
		if !ok {
			serviceType = &HierarchyServiceType{
				Order:    []uint{},
				RankMap:  make(map[uint]int),
				OrderMap: make(map[uint]int),
			}

			variationHierarchy.serviceTypes[serviceTypeProperty.ServiceTypeID] = serviceType
		}

		serviceType.Order = append(serviceType.Order, serviceTypeProperty.VariationPropertyID)
		serviceType.RankMap[serviceTypeProperty.VariationPropertyID] = accumulatedDepth
		serviceType.OrderMap[serviceTypeProperty.VariationPropertyID] = len(serviceType.Order) - 1

		accumulatedDepth += variationHierarchy.properties[serviceTypeProperty.VariationPropertyID].MaxDepth + 1
	}

	return variationHierarchy
}

func assignOrderToPropertyValues(propertyValues []*HierarchyValue, order *int) {
	for _, propertyValue := range propertyValues {
		propertyValue.Order = *order
		*order++

		if len(propertyValue.Children) > 0 {
			assignOrderToPropertyValues(propertyValue.Children, order)
		}
	}
}

func (v *Hierarchy) GetParents(propertyId uint, value string) []string {
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

func (v *Hierarchy) GetProperties(serviceTypeID uint) []*HierarchyProperty {
	properties := []*HierarchyProperty{}

	for _, propertyID := range v.serviceTypes[serviceTypeID].Order {
		properties = append(properties, v.properties[propertyID])
	}

	return properties
}

func (v *Hierarchy) GetAllProperties() []*HierarchyProperty {
	properties := []*HierarchyProperty{}

	for _, property := range v.properties {
		properties = append(properties, property)
	}

	return properties
}

func (v *Hierarchy) GetProperty(propertyID uint) (*HierarchyProperty, error) {
	property, ok := v.properties[propertyID]

	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Property with ID %d not found", propertyID))
	}

	return property, nil
}

func (v *Hierarchy) GetPropertyID(property string) (uint, error) {
	propertyID, ok := v.propertyLookup[property]

	if !ok {
		return 0, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Property %s not found", property))
	}

	return propertyID, nil
}

func (v *Hierarchy) GetValue(valueID uint) (*HierarchyValue, error) {
	value, ok := v.values[valueID]

	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Value with ID %d not found", valueID))
	}

	return value, nil
}

func (v *Hierarchy) GetPropertyValue(propertyID uint, value string) (*HierarchyValue, error) {
	property, err := v.GetProperty(propertyID)
	if err != nil {
		return nil, err
	}

	hierarchyValue, ok := v.lookup[propertyID][value]

	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Value %s not found for property %s", value, property.Name))
	}

	return hierarchyValue, nil
}

type GetValuesWithSameParentOptions struct {
	ValueID  uint
	ParentID uint
}

type GetValuesWithSameParentOptionsFunc func(options *GetValuesWithSameParentOptions)

func ByValueID(valueID uint) GetValuesWithSameParentOptionsFunc {
	return func(options *GetValuesWithSameParentOptions) {
		options.ValueID = valueID
	}
}

func ByParentID(parentID uint) GetValuesWithSameParentOptionsFunc {
	return func(options *GetValuesWithSameParentOptions) {
		options.ParentID = parentID
	}
}

func (v *Hierarchy) GetValuesWithSameParent(propertyID uint, by GetValuesWithSameParentOptionsFunc) ([]*HierarchyValue, error) {
	opts := GetValuesWithSameParentOptions{}

	by(&opts)

	if opts.ValueID != 0 {
		value, err := v.GetValue(opts.ValueID)
		if err != nil {
			return nil, err
		}

		if value.Parent != nil {
			return value.Parent.Children, nil
		}
	} else if opts.ParentID != 0 {
		parent, err := v.GetValue(opts.ParentID)
		if err != nil {
			return nil, err
		}

		return parent.Children, nil
	}

	property, err := v.GetProperty(propertyID)
	if err != nil {
		return nil, err
	}

	return property.Values, nil
}

func (v *Hierarchy) VariationMapToIDs(serviceTypeID uint, variation map[uint]string) ([]uint, error) {
	ids := []uint{}

	properties := v.GetProperties(serviceTypeID)

	for _, property := range properties {
		value, ok := variation[property.ID]

		if !ok {
			continue
		}

		hierarchyValue, ok := v.lookup[property.ID][value]
		if !ok {
			return nil, core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Value %s not found for property %s", variation[property.ID], property.Name))
		}

		if hierarchyValue.Archived {
			return nil, core.NewServiceError(core.ErrorCodeInvalidOperation, fmt.Sprintf("Value %s for property %s is archived", variation[property.ID], property.Name))
		}

		ids = append(ids, hierarchyValue.ID)
	}

	return ids, nil
}

func (v *Hierarchy) GetVariationStringMap(variation map[uint]string) map[string]string {
	variationMap := make(map[string]string)

	for propertyID, value := range variation {
		property, err := v.GetProperty(propertyID)
		if err != nil {
			continue
		}

		variationMap[property.Name] = value
	}

	return variationMap
}

func (v *Hierarchy) GetVariationIDMap(variation map[string]string) map[uint]string {
	variationIDMap := make(map[uint]string)

	for propertyName, value := range variation {
		propertyID, err := v.GetPropertyID(propertyName)
		if err != nil {
			continue
		}

		variationIDMap[propertyID] = value
	}

	return variationIDMap
}

func (v *Hierarchy) GetRank(serviceTypeID uint, variation map[uint]string) int {
	if _, ok := v.serviceTypes[serviceTypeID]; !ok {
		panic(fmt.Sprintf("Service type %d not found", serviceTypeID))
	}

	rank := 0

	for propertyID, value := range variation {
		if value == "any" {
			continue
		}

		baseRank := v.serviceTypes[serviceTypeID].RankMap[propertyID]
		depth := v.lookup[propertyID][value].Depth

		rank += 1<<baseRank + depth
	}

	return rank
}

func (v *Hierarchy) GetOrder(serviceTypeID uint, variation map[uint]string) []int {
	if _, ok := v.serviceTypes[serviceTypeID]; !ok {
		panic(fmt.Sprintf("Service type %d not found", serviceTypeID))
	}

	order := make([]int, len(v.serviceTypes[serviceTypeID].Order))

	for propertyID, value := range variation {
		if value == "any" {
			continue
		}

		order[v.serviceTypes[serviceTypeID].OrderMap[propertyID]] = v.lookup[propertyID][value].Order
	}

	return order
}

func (v *Hierarchy) Filter(valueVariation map[uint]string, filterVariation map[uint]string) (bool, map[uint]string, error) {
	unresolved := make(map[uint]string)

	for propertyID, value := range valueVariation {
		_, err := v.GetProperty(propertyID)
		if err != nil {
			return false, nil, err
		}

		filterValue, ok := filterVariation[propertyID]
		if !ok {
			unresolved[propertyID] = value
		} else {
			match := value == filterValue

			if !match {
				parents := v.GetParents(propertyID, value)
				if slices.Contains(parents, filterValue) {
					match = true
				}
			}

			if !match {
				return false, nil, nil
			}
		}
	}

	return true, unresolved, nil
}

type ServiceTypePropertyPriority struct {
	PropertyID uint
	Priority   int
}
