package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/service"
	feature_views "github.com/necroskillz/config-service/views/features"
)

func (h *Handler) populateCreateFeatureViewData(data *feature_views.CreateFeatureData, serviceVersion *model.ServiceVersion) {
	data.ServiceVersion = serviceVersion
}

func (h *Handler) CreateFeature(c echo.Context) error {
	var serviceVersion model.ServiceVersion

	err := h.LoadBasicData(c, &serviceVersion)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForService(serviceVersion.Service.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create features for service %s", serviceVersion.Service.Name)
	}

	data := feature_views.CreateFeatureData{}

	h.populateCreateFeatureViewData(&data, &serviceVersion)

	return h.RenderPage(c, http.StatusOK, feature_views.CreateFeaturePage(data), "Create New Feature")
}

func (h *Handler) CreateFeatureSubmit(c echo.Context) error {
	var serviceVersion model.ServiceVersion

	err := h.LoadBasicData(c, &serviceVersion)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForService(serviceVersion.Service.ID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create features for service %s", serviceVersion.Service.Name)
	}

	var data feature_views.CreateFeatureData

	valid, err := h.BindAndValidate(c, &data)
	if err != nil {
		return err
	}

	if !valid {
		h.populateCreateFeatureViewData(&data, &serviceVersion)

		return h.RenderPartial(c, http.StatusUnprocessableEntity, feature_views.CreateFeatureForm(data))
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	featureVersion, err := h.FeatureService.CreateFeature(c.Request().Context(), service.CreateFeatureParams{
		ChangesetID:      changesetID,
		ServiceVersionID: serviceVersion.ID,
		Name:             data.Name,
		Description:      data.Description,
		ServiceID:        serviceVersion.Service.ID,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create feature").WithInternal(err)
	}

	return Redirect(c, fmt.Sprintf("/services/%d/features/%d", serviceVersion.ID, featureVersion.ID))
}

func (h *Handler) FeatureDetail(c echo.Context) error {
	var serviceVersion model.ServiceVersion
	var featureVersion model.FeatureVersion

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion)
	if err != nil {
		return err
	}

	allFeatureVersions, err := h.FeatureService.GetFeatureVersionsLinkedToServiceVersion(c.Request().Context(), featureVersion.FeatureID, serviceVersion.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get feature versions").WithInternal(err)
	}

	keys, err := h.KeyService.GetFeatureKeys(c.Request().Context(), featureVersion.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get keys").WithInternal(err)
	}

	data := feature_views.FeatureDetailData{
		ServiceVersion:       &serviceVersion,
		FeatureVersion:       &featureVersion,
		OtherFeatureVersions: allFeatureVersions,
		Keys:                 keys,
	}

	return h.RenderPage(c, http.StatusOK, feature_views.FeatureDetailPage(data), fmt.Sprintf("Service %s - Feature %s", serviceVersion.Service.Name, featureVersion.Feature.Name))
}
