package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get keys for a feature
// @Description Get keys for a feature
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 200 {array} service.KeyDto
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys [get]
func (h *Handler) Keys(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	keys, err := h.KeyService.GetFeatureKeys(c.Request().Context(), serviceVersionID, featureVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, keys)
}

// @Summary Get a key
// @Description Get a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param key_id path int true "Key ID"
// @Success 200 {object} service.KeyDto
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id} [get]
func (h *Handler) Key(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		MustUint("key_id", &keyID).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	key, err := h.KeyService.GetKey(c.Request().Context(), serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, key)
}

type CreateKeyRequest struct {
	Name         string `json:"name" validate:"required"`
	Description  string `json:"description"`
	DefaultValue string `json:"defaultValue"`
	ValueTypeID  uint   `json:"valueTypeId" validate:"required"`
}

// @Summary Create a key
// @Description Create a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param key_dto body CreateKeyRequest true "Key DTO"
// @Success 200 {object} CreateResponse
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys [post]
func (h *Handler) CreateKey(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).MustUint("feature_version_id", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var data CreateKeyRequest
	err = c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	keyID, err := h.KeyService.CreateKey(c.Request().Context(), service.CreateKeyParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		Name:             data.Name,
		Description:      data.Description,
		DefaultValue:     data.DefaultValue,
		ValueTypeID:      data.ValueTypeID,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(keyID))
}

// @Summary Check if key name is taken
// @Description Check if key name is taken
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param name path string true "Key name"
// @Success 200 {object} BooleanResponse
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/name-taken/{name} [get]
func (h *Handler) IsKeyNameTaken(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var name string

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		String("name", &name).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	exists, err := h.ValidationService.IsKeyNameTaken(c.Request().Context(), featureVersionID, name)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewBooleanResponse(exists))
}
