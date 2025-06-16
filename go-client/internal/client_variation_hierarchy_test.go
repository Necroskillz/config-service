package internal

import (
	"context"
	"testing"

	"github.com/necroskillz/config-service/go-client/internal/test"
	"gotest.tools/v3/assert"
)

func TestVariationHierarchy(t *testing.T) {
	DefaultResponse := func() *test.TestVariationHierarchyResponseBuilder {
		return test.NewTestVariationHierarchyResponseBuilder().
			WithValue("env", "dev").
			WithValue("env", "qa").
			WithChildValue("env", "qa", "qa1").
			WithChildValue("env", "qa", "qa2").
			WithValue("env", "prod").
			WithValue("domain", "necroskillz.io")
	}

	t.Run("GetPropertyNames", func(t *testing.T) {
		response := DefaultResponse().Response()

		variationHierarchy := NewVariationHierarchy(response)

		assert.DeepEqual(t, variationHierarchy.getPropertyNames(), []string{"domain", "env"})
	})

	t.Run("GetPropertyValues", func(t *testing.T) {
		response := DefaultResponse().Response()

		variationHierarchy := NewVariationHierarchy(response)

		assert.DeepEqual(t, variationHierarchy.getPropertyValues("env"), []string{"dev", "prod", "qa1", "qa2"})
		assert.DeepEqual(t, variationHierarchy.getPropertyValues("domain"), []string{"necroskillz.io"})
	})

	t.Run("Validate", func(t *testing.T) {
		type testCase struct {
			staticVariation           map[string]string
			dynamicVariationResolvers map[string]PropertyResolverFunc
			expectedError             string
		}

		run := func(t *testing.T, testCase testCase) {
			response := DefaultResponse()
			variationHierarchy := NewVariationHierarchy(response.Response())
			err := variationHierarchy.Validate(testCase.staticVariation, testCase.dynamicVariationResolvers)
			if testCase.expectedError != "" {
				assert.Error(t, err, testCase.expectedError)
			} else {
				assert.NilError(t, err)
			}
		}

		cases := map[string]testCase{
			"valid": {
				staticVariation:           map[string]string{"domain": "necroskillz.io"},
				dynamicVariationResolvers: map[string]PropertyResolverFunc{},
			},
			"invalid property in static variation": {
				staticVariation:           map[string]string{"invalid": "necroskillz.io"},
				dynamicVariationResolvers: map[string]PropertyResolverFunc{},
				expectedError:             "provided static variation is invalid: property invalid is not defined in the configuration system. available properties: domain, env",
			},
			"invalid property in dynamic variation resolver": {
				staticVariation:           map[string]string{"domain": "necroskillz.io"},
				dynamicVariationResolvers: map[string]PropertyResolverFunc{"invalid": func(ctx context.Context) (string, error) { return "qa1", nil }},
				expectedError:             "dynamic variation property invalid is not defined in the configuration system. available properties: domain, env",
			},
			"invalid value in static variation": {
				staticVariation:           map[string]string{"env": "invalid"},
				dynamicVariationResolvers: map[string]PropertyResolverFunc{},
				expectedError:             "provided static variation is invalid: value invalid of property env is not defined in the configuration system. available values: dev, prod, qa1, qa2",
			},
		}

		test.RunCases(t, run, cases)
	})

	t.Run("GetParents", func(t *testing.T) {
		type testCase struct {
			property        string
			value           string
			expectedParents []string
			expectedError   string
			setupResponse   func(*test.TestVariationHierarchyResponseBuilder) *test.TestVariationHierarchyResponseBuilder
		}

		run := func(t *testing.T, testCase testCase) {
			response := DefaultResponse()
			if testCase.setupResponse != nil {
				response = testCase.setupResponse(response)
			}

			variationHierarchy := NewVariationHierarchy(response.Response())
			parents, err := variationHierarchy.GetParents(testCase.property, testCase.value)
			if testCase.expectedError != "" {
				assert.Error(t, err, testCase.expectedError)
			} else {
				assert.NilError(t, err)
				assert.DeepEqual(t, parents, testCase.expectedParents)
			}
		}

		cases := map[string]testCase{
			"valid root leaf": {
				property:        "env",
				value:           "prod",
				expectedParents: []string{},
			},
			"valid child leaf": {
				property:        "env",
				value:           "qa1",
				expectedParents: []string{"qa"},
			},
			"valid deep child leaf": {
				property:        "env",
				value:           "qa11",
				expectedParents: []string{"qa", "qa1"},
				setupResponse: func(response *test.TestVariationHierarchyResponseBuilder) *test.TestVariationHierarchyResponseBuilder {
					return response.WithChildValue("env", "qa1", "qa11")
				},
			},
			"invalid property": {
				property:      "invalid",
				value:         "qa1",
				expectedError: "property invalid is not defined in the configuration system. available properties: domain, env",
			},
			"invalid value": {
				property:      "env",
				value:         "invalid",
				expectedError: "value invalid of property env is not defined in the configuration system. available values: dev, prod, qa1, qa2",
			},
		}

		test.RunCases(t, run, cases)
	})
}
