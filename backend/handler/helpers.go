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

func (h *Handler) parseVariationParams(c echo.Context, keyFunc func(string) (uint, error)) (map[uint]string, error) {
	variation := make(map[uint]string)
	var variationParams []string

	err := echo.QueryParamsBinder(c).Strings("variation[]", &variationParams).BindError()
	if err != nil {
		return nil, err
	}

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
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

		propertyID, err := keyFunc(key)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		_, err = variationHierarchy.GetPropertyValue(propertyID, value)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		// Check for duplicate keys
		if _, exists := variation[propertyID]; exists {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Duplicate variation parameter key: '%s'", key))
		}

		variation[propertyID] = value
	}

	return variation, nil
}

func (h *Handler) GetVariationFromQuery(c echo.Context) (map[uint]string, error) {
	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return nil, err
	}

	return h.parseVariationParams(c, func(name string) (uint, error) {
		propertyID, err := variationHierarchy.GetPropertyID(name)
		if err != nil {
			return 0, err
		}

		return propertyID, nil
	})
}

func (h *Handler) GetVariationFromQueryIds(c echo.Context) (map[uint]string, error) {
	return h.parseVariationParams(c, func(id string) (uint, error) {
		propertyID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			return 0, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid property ID %s", id)).WithInternal(err)
		}

		return uint(propertyID), nil
	})
}

func ToHTTPError(err error) *echo.HTTPError {
	if errors.Is(err, core.ErrPermissionDenied) {
		return echo.NewHTTPError(http.StatusForbidden, err.Error()).WithInternal(err)
	}

	if errors.Is(err, core.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error()).WithInternal(err)
	}

	if errors.Is(err, core.ErrInvalidOperation) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).WithInternal(err)
	}

	var validationError *validator.ValidationError
	if errors.As(err, &validationError) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, validationError.Error()).WithInternal(err)
	}

	if errors.Is(err, core.ErrInvalidInput) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error()).WithInternal(err)
	}

	if errors.Is(err, core.ErrDuplicateVariation) {
		return echo.NewHTTPError(http.StatusConflict, err.Error()).WithInternal(err)
	}

	if errors.Is(err, core.ErrUnexpectedError) {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).WithInternal(err)
	}

	var bindingError *echo.BindingError
	if errors.As(err, &bindingError) {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Parameter %s is invalid: %s", bindingError.Field, bindingError.Message)).WithInternal(err)
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
