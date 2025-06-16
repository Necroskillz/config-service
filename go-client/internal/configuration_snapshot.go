package internal

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"time"

	grpcgen "github.com/necroskillz/config-service/go-client/grpc/gen"
)

const (
	DataTypeString  = "string"
	DataTypeInteger = "integer"
	DataTypeBoolean = "boolean"
	DataTypeDecimal = "decimal"
	DataTypeJson    = "json"
)

type ValueSnapshot struct {
	Data      string            `json:"data"`
	Variation map[string]string `json:"variation"`
	Rank      int32             `json:"rank"`
}

func (v *ValueSnapshot) matchVariation(variationWithParents map[string][]string) bool {
	for property, propertyValue := range v.Variation {
		variationPropertyValue, ok := variationWithParents[property]

		if !ok || !slices.Contains(variationPropertyValue, propertyValue) {
			return false
		}
	}

	return true
}

type KeySnapshot struct {
	DataType string           `json:"dataType"`
	Values   []*ValueSnapshot `json:"values"`
}

func NewKeySnapshot(key *grpcgen.ConfigKey) *KeySnapshot {
	values := make([]*ValueSnapshot, len(key.Values))
	for i, value := range key.Values {
		values[i] = &ValueSnapshot{Data: value.Data, Variation: value.Variation, Rank: value.Rank}
	}

	slices.SortFunc(values, func(i, j *ValueSnapshot) int {
		return int(j.Rank - i.Rank)
	})

	return &KeySnapshot{DataType: key.DataType, Values: values}
}

func (k *KeySnapshot) getValues(variationWithParents map[string][]string) []*ValueSnapshot {
	values := make([]*ValueSnapshot, 0, len(k.Values))

	if k.DataType == "json" {
		for _, value := range slices.Backward(k.Values) {
			if value.matchVariation(variationWithParents) {
				values = append(values, value)
			}
		}
	} else {
		for _, value := range k.Values {
			if value.matchVariation(variationWithParents) {
				values = append(values, value)

				return values
			}
		}
	}

	return values
}

type ConfigurationSnapshot struct {
	ChangesetId uint32                             `json:"changesetId"`
	Features    map[string]map[string]*KeySnapshot `json:"features"`
	AppliedAt   *time.Time                         `json:"appliedAt"`
	Errors      []string                           `json:"-"`
	Warnings    []string                           `json:"-"`
}

func NewConfigurationSnapshot(response *grpcgen.GetConfigurationResponse) *ConfigurationSnapshot {
	features := make(map[string]map[string]*KeySnapshot, len(response.Features))
	for _, feature := range response.Features {
		features[feature.Name] = make(map[string]*KeySnapshot, len(feature.Keys))

		for _, key := range feature.Keys {
			features[feature.Name][key.Name] = NewKeySnapshot(key)
		}
	}

	snapshot := &ConfigurationSnapshot{
		ChangesetId: response.ChangesetId,
		Features:    features,
		Errors:      []string{},
		Warnings:    []string{},
	}

	if response.AppliedAt != nil {
		appliedAt := response.AppliedAt.AsTime()
		snapshot.AppliedAt = &appliedAt
	}

	return snapshot
}

type FeatureField struct {
	Value reflect.Value
	Field reflect.StructField
}

func getFeatureFields(feature Feature) ([]FeatureField, error) {
	featureValue := reflect.ValueOf(feature)
	if featureValue.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("feature must be a pointer to a struct")
	}

	featureElem := featureValue.Elem()
	if featureElem.Kind() != reflect.Struct {
		return nil, fmt.Errorf("feature must be a pointer to a struct")
	}

	fields := make([]FeatureField, featureElem.NumField())
	for i := range featureElem.NumField() {
		fields[i] = FeatureField{Value: featureElem.Field(i), Field: featureElem.Type().Field(i)}
	}

	return fields, nil
}

func validateDataType(keyName string, dataType string, fieldType reflect.Type) error {
	switch dataType {
	case DataTypeString:
		if fieldType.Kind() != reflect.String {
			return fmt.Errorf("field %s is defined as %s, but configuration type %s requires string", keyName, fieldType.Kind(), dataType)
		}
	case DataTypeInteger:
		if fieldType.Kind() != reflect.Int && fieldType.Kind() != reflect.Int32 && fieldType.Kind() != reflect.Int64 {
			return fmt.Errorf("field %s is defined as %s, but configuration type %s requires integer", keyName, fieldType.Kind(), dataType)
		}
	case DataTypeBoolean:
		if fieldType.Kind() != reflect.Bool {
			return fmt.Errorf("field %s is defined as %s, but configuration type %s requires boolean", keyName, fieldType.Kind(), dataType)
		}
	case DataTypeDecimal:
		if fieldType.Kind() != reflect.Float32 && fieldType.Kind() != reflect.Float64 {
			return fmt.Errorf("field %s is defined as %s, but configuration type %s requires decimal", keyName, fieldType.Kind(), dataType)
		}
	case DataTypeJson:
		return nil
	default:
		return fmt.Errorf("field %s is defined as %s, but configuration type %s is not supported", keyName, fieldType.Name(), dataType)
	}
	return nil
}

