package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get variation properties
// @Description Get all variation properties
// @Produce json
// @Security BearerAuth
// @Success 200 {array} service.VariationPropertyItemDto
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties [get]
func (h *Handler) VariationProperties(c echo.Context) error {
	variationProperties, err := h.VariationPropertyService.GetVariationProperties(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, variationProperties)
}

// @Summary Get variation property
// @Description Get variation property by ID
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Success 200 {object} service.VariationPropertyDto
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id} [get]
func (h *Handler) VariationProperty(c echo.Context) error {
	var propertyID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	variationProperty, err := h.VariationPropertyService.GetVariationProperty(c.Request().Context(), propertyID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, variationProperty)
}

type CreateVariationPropertyRequest struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

// @Summary Create variation property
// @Description Create a new variation property
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param variation_property body CreateVariationPropertyRequest true "Variation property"
// @Success 200 {object} CreateResponse
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties [post]
func (h *Handler) CreateVariationProperty(c echo.Context) error {
	var data CreateVariationPropertyRequest
	if err := c.Bind(&data); err != nil {
		return ToHTTPError(err)
	}

	variationPropertyID, err := h.VariationPropertyService.CreateVariationProperty(c.Request().Context(), service.CreateVariationPropertyParams{
		Name:        data.Name,
		DisplayName: data.DisplayName,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(variationPropertyID))
}

type UpdateVariationPropertyRequest struct {
	DisplayName string `json:"displayName" validate:"required"`
}

// @Summary Update variation property
// @Description Update variation property
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param variation_property body UpdateVariationPropertyRequest true "Variation property"
// @Success 204
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 400 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id} [put]
func (h *Handler) UpdateVariationProperty(c echo.Context) error {
	var data UpdateVariationPropertyRequest
	var propertyID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := c.Bind(&data); err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.UpdateVariationProperty(c.Request().Context(), propertyID, service.UpdateVariationPropertyParams{
		DisplayName: data.DisplayName,
	}); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Delete variation property
// @Description Delete a variation property
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id} [delete]
func (h *Handler) DeleteVariationProperty(c echo.Context) error {
	var propertyID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.DeleteVariationProperty(c.Request().Context(), propertyID); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Check if variation property name is taken
// @Description Check if variation property name is taken
// @Produce json
// @Security BearerAuth
// @Param name path string true "Variation property name"
// @Success 200 {object} BooleanResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/name-taken/{name} [get]
func (h *Handler) IsVariationPropertyNameTaken(c echo.Context) error {
	var name string
	err := echo.PathParamsBinder(c).MustString("name", &name).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	taken, err := h.ValidationService.IsVariationPropertyNameTaken(c.Request().Context(), name)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewBooleanResponse(taken))
}

type CreateVariationPropertyValueRequest struct {
	ParentID uint   `json:"parentId"`
	Value    string `json:"value"`
}

// @Summary Create variation property value
// @Description Create a new variation property value
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param variation_property_value body CreateVariationPropertyValueRequest true "Variation property value"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/values [post]
func (h *Handler) CreateVariationPropertyValue(c echo.Context) error {
	var data CreateVariationPropertyValueRequest
	var propertyID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := c.Bind(&data); err != nil {
		return ToHTTPError(err)
	}

	variationPropertyValueID, err := h.VariationPropertyService.CreateVariationPropertyValue(c.Request().Context(), service.CreateVariationPropertyValueParams{
		PropertyID: propertyID,
		ParentID:   data.ParentID,
		Value:      data.Value,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(variationPropertyValueID))
}

type UpdateVariationPropertyValueOrderRequest struct {
	Order int `json:"order" validate:"required"`
}

// @Summary Update variation property value order
// @Description Update variation property value order
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param value_id path int true "Value ID"
// @Param variation_property_value_order body UpdateVariationPropertyValueOrderRequest true "Variation property value order"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/values/{value_id}/order [put]
func (h *Handler) UpdateVariationPropertyValueOrder(c echo.Context) error {
	var data UpdateVariationPropertyValueOrderRequest
	var propertyID uint
	var valueID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).MustUint("value_id", &valueID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := c.Bind(&data); err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.UpdateVariationPropertyValueOrder(c.Request().Context(), service.UpdateVariationPropertyValueOrderParams{
		PropertyID: propertyID,
		ValueID:    valueID,
		Order:      data.Order,
	}); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Check if variation property name is taken
// @Description Check if variation property name is taken
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param value path string true "Variation property value"
// @Success 200 {object} BooleanResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/value-taken/{value} [get]
func (h *Handler) IsVariationPropertyValueTaken(c echo.Context) error {
	var value string
	var propertyID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).MustString("value", &value).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	taken, err := h.ValidationService.IsVariationPropertyValueTaken(c.Request().Context(), propertyID, value)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewBooleanResponse(taken))
}

// @Summary Delete variation property value
// @Description Delete a variation property value
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param value_id path int true "Value ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/values/{value_id} [delete]
func (h *Handler) DeleteVariationPropertyValue(c echo.Context) error {
	var propertyID uint
	var valueID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).MustUint("value_id", &valueID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.DeleteVariationPropertyValue(c.Request().Context(), service.VariationPropertyValueParams{
		PropertyID: propertyID,
		ValueID:    valueID,
	}); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Archive variation property value
// @Description Archive a variation property value
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param value_id path int true "Value ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/values/{value_id}/archive [put]
func (h *Handler) ArchiveVariationPropertyValue(c echo.Context) error {
	var propertyID uint
	var valueID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).MustUint("value_id", &valueID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.ArchiveVariationPropertyValue(c.Request().Context(), service.VariationPropertyValueParams{
		PropertyID: propertyID,
		ValueID:    valueID,
	}); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Unarchive variation property value
// @Description Unarchive a variation property value
// @Produce json
// @Security BearerAuth
// @Param property_id path int true "Property ID"
// @Param value_id path int true "Value ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /variation-properties/{property_id}/values/{value_id}/unarchive [put]
func (h *Handler) UnarchiveVariationPropertyValue(c echo.Context) error {
	var propertyID uint
	var valueID uint
	err := echo.PathParamsBinder(c).MustUint("property_id", &propertyID).MustUint("value_id", &valueID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if err := h.VariationPropertyService.UnarchiveVariationPropertyValue(c.Request().Context(), service.VariationPropertyValueParams{
		PropertyID: propertyID,
		ValueID:    valueID,
	}); err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
