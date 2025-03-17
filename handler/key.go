package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/service"
	key_views "github.com/necroskillz/config-service/views/keys"
)

func (h *Handler) populateCreateKeyViewData(c echo.Context, data *key_views.CreateKeyData, serviceVersion *model.ServiceVersion, featureVersion *model.FeatureVersion) error {
	data.ServiceVersion = serviceVersion
	data.FeatureVersion = featureVersion

	valueTypes, err := h.KeyService.GetValueTypes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get value types").WithInternal(err)
	}

	data.ValueTypeOptions = MakeSelectOptions(valueTypes, func(item model.ValueType) (uint, string) {
		return item.ID, item.Name
	})

	return nil
}

func (h *Handler) CreateKey(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint

	err := echo.PathParamsBinder(c).Uint("service_version_id", &serviceVersionID).Uint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid service version ID or feature version ID")
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

	if h.User(c).GetPermissionForFeature(serviceVersion.Service.ID, featureVersion.Feature.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create keys for feature %s", featureVersion.Feature.Name)
	}

	data := key_views.CreateKeyData{}

	err = h.populateCreateKeyViewData(c, &data, serviceVersion, featureVersion)
	if err != nil {
		return err
	}

	return h.RenderPage(c, http.StatusOK, key_views.CreateKeyPage(data), fmt.Sprintf("Service %s - Feature %s - Create New Key", serviceVersion.Service.Name, featureVersion.Feature.Name))
}

func (h *Handler) CreateKeySubmit(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint

	err := echo.PathParamsBinder(c).Uint("service_version_id", &serviceVersionID).Uint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid service version ID or feature version ID")
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

	if h.User(c).GetPermissionForFeature(serviceVersion.Service.ID, featureVersion.Feature.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create keys for feature %s", featureVersion.Feature.Name)
	}

	var data key_views.CreateKeyData

	valid, err := h.BindAndValidate(c, &data,
		h.CollectServiceErrors(func(sec *ServiceErrorCollector) {
			sec.Collect(h.ValidationService.ValidateKeyNameUniqueness(c.Request().Context(), featureVersionID, data.Name))
		}),
	)
	if err != nil {
		return err
	}

	if !valid {
		err = h.populateCreateKeyViewData(c, &data, serviceVersion, featureVersion)
		if err != nil {
			return err
		}

		return h.RenderPartial(c, http.StatusUnprocessableEntity, key_views.CreateKeyForm(data))
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	err = h.KeyService.CreateKey(c.Request().Context(), service.CreateKeyParams{
		ChangesetID:      changesetID,
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		Name:             data.Name,
		Description:      data.Description,
		DefaultValue:     data.DefaultValue,
		ValueTypeID:      data.ValueTypeID,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create feature").WithInternal(err)
	}

	return Redirect(c, fmt.Sprintf("/services/%d/features/%d", serviceVersionID, featureVersionID))
}
