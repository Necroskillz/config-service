package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
	views "github.com/necroskillz/config-service/views/changesets"
)

func (h *Handler) ChangesetDetail(c echo.Context) error {
	var changesetID uint

	err := echo.PathParamsBinder(c).Uint("changeset_id", &changesetID).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid changeset ID")
	}

	changeset, err := h.ChangesetService.GetChangeset(c.Request().Context(), changesetID)
	if err != nil {
		if errors.Is(err, service.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "Changeset not found")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get changeset").WithInternal(err)
	}

	data := views.ChangesetDetailData{
		Changeset: changeset,
	}

	return h.RenderPage(c, http.StatusOK, views.ChangesetDetailPage(data), "Changeset Detail")
}
