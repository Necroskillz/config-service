package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

func getVariation(serviceTypeID uint, variationHierarchy *service.VariationHierarchy, getter func(string) string) map[string]string {
	properties := variationHierarchy.GetProperties(serviceTypeID)
	variation := make(map[string]string)

	for _, property := range properties {
		propertyValue := getter(property.Name)
		if propertyValue != "" && propertyValue != "any" {
			variation[property.Name] = propertyValue
		}
	}

	return variation
}

func GetVariationFromForm(c echo.Context, serviceTypeID uint, variationHierarchy *service.VariationHierarchy) map[string]string {
	return getVariation(serviceTypeID, variationHierarchy, func(name string) string {
		return c.FormValue(name)
	})
}

func GetVariationFromQuery(c echo.Context, serviceTypeID uint, variationHierarchy *service.VariationHierarchy) map[string]string {
	return getVariation(serviceTypeID, variationHierarchy, func(name string) string {
		return c.QueryParam(name)
	})
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
	if errors.Is(err, service.ErrPermissionDenied) {
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	}

	if errors.Is(err, service.ErrRecordNotFound) {
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	}

	if errors.Is(err, service.ErrInvalidOperation) {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if errors.Is(err, service.ErrInvalidInput) {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	if errors.Is(err, service.ErrDuplicateVariation) {
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	if errors.Is(err, service.ErrUnknownError) {
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
