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

func (p *HierarchyProperty) GetAllValues() []*HierarchyValue {
	values := []*HierarchyValue{}
	stack := make([]*HierarchyValue, len(p.Values))
	copy(stack, p.Values)

	for len(stack) > 0 {
		value := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		values = append(values, value)
		stack = append(stack, value.Children...)
	}

	return values
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

func (h *Hierarchy) GetParents(propertyId uint, value string) ([]string, error) {
	hierachyValue, err := h.GetPropertyValue(propertyId, value)
	if err != nil {
		return nil, err
	}

	values := []string{}

	for hierachyValue.Parent != nil {
		values = append(values, hierachyValue.Parent.Value)
		hierachyValue = hierachyValue.Parent
	}

	return values, nil
}

func (h *Hierarchy) GetServiceType(serviceTypeID uint) (*HierarchyServiceType, error) {
	serviceType, ok := h.serviceTypes[serviceTypeID]
	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Service type with ID %d not found", serviceTypeID))
	}

	return serviceType, nil
}

func (h *Hierarchy) GetProperties(serviceTypeID uint) ([]*HierarchyProperty, error) {
	properties := []*HierarchyProperty{}

	serviceType, err := h.GetServiceType(serviceTypeID)
	if err != nil {
		return nil, err
	}

	for _, propertyID := range serviceType.Order {
		property, err := h.GetProperty(propertyID)
		if err != nil {
			return nil, err
		}

		properties = append(properties, property)
	}

	return properties, nil
}

func (h *Hierarchy) GetAllProperties() []*HierarchyProperty {
	properties := []*HierarchyProperty{}

	for _, property := range h.properties {
		properties = append(properties, property)
	}

	return properties
}

func (h *Hierarchy) GetProperty(propertyID uint) (*HierarchyProperty, error) {
	property, ok := h.properties[propertyID]

	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Property with ID %d not found", propertyID))
	}

	return property, nil
}

func (h *Hierarchy) GetPropertyID(property string) (uint, error) {
	propertyID, ok := h.propertyLookup[property]

	if !ok {
		return 0, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Property %s not found", property))
	}

	return propertyID, nil
}

func (h *Hierarchy) GetValue(valueID uint) (*HierarchyValue, error) {
	value, ok := h.values[valueID]

	if !ok {
		return nil, core.NewServiceError(core.ErrorCodeRecordNotFound, fmt.Sprintf("Value with ID %d not found", valueID))
	}

	return value, nil
}

func (h *Hierarchy) GetPropertyValue(propertyID uint, value string) (*HierarchyValue, error) {
	property, err := h.GetProperty(propertyID)
	if err != nil {
		return nil, err
	}

	hierarchyValue, ok := h.lookup[propertyID][value]

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

func (h *Hierarchy) GetValuesWithSameParent(propertyID uint, by GetValuesWithSameParentOptionsFunc) ([]*HierarchyValue, error) {
	opts := GetValuesWithSameParentOptions{}

	by(&opts)

	if opts.ValueID != 0 {
		value, err := h.GetValue(opts.ValueID)
		if err != nil {
			return nil, err
		}

		if value.Parent != nil {
			return value.Parent.Children, nil
		}
	} else if opts.ParentID != 0 {
		parent, err := h.GetValue(opts.ParentID)
		if err != nil {
			return nil, err
		}

		return parent.Children, nil
	}

	property, err := h.GetProperty(propertyID)
	if err != nil {
		return nil, err
	}

	return property.Values, nil
}

func (h *Hierarchy) GetVariationStringMap(variation map[uint]string) (map[string]string, error) {
	variationMap := make(map[string]string)

	for propertyID, value := range variation {
		property, err := h.GetProperty(propertyID)
		if err != nil {
			return nil, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
		}

		variationMap[property.Name] = value
	}

	return variationMap, nil
}

func (h *Hierarchy) GetVariationIDMap(variation map[string]string) (map[uint]string, error) {
	variationIDMap := make(map[uint]string)

	for propertyName, value := range variation {
		propertyID, err := h.GetPropertyID(propertyName)
		if err != nil {
			return nil, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
		}

		variationIDMap[propertyID] = value
	}

	return variationIDMap, nil
}

func (h *Hierarchy) GetRank(serviceTypeID uint, variation map[uint]string) (int, error) {
	serviceType, err := h.GetServiceType(serviceTypeID)
	if err != nil {
		return 0, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
	}

	rank := 0

	for propertyID, value := range variation {
		baseRank, ok := serviceType.RankMap[propertyID]
		if !ok {
			return 0, core.NewServiceError(core.ErrorCodeUnexpectedError, fmt.Sprintf("Property with ID %d not found in RankMap of service type %d", propertyID, serviceTypeID))
		}

		value, err := h.GetPropertyValue(propertyID, value)
		if err != nil {
			return 0, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
		}

		rank += 1<<baseRank + value.Depth
	}

	return rank, nil
}

func (h *Hierarchy) GetOrder(serviceTypeID uint, variation map[uint]string) ([]int, error) {
	serviceType, err := h.GetServiceType(serviceTypeID)
	if err != nil {
		return nil, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
	}

	order := make([]int, len(serviceType.Order))

	for propertyID, value := range variation {
		orderIndex, ok := serviceType.OrderMap[propertyID]
		if !ok {
			return nil, core.NewServiceError(core.ErrorCodeUnexpectedError, fmt.Sprintf("Property with ID %d not found in OrderMap of service type %d", propertyID, serviceTypeID))
		}

		order[orderIndex] = h.lookup[propertyID][value].Order
	}

	return order, nil
}

func (h *Hierarchy) Filter(valueVariation map[uint]string, filterVariation map[uint]string) (bool, map[uint]string, error) {
	unresolved := make(map[uint]string)

	for propertyID, value := range valueVariation {
		filterValue, ok := filterVariation[propertyID]
		if !ok {
			unresolved[propertyID] = value
		} else {
			match := value == filterValue

			if !match {
				parents, err := h.GetParents(propertyID, value)
				if err != nil {
					return false, nil, core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
				}

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

func validateVariation[T comparable](h *Hierarchy, serviceTypeID uint, variation map[T]string, selector func(property *HierarchyProperty) T) error {
	if variation == nil {
		return core.NewServiceError(core.ErrorCodeInvalidInput, "Variation is required")
	}

	properties, err := h.GetProperties(serviceTypeID)
	if err != nil {
		return core.NewServiceError(core.ErrorCodeUnexpectedError, err.Error())
	}

	validated := 0

	for _, property := range properties {
		variationValue, ok := variation[selector(property)]
		if !ok {
			continue
		}

		_, err := h.GetPropertyValue(property.ID, variationValue)
		if err != nil {
			return core.NewServiceError(core.ErrorCodeInvalidInput, err.Error())
		}

		validated++
	}

	if validated != len(variation) {
		return core.NewServiceError(core.ErrorCodeInvalidInput, fmt.Sprintf("Variation %+v contains invalid properties for service type %d", variation, serviceTypeID))
	}

	return nil
}

func (h *Hierarchy) ValidateIDVariation(serviceTypeID uint, variation map[uint]string) error {
	return validateVariation(h, serviceTypeID, variation, func(property *HierarchyProperty) uint {
		return property.ID
	})
}

func (h *Hierarchy) ValidateStringVariation(serviceTypeID uint, variation map[string]string) error {
	return validateVariation(h, serviceTypeID, variation, func(property *HierarchyProperty) string {
		return property.Name
	})
}

type ServiceTypePropertyPriority struct {
	PropertyID uint
	Priority   int
}
