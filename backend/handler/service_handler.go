package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/service"
)

// @Summary Get services
// @Description Get list of services
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []service.ServiceDto
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services [get]
func (h *Handler) Services(c echo.Context) error {
	serviceVersions, err := h.ServiceService.GetServices(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceVersions)
}

type CreateServiceRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description" validate:"required"`
	ServiceTypeID uint   `json:"serviceTypeId" validate:"required"`
}

// @Summary Create service
// @Description Create service
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param createServiceRequest body CreateServiceRequest true "Create service request"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services [post]
func (h *Handler) CreateService(c echo.Context) error {
	var data CreateServiceRequest
	err := c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	serviceId, err := h.ServiceService.CreateService(c.Request().Context(), service.CreateServiceParams{
		Name:          data.Name,
		Description:   data.Description,
		ServiceTypeID: data.ServiceTypeID,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(serviceId))
}

type UpdateServiceRequest struct {
	Description string `json:"description" validate:"required"`
}

// @Summary Update service
// @Description Update service
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID (for url consistency, the underling service will be updated)"
// @Param updateServiceRequest body UpdateServiceRequest true "Update service request"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id} [put]
func (h *Handler) UpdateService(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var data UpdateServiceRequest
	err = c.Bind(&data)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ServiceService.UpdateService(c.Request().Context(), service.UpdateServiceParams{
		ServiceVersionID: serviceVersionID,
		Description:      data.Description,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Publish service version
// @Description Publish service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/publish [put]
func (h *Handler) PublishServiceVersion(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ServiceService.PublishServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Get service
// @Description Get service
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 200 {object} service.ServiceVersionDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id} [get]
func (h *Handler) Service(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersion, err := h.ServiceService.GetServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceVersion)
}

// @Summary Get service versions
// @Description Get service versions
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 200 {object} []service.ServiceVersionLinkDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/versions [get]
func (h *Handler) ServiceVersions(c echo.Context) error {
	var serviceVersionID uint

	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersions, err := h.ServiceService.GetServiceVersionsForServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceVersions)
}

// @Summary Create service version
// @Description Create service version
// @Produce json
// @Security BearerAuth
// @Param service_version_id path int true "Service version ID"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/{service_version_id}/versions [post]
func (h *Handler) CreateServiceVersion(c echo.Context) error {
	var serviceVersionID uint
	err := echo.PathParamsBinder(c).MustUint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersionID, err = h.ServiceService.CreateServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(serviceVersionID))
}

// @Summary Check if service name is taken
// @Description Check if service name is taken
// @Produce json
// @Security BearerAuth
// @Param name path string true "Service name"
// @Success 200 {object} BooleanResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /services/name-taken/{name} [get]
func (h *Handler) IsServiceNameTaken(c echo.Context) error {
	var name string
	err := echo.PathParamsBinder(c).MustString("name", &name).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	exists, err := h.ValidationService.IsServiceNameTaken(c.Request().Context(), name)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewBooleanResponse(exists))
}
