package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get values for a key
// @Description Get values for a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Success 200 {array} service.VariationValue
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values [get]
func (h *Handler) Values(c echo.Context) error {
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

	values, err := h.ValueService.GetKeyValues(c.Request().Context(), serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, values)
}

type ValueRequest struct {
	Data      string          `json:"data" validate:"required"`
	Variation map[uint]string `json:"variation" validate:"required"`
}

// @Summary Create value
// @Description Create value
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Param valueRequest body ValueRequest true "Value request"
// @Success 200 {object} service.NewValueInfo
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values [post]
func (h *Handler) CreateValue(c echo.Context) error {
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

	var data ValueRequest
	err = c.Bind(&data)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to bind form data")
	}

	info, err := h.ValueService.CreateValue(c.Request().Context(), service.CreateValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		KeyID:            keyID,
		Data:             data.Data,
		Variation:        data.Variation,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, info)
}

// @Summary Update value
// @Description Update value
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Param value_id path int true "Value ID"
// @Param valueRequest body ValueRequest true "Value request"
// @Success 200 {object} service.NewValueInfo
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values/{value_id} [put]
func (h *Handler) UpdateValue(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint
	var valueID uint

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		MustUint("key_id", &keyID).
		MustUint("value_id", &valueID).
		BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to bind path params").WithInternal(err)
	}

	var data ValueRequest
	err = c.Bind(&data)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to bind form data")
	}

	info, err := h.ValueService.UpdateValue(c.Request().Context(), service.UpdateValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		KeyID:            keyID,
		ValueID:          valueID,
		Data:             data.Data,
		Variation:        data.Variation,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, info)
}

// @Summary Delete value
// @Description Delete value
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Param value_id path int true "Value ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values/{value_id} [delete]
func (h *Handler) DeleteValue(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint
	var valueID uint

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		MustUint("key_id", &keyID).
		MustUint("value_id", &valueID).
		BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to bind path params").WithInternal(err)
	}

	err = h.ValueService.DeleteValue(c.Request().Context(), service.DeleteValueParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		KeyID:            keyID,
		ValueID:          valueID,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Can operate value with variation
// @Description Can operate value with variation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 409 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values/can-add [get]
func (h *Handler) CanAddValue(c echo.Context) error {
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

	variation, err := GetVariationFromQueryIds(c)
	if err != nil {
		return err
	}

	err = h.ValidationService.CanAddValue(c.Request().Context(), serviceVersionID, featureVersionID, keyID, variation)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Can edit value with variation
// @Description Can edit value with variation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service Version ID"
// @Param feature_version_id path int true "Feature Version ID"
// @Param key_id path int true "Key ID"
// @Param value_id path int true "Value ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 409 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id}/values/{value_id}/can-edit [get]
func (h *Handler) CanEditValue(c echo.Context) error {
	var serviceVersionID uint
	var featureVersionID uint
	var keyID uint
	var valueID uint

	err := echo.PathParamsBinder(c).
		MustUint("service_version_id", &serviceVersionID).
		MustUint("feature_version_id", &featureVersionID).
		MustUint("key_id", &keyID).
		MustUint("value_id", &valueID).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	variation, err := GetVariationFromQueryIds(c)
	if err != nil {
		return err
	}

	err = h.ValidationService.CanEditValue(c.Request().Context(), serviceVersionID, featureVersionID, keyID, valueID, variation)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
