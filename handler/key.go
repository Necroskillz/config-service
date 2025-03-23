package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/service"
	key_views "github.com/necroskillz/config-service/views/keys"
)

func (h *Handler) populateCreateKeyViewData(c echo.Context, data *key_views.CreateKeyData, serviceVersion db.GetServiceVersionRow, featureVersion db.GetFeatureVersionRow) error {
	data.ServiceVersion = serviceVersion
	data.FeatureVersion = featureVersion

	valueTypes, err := h.KeyService.GetValueTypes(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get value types").WithInternal(err)
	}

	data.ValueTypeOptions = MakeSelectOptions(valueTypes, func(item db.ValueType) (uint, string) {
		return item.ID, item.Name
	})

	return nil
}

func (h *Handler) CreateKey(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForFeature(serviceVersion.ServiceID, featureVersion.FeatureID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create keys for feature %s", featureVersion.FeatureName)
	}

	data := key_views.CreateKeyData{}

	err = h.populateCreateKeyViewData(c, &data, serviceVersion, featureVersion)
	if err != nil {
		return err
	}

	return h.RenderPage(c, http.StatusOK, key_views.CreateKeyPage(data), fmt.Sprintf("Service %s - Feature %s - Create New Key", serviceVersion.ServiceName, featureVersion.FeatureName))
}

func (h *Handler) CreateKeySubmit(c echo.Context) error {
	var serviceVersion db.GetServiceVersionRow
	var featureVersion db.GetFeatureVersionRow

	err := h.LoadBasicData(c, &serviceVersion, &featureVersion)
	if err != nil {
		return err
	}

	if h.User(c).GetPermissionForFeature(serviceVersion.ServiceID, featureVersion.FeatureID) != constants.PermissionAdmin {
		return echo.NewHTTPError(http.StatusUnauthorized, "You are not authorized to create keys for feature %s", featureVersion.FeatureName)
	}

	var data key_views.CreateKeyData

	valid, err := h.BindAndValidate(c, &data,
		h.CollectServiceErrors(func(sec *ServiceErrorCollector) {
			sec.Collect(h.ValidationService.ValidateKeyNameUniqueness(c.Request().Context(), featureVersion.ID, data.Name))
		}),
	)
	if err != nil {
		return err
	}

	if !valid {
		err = h.populateCreateKeyViewData(c, &data, serviceVersion, featureVersion)
		if err != nil {
			return err
		}

		return h.RenderPartial(c, http.StatusUnprocessableEntity, key_views.CreateKeyForm(data))
	}

	changesetID, err := h.EnsureChangesetID(c)
	if err != nil {
		return err
	}

	keyID, err := h.KeyService.CreateKey(c.Request().Context(), service.CreateKeyParams{
		ChangesetID:      changesetID,
		FeatureVersionID: featureVersion.ID,
		Name:             data.Name,
		Description:      data.Description,
		DefaultValue:     data.DefaultValue,
		ValueTypeID:      data.ValueTypeID,
		ServiceVersionID: serviceVersion.ID,
	})

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create feature").WithInternal(err)
	}

	return Redirect(c, fmt.Sprintf("/services/%d/features/%d/keys/%d/values", serviceVersion.ID, featureVersion.ID, keyID))
}
