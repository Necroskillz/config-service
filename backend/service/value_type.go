package service

import (
	"context"

	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/util/str"
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

type ValueTypeDto struct {
	ID                uint                            `json:"id" validate:"required"`
	Name              string                          `json:"name" validate:"required"`
	Kind              db.ValueTypeKind                `json:"kind" validate:"required"`
	Validators        []ValidatorWithParameterTypeDto `json:"validators" validate:"required"`
	AllowedValidators []AllowedValidatorDto           `json:"allowedValidators" validate:"required"`
}

func (s *ValueTypeService) GetValueTypes(ctx context.Context) ([]ValueTypeDto, error) {
	valueTypes, err := s.queries.GetValueTypes(ctx)
	if err != nil {
		return nil, err
	}

	valueValidators, err := s.queries.GetValueTypeValueValidators(ctx)
	if err != nil {
		return nil, err
	}

	validatorMap := make(map[uint][]ValidatorWithParameterTypeDto, len(valueTypes))
	for _, validator := range valueValidators {
		validatorMap[*validator.ValueTypeID] = append(validatorMap[*validator.ValueTypeID], ValidatorWithParameterTypeDto{
			ValidatorDto: ValidatorDto{
				ValidatorType: validator.ValidatorType,
				Parameter:     str.FromPtr(validator.Parameter),
				ErrorText:     str.FromPtr(validator.ErrorText),
			},
			ParameterType: s.valueValidatorService.GetValidatorParameterType(validator.ValidatorType),
		})
	}

	result := make([]ValueTypeDto, len(valueTypes))
	for i, valueType := range valueTypes {
		validators, ok := validatorMap[valueType.ID]
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

		result[i] = ValueTypeDto{
			ID:                valueType.ID,
			Name:              valueType.Name,
			Kind:              valueType.Kind,
			Validators:        validators,
			AllowedValidators: allowedValidatorsDtos,
		}
	}

	return result, nil
}
