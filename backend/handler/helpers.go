package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/util/validator"
)

func parseVariationParams[T comparable](c echo.Context, keyFunc func(string) (T, error), validator func(T, string) error) (map[T]string, error) {
	variation := make(map[T]string)
	var variationParams []string

	err := echo.QueryParamsBinder(c).Strings("variation[]", &variationParams).BindError()
	if err != nil {
		return nil, err
	}

	for i, param := range variationParams {
		parts := strings.SplitN(param, ":", 2)
		if len(parts) != 2 {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid variation parameter at index %d: '%s'. Expected format: key:value", i, param))
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Variation parameter key cannot be empty at index %d: '%s'", i, param))
		}

		processedKey, err := keyFunc(key)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		err = validator(processedKey, value)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// Check for duplicate keys
		if _, exists := variation[processedKey]; exists {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Duplicate variation parameter key: '%s'", key))
		}

		variation[processedKey] = value
	}

	return variation, nil
}

func (h *Handler) GetVariationFromQuery(c echo.Context) (map[string]string, error) {
	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return nil, err
	}

	return parseVariationParams(c, func(name string) (string, error) {
		return name, nil
	}, func(propertyName string, value string) error {
		propertyID, err := variationHierarchy.GetPropertyID(propertyName)
		if err != nil {
			return err
		}

		_, err = variationHierarchy.GetPropertyValue(propertyID, value)
		if err != nil {
			return err
		}

		return nil
	})
}

func (h *Handler) GetVariationFromQueryIds(c echo.Context) (map[uint]string, error) {
	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return nil, err
	}

	return parseVariationParams(c, func(id string) (uint, error) {
		propertyID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid property ID %s", id)).WithInternal(err)
		}

		return uint(propertyID), nil
	}, func(propertyID uint, value string) error {
		_, err = variationHierarchy.GetPropertyValue(propertyID, value)
		if err != nil {
			return err
		}

		return nil
	})
}

func ToHTTPError(err error) *echo.HTTPError {
	if errors.Is(err, core.ErrPermissionDenied) {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	if errors.Is(err, core.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if errors.Is(err, core.ErrInvalidOperation) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if errors.Is(err, core.ErrInvalidInput) || errors.Is(err, &validator.ValidationError{}) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	if errors.Is(err, core.ErrDuplicateVariation) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	if errors.Is(err, core.ErrUnknownError) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).WithInternal(err)
	}

	if errors.Is(err, &echo.BindingError{}) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).WithInternal(err)
	}

	return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).WithInternal(err)
}

type CreateResponse struct {
	NewId uint `json:"newId" validate:"required"`
}

func NewCreateResponse(id uint) *CreateResponse {
	return &CreateResponse{
		NewId: id,
	}
}

type BooleanResponse struct {
	Value bool `json:"value" validate:"required"`
}

func NewBooleanResponse(value bool) *BooleanResponse {
	return &BooleanResponse{
		Value: value,
	}
}
