package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/service"
)

// @Summary Get service types
// @Description Get all service types
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} SelectOption
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types [get]
func (h *Handler) GetServiceTypes(c echo.Context) error {
	serviceTypes, err := h.ServiceService.GetServiceTypes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, MakeSelectOptions(serviceTypes, func(item db.ServiceType) (uint, string) {
		return item.ID, item.Name
	}))
}

type VariationProperty struct {
	ID          uint                         `json:"id" validate:"required"`
	Name        string                       `json:"name" validate:"required"`
	DisplayName string                       `json:"displayName" validate:"required"`
	Values      []VariationValueSelectOption `json:"values" validate:"required"`
}

type VariationValueSelectOption struct {
	Value string `json:"value" validate:"required"`
	Depth int    `json:"depth" validate:"required"`
}

// @Summary Get variation properties
// @Description Get variation properties for a service type
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_type_id path uint true "Service type ID"
// @Success 200 {array} VariationProperty
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /service-types/{service_type_id}/variation-properties [get]
func (h *Handler) GetProperties(c echo.Context) error {
	var serviceTypeId uint
	if err := echo.PathParamsBinder(c).Uint("service_type_id", &serviceTypeId).BindError(); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// 404 check
	_, err := h.ServiceService.GetServiceType(c.Request().Context(), serviceTypeId)
	if err != nil {
		return ToHTTPError(err)
	}

	variationHierarchy, err := h.VariationHierarchyService.GetVariationHierarchy(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
	}

	properties := variationHierarchy.GetProperties(serviceTypeId)

	response := []VariationProperty{}

	for _, property := range properties {
		response = append(response, VariationProperty{
			ID:          property.ID,
			Name:        property.Name,
			DisplayName: property.DisplayName,
			Values:      variationSelectOptions(property),
		})
	}

	return c.JSON(http.StatusOK, response)
}

func makeIndentedSelectOptions(indent int, values []*service.VariationHierarchyValue) []VariationValueSelectOption {
	options := []VariationValueSelectOption{}

	for _, value := range values {
		options = append(options, VariationValueSelectOption{
			Value: value.Value,
			Depth: value.Depth,
		})

		if len(value.Children) > 0 {
			options = append(options, makeIndentedSelectOptions(indent+1, value.Children)...)
		}
	}

	return options
}

func variationSelectOptions(property *service.VariationHierarchyProperty) []VariationValueSelectOption {
	options := []VariationValueSelectOption{
		{
			Value: "any",
			Depth: 0,
		},
	}

	options = append(options, makeIndentedSelectOptions(0, property.Values)...)

	return options
}
