package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/db"
	views "github.com/necroskillz/config-service/views/services"
)

func (h *Handler) Services(c echo.Context) error {
	serviceVersions, err := h.ServiceService.GetCurrentServiceVersions(c.Request().Context(), h.User(c).ChangesetID)
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

	data.ServiceTypeOptions = MakeSelectOptions(serviceTypes, func(item db.ServiceType) (uint, string) {
		return item.ID, item.Name
	})

	return nil
}

func (h *Handler) CreateService(c echo.Context) error {
	if !h.User(c).IsGlobalAdmin {
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to create services")
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
		return echo.NewHTTPError(http.StatusForbidden, "You are not authorized to create services")
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

	changesetId, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	serviceId, err := h.ServiceService.CreateService(c.Request().Context(), data.Name, data.Description, data.ServiceTypeID, changesetId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create service").WithInternal(err)
	}

	return Redirect(c, fmt.Sprintf("/services/%d", serviceId))
}

func (h *Handler) ServiceDetail(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow

	err := h.LoadBasicData(c, &serviceVersion)
	if err != nil {
		return err
	}

	serviceFeatures, err := h.FeatureService.GetServiceFeatures(c.Request().Context(), serviceVersion.ID, h.User(c).ChangesetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get features for service version %d", serviceVersion.ID).WithInternal(err)
	}

	allServiceVersions, err := h.ServiceService.GetServiceVersions(c.Request().Context(), serviceVersion.ServiceID, h.User(c).ChangesetID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get service versions for service %d", serviceVersion.ServiceID).WithInternal(err)
	}

	data := views.ServiceDetailData{
		ServiceVersion:       serviceVersion,
		ServiceFeatures:      serviceFeatures,
		OtherServiceVersions: allServiceVersions,
	}

	return h.RenderPage(c, http.StatusOK, views.ServiceDetailPage(data), fmt.Sprintf("Service %s", serviceVersion.ServiceName))
}
