package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/variation"
	"github.com/necroskillz/config-service/util/validator"
)

func getVariation(variationHierarchy *variation.Hierarchy, getter func(string) string) map[string]string {
	properties := variationHierarchy.GetAllProperties()
	variation := make(map[string]string)

	for _, property := range properties {
		propertyValue := getter(property.Name)
		if propertyValue != "" && propertyValue != "any" {
			variation[property.Name] = propertyValue
		}
	}

	return variation
}

func (h *Handler) GetVariationFromForm(c echo.Context) (map[string]string, error) {
	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return nil, err
	}

	return getVariation(variationHierarchy, func(name string) string {
		return c.FormValue(name)
	}), nil
}

func (h *Handler) GetVariationFromQuery(c echo.Context) (map[string]string, error) {
	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return nil, err
	}

	return getVariation(variationHierarchy, func(name string) string {
		return c.QueryParam(name)
	}), nil
}

func GetVariationFromQueryIds(c echo.Context) (map[uint]string, error) {
	variation := make(map[uint]string)

	for key, value := range c.QueryParams() {
		propertyID, err := strconv.ParseUint(key, 10, 32)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid property ID %s", key)).WithInternal(err)
		}

		if len(value) != 1 {
			return nil, echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid value %s", value)).WithInternal(err)
		}

		variation[uint(propertyID)] = value[0]
	}

	return variation, nil
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
