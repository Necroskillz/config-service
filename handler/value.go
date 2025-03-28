package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/service"
	"github.com/necroskillz/config-service/views/layouts"
	value_views "github.com/necroskillz/config-service/views/values"
)

func (h *Handler) populateValueMatrixViewData(c echo.Context, data *value_views.ValueMatrixData, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow, key db.Key) error {
	data.ServiceVersionID = serviceVersion.ID
	data.FeatureVersionID = featureVersion.ID
	data.Key = key

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	values, err := h.ValueService.GetKeyValues(c.Request().Context(), key.ID, h.User(c).ChangesetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get values for key %d", key.ID)).WithInternal(err)
	}

	filter := service.VariationValueFilter{
		Filter:          map[uint]string{},
		IncludeChildren: false,
		ValueSortOrder:  service.ValueSortOrderTree,
	}

	properties := variationHierarchy.GetProperties(serviceVersion.ServiceTypeID)
	for _, property := range properties {
		filter.Filter[property.ID] = c.QueryParam(property.Name)
	}

	evaluatedValues := variationHierarchy.SortAndFilterValues(serviceVersion.ServiceTypeID, values, filter)

	data.Values = evaluatedValues
	data.Properties = properties

	return nil
}

func (h *Handler) ValueMatrix(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.Key

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	var data value_views.ValueMatrixData
	err = h.populateValueMatrixViewData(c, &data, serviceVersion, featureVersion, key)
	if err != nil {
		return err
	}

	return h.RenderPage(c, http.StatusOK, value_views.ValueMatrixPage(data), fmt.Sprintf("Service %s - Feature %s - Key %s", serviceVersion.ServiceName, featureVersion.FeatureName, key.Name))
}

func (h *Handler) CreateValueSubmit(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.Key

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to create values for this key")
	}

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	var data value_views.ValueFormData
	data.Variation = GetVariationFromForm(c, serviceVersion.ServiceTypeID, variationHierarchy)

	variationIds, err := variationHierarchy.VariationMapToIds(serviceVersion.ServiceTypeID, data.Variation)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to convert variation to ids").WithInternal(err)
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	valid, err := h.BindAndValidate(c, &data,
		h.CollectServiceErrors(func(sec *ServiceErrorCollector) {
			sec.Collect(h.ValidationService.ValidateVariationUniqueness(c.Request().Context(), key.ID, variationIds, changesetID))
		}),
	)
	if err != nil {
		return err
	}

	if !valid {
		data.Properties = variationHierarchy.GetProperties(serviceVersion.ServiceTypeID)
		data.ServiceVersionID = serviceVersion.ID
		data.FeatureVersionID = featureVersion.ID
		data.KeyID = key.ID

		return h.RenderPartial(c, http.StatusUnprocessableEntity, value_views.ValueForm(data))
	}

	err = h.ValueService.CreateValue(c.Request().Context(), service.CreateValueParams{
		ChangesetID:      changesetID,
		KeyID:            key.ID,
		FeatureVersionID: featureVersion.ID,
		Value:            data.Value,
		Variation:        variationIds,
		ServiceVersionID: serviceVersion.ID,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create value").WithInternal(err)
	}

	var matrixData value_views.ValueMatrixData
	err = h.populateValueMatrixViewData(c, &matrixData, serviceVersion, featureVersion, key)
	if err != nil {
		return err
	}

	return h.RenderPartial(c, http.StatusOK, value_views.ValueMatrix(matrixData))
}

func (h *Handler) DeleteValueSubmit(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow
	var key db.Key
	var valueID uint

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	err = echo.PathParamsBinder(c).MustUint("value_id", &valueID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid value id").WithInternal(err)
	}

	if h.User(c).GetPermissionForKey(serviceVersion.ServiceID, featureVersion.FeatureID, key.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "You do not have permission to delete values for this key")
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	err = h.ValueService.DeleteValue(c.Request().Context(), service.DeleteValueParams{
		ChangesetID:      changesetID,
		ServiceVersionID: serviceVersion.ID,
		FeatureVersionID: featureVersion.ID,
		KeyID:            key.ID,
		ValueID:          valueID,
	})

	if err != nil {
		if errors.Is(err, service.ErrCannotDeleteDefaultValue) {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot delete default value")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete value").WithInternal(err)
	}

	var matrixData value_views.ValueMatrixData
	err = h.populateValueMatrixViewData(c, &matrixData, serviceVersion, featureVersion, key)
	if err != nil {
		return err
	}

	return h.RenderPartial(c, http.StatusOK, layouts.Empty())
}
