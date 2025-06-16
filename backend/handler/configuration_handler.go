package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/configuration"
	"github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/util/ptr"
)

// @Summary Get next changesets
// @Description Get next changesets
// @Produce json
// @Param after query uint true "After changeset ID"
// @Param services[] query []string true "Service versions"
// @Success 200 {array} []uint
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /configuration/changesets [get]
func (h *Handler) GetNextChangesets(c echo.Context) error {
	var afterChangesetID uint
	var serviceVersions []string
	err := echo.QueryParamsBinder(c).MustUint("after", &afterChangesetID).MustStrings("services[]", &serviceVersions).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(serviceVersions)
	if err != nil {
		return err
	}

	changesets, err := h.ConfigurationService.GetNextChangesets(c.Request().Context(), serviceVersionSpecifiers, afterChangesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, changesets)
}

// @Summary Get variation hierarchy
// @Description Get variation hierarchy
// @Produce json
// @Param services[] query []string true "Service versions" example(TestService:1) collectionFormat(multi)
// @Success 200 {object} configuration.VariationHierarchyDto
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /configuration/variation-hierarchy [get]
func (h *Handler) GetVariationHierarchy(c echo.Context) error {
	var serviceVersions []string
	err := echo.QueryParamsBinder(c).MustStrings("services[]", &serviceVersions).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(serviceVersions)
	if err != nil {
		return ToHTTPError(err)
	}

	variationHierarchy, err := h.ConfigurationService.GetVariationHierarchy(c.Request().Context(), serviceVersionSpecifiers)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, variationHierarchy)
}

// @Summary Get configuration
// @Description Get configuration
// @Produce json
// @Param changesetId query uint false "Changeset ID"
// @Param services[] query []string true "Service versions in format service:version" example(TestService:1) collectionFormat(multi)
// @Param mode query string false "Mode" Enums(production)
// @Param variation[] query []string false "Variation" example(env:prod) collectionFormat(multi)
// @Success 200 {object} configuration.ConfigurationDto
// @Failure 400 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /configuration [get]
func (h *Handler) GetConfiguration(c echo.Context) error {
	var changesetID uint
	var serviceVersions []string
	var mode string
	err := echo.QueryParamsBinder(c).Uint("changesetId", &changesetID).MustStrings("services[]", &serviceVersions).String("mode", &mode).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	variation, err := h.GetVariationFromQuery(c)
	if err != nil {
		return err
	}

	serviceVersionSpecifiers, err := core.ParseServiceVersionSpecifiers(serviceVersions)
	if err != nil {
		return err
	}

	configuration, err := h.ConfigurationService.GetConfiguration(c.Request().Context(), configuration.GetConfigurationParams{
		ServiceVersionSpecifiers: serviceVersionSpecifiers,
		ChangesetID:              ptr.To(changesetID, ptr.NilIfZero()),
		Mode:                     mode,
		Variation:                variation,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, configuration)
}