func (c *ConfigurationSnapshot) Validate(features []Feature) {
	visitedFeatures := make(map[string]bool)

	for _, feature := range features {
		featureName := feature.FeatureName()

		visitedFeatures[featureName] = true

		fields, err := getFeatureFields(feature)
		if err != nil {
			c.Errors = append(c.Errors, err.Error())
			continue
		}

		configFeature, ok := c.Features[featureName]
		if !ok {
			c.Errors = append(c.Errors, fmt.Sprintf("Feature %s not found in the configuration", featureName))
			continue
		}

		visitedKeys := make(map[string]bool)

		for _, field := range fields {
			key, ok := configFeature[field.Field.Name]
			if !ok {
				c.Errors = append(c.Errors, fmt.Sprintf("Key %s not found in the configuration", field.Field.Name))
				continue
			}

			visitedKeys[field.Field.Name] = true

			if err := validateDataType(field.Field.Name, key.DataType, field.Field.Type); err != nil {
				c.Errors = append(c.Errors, err.Error())
				continue
			}

			// TODO: Validate values
		}

		if len(visitedKeys) != len(configFeature) {
			for keyName := range configFeature {
				if !visitedKeys[keyName] {
					c.Warnings = append(c.Warnings, fmt.Sprintf("Configuration contains unused key %s in feature %s", keyName, featureName))
				}
			}
		}
	}

	if len(visitedFeatures) != len(c.Features) {
		for featureName := range c.Features {
			if !visitedFeatures[featureName] {
				c.Warnings = append(c.Warnings, fmt.Sprintf("Configuration contains unused feature %s", featureName))
			}
		}
	}
}

func (c *ConfigurationSnapshot) BindFeature(feature Feature, variationWithParents map[string][]string, overrides Overrides) error {
	featureName := feature.FeatureName()
	configFeature, ok := c.Features[featureName]
	if !ok {
		return fmt.Errorf("feature %s not found in configuration", featureName)
	}

	fields, err := getFeatureFields(feature)
	if err != nil {
		return fmt.Errorf("failed to get feature fields: %w", err)
	}

	for _, field := range fields {
		fieldName := field.Field.Name

		if value, overriden := overrides.Get(featureName, fieldName); overriden {
			if field.Field.Type != reflect.TypeOf(value) {
				return fmt.Errorf("field %s is defined as %s, but override value is %s", fieldName, field.Field.Type, reflect.TypeOf(value))
			}

			field.Value.Set(reflect.ValueOf(value))
		} else {
			key, ok := configFeature[fieldName]
			if !ok {
				return fmt.Errorf("key %s not found in configuration", fieldName)
			}

			values := key.getValues(variationWithParents)
			if len(values) == 0 {
				return fmt.Errorf("no value found for key %s", fieldName)
			}

			var data string
			if key.DataType == "json" && len(values) > 1 {
				var mergedValue any
				if err := json.Unmarshal([]byte(values[0].Data), &mergedValue); err != nil {
					return fmt.Errorf("failed to unmarshal JSON: %w", err)
				}

				for _, value := range values[1:] {
					var nextValue any
					if err := json.Unmarshal([]byte(value.Data), &nextValue); err != nil {
						return fmt.Errorf("failed to unmarshal JSON: %w", err)
					}

					mergedValue = MergeObjects(mergedValue, nextValue)
				}

				bytes, err := json.Marshal(mergedValue)
				if err != nil {
					return fmt.Errorf("failed to marshal JSON: %w", err)
				}

				data = string(bytes)
			} else {
				data = values[0].Data
			}

			if err := c.setFieldValue(field, key.DataType, data); err != nil {
				return fmt.Errorf("failed to set field %s: %w", fieldName, err)
			}
		}
	}

	return nil
}

func (c *ConfigurationSnapshot) setFieldValue(field FeatureField, dataType string, value string) error {
	if !field.Value.CanSet() {
		return fmt.Errorf("field %s is not settable", field.Field.Name)
	}

	switch dataType {
	case DataTypeString:
		field.Value.SetString(value)
	case DataTypeInteger:
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s for field %s", value, field.Field.Name)
		}
		field.Value.SetInt(int64(intValue))
	case DataTypeBoolean:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.Value.SetBool(boolValue)
	case DataTypeDecimal:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid decimal value: %s", value)
		}
		field.Value.SetFloat(floatValue)
	case DataTypeJson:
		newValue := reflect.New(field.Field.Type)
		if err := json.Unmarshal([]byte(value), newValue.Interface()); err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %w", err)
		}
		field.Value.Set(newValue.Elem())
	default:
		return fmt.Errorf("unsupported data type: %s", dataType)
	}
	return nil
}
