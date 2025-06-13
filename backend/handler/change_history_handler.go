package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/changeset"
	_ "github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/util/ptr"
)

// @Summary Get change history
// @Description Get change history for a service, feature, key, or variation
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page"
// @Param pageSize query int false "Page size"
// @Param serviceId query uint false "Service ID"
// @Param serviceVersionId query uint false "Service version ID"
// @Param featureId query uint false "Feature ID"
// @Param featureVersionId query uint false "Feature version ID"
// @Param keyName query string false "Key name"
// @Param applyVariation query bool false "Apply variation"
// @Param kinds[] query []string false "Kinds" collectionFormat(multi)
// @Param variation[] query []string false "Variation" collectionFormat(multi)
// @Success 200 {object} core.PaginatedResult[changeset.ChangeHistoryItemDto]
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history [get]
func (h *Handler) GetChangeHistory(c echo.Context) error {
	page := 1
	pageSize := 20
	var serviceID uint
	var serviceVersionID uint
	var featureID uint
	var featureVersionID uint
	var keyName string
	var applyVariation bool
	var variation map[uint]string
	var kinds []string
	var from time.Time
	var to time.Time

	err := echo.QueryParamsBinder(c).
		Int("page", &page).
		Int("pageSize", &pageSize).
		Uint("serviceId", &serviceID).
		Uint("serviceVersionId", &serviceVersionID).
		Uint("featureId", &featureID).
		Uint("featureVersionId", &featureVersionID).
		String("keyName", &keyName).
		Bool("applyVariation", &applyVariation).
		Strings("kinds[]", &kinds).
		Time("from", &from, time.RFC3339).
		Time("to", &to, time.RFC3339).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	if applyVariation {
		variation, err = h.GetVariationFromQuery(c)
		if err != nil {
			return ToHTTPError(err)
		}
	}

	changeHistory, err := h.ChangesetService.GetChangeHistory(c.Request().Context(), changeset.GetChangeHistoryFilter{
		ServiceID:        ptr.To(serviceID, ptr.NilIfZero()),
		ServiceVersionID: ptr.To(serviceVersionID, ptr.NilIfZero()),
		FeatureID:        ptr.To(featureID, ptr.NilIfZero()),
		FeatureVersionID: ptr.To(featureVersionID, ptr.NilIfZero()),
		KeyName:          ptr.To(keyName, ptr.NilIfZero()),
		From:             ptr.To(from, ptr.NilIfZero()),
		To:               ptr.To(to, ptr.NilIfZero()),
		Variation:        variation,
		Kinds:            kinds,
		Page:             page,
		PageSize:         pageSize,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, changeHistory)
}

// @Summary Get applied services
// @Description Get applied services
// @Produce json
// @Security BearerAuth
// @Success 200 {object} []service.AppliedServiceDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history/services [get]
func (h *Handler) GetAppliedServices(c echo.Context) error {
	services, err := h.ServiceService.GetAppliedServices(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, services)
}

// @Summary Get applied service versions
// @Description Get applied service versions
// @Produce json
// @Security BearerAuth
// @Param service_id path uint true "Service ID"
// @Success 200 {object} []service.ServiceVersionLinkDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history/services/{service_id}/versions [get]
func (h *Handler) GetAppliedServiceVersions(c echo.Context) error {
	var serviceID uint
	err := echo.PathParamsBinder(c).MustUint("service_id", &serviceID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersions, err := h.ServiceService.GetAppliedServiceVersionsForService(c.Request().Context(), serviceID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, serviceVersions)
}

// @Summary Get features
// @Description Get features
// @Produce json
// @Security BearerAuth
// @Param serviceId query uint false "Service ID"
// @Param serviceVersionId query uint false "Service version ID"
// @Success 200 {object} []feature.FeatureDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history/features [get]
func (h *Handler) GetAppliedFeatures(c echo.Context) error {
	var serviceID uint
	var serviceVersionID uint
	err := echo.QueryParamsBinder(c).Uint("serviceId", &serviceID).Uint("serviceVersionId", &serviceVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	features, err := h.FeatureService.GetAppliedFeatures(c.Request().Context(), ptr.To(serviceID, ptr.NilIfZero()), ptr.To(serviceVersionID, ptr.NilIfZero()))
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, features)
}

// @Summary Get applied feature versions
// @Description Get applied feature versions
// @Produce json
// @Security BearerAuth
// @Param feature_id path uint true "Feature ID"
// @Success 200 {object} []feature.AppliedFeatureVersionDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history/features/{feature_id}/versions [get]
func (h *Handler) GetAppliedFeatureVersions(c echo.Context) error {
	var featureID uint
	err := echo.PathParamsBinder(c).MustUint("feature_id", &featureID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	featureVersions, err := h.FeatureService.GetAppliedFeatureVersionsForFeature(c.Request().Context(), featureID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, featureVersions)
}

// @Summary Get applied keys
// @Description Get applied keys
// @Produce json
// @Security BearerAuth
// @Param featureId query uint false "Feature ID"
// @Param featureVersionId query uint false "Feature version ID"
// @Success 200 {object} []key.AppliedKeyDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /change-history/keys [get]
func (h *Handler) GetAppliedKeys(c echo.Context) error {
	var featureID uint
	var featureVersionID uint
	err := echo.QueryParamsBinder(c).Uint("featureId", &featureID).Uint("featureVersionId", &featureVersionID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	keys, err := h.KeyService.GetAppliedKeys(c.Request().Context(), ptr.To(featureVersionID, ptr.NilIfZero()), ptr.To(featureID, ptr.NilIfZero()))
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, keys)
}
