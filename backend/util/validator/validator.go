package validator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

type ValidationError struct {
	Field   string
	Message string
}

func NewValidationError(field string, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func (v *ValidationError) Error() string {
	return v.Message
}

type RuleID string

const (
	RuleIDRequired        RuleID = "required"
	RuleIDMin             RuleID = "min"
	RuleIDMax             RuleID = "max"
	RuleIDMinFloat        RuleID = "min_float"
	RuleIDMaxFloat        RuleID = "max_float"
	RuleIDMinLength       RuleID = "min_length"
	RuleIDMaxLength       RuleID = "max_length"
	RuleIDRegex           RuleID = "regex"
	RuleIDJsonSchema      RuleID = "json_schema"
	RuleIDValidJson       RuleID = "valid_json"
	RuleIDValidJsonSchema RuleID = "valid_json_schema"
	RuleIDValidInteger    RuleID = "valid_integer"
	RuleIDValidFloat      RuleID = "valid_float"
	RuleIDValidRegex      RuleID = "valid_regex"
)

var (
	ErrNumberParseError = errors.New("unable to parse number")
)

type RuleFunc func(ctx context.Context, value any, fieldName string, options ...any) error

type Validator struct {
	rules map[RuleID]RuleFunc
}

func New() *Validator {
	validator := &Validator{
		rules: make(map[RuleID]RuleFunc),
	}

	validator.registerRules()

	return validator
}

func normalizeInt(value any) (int64, error) {
	switch x := value.(type) {
	case int:
		return int64(x), nil
	case uint:
		return int64(x), nil
	case int64:
		return x, nil
	case uint64:
		return int64(x), nil
	case string:
		parsed, err := strconv.ParseInt(x, 10, 64)
		if err != nil {
			return 0, ErrNumberParseError
		}
		return parsed, nil
	}

	return 0, fmt.Errorf("type %T cannot be converted to int64", value)
}

func normalizeFloat(value any) (float64, error) {
	switch x := value.(type) {
	case float64:
		return x, nil
	case float32:
		return float64(x), nil
	case string:
		parsed, err := strconv.ParseFloat(x, 64)
		if err != nil {
			return 0, ErrNumberParseError
		}
		return parsed, nil
	}

	return 0, fmt.Errorf("type %T cannot be converted to float64", value)
}

func param[T any](options []any, index int) (T, error) {
	if index >= len(options) {
		return *new(T), fmt.Errorf("required validator param at index %d not provided", index)
	}

	param, ok := options[index].(T)
	if !ok {
		return *new(T), fmt.Errorf("validator param at index %d is not of type %T", index, new(T))
	}

	return param, nil
}

func (v *Validator) registerRules() {
	v.registerRule(RuleIDRequired, func(ctx context.Context, value any, fieldName string, options ...any) error {
		valid := false

		switch x := value.(type) {
		case string:
			if x != "" {
				valid = true
			}
		default:
			if x != nil {
				valid = true
			}
		}

		if !valid {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s is required", fieldName))
		}

		return nil
	})

	v.registerRule(RuleIDMin, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		num, err := normalizeInt(value)
		if err != nil {
			if errors.Is(err, ErrNumberParseError) {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid integer", fieldName))
			}

			return err
		}

		min, err := param[int](options, 0)
		if err != nil {
			return err
		}

		if num < int64(min) {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s must be at least %d", fieldName, min))
		}

		return nil
	})

	v.registerRule(RuleIDMax, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		num, err := normalizeInt(value)
		if err != nil {
			if errors.Is(err, ErrNumberParseError) {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid integer", fieldName))
			}

			return err
		}

		max, err := param[int](options, 0)
		if err != nil {
			return err
		}

		if num > int64(max) {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s must be at most %d", fieldName, max))
		}

		return nil
	})

	v.registerRule(RuleIDMinFloat, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		num, err := normalizeFloat(value)
		if err != nil {
			if errors.Is(err, ErrNumberParseError) {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid float", fieldName))
			}

			return err
		}

		min, err := param[float64](options, 0)
		if err != nil {
			return err
		}

		if num < min {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s must be at least %f", fieldName, min))
		}

		return nil
	})

	v.registerRule(RuleIDMaxFloat, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		num, err := normalizeFloat(value)
		if err != nil {
			if errors.Is(err, ErrNumberParseError) {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid float", fieldName))
			}

			return err
		}

		max, err := param[float64](options, 0)
		if err != nil {
			return err
		}

		if num > max {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s must be at most %f", fieldName, max))
		}

		return nil
	})

	v.registerRule(RuleIDMinLength, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		switch x := value.(type) {
		case string:
			if len(x) == 0 {
				return nil
			}

			min, err := param[int](options, 0)
			if err != nil {
				return err
			}

			if len(x) < min {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be at least %d characters", fieldName, min))
			}
		default:
			return fmt.Errorf("invalid type for min length validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDMaxLength, func(ctx context.Context, value any, fieldName string, options ...any) error {
		if value == nil {
			return nil
		}

		switch x := value.(type) {
		case string:
			max, err := param[int](options, 0)
			if err != nil {
				return err
			}

			if len(x) > max {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be less than or equal to %d characters", fieldName, max))
			}
		default:
			return fmt.Errorf("invalid type for max length validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDValidRegex, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			_, err := regexp.Compile(x)
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid regex", fieldName))
			}
		default:
			return fmt.Errorf("invalid type for regex validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDRegex, func(ctx context.Context, value any, fieldName string, options ...any) error {
		regex, err := param[string](options, 0)
		if err != nil {
			return err
		}

		switch x := value.(type) {
		case string:
			if x == "" {
				return nil
			}

			match, err := regexp.MatchString(regex, x)
			if err != nil {
				return err
			}

			if !match {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must match the regex %s", fieldName, regex))
			}
		default:
			return fmt.Errorf("invalid type for regex validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDValidJson, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			var jsonObj any
			err := json.Unmarshal([]byte(x), &jsonObj)
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be valid JSON", fieldName))
			}
		default:
			return fmt.Errorf("invalid type for valid JSON validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDValidJsonSchema, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			schema, err := jsonschema.UnmarshalJSON(strings.NewReader(x))
			if err != nil {
				return err
			}

			c := jsonschema.NewCompiler()
			c.AddResource("schema.json", schema)
			_, err = c.Compile("schema.json")
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid JSON schema", fieldName))
			}
		default:
			return fmt.Errorf("invalid type for JSON schema validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDJsonSchema, func(ctx context.Context, value any, fieldName string, options ...any) error {
		schemaStr, err := param[string](options, 0)
		if err != nil {
			return err
		}

		switch x := value.(type) {
		case string:
			schema, err := jsonschema.UnmarshalJSON(strings.NewReader(schemaStr))
			if err != nil {
				return err
			}

			inst, err := jsonschema.UnmarshalJSON(strings.NewReader(x))
			if err != nil {
				return err
			}

			c := jsonschema.NewCompiler()
			c.AddResource("schema.json", schema)
			sch, err := c.Compile("schema.json")
			if err != nil {
				return fmt.Errorf("invalid JSON schema: %w", err)
			}

			err = sch.Validate(inst)
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must match the JSON schema %s", fieldName, schemaStr))
			}
		default:
			return fmt.Errorf("invalid type for JSON schema validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDValidInteger, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			_, err := strconv.ParseInt(x, 10, 64)
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid integer", fieldName))
			}
		default:
			return fmt.Errorf("invalid type for valid integer validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDValidFloat, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			_, err := strconv.ParseFloat(x, 64)
			if err != nil {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be a valid float", fieldName))
			}
		default:
			return fmt.Errorf("invalid type for valid float validator %T", value)
		}

		return nil
	})
}

func (v *Validator) registerRule(id RuleID, rule RuleFunc) {
	v.rules[id] = rule
}

type RuleExecFunc func(ctx context.Context, fieldName string, value any) error
type RuleExecContext struct {
	fieldName string
	value     any
	fn        RuleExecFunc
}

type Context struct {
	validator *Validator
	fieldName string
	value     any
	rules     []RuleExecContext
}

func (v *Validator) Validate(value any, fieldName string) *Context {
	return &Context{
		fieldName: fieldName,
		value:     value,
		validator: v,
		rules:     []RuleExecContext{},
	}
}

func (v *Context) Rule(ruleID RuleID, options ...any) *Context {
	v.rules = append(v.rules, RuleExecContext{
		fieldName: v.fieldName,
		value:     v.value,
		fn: func(ctx context.Context, fieldName string, value any) error {
			rule, ok := v.validator.rules[ruleID]
			if !ok {
				return fmt.Errorf("Rule %s not found", ruleID)
			}

			return rule(ctx, value, fieldName, options...)
		},
	})

	return v
}

type ValidatorFunc func(vc *Context) *Context

func (v *Context) Func(fn ValidatorFunc) *Context {
	return fn(v)
}

func (v *Context) Error(ctx context.Context) error {
	for _, rule := range v.rules {
		err := rule.fn(ctx, rule.fieldName, rule.value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (v *Context) Validate(value any, fieldName string) *Context {
	v.fieldName = fieldName
	v.value = value

	return v
}

func (v *Context) Required() *Context {
	return v.Rule(RuleIDRequired)
}

func (v *Context) Min(min int) *Context {
	return v.Rule(RuleIDMin, min)
}

func (v *Context) Max(max int) *Context {
	return v.Rule(RuleIDMax, max)
}

func (v *Context) MinFloat(min float64) *Context {
	return v.Rule(RuleIDMinFloat, min)
}

func (v *Context) MaxFloat(max float64) *Context {
	return v.Rule(RuleIDMaxFloat, max)
}

func (v *Context) MinLength(min int) *Context {
	return v.Rule(RuleIDMinLength, min)
}

func (v *Context) MaxLength(max int) *Context {
	return v.Rule(RuleIDMaxLength, max)
}

func (v *Context) Regex(regex string) *Context {
	return v.Rule(RuleIDRegex, regex)
}

func (v *Context) ValidJson() *Context {
	return v.Rule(RuleIDValidJson)
}

func (v *Context) ValidJsonSchema() *Context {
	return v.Rule(RuleIDValidJsonSchema)
}

func (v *Context) ValidRegex() *Context {
	return v.Rule(RuleIDValidRegex)
}

func (v *Context) JsonSchema(schema string) *Context {
	return v.Rule(RuleIDJsonSchema, schema)
}

func (v *Context) ValidInteger() *Context {
	return v.Rule(RuleIDValidInteger)
}

func (v *Context) ValidFloat() *Context {
	return v.Rule(RuleIDValidFloat)
}
