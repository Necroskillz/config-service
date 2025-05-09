package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get keys for a feature
// @Description Get keys for a feature
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Success 200 {array} service.KeyItemDto
// @Failure 400 {object} echo.HTTPError
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
// @Failure 400 {object} echo.HTTPError
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

type ValidatorRequest struct {
	ValidatorType db.ValueValidatorType `json:"validatorType" validate:"required"`
	Parameter     string                `json:"parameter"`
	ErrorText     string                `json:"errorText"`
}

type ValidatorRequestList []ValidatorRequest

func (v ValidatorRequestList) ToDto() []service.ValidatorDto {
	validators := make([]service.ValidatorDto, len(v))
	for i, v := range v {
		validators[i] = service.ValidatorDto{
			ValidatorType: v.ValidatorType,
			Parameter:     v.Parameter,
			ErrorText:     v.ErrorText,
		}
	}
	return validators
}

type CreateKeyRequest struct {
	Name         string               `json:"name" validate:"required"`
	Description  string               `json:"description"`
	DefaultValue string               `json:"defaultValue"`
	ValueTypeID  uint                 `json:"valueTypeId" validate:"required"`
	Validators   ValidatorRequestList `json:"validators" validate:"required"`
}

// @Summary Create a key
// @Description Create a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param createKeyRequest body CreateKeyRequest true "Create key request"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
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
		Validators:       data.Validators.ToDto(),
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(keyID))
}

type UpdateKeyRequest struct {
	Description string               `json:"description"`
	Validators  ValidatorRequestList `json:"validators" validate:"required"`
}

// @Summary Create a key
// @Description Create a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param key_id path int true "Key ID"
// @Param updateKeyRequest body UpdateKeyRequest true "Update key request"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id} [put]
func (h *Handler) UpdateKey(c echo.Context) error {
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

	var data UpdateKeyRequest
	err = c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.KeyService.UpdateKey(c.Request().Context(), service.UpdateKeyParams{
		ServiceVersionID: serviceVersionID,
		FeatureVersionID: featureVersionID,
		KeyID:            keyID,
		Description:      data.Description,
		Validators:       data.Validators.ToDto(),
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Delete a key
// @Description Delete a key
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Param feature_version_id path int true "Feature version ID"
// @Param key_id path int true "Key ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/features/{feature_version_id}/keys/{key_id} [delete]
func (h *Handler) DeleteKey(c echo.Context) error {
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

	err = h.KeyService.DeleteKey(c.Request().Context(), serviceVersionID, featureVersionID, keyID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
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
// @Failure 400 {object} echo.HTTPError
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
		MustString("name", &name).
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
