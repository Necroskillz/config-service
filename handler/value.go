package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/service"
	value_views "github.com/necroskillz/config-service/views/values"
)

func (h *Handler) populateValueMatrixViewData(c echo.Context, data *value_views.ValueMatrixData, serviceVersion *model.ServiceVersion, featureVersion *model.FeatureVersion, key *model.Key) error {
	data.ServiceVersionID = serviceVersion.ID
	data.FeatureVersionID = featureVersion.ID
	data.Key = key

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	values, err := h.ValueService.GetKeyValues(c.Request().Context(), key.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to get values for key %d", key.ID)).WithInternal(err)
	}

	filter := service.VariationValueFilter{
		Filter:          map[string]string{},
		IncludeChildren: false,
	}

	properties := variationHierarchy.GetProperties(serviceVersion.Service.ServiceTypeID)
	for _, property := range properties {
		filter.Filter[property.Name] = c.QueryParam(property.Name)
	}

	evaluatedValues := variationHierarchy.SortAndFilterValues(serviceVersion.Service.ServiceTypeID, values, filter)

	data.Values = evaluatedValues
	data.Properties = properties

	return nil
}

func (h *Handler) ValueMatrix(c echo.Context) error {
	var serviceVersion model.ServiceVersion
	var featureVersion model.FeatureVersion
	var key model.Key

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	var data value_views.ValueMatrixData
	err = h.populateValueMatrixViewData(c, &data, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	return h.RenderPage(c, http.StatusOK, value_views.ValueMatrixPage(data), fmt.Sprintf("Service %s - Feature %s - Key %s", serviceVersion.Service.Name, featureVersion.Feature.Name, key.Name))
}

func (h *Handler) CreateValueSubmit(c echo.Context) error {
	var serviceVersion model.ServiceVersion
	var featureVersion model.FeatureVersion
	var key model.Key

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForKey(serviceVersion.Service.ServiceTypeID, featureVersion.Feature.ID, key.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You do not have permission to create values for this key")
	}

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	var data value_views.ValueFormData
	data.Variation = GetVariationFromForm(c, serviceVersion.Service.ServiceTypeID, variationHierarchy)

	variationIds, err := variationHierarchy.VariationMapToIds(serviceVersion.Service.ServiceTypeID, data.Variation)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to convert variation to ids").WithInternal(err)
	}

	valid, err := h.BindAndValidate(c, &data,
		h.CollectServiceErrors(func(sec *ServiceErrorCollector) {
			sec.Collect(h.ValidationService.ValidateVariationUniqueness(c.Request().Context(), key.ID, variationIds))
		}),
	)
	if err != nil {
		return err
	}

	if !valid {
		data.Properties = variationHierarchy.GetProperties(serviceVersion.Service.ServiceTypeID)
		data.ServiceVersionID = serviceVersion.ID
		data.FeatureVersionID = featureVersion.ID
		data.KeyID = key.ID

		return h.RenderPartial(c, http.StatusUnprocessableEntity, value_views.ValueForm(data))
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	err = h.ValueService.CreateValue(c.Request().Context(), service.CreateValueParams{
		ChangesetID:      changesetID,
		KeyID:            key.ID,
		FeatureVersionID: featureVersion.ID,
		Value:            data.Value,
		Variation:        variationIds,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create value").WithInternal(err)
	}

	var matrixData value_views.ValueMatrixData
	err = h.populateValueMatrixViewData(c, &matrixData, &serviceVersion, &featureVersion, &key)
	if err != nil {
		return err
	}

	return h.RenderPartial(c, http.StatusOK, value_views.ValueMatrix(matrixData))
}
