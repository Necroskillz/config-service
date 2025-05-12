package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/util/str"
)

type ValueValidatorParameterType string

const (
	ValueValidatorParameterTypeNone       ValueValidatorParameterType = "none"
	ValueValidatorParameterTypeInteger    ValueValidatorParameterType = "integer"
	ValueValidatorParameterTypeFloat      ValueValidatorParameterType = "float"
	ValueValidatorParameterTypeRegex      ValueValidatorParameterType = "regex"
	ValueValidatorParameterTypeJsonSchema ValueValidatorParameterType = "json_schema"
)

type ValidatorDto struct {
	ValidatorType db.ValueValidatorType `json:"validatorType" validate:"required"`
	Parameter     string                `json:"parameter" validate:"required"`
	ErrorText     string                `json:"errorText" validate:"required"`
	IsBuiltIn     bool                  `json:"isBuiltIn" validate:"required"`
}

func NewValidatorDto(validator db.ValueValidator) ValidatorDto {
	return ValidatorDto{
		ValidatorType: validator.ValidatorType,
		Parameter:     str.FromPtr(validator.Parameter),
		ErrorText:     str.FromPtr(validator.ErrorText),
		IsBuiltIn:     validator.ValueTypeID != nil,
	}
}

type ValueValidatorService struct {
	allowedKeyValidators    map[db.ValueTypeKind][]db.ValueValidatorType
	valueValidators         map[db.ValueValidatorType]ValueValidatorFunc
	validatorParameterTypes map[db.ValueValidatorType]ValueValidatorParameterType
	queries                 *db.Queries
}

func NewValueValidatorService(queries *db.Queries) *ValueValidatorService {
	s := &ValueValidatorService{
		allowedKeyValidators: map[db.ValueTypeKind][]db.ValueValidatorType{
			db.ValueTypeKindString:  {db.ValueValidatorTypeRequired, db.ValueValidatorTypeMinLength, db.ValueValidatorTypeMaxLength, db.ValueValidatorTypeRegex, db.ValueValidatorTypeValidRegex},
			db.ValueTypeKindInteger: {db.ValueValidatorTypeMin, db.ValueValidatorTypeMax, db.ValueValidatorTypeRegex},
			db.ValueTypeKindDecimal: {db.ValueValidatorTypeMinDecimal, db.ValueValidatorTypeMaxDecimal, db.ValueValidatorTypeRegex},
			db.ValueTypeKindBoolean: {},
			db.ValueTypeKindJson:    {db.ValueValidatorTypeRegex, db.ValueValidatorTypeJsonSchema},
		},
		valueValidators: map[db.ValueValidatorType]ValueValidatorFunc{},
		validatorParameterTypes: map[db.ValueValidatorType]ValueValidatorParameterType{
			db.ValueValidatorTypeRequired:     ValueValidatorParameterTypeNone,
			db.ValueValidatorTypeMin:          ValueValidatorParameterTypeInteger,
			db.ValueValidatorTypeMax:          ValueValidatorParameterTypeInteger,
			db.ValueValidatorTypeMinDecimal:   ValueValidatorParameterTypeFloat,
			db.ValueValidatorTypeMaxDecimal:   ValueValidatorParameterTypeFloat,
			db.ValueValidatorTypeMinLength:    ValueValidatorParameterTypeInteger,
			db.ValueValidatorTypeMaxLength:    ValueValidatorParameterTypeInteger,
			db.ValueValidatorTypeRegex:        ValueValidatorParameterTypeRegex,
			db.ValueValidatorTypeJsonSchema:   ValueValidatorParameterTypeJsonSchema,
			db.ValueValidatorTypeValidJson:    ValueValidatorParameterTypeNone,
			db.ValueValidatorTypeValidInteger: ValueValidatorParameterTypeNone,
			db.ValueValidatorTypeValidDecimal: ValueValidatorParameterTypeNone,
			db.ValueValidatorTypeValidRegex:   ValueValidatorParameterTypeNone,
		},
		queries: queries,
	}

	s.registerValueValidators()

	return s
}

type ValueValidatorFunc func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error)

func (s *ValueValidatorService) parseIntParam(param string) (int, error) {
	parsed, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int parameter: %w", err)
	}

	return parsed, nil
}

func (s *ValueValidatorService) parseFloatParam(param string) (float64, error) {
	parsed, err := strconv.ParseFloat(param, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse float parameter: %w", err)
	}

	return parsed, nil
}

func (s *ValueValidatorService) registerValueValidators() {
	s.registerValueValidator(db.ValueValidatorTypeRequired, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.Required()
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMin, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		min, err := s.parseIntParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.Min(min)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMax, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		max, err := s.parseIntParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.Max(max)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMinDecimal, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		min, err := s.parseFloatParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.MinFloat(min)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMaxDecimal, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		max, err := s.parseFloatParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.MaxFloat(max)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMinLength, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		minLength, err := s.parseIntParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.MinLength(minLength)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeMaxLength, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		maxLength, err := s.parseIntParam(param)
		if err != nil {
			return nil, err
		}

		return func(v *ValidatorContext) *ValidatorContext {
			return v.MaxLength(maxLength)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeRegex, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.Regex(param)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeJsonSchema, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.JsonSchema(param)
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeValidJson, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.ValidJson()
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeValidInteger, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.ValidInteger()
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeValidDecimal, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.ValidFloat()
		}, nil
	})

	s.registerValueValidator(db.ValueValidatorTypeValidRegex, func(param string, errorText string) (func(v *ValidatorContext) *ValidatorContext, error) {
		return func(v *ValidatorContext) *ValidatorContext {
			return v.ValidRegex()
		}, nil
	})
}

func (s *ValueValidatorService) registerValueValidator(validatorType db.ValueValidatorType, validatorFunc ValueValidatorFunc) {
	s.valueValidators[validatorType] = validatorFunc
}

func (s *ValueValidatorService) GetAllowedKeyValidators(valueTypeKind db.ValueTypeKind) []db.ValueValidatorType {
	return s.allowedKeyValidators[valueTypeKind]
}

func (s *ValueValidatorService) GetValidatorParameterType(validatorType db.ValueValidatorType) ValueValidatorParameterType {
	return s.validatorParameterTypes[validatorType]
}

type ValueValidatorParams struct {
	ValidatorType db.ValueValidatorType
	Parameter     *string
	ErrorText     *string
}

func (s *ValueValidatorService) CreateValueValidatorFunc(params []ValidatorDto) (ValidatorFunc, error) {
	fns := make([]ValidatorFunc, len(params))

	for i, param := range params {
		validatorFunc, ok := s.valueValidators[param.ValidatorType]
		if !ok {
			return nil, fmt.Errorf("validator type %s not found", param.ValidatorType)
		}

		fn, err := validatorFunc(param.Parameter, param.ErrorText)
		if err != nil {
			return nil, err
		}

		fns[i] = fn
	}

	return func(v *ValidatorContext) *ValidatorContext {
		for _, fn := range fns {
			v = fn(v)
		}

		return v
	}, nil
}

func (s *ValueValidatorService) GetValueValidators(ctx context.Context, keyID *uint, valueTypeID *uint) ([]ValidatorDto, error) {
	valueValidators, err := s.queries.GetValueValidators(ctx, db.GetValueValidatorsParams{
		KeyID:       keyID,
		ValueTypeID: valueTypeID,
	})
	if err != nil {
		return nil, err
	}

	validators := make([]ValidatorDto, len(valueValidators))
	for i, validator := range valueValidators {
		validators[i] = NewValidatorDto(validator)
	}

	return validators, nil
}
