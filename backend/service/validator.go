package service

import (
	"context"
	"fmt"
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
	RuleIDRequired            RuleID = "required"
	RuleIDMin                 RuleID = "min"
	RuleIDServiceNameNotTaken RuleID = "service_name_not_taken"
	RuleIDFeatureNameNotTaken RuleID = "feature_name_not_taken"
	RuleIDKeyNameNotTaken     RuleID = "key_name_not_taken"
	RuleIDMaxLength           RuleID = "max_length"
)

type RuleFunc func(ctx context.Context, value any, fieldName string, options ...any) error

type DBValidationService interface {
	IsServiceNameTaken(ctx context.Context, name string) (bool, error)
	IsFeatureNameTaken(ctx context.Context, name string) (bool, error)
	IsKeyNameTaken(ctx context.Context, featureVersionID uint, name string) (bool, error)
	DoesVariationExist(ctx context.Context, keyID uint, serviceTypeID uint, variation map[uint]string) (uint, error)
}

type Validator struct {
	rules             map[RuleID]RuleFunc
	validationService DBValidationService
}

func NewValidator(validationService DBValidationService) *Validator {
	validator := &Validator{
		rules:             make(map[RuleID]RuleFunc),
		validationService: validationService,
	}

	validator.registerRules()

	return validator
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
		case int:
			if x != 0 {
				valid = true
			}
		case uint:
			if x != 0 {
				valid = true
			}
		default:
			if value != nil {
				valid = true
			}
		}

		if !valid {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s is required", fieldName))
		}

		return nil
	})

	v.registerRule(RuleIDMin, func(ctx context.Context, value any, fieldName string, options ...any) error {
		var num int64
		switch x := value.(type) {
		case int:
			num = int64(x)
		case uint:
			num = int64(x)
		case int64:
			num = x
		case uint64:
			num = int64(x)
		default:
			return fmt.Errorf("invalid type for min validator %T", value)
		}

		min, err := param[int](options, 0)
		if err != nil {
			return err
		}

		if num < int64(min) {
			return NewValidationError(fieldName, fmt.Sprintf("Field %s must be greater than or equal to %d", fieldName, min))
		}

		return nil
	})

	v.registerRule(RuleIDMaxLength, func(ctx context.Context, value any, fieldName string, options ...any) error {
		max, err := param[int](options, 0)
		if err != nil {
			return err
		}

		switch x := value.(type) {
		case string:
			if len(x) > max {
				return NewValidationError(fieldName, fmt.Sprintf("Field %s must be less than or equal to %d characters", fieldName, max))
			}
		default:
			return fmt.Errorf("invalid type for max length validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDServiceNameNotTaken, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			taken, err := v.validationService.IsServiceNameTaken(ctx, x)
			if err != nil {
				return err
			}

			if taken {
				return NewValidationError(fieldName, fmt.Sprintf("Service name %s is already taken", x))
			}
		default:
			return fmt.Errorf("invalid type for service name not taken validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDFeatureNameNotTaken, func(ctx context.Context, value any, fieldName string, options ...any) error {
		switch x := value.(type) {
		case string:
			taken, err := v.validationService.IsFeatureNameTaken(ctx, x)
			if err != nil {
				return err
			}

			if taken {
				return NewValidationError(fieldName, fmt.Sprintf("Feature name %s is already taken", x))
			}
		default:
			return fmt.Errorf("invalid type for feature name not taken validator %T", value)
		}

		return nil
	})

	v.registerRule(RuleIDKeyNameNotTaken, func(ctx context.Context, value any, fieldName string, options ...any) error {
		featureVersionID, err := param[uint](options, 0)
		if err != nil {
			return err
		}

		switch x := value.(type) {
		case string:
			taken, err := v.validationService.IsKeyNameTaken(ctx, featureVersionID, x)
			if err != nil {
				return err
			}

			if taken {
				return NewValidationError(fieldName, fmt.Sprintf("Key name %s is already taken", x))
			}
		default:
			return fmt.Errorf("invalid type for key name not taken validator %T", value)
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

type ValidatorContext struct {
	validator *Validator
	fieldName string
	value     any
	rules     []RuleExecContext
}

func (v *Validator) Validate(value any, fieldName string) *ValidatorContext {
	return &ValidatorContext{
		fieldName: fieldName,
		value:     value,
		validator: v,
		rules:     []RuleExecContext{},
	}
}

func (v *ValidatorContext) Rule(ruleID RuleID, options ...any) *ValidatorContext {
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

func (v *ValidatorContext) Func(fn func(vc *ValidatorContext) *ValidatorContext) *ValidatorContext {
	return fn(v)
}

func (v *ValidatorContext) Error(ctx context.Context) error {
	for _, rule := range v.rules {
		err := rule.fn(ctx, rule.fieldName, rule.value)
		if err != nil {
			return NewServiceError(ErrorCodeInvalidInput, err.Error()).WithErr(err)
		}
	}

	return nil
}

func (v *ValidatorContext) Validate(value any, fieldName string) *ValidatorContext {
	v.fieldName = fieldName
	v.value = value

	return v
}

func (v *ValidatorContext) Required() *ValidatorContext {
	return v.Rule(RuleIDRequired)
}

func (v *ValidatorContext) ServiceNameNotTaken() *ValidatorContext {
	return v.Rule(RuleIDServiceNameNotTaken)
}

func (v *ValidatorContext) FeatureNameNotTaken() *ValidatorContext {
	return v.Rule(RuleIDFeatureNameNotTaken)
}

func (v *ValidatorContext) KeyNameNotTaken(featureVersionID uint) *ValidatorContext {
	return v.Rule(RuleIDKeyNameNotTaken, featureVersionID)
}

func (v *ValidatorContext) Min(min int) *ValidatorContext {
	return v.Rule(RuleIDMin, min)
}

func (v *ValidatorContext) MaxLength(max int) *ValidatorContext {
	return v.Rule(RuleIDMaxLength, max)
}
