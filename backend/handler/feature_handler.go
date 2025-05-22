package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get features for a service version
// @Description Get features for a service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 200 {object} []service.FeatureVersionItemDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features [get]
func (h *Handler) Features(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceFeatures, err := h.FeatureService.GetServiceFeatures(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceFeatures)
}

type CreateFeatureRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

// @Summary Create feature
// @Description Create feature
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param createFeatureRequest body CreateFeatureRequest true "Create feature request"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features [post]
func (h *Handler) CreateFeature(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var data CreateFeatureRequest

	err = c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	featureVersionID, err := h.FeatureService.CreateFeature(c.Request().Context(), service.CreateFeatureParams{
		ServiceVersionID: serviceVersionID,
		Name:             data.Name,
		Description:      data.Description,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(featureVersionID))
}

type UpdateFeatureRequest struct {
	Description string `json:"description" validate:"required"`
}

// @Summary Create feature
// @Description Create feature
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param updateFeatureRequest body UpdateFeatureRequest true "Update feature request"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id} [put]
func (h *Handler) UpdateFeature(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var data UpdateFeatureRequest

	err = c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.FeatureService.UpdateFeature(c.Request().Context(), service.UpdateFeatureParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		Description:      data.Description,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Get feature
// @Description Get feature
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 200 {object} service.FeatureVersionDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id} [get]
func (h *Handler) Feature(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	featureVersion, err := h.FeatureService.GetFeatureVersion(c.Request().Context(), serviceVersionID, featureVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, featureVersion)
}

// @Summary Check if feature name is taken
// @Description Check if feature name is taken
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param name path string true "Feature name"
// @Success 200 {object} BooleanResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/name-taken/{name} [get]
func (h *Handler) IsFeatureNameTaken(c echo.Context) error {
	var name string
	err := echo.PathParamsBinder(c).MustString("name", &name).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	exists, err := h.ValidationService.IsFeatureNameTaken(c.Request().Context(), name)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewBooleanResponse(exists))
}

// @Summary Get feature versions
// @Description Get feature versions
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 200 {object} []service.FeatureVersionLinkDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/versions [get]
func (h *Handler) FeatureVersions(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	featureVersions, err := h.FeatureService.GetVersionsOfFeatureForServiceVersion(c.Request().Context(), featureVersionID, serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, featureVersions)
}

// @Summary Get linkable features for a service version
// @Description Get feature versions that can be linked to a service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 200 {object} []service.FeatureVersionDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/linkable [get]
func (h *Handler) LinkableFeatures(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	linkableFeatures, err := h.FeatureService.GetFeatureVersionsLinkableToServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, linkableFeatures)
}

// @Summary Unlink feature version
// @Description Unlink feature version from a service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/unlink [delete]
func (h *Handler) UnlinkFeatureVersion(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.FeatureService.UnlinkFeatureVersion(c.Request().Context(), serviceVersionID, featureVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Link feature version
// @Description Link feature version to a service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 204
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/link [post]
func (h *Handler) LinkFeatureVersion(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.FeatureService.LinkFeatureVersion(c.Request().Context(), serviceVersionID, featureVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Create feature version
// @Description Create feature version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/versions [post]
func (h *Handler) CreateFeatureVersion(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	newId, err := h.FeatureService.CreateFeatureVersion(c.Request().Context(), service.CreateFeatureVersionParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(newId))
}
