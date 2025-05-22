package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get service types
// @Description Get all service types
// @Produce json
// @Security BearerAuth
// @Success 200 {array} service.ServiceTypeDto
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types [get]
func (h *Handler) GetServiceTypes(c echo.Context) error {
	serviceTypes, err := h.ServiceTypeService.GetServiceTypes(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceTypes)
}

// @Summary Get service type
// @Description Get a service type
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Success 200 {object} service.ServiceTypeDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id} [get]
func (h *Handler) GetServiceType(c echo.Context) error {
	var serviceTypeId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).BindError(); err != nil {
		return ToHTTPError(err)
	}

	serviceType, err := h.ServiceTypeService.GetServiceType(c.Request().Context(), serviceTypeId)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceType)
}

// @Summary Get variation properties
// @Description Get variation properties for a service type
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Success 200 {array} service.ServiceTypeVariationPropertyDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id}/variation-properties [get]
func (h *Handler) GetProperties(c echo.Context) error {
	var serviceTypeId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).BindError(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	properties, err := h.VariationPropertyService.GetVariationPropertiesForServiceType(c.Request().Context(), serviceTypeId)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, properties)
}

type CreateServiceTypeRequest struct {
	Name string `json:"name" validate:"required"`
}

// @Summary Create service type
// @Description Create a new service type
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_type_dto body CreateServiceTypeRequest true "Service type"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types [post]
func (h *Handler) CreateServiceType(c echo.Context) error {
	var data CreateServiceTypeRequest
	if err := c.Bind(&data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	serviceTypeID, err := h.ServiceTypeService.CreateServiceType(c.Request().Context(), service.CreateServiceTypeParams{
		Name: data.Name,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(serviceTypeID))
}

// @Summary Delete service type
// @Description Delete a service type
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id} [delete]
func (h *Handler) DeleteServiceType(c echo.Context) error {
	var serviceTypeId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).BindError(); err != nil {
		return ToHTTPError(err)
	}

	err := h.ServiceTypeService.DeleteServiceType(c.Request().Context(), serviceTypeId)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type LinkVariationPropertyToServiceTypeRequest struct {
	VariationPropertyID uint `json:"variation_property_id" validate:"required"`
}

// @Summary Link variation property to service type
// @Description Link a variation property to a service type
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Param link_variation_property_to_service_type_request body LinkVariationPropertyToServiceTypeRequest true "Link variation property to service type request"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id}/variation-properties [post]
func (h *Handler) LinkVariationPropertyToServiceType(c echo.Context) error {
	var serviceTypeId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).BindError(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var data LinkVariationPropertyToServiceTypeRequest
	if err := c.Bind(&data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.ServiceTypeService.LinkVariationPropertyToServiceType(c.Request().Context(), service.LinkVariationPropertyToServiceTypeParams{
		ServiceTypeID:       serviceTypeId,
		VariationPropertyID: data.VariationPropertyID,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Unlink variation property from service type
// @Description Unlink a variation property from a service type
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Param variation_property_id path uint true "Variation property ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id}/variation-properties/{variation_property_id} [delete]
func (h *Handler) UnlinkVariationPropertyFromServiceType(c echo.Context) error {
	var serviceTypeId uint
	var variationPropertyId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).Uint("variation_property_id", &variationPropertyId).BindError(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.ServiceTypeService.UnlinkVariationPropertyToServiceType(c.Request().Context(), service.LinkVariationPropertyToServiceTypeParams{
		ServiceTypeID:       serviceTypeId,
		VariationPropertyID: variationPropertyId,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type UpdateServiceTypeVariationPropertyPriorityRequest struct {
	Priority int `json:"priority" validate:"required"`
}

// @Summary Update service type variation property priority
// @Description Update the priority of a variation property in a service type
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Param variation_property_id path uint true "Variation property ID"
// @Param update_service_type_variation_property_priority_request body UpdateServiceTypeVariationPropertyPriorityRequest true "Update service type variation property priority request"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id}/variation-properties/{variation_property_id}/priority [put]
func (h *Handler) UpdateServiceTypeVariationPropertyPriority(c echo.Context) error {
	var serviceTypeId uint
	var variationPropertyId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).Uint("variation_property_id", &variationPropertyId).BindError(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var data UpdateServiceTypeVariationPropertyPriorityRequest
	if err := c.Bind(&data); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	err := h.ServiceTypeService.UpdateServiceTypeVariationPropertyPriority(c.Request().Context(), service.UpdateServiceTypeVariationPropertyPriorityParams{
		ServiceTypeID:       serviceTypeId,
		VariationPropertyID: variationPropertyId,
		Priority:            data.Priority,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
