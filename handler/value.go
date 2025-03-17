package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
	value_views "github.com/necroskillz/config-service/views/values"
)

func (h *Handler) ValueMatrix(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint

	err := echo.PathParamsBinder(c).Uint("service_version_id", &serviceVersionID).Uint("feature_version_id", &featureVersionID).Uint("key_id", &keyID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid service version ID or feature version ID or key ID")
	}

	serviceVersion, err := h.ServiceService.GetServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get service version").WithInternal(err)
	}

	if serviceVersion == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Service version not found")
	}

	featureVersion, err := h.FeatureService.GetFeatureVersion(c.Request().Context(), featureVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get feature version").WithInternal(err)
	}

	if featureVersion == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Feature version not found")
	}

	key, err := h.KeyService.GetKey(c.Request().Context(), keyID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get key").WithInternal(err)
	}

	if key == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Key not found")
	}

	values, err := h.ValueService.GetKeyValues(c.Request().Context(), keyID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get values for key %d", keyID)).WithInternal(err)
	}

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	properties := variationHierarchy.GetProperties(serviceVersion.Service.ServiceTypeID)

	filter := service.VariationValueFilter{
		Filter:          map[string]string{},
		IncludeChildren: false,
	}

	for _, property := range properties {
		filter.Filter[property.Name] = c.QueryParam(property.Name)
	}

	evaluatedValues := variationHierarchy.SortAndFilterValues(serviceVersion.Service.ServiceTypeID, values, filter)

	data := value_views.ValueMatrixData{
		Key:        key,
		Properties: properties,
		Values:     evaluatedValues,
	}

	return h.RenderPage(c, http.StatusOK, value_views.ValueMatrix(data), fmt.Sprintf("Service %s - Feature %s - Key %s", serviceVersion.Service.Name, featureVersion.Feature.Name, key.Name))
}
