package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockValidationService struct {
	mock.Mock
}

func (m *MockValidationService) IsServiceNameTaken(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockValidationService) IsFeatureNameTaken(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockValidationService) IsKeyNameTaken(ctx context.Context, featureVersionID uint, name string) (bool, error) {
	args := m.Called(ctx, featureVersionID, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockValidationService) DoesVariationExist(ctx context.Context, keyID uint, serviceTypeID uint, variation map[uint]string) (uint, error) {
	args := m.Called(ctx, keyID, serviceTypeID, variation)
	return args.Get(0).(uint), args.Error(1)
}

type ValidatorTestSuite struct {
	suite.Suite
	validator   *Validator
	mockService *MockValidationService
	ctx         context.Context
}

func (s *ValidatorTestSuite) SetupTest() {
	s.mockService = new(MockValidationService)
	s.validator = NewValidator(s.mockService)
	s.ctx = context.Background()
}

func (s *ValidatorTestSuite) TestRequired() {
	tests := []struct {
		name        string
		value       any
		fieldName   string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "empty string",
			value:       "",
			fieldName:   "test",
			expectError: true,
			errorMsg:    "Field test is required",
		},
		{
			name:        "non-empty string",
			value:       "value",
			fieldName:   "test",
			expectError: false,
		},
		{
			name:        "non-string value",
			value:       true,
			fieldName:   "test",
			expectError: true,
			errorMsg:    "invalid type for required validator bool",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.validator.Validate(tt.value, tt.fieldName).Required().Error(s.ctx)

			if tt.expectError {
				if s.Error(err) {
					s.Equal(tt.errorMsg, err.Error())
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ValidatorTestSuite) TestServiceNameNotTaken() {
	tests := []struct {
		name        string
		value       any
		fieldName   string
		mockReturn  bool
		mockError   error
		expectError bool
		errorMsg    string
	}{
		{
			name:        "name available",
			value:       "service1",
			fieldName:   "name",
			mockReturn:  false,
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "name taken",
			value:       "taken-service",
			fieldName:   "name",
			mockReturn:  true,
			mockError:   nil,
			expectError: true,
			errorMsg:    "Service name taken-service is already taken",
		},
		{
			name:        "non-string value",
			value:       123,
			fieldName:   "name",
			expectError: true,
			errorMsg:    "invalid type for service name not taken validator int",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			switch v := tt.value.(type) {
			case string:
				s.mockService.On("IsServiceNameTaken", s.ctx, v).
					Return(tt.mockReturn, tt.mockError).
					Once()
			}

			err := s.validator.Validate(tt.value, tt.fieldName).
				ServiceNameNotTaken().
				Error(s.ctx)

			if tt.expectError {
				if s.Error(err) {
					s.Equal(tt.errorMsg, err.Error())
				}
			} else {
				s.NoError(err)
			}

			s.mockService.AssertExpectations(s.T())
		})
	}
}

func (s *ValidatorTestSuite) TestFeatureNameNotTaken() {
	tests := []struct {
		name        string
		value       any
		fieldName   string
		mockReturn  bool
		mockError   error
		expectError bool
		errorMsg    string
	}{
		{
			name:        "name available",
			value:       "feature1",
			fieldName:   "name",
			mockReturn:  false,
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "name taken",
			value:       "taken-feature",
			fieldName:   "name",
			mockReturn:  true,
			mockError:   nil,
			expectError: true,
			errorMsg:    "Feature name taken-feature is already taken",
		},
		{
			name:        "non-string value",
			value:       123,
			fieldName:   "name",
			expectError: true,
			errorMsg:    "invalid type for feature name not taken validator int",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			switch v := tt.value.(type) {
			case string:
				s.mockService.On("IsFeatureNameTaken", s.ctx, v).
					Return(tt.mockReturn, tt.mockError).
					Once()
			}

			err := s.validator.Validate(tt.value, tt.fieldName).
				FeatureNameNotTaken().
				Error(s.ctx)

			if tt.expectError {
				if s.Error(err) {
					s.Equal(tt.errorMsg, err.Error())
				}
			} else {
				s.NoError(err)
			}

			s.mockService.AssertExpectations(s.T())
		})
	}
}

func (s *ValidatorTestSuite) TestMin() {
	tests := []struct {
		name        string
		value       any
		fieldName   string
		min         int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "value above minimum",
			value:       10,
			fieldName:   "age",
			min:         5,
			expectError: false,
		},
		{
			name:        "value equal to minimum",
			value:       5,
			fieldName:   "age",
			min:         5,
			expectError: false,
		},
		{
			name:        "value below minimum",
			value:       3,
			fieldName:   "age",
			min:         5,
			expectError: true,
			errorMsg:    "Field must be greater than or equal to 5",
		},
		{
			name:        "uint value",
			value:       uint(6),
			fieldName:   "age",
			min:         5,
			expectError: false,
		},
		{
			name:        "non-number value",
			value:       "string",
			fieldName:   "age",
			min:         5,
			expectError: true,
			errorMsg:    "invalid type for min validator string",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := s.validator.Validate(tt.value, tt.fieldName).
				Min(tt.min).
				Error(s.ctx)

			if tt.expectError {
				if s.Error(err) {
					s.Equal(tt.errorMsg, err.Error())
				}
			} else {
				s.NoError(err)
			}
		})
	}
}

func (s *ValidatorTestSuite) TestChainedValidation() {
	s.Run("successful chain", func() {
		s.mockService.On("IsServiceNameTaken", s.ctx, "valid-service").
			Return(false, nil).
			Once()

		err := s.validator.Validate("valid-service", "name").
			Required().
			ServiceNameNotTaken().
			Error(s.ctx)

		s.NoError(err)
		s.mockService.AssertExpectations(s.T())
	})

	s.Run("successful chain - multiple values", func() {
		err := s.validator.Validate("val1", "name1").
			Required().
			Validate(2, "name2").
			Min(1).
			Error(s.ctx)

		s.NoError(err)
		s.mockService.AssertExpectations(s.T())
	})

	s.Run("failing chain - first rule", func() {
		err := s.validator.Validate("", "name").
			Required().
			ServiceNameNotTaken().
			Error(s.ctx)

		s.Error(err)
		s.Equal("Field name is required", err.Error())
		s.mockService.AssertNotCalled(s.T(), "IsServiceNameTaken")
	})

	s.Run("failing chain - second rule", func() {
		s.mockService.On("IsServiceNameTaken", s.ctx, "taken-name").
			Return(true, nil).
			Once()

		err := s.validator.Validate("taken-name", "name").
			Required().
			ServiceNameNotTaken().
			Error(s.ctx)

		s.Error(err)
		s.Equal("Service name taken-name is already taken", err.Error())
		s.mockService.AssertExpectations(s.T())
	})
}

func TestValidatorSuite(t *testing.T) {
	suite.Run(t, new(ValidatorTestSuite))
}
