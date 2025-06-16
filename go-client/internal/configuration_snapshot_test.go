package internal

import (
	"fmt"
	"testing"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"gotest.tools/v3/assert"
)

type TestJSONStruct struct {
	Field1 string
}

type TestFeature struct {
	StringKey  string
	IntKey     int
	BoolKey    bool
	DecimalKey float64
	JsonKey    TestJSONStruct
}

func (f *TestFeature) FeatureName() string {
	return "Feature1"
}

type NonStructFeature string

func (f NonStructFeature) FeatureName() string {
	return "Feature1"
}

type NonPointerFeature struct {
	StringKey string
}

func (f NonPointerFeature) FeatureName() string {
	return "Feature1"
}

func TestConfigurationSnapshot(t *testing.T) {
	DefaultResponse := func() *test.TestConfigurationReponseBuilder {
		return test.NewTestConfigurationReponseBuilder().
			WithDefaultValue("Feature1", "StringKey", DataTypeString, "test").
			WithDefaultValue("Feature1", "IntKey", DataTypeInteger, "1").
			WithDefaultValue("Feature1", "BoolKey", DataTypeBoolean, "true").
			WithDefaultValue("Feature1", "DecimalKey", DataTypeDecimal, "1.0").
			WithDefaultValue("Feature1", "JsonKey", DataTypeJson, "{\"field1\":\"test\"}")
	}

	t.Run("Validate", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			response := DefaultResponse().Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{})
			assert.DeepEqual(t, snapshot.Warnings, []string{})
		})

		t.Run("Error - Missing Feature", func(t *testing.T) {
			response := DefaultResponse().WithoutFeature("Feature1").Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{"Feature Feature1 not found in the configuration"})
			assert.DeepEqual(t, snapshot.Warnings, []string{})
		})

		t.Run("Error - Missing Key", func(t *testing.T) {
			response := DefaultResponse().WithoutKey("Feature1", "StringKey").Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{"Key StringKey not found in the configuration"})
			assert.DeepEqual(t, snapshot.Warnings, []string{})
		})

		t.Run("Error - Invalid Data Type", func(t *testing.T) {
			type testCase struct {
				configurationDataType string
				keyName               string
				actualDataType        string
				errorMessage          string
			}

			run := func(t *testing.T, testCase testCase) {
				response := DefaultResponse().
					WithoutKey("Feature1", testCase.keyName).
					WithDefaultValue("Feature1", testCase.keyName, testCase.configurationDataType, "??").
					Response()

				snapshot := NewConfigurationSnapshot(response)

				snapshot.Validate([]Feature{&TestFeature{}})

				errorMessage := testCase.errorMessage
				if errorMessage == "" {
					errorMessage = fmt.Sprintf("field %s is defined as %s, but configuration type %s requires %s", testCase.keyName, testCase.actualDataType, testCase.configurationDataType, testCase.configurationDataType)
				}

				assert.DeepEqual(t, snapshot.Errors, []string{errorMessage})
				assert.DeepEqual(t, snapshot.Warnings, []string{})
			}

			cases := map[string]testCase{
				"int":     {keyName: "StringKey", configurationDataType: DataTypeInteger, actualDataType: "string"},
				"bool":    {keyName: "StringKey", configurationDataType: DataTypeBoolean, actualDataType: "string"},
				"decimal": {keyName: "StringKey", configurationDataType: DataTypeDecimal, actualDataType: "string"},
				"string":  {keyName: "IntKey", configurationDataType: DataTypeString, actualDataType: "int"},
				"unknown": {keyName: "StringKey", configurationDataType: "unknown", errorMessage: "field StringKey is defined as string, but configuration type unknown is not supported"},
			}

			test.RunCases(t, run, cases)
		})

		t.Run("Warning - Extra Key", func(t *testing.T) {
			response := DefaultResponse().WithDefaultValue("Feature1", "ExtraKey", DataTypeString, "test").Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{})
			assert.DeepEqual(t, snapshot.Warnings, []string{"Configuration contains unused key ExtraKey in feature Feature1"})
		})

		t.Run("Warning - Extra Feature", func(t *testing.T) {
			response := DefaultResponse().WithDefaultValue("Feature2", "StringKey", DataTypeString, "test").Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{})
			assert.DeepEqual(t, snapshot.Warnings, []string{"Configuration contains unused feature Feature2"})
		})

		t.Run("Warning - Extra Feature", func(t *testing.T) {
			response := DefaultResponse().WithDefaultValue("Feature2", "StringKey", DataTypeString, "test").Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{&TestFeature{}})

			assert.DeepEqual(t, snapshot.Errors, []string{})
			assert.DeepEqual(t, snapshot.Warnings, []string{"Configuration contains unused feature Feature2"})
		})

		t.Run("Error - Not a pointer", func(t *testing.T) {
			response := DefaultResponse().Response()

			snapshot := NewConfigurationSnapshot(response)

			snapshot.Validate([]Feature{NonPointerFeature{StringKey: "test"}})

			assert.DeepEqual(t, snapshot.Errors, []string{"feature must be a pointer to a struct"})
			assert.DeepEqual(t, snapshot.Warnings, []string{})
		})

		t.Run("Error - Not a struct", func(t *testing.T) {
			response := DefaultResponse().Response()

			snapshot := NewConfigurationSnapshot(response)
			var nonStruct NonStructFeature = "test"

			snapshot.Validate([]Feature{&nonStruct})

			assert.DeepEqual(t, snapshot.Errors, []string{"feature must be a pointer to a struct"})
			assert.DeepEqual(t, snapshot.Warnings, []string{})
		})
	})

	t.Run("BindFeature", func(t *testing.T) {
		t.Run("Valid", func(t *testing.T) {
			response := DefaultResponse().Response()

			snapshot := NewConfigurationSnapshot(response)

			feature := TestFeature{}

			snapshot.BindFeature(&feature, map[string][]string{}, Overrides{})

			assert.Equal(t, feature.StringKey, "test")
			assert.Equal(t, feature.IntKey, 1)
			assert.Equal(t, feature.BoolKey, true)
			assert.Equal(t, feature.DecimalKey, 1.0)
			assert.DeepEqual(t, feature.JsonKey, TestJSONStruct{Field1: "test"})
		})

		t.Run("Error - Feature not found", func(t *testing.T) {
			response := test.NewTestConfigurationReponseBuilder().Response()
			snapshot := NewConfigurationSnapshot(response)
			feature := &TestFeature{}

			err := snapshot.BindFeature(feature, map[string][]string{}, Overrides{})
			assert.ErrorContains(t, err, "feature Feature1 not found in configuration")
		})

		t.Run("Error - Override wrong type", func(t *testing.T) {
			response := DefaultResponse().Response()
			snapshot := NewConfigurationSnapshot(response)
			feature := &TestFeature{}

			overrides := Overrides{
				"Feature1": {
					"StringKey": 123,
				},
			}

			err := snapshot.BindFeature(feature, map[string][]string{}, overrides)
			assert.ErrorContains(t, err, "field StringKey is defined as string, but override value is int")
		})

		t.Run("Error - Key not found", func(t *testing.T) {
			response := DefaultResponse().WithoutKey("Feature1", "StringKey").Response()
			snapshot := NewConfigurationSnapshot(response)
			feature := &TestFeature{}

			err := snapshot.BindFeature(feature, map[string][]string{}, Overrides{})
			assert.ErrorContains(t, err, "key StringKey not found in configuration")
		})

		t.Run("Error - No matching variation", func(t *testing.T) {
			response := DefaultResponse().
				WithoutKey("Feature1", "StringKey").
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "prod_value", map[string]string{"env": "prod"}, 1).
				Response()
			snapshot := NewConfigurationSnapshot(response)
			feature := &TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(feature, variation, Overrides{})
			assert.ErrorContains(t, err, "no value found for key StringKey")
		})

		t.Run("Error - Invalid Data Type Values", func(t *testing.T) {
			type testCase struct {
				keyName       string
				dataType      string
				value         string
				expectedError string
			}

			run := func(t *testing.T, tc testCase) {
				response := DefaultResponse().
					WithoutKey("Feature1", tc.keyName).
					WithDefaultValue("Feature1", tc.keyName, tc.dataType, tc.value).
					Response()

				snapshot := NewConfigurationSnapshot(response)
				feature := &TestFeature{}

				err := snapshot.BindFeature(feature, map[string][]string{}, Overrides{})
				assert.ErrorContains(t, err, tc.expectedError)
			}

			cases := map[string]testCase{
				"invalid integer": {keyName: "IntKey", dataType: DataTypeInteger, value: "invalid", expectedError: "invalid integer value"},
				"invalid boolean": {keyName: "BoolKey", dataType: DataTypeBoolean, value: "invalid", expectedError: "invalid boolean value"},
				"invalid decimal": {keyName: "DecimalKey", dataType: DataTypeDecimal, value: "invalid", expectedError: "invalid decimal value"},
				"invalid json":    {keyName: "JsonKey", dataType: DataTypeJson, value: "invalid json", expectedError: "failed to unmarshal JSON"},
				"unsupported":     {keyName: "StringKey", dataType: "unsupported", value: "test", expectedError: "unsupported data type"},
			}

			test.RunCases(t, run, cases)
		})

		t.Run("Valid with overrides", func(t *testing.T) {
			response := DefaultResponse().Response()
			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			overrides := Overrides{
				"Feature1": {
					"StringKey": "overridden",
					"IntKey":    42,
				},
			}

			err := snapshot.BindFeature(&feature, map[string][]string{}, overrides)
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "overridden")
			assert.Equal(t, feature.IntKey, 42)
			assert.Equal(t, feature.BoolKey, true)
			assert.Equal(t, feature.DecimalKey, 1.0)
			assert.DeepEqual(t, feature.JsonKey, TestJSONStruct{Field1: "test"})
		})

		t.Run("Valid with variation", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "dev_value", map[string]string{"env": "dev"}, 1).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "dev_value")
		})

		t.Run("Should ignore variation that doesnt match", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "dev_value", map[string]string{"env": "dev"}, 1).
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "specific_value", map[string]string{"env": "dev", "domain": "example.com"}, 2).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "dev_value")
		})

		t.Run("Valid with variation (parent)", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "qa_value", map[string]string{"env": "qa"}, 1).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"qa", "qa1"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "qa_value")
		})

		t.Run("Pick highest rank", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "qa_value", map[string]string{"env": "qa"}, 1).
				WithDynamicVariationValue("Feature1", "StringKey", DataTypeString, "qa1_value", map[string]string{"env": "qa1"}, 2).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"qa", "qa1"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "qa1_value")
		})

		t.Run("JSON merging with multiple values", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "JsonKey", DataTypeJson, "{\"field1\":\"override\"}", map[string]string{"env": "dev"}, 1).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.NilError(t, err)

			assert.Equal(t, feature.StringKey, "test")
			assert.Equal(t, feature.IntKey, 1)
			assert.Equal(t, feature.BoolKey, true)
			assert.Equal(t, feature.DecimalKey, 1.0)
			assert.DeepEqual(t, feature.JsonKey, TestJSONStruct{Field1: "override"})
		})

		t.Run("JSON merging error in first value", func(t *testing.T) {
			response := DefaultResponse().
				WithoutKey("Feature1", "JsonKey").
				WithDynamicVariationValue("Feature1", "JsonKey", DataTypeJson, "{\"field1\":\"valid\"}", map[string]string{"env": "dev"}, 1).
				WithDefaultValue("Feature1", "JsonKey", DataTypeJson, "invalid json").
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.ErrorContains(t, err, "failed to unmarshal JSON:")
		})

		t.Run("JSON merging error in subsequent value", func(t *testing.T) {
			response := DefaultResponse().
				WithDynamicVariationValue("Feature1", "JsonKey", DataTypeJson, "invalid json", map[string]string{"env": "dev"}, 1).
				Response()

			snapshot := NewConfigurationSnapshot(response)
			feature := TestFeature{}

			variation := map[string][]string{
				"env": {"dev"},
			}

			err := snapshot.BindFeature(&feature, variation, Overrides{})
			assert.ErrorContains(t, err, "failed to unmarshal JSON:")
		})
	})
}
