package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/core"
)

func parseServiceVersionSpecifiers(serviceVersions []string) ([]core.ServiceVersionSpecifier, error) {
	serviceVersionSpecifiers := make([]core.ServiceVersionSpecifier, len(serviceVersions))
	for i, serviceVersion := range serviceVersions {
		specifier, err := core.ParseServiceVersionSpecifier(serviceVersion)
		if err != nil {
			return nil, echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		serviceVersionSpecifiers[i] = specifier
	}

	return serviceVersionSpecifiers, nil
}

// @Summary Get next changesets
// @Description Get next changesets
// @Accept json
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

	serviceVersionSpecifiers, err := parseServiceVersionSpecifiers(serviceVersions)
	if err != nil {
		return err
	}

	changesets, err := h.ConfigurationService.GetNextChangesets(c.Request().Context(), serviceVersionSpecifiers, afterChangesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, changesets)
}

// @Summary Get configuration
// @Description Get configuration
// @Accept json
// @Produce json
// @Param changeset_id query uint false "Changeset ID"
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
	err := echo.QueryParamsBinder(c).Uint("changeset_id", &changesetID).MustStrings("services[]", &serviceVersions).String("mode", &mode).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	variation, err := h.GetVariationFromQuery(c)
	if err != nil {
		return err
	}

	serviceVersionSpecifiers, err := parseServiceVersionSpecifiers(serviceVersions)
	if err != nil {
		return err
	}

	configuration, err := h.ConfigurationService.GetConfiguration(c.Request().Context(), serviceVersionSpecifiers, changesetID, mode, variation)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, configuration)
}
