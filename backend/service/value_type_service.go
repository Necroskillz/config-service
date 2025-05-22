package service

import (
	"context"

	"github.com/necroskillz/config-service/db"
)

type ValueTypeService struct {
	queries               *db.Queries
	valueValidatorService *ValueValidatorService
}

func NewValueTypeService(queries *db.Queries, valueValidatorService *ValueValidatorService) *ValueTypeService {
	return &ValueTypeService{
		queries:               queries,
		valueValidatorService: valueValidatorService,
	}
}

type AllowedValidatorDto struct {
	ValidatorType db.ValueValidatorType       `json:"validatorType" validate:"required"`
	ParameterType ValueValidatorParameterType `json:"parameterType" validate:"required"`
}

type ValidatorWithParameterTypeDto struct {
	ValidatorDto
	ParameterType ValueValidatorParameterType `json:"parameterType" validate:"required"`
}

type ValueTypeDto struct {
	ID                uint                            `json:"id" validate:"required"`
	Name              string                          `json:"name" validate:"required"`
	Kind              db.ValueTypeKind                `json:"kind" validate:"required"`
	Validators        []ValidatorWithParameterTypeDto `json:"validators" validate:"required"`
	AllowedValidators []AllowedValidatorDto           `json:"allowedValidators" validate:"required"`
}

func (s *ValueTypeService) getBuiltInValidatorMap(ctx context.Context) (map[uint][]ValidatorWithParameterTypeDto, error) {
	valueValidators, err := s.queries.GetValueTypeValueValidators(ctx)
	if err != nil {
		return nil, err
	}

	validatorMap := make(map[uint][]ValidatorWithParameterTypeDto)
	for _, validator := range valueValidators {
		validatorMap[*validator.ValueTypeID] = append(validatorMap[*validator.ValueTypeID], ValidatorWithParameterTypeDto{
			ValidatorDto:  NewValidatorDto(validator),
			ParameterType: s.valueValidatorService.GetValidatorParameterType(validator.ValidatorType),
		})
	}

	return validatorMap, nil
}

func (s *ValueTypeService) createValueTypeDto(valueType db.ValueType, builtInValidatorMap map[uint][]ValidatorWithParameterTypeDto) ValueTypeDto {
	validators, ok := builtInValidatorMap[valueType.ID]
	if !ok {
		validators = []ValidatorWithParameterTypeDto{}
	}

	allowedValidators := s.valueValidatorService.GetAllowedKeyValidators(valueType.Kind)
	allowedValidatorsDtos := make([]AllowedValidatorDto, len(allowedValidators))
	for i, validator := range allowedValidators {
		allowedValidatorsDtos[i] = AllowedValidatorDto{
			ValidatorType: validator,
			ParameterType: s.valueValidatorService.GetValidatorParameterType(validator),
		}
	}

	return ValueTypeDto{
		ID:                valueType.ID,
		Name:              valueType.Name,
		Kind:              valueType.Kind,
		Validators:        validators,
		AllowedValidators: allowedValidatorsDtos,
	}
}

func (s *ValueTypeService) GetValueTypes(ctx context.Context) ([]ValueTypeDto, error) {
	valueTypes, err := s.queries.GetValueTypes(ctx)
	if err != nil {
		return nil, err
	}

	validatorMap, err := s.getBuiltInValidatorMap(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]ValueTypeDto, len(valueTypes))
	for i, valueType := range valueTypes {
		result[i] = s.createValueTypeDto(valueType, validatorMap)
	}

	return result, nil
}

func (s *ValueTypeService) GetValueType(ctx context.Context, id uint) (ValueTypeDto, error) {
	valueType, err := s.queries.GetValueType(ctx, id)
	if err != nil {
		return ValueTypeDto{}, err
	}

	validatorMap, err := s.getBuiltInValidatorMap(ctx)
	if err != nil {
		return ValueTypeDto{}, err
	}

	return s.createValueTypeDto(valueType, validatorMap), nil
}
