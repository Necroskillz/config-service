package server

import (
	"context"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/necroskillz/config-service/service"
)

type CustomValidator struct {
	validator         *validator.Validate
	validationService *service.ValidationService
}

func NewCustomValidator(validator *validator.Validate, validationService *service.ValidationService) *CustomValidator {
	cv := &CustomValidator{
		validator:         validator,
		validationService: validationService,
	}

	return cv
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func (cv *CustomValidator) RegisterCustomValidation(trans ut.Translator) error {
	if err := cv.validator.RegisterValidationCtx("service_name_free", func(ctx context.Context, fl validator.FieldLevel) bool {
		value := fl.Field().String()
		err := cv.validationService.ValidateServiceNameUniqueness(ctx, value)
		return err == nil
	}); err != nil {
		return err
	}

	if err := cv.validator.RegisterValidationCtx("feature_name_free", func(ctx context.Context, fl validator.FieldLevel) bool {
		value := fl.Field().String()
		err := cv.validationService.ValidateFeatureNameUniqueness(ctx, value)
		return err == nil
	}); err != nil {
		return err
	}

	if err := cv.validator.RegisterTranslation("service_name_free", trans, func(ut ut.Translator) error {
		return ut.Add("service_name_free", "Service with this name already exists", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("service_name_free")
		return t
	}); err != nil {
		return err
	}

	if err := cv.validator.RegisterTranslation("feature_name_free", trans, func(ut ut.Translator) error {
		return ut.Add("feature_name_free", "Feature with this name already exists", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("feature_name_free")
		return t
	}); err != nil {
		return err
	}

	return nil
}
