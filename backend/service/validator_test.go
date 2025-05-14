package service

import (
	"context"
	"testing"

	"github.com/necroskillz/config-service/util/test"
	"gotest.tools/v3/assert"
)

const (
	testFieldName = "FieldName"
)

func assertValidatorError(t *testing.T, err error, expectError bool, errorText string) {
	if expectError {
		assert.Error(t, err, errorText)
	} else {
		assert.NilError(t, err)
	}
}

func TestValidator(t *testing.T) {
	validator := NewValidator()

	t.Run("Required", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).Required().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":                {value: "test", expectError: false},
			"invalid empty string": {value: "", expectError: true, errorText: "Field FieldName is required"},
			"invalid nil":          {value: nil, expectError: true, errorText: "Field FieldName is required"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("Min", func(t *testing.T) {
		type testCase struct {
			value       any
			min         int
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).Min(tc.min).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":                   {value: 6, min: 5, expectError: false},
			"valid same value as min": {value: 5, min: 5, expectError: false},
			"invalid":                 {value: 0, min: 1, expectError: true, errorText: "Field FieldName must be at least 1"},
			"valid uint":              {value: uint(6), min: 5, expectError: false},
			"valid int64":             {value: int64(6), min: 5, expectError: false},
			"valid uint64":            {value: uint64(6), min: 5, expectError: false},
			"valid string":            {value: "6", min: 5, expectError: false},
			"invalid string":          {value: "0", min: 1, expectError: true, errorText: "Field FieldName must be at least 1"},
			"not a number string":     {value: "not a number", min: 1, expectError: true, errorText: "Field FieldName must be a valid integer"},
			"wrong type":              {value: 6.0, min: 1, expectError: true, errorText: "type float64 cannot be converted to int64"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("Max", func(t *testing.T) {
		type testCase struct {
			value       any
			max         int
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).Max(tc.max).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":               {value: 4, max: 5, expectError: false},
			"valid same as max":   {value: 5, max: 5, expectError: false},
			"invalid":             {value: 6, max: 5, expectError: true, errorText: "Field FieldName must be at most 5"},
			"valid uint":          {value: uint(4), max: 5, expectError: false},
			"valid int64":         {value: int64(4), max: 5, expectError: false},
			"valid uint64":        {value: uint64(4), max: 5, expectError: false},
			"valid string":        {value: "4", max: 5, expectError: false},
			"invalid string":      {value: "6", max: 5, expectError: true, errorText: "Field FieldName must be at most 5"},
			"not a number string": {value: "not a number", max: 5, expectError: true, errorText: "Field FieldName must be a valid integer"},
			"wrong type":          {value: 6.0, max: 5, expectError: true, errorText: "type float64 cannot be converted to int64"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MinFloat", func(t *testing.T) {
		type testCase struct {
			value       any
			min         float64
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).MinFloat(tc.min).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":               {value: 6.1, min: 6.0, expectError: false},
			"valid same as min":   {value: 6.0, min: 6.0, expectError: false},
			"invalid":             {value: 5.9, min: 6.0, expectError: true, errorText: "Field FieldName must be at least 6.000000"},
			"valid string":        {value: "6.1", min: 6.0, expectError: false},
			"invalid string":      {value: "5.9", min: 6.0, expectError: true, errorText: "Field FieldName must be at least 6.000000"},
			"not a number string": {value: "not a number", min: 6.0, expectError: true, errorText: "Field FieldName must be a valid float"},
			"wrong type":          {value: 6, min: 6.0, expectError: true, errorText: "type int cannot be converted to float64"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MaxFloat", func(t *testing.T) {
		type testCase struct {
			value       any
			max         float64
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).MaxFloat(tc.max).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":               {value: 5.9, max: 6.0, expectError: false},
			"valid same as max":   {value: 6.0, max: 6.0, expectError: false},
			"invalid":             {value: 6.1, max: 6.0, expectError: true, errorText: "Field FieldName must be at most 6.000000"},
			"valid string":        {value: "5.9", max: 6.0, expectError: false},
			"invalid string":      {value: "6.1", max: 6.0, expectError: true, errorText: "Field FieldName must be at most 6.000000"},
			"not a number string": {value: "not a number", max: 6.0, expectError: true, errorText: "Field FieldName must be a valid float"},
			"wrong type":          {value: 6, max: 6.0, expectError: true, errorText: "type int cannot be converted to float64"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MinLength", func(t *testing.T) {
		type testCase struct {
			value       any
			min         int
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).MinLength(tc.min).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":        {value: "abcde", min: 5, expectError: false},
			"valid min":    {value: "abcde", min: 5, expectError: false},
			"invalid":      {value: "abc", min: 5, expectError: true, errorText: "Field FieldName must be at least 5 characters"},
			"wrong type":   {value: 123, min: 5, expectError: true, errorText: "invalid type for min length validator int"},
			"empty string": {value: "", min: 5, expectError: false},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("MaxLength", func(t *testing.T) {
		type testCase struct {
			value       any
			max         int
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).MaxLength(tc.max).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":      {value: "abc", max: 5, expectError: false},
			"valid max":  {value: "abcde", max: 5, expectError: false},
			"invalid":    {value: "abcdef", max: 5, expectError: true, errorText: "Field FieldName must be less than or equal to 5 characters"},
			"wrong type": {value: 123, max: 5, expectError: true, errorText: "invalid type for max length validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("Regex", func(t *testing.T) {
		type testCase struct {
			value       any
			regex       string
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).Regex(tc.regex).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":         {value: "abc123", regex: "^[a-z0-9]+$", expectError: false},
			"invalid":       {value: "abc-123", regex: "^[a-z0-9]+$", expectError: true, errorText: "Field FieldName must match the regex ^[a-z0-9]+$"},
			"wrong type":    {value: 123, regex: "^[a-z0-9]+$", expectError: true, errorText: "invalid type for regex validator int"},
			"invalid regex": {value: "abc", regex: "[", expectError: true, errorText: "error parsing regexp: missing closing ]: `[`"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("ValidJson", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).ValidJson().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":      {value: `{"a":1}`, expectError: false},
			"invalid":    {value: `{"a":`, expectError: true, errorText: "Field FieldName must be valid JSON"},
			"wrong type": {value: 123, expectError: true, errorText: "invalid type for valid JSON validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("ValidJsonSchema", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).ValidJsonSchema().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		validSchema := `{"type":"object","properties":{"a":{"type":"integer"}}}`
		invalidSchema := `{"type":"object","properties":{"a":{"type":"invalid"}}}`
		testCases := map[string]testCase{
			"valid":      {value: validSchema, expectError: false},
			"invalid":    {value: invalidSchema, expectError: true, errorText: "Field FieldName must be a valid JSON schema"},
			"wrong type": {value: 123, expectError: true, errorText: "invalid type for JSON schema validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("ValidRegex", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).ValidRegex().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":      {value: "^[a-z0-9]+$", expectError: false},
			"invalid":    {value: "[", expectError: true, errorText: "Field FieldName must be a valid regex"},
			"wrong type": {value: 123, expectError: true, errorText: "invalid type for regex validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("JsonSchema", func(t *testing.T) {
		type testCase struct {
			value       any
			schema      string
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).JsonSchema(tc.schema).Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		validSchema := `{"type":"object","properties":{"a":{"type":"integer"}}}`
		invalidSchema := `{"type":"object","properties":{"a":{"type":"invalid"}}}`
		validInstance := `{"a":1}`
		invalidInstance := `{"a":"string"}`
		testCases := map[string]testCase{
			"valid":          {value: validInstance, schema: validSchema, expectError: false},
			"invalid":        {value: invalidInstance, schema: validSchema, expectError: true, errorText: "Field FieldName must match the JSON schema {\"type\":\"object\",\"properties\":{\"a\":{\"type\":\"integer\"}}}"},
			"invalid schema": {value: validInstance, schema: invalidSchema, expectError: true, errorText: "Invalid JSON schema"},
			"wrong type":     {value: 123, schema: validSchema, expectError: true, errorText: "invalid type for JSON schema validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("ValidInteger", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).ValidInteger().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":      {value: "123", expectError: false},
			"invalid":    {value: "abc", expectError: true, errorText: "Field FieldName must be a valid integer"},
			"wrong type": {value: 123, expectError: true, errorText: "invalid type for valid integer validator int"},
		}

		test.RunCases(t, run, testCases)
	})

	t.Run("ValidFloat", func(t *testing.T) {
		type testCase struct {
			value       any
			expectError bool
			errorText   string
		}

		run := func(t *testing.T, tc testCase) {
			err := validator.Validate(tc.value, testFieldName).ValidFloat().Error(context.Background())
			assertValidatorError(t, err, tc.expectError, tc.errorText)
		}

		testCases := map[string]testCase{
			"valid":      {value: "123.45", expectError: false},
			"invalid":    {value: "abc", expectError: true, errorText: "Field FieldName must be a valid float"},
			"wrong type": {value: 123, expectError: true, errorText: "invalid type for valid float validator int"},
		}

		test.RunCases(t, run, testCases)
	})
}
