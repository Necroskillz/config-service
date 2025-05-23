package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get value types
// @Description Get all value types
// @Produce json
// @Security BearerAuth
// @Success 200 {array} valuetype.ValueTypeDto
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /value-types [get]
func (h *Handler) GetValueTypes(c echo.Context) error {
	valueTypes, err := h.ValueTypeService.GetValueTypes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get value types").WithInternal(err)
	}

	return c.JSON(http.StatusOK, valueTypes)
}

// @Summary Get value type
// @Description Get a value type by ID
// @Produce json
// @Security BearerAuth
// @Param value_type_id path uint true "Value type ID"
// @Success 200 {object} valuetype.ValueTypeDto
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /value-types/{value_type_id} [get]
func (h *Handler) GetValueType(c echo.Context) error {
	var id uint
	err := echo.PathParamsBinder(c).MustUint("value_type_id", &id).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	valueType, err := h.ValueTypeService.GetValueType(c.Request().Context(), id)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, valueType)
}
