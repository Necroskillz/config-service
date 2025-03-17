package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	value_views "github.com/necroskillz/config-service/views/values"
)

func (h *Handler) ValueMatrix(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint

	err := echo.PathParamsBinder(c).Uint("service_version_id", &serviceVersionID).Uint("feature_version_id", &featureVersionID).Uint("key_version_id", &keyID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid service version ID or feature version ID or key ID")
	}

	values, err := h.ValueService.GetKeyValues(c.Request().Context(), keyID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get values for key %d", keyID)).WithInternal(err)
	}

	fmt.Println(values)

	data := value_views.ValueMatrixData{}

	return h.RenderPartial(c, http.StatusOK, value_views.ValueMatrix(data))
}
