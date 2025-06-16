package internal

import (
	"fmt"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
)

type VariationHierarchy struct {
	Properties map[string]map[string][]string `json:"properties"`
}

func populateValuesWithParents(values []*grpcgen.VariationHierarchyPropertyValue, parents []string, result map[string][]string) {
	for _, value := range values {
		if value.Children == nil {
			result[value.Value] = parents
		} else {
			populateValuesWithParents(value.Children, append(parents, value.Value), result)
		}
	}
}

func NewVariationHierarchy(res *grpcgen.GetVariationHierarchyResponse) *VariationHierarchy {
	properties := make(map[string]map[string][]string)

	for _, property := range res.Properties {
		properties[property.Name] = make(map[string][]string)
		populateValuesWithParents(property.Values, []string{}, properties[property.Name])
	}

	return &VariationHierarchy{Properties: properties}
}

func (v *VariationHierarchy) GetPropertyNames() []string {
	propertyNames := make([]string, 0, len(v.Properties))
	for propertyName := range v.Properties {
		propertyNames = append(propertyNames, propertyName)
	}
	return propertyNames
}

func (v *VariationHierarchy) GetPropertyValues(property string) []string {
	propertyValues := make([]string, 0, len(v.Properties[property]))
	for value := range v.Properties[property] {
		propertyValues = append(propertyValues, value)
	}
	return propertyValues
}

func (v *VariationHierarchy) GetParents(property string, value string) ([]string, error) {
	propertyValues, ok := v.Properties[property]
	if !ok {
		return nil, fmt.Errorf("property %s is not defined in the configuration system. available properties: %v", property, v.GetPropertyNames())
	}

	parents, ok := propertyValues[value]
	if !ok {
		return nil, fmt.Errorf("value %s of property %s is not defined in the configuration system. available values: %v", value, property, v.GetPropertyValues(property))
	}

	return parents, nil
}

func (v *VariationHierarchy) Validate(staticVariation map[string]string, dynamicVariationResolvers map[string]PropertyResolverFunc) error {
	for property, value := range staticVariation {
		_, err := v.GetParents(property, value)
		if err != nil {
			return fmt.Errorf("provided static variation is invalid: %w", err)
		}
	}

	for property := range dynamicVariationResolvers {
		_, ok := v.Properties[property]
		if !ok {
			return fmt.Errorf("dynamic variation property %s is not defined in the configuration system", property)
		}
	}

	return nil
}
