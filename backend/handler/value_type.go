package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get value types
// @Description Get all value types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} service.ValueTypeDto
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
