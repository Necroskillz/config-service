package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/model"
	views "github.com/necroskillz/config-service/views/services"
)

func (h *Handler) Services(c echo.Context) error {
	serviceVersions, err := h.ServiceService.GetCurrentServiceVersions(c.Request().Context())
	if err != nil {
		return err
	}

	data := views.ServiceListData{
		ServiceVersions: serviceVersions,
	}

	return h.RenderPage(c, http.StatusOK, views.ServiceList(data), "Services")
}

func (h *Handler) createServiceViewData(c echo.Context, data *views.CreateServiceData) error {
	serviceTypes, err := h.ServiceService.GetServiceTypes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get service types").WithInternal(err)
	}

	data.ServiceTypeOptions = MakeSelectOptions(serviceTypes, func(item model.ServiceType) (uint, string) {
		return item.ID, item.Name
	})

	return nil
}

func (h *Handler) CreateService(c echo.Context) error {
	if !h.User(c).IsGlobalAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create services")
	}

	data := views.CreateServiceData{}

	err := h.createServiceViewData(c, &data)
	if err != nil {
		return err
	}

	return h.RenderPage(c, http.StatusOK, views.CreateServicePage(data), "Create New Service")
}

func (h *Handler) CreateServiceSubmit(c echo.Context) error {
	if !h.User(c).IsGlobalAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create services")
	}

	data := views.CreateServiceData{}

	valid, err := h.BindAndValidate(c, &data)
	if err != nil {
		return err
	}

	if !valid {
		err := h.createServiceViewData(c, &data)
		if err != nil {
			return err
		}

		return h.RenderPartial(c, http.StatusUnprocessableEntity, views.CreateServiceForm(data))
	}

	err = h.ServiceService.CreateService(c.Request().Context(), data.Name, data.Description, data.ServiceTypeID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create service").WithInternal(err)
	}

	return Redirect(c, "/services")
}

func (h *Handler) ServiceDetail(c echo.Context) error {
	var serviceVersionID uint

	err := echo.PathParamsBinder(c).Uint("service_version_id", &serviceVersionID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid service version ID")
	}

	serviceVersion, err := h.ServiceService.GetServiceVersion(c.Request().Context(), serviceVersionID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get service version %d", serviceVersionID).WithInternal(err)
	}

	if serviceVersion == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Service version not found")
	}

	serviceFeatures, err := h.FeatureService.GetServiceFeatures(c.Request().Context(), serviceVersion.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get features for service version %d", serviceVersionID).WithInternal(err)
	}

	allServiceVersions, err := h.ServiceService.GetServiceVersions(c.Request().Context(), serviceVersion.ServiceID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get service versions for service %d", serviceVersion.ServiceID).WithInternal(err)
	}

	data := views.ServiceDetailData{
		ServiceVersion:       serviceVersion,
		ServiceFeatures:      serviceFeatures,
		OtherServiceVersions: allServiceVersions,
	}

	return h.RenderPage(c, http.StatusOK, views.ServiceDetailPage(data), fmt.Sprintf("Service %s", serviceVersion.Service.Name))
}
