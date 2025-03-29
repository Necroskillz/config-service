package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/service"
	views "github.com/necroskillz/config-service/views/changesets"
)

func (h *Handler) getChangeset(c echo.Context) (service.Changeset, error) {
	var changesetID uint

	err := echo.PathParamsBinder(c).Uint("changeset_id", &changesetID).BindError()
	if err != nil {
		return service.Changeset{}, echo.NewHTTPError(http.StatusBadRequest, "Invalid changeset ID")
	}

	changeset, err := h.ChangesetService.GetChangeset(c.Request().Context(), changesetID)
	if err != nil {
		if errors.Is(err, service.ErrRecordNotFound) {
			return service.Changeset{}, echo.NewHTTPError(http.StatusNotFound, "Changeset not found")
		}

		return service.Changeset{}, echo.NewHTTPError(http.StatusInternalServerError, "Failed to get changeset").WithInternal(err)
	}

	return changeset, nil
}
func (h *Handler) ChangesetDetail(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	data := views.ChangesetDetailData{
		Changeset: changeset,
	}

	return h.RenderPage(c, http.StatusOK, views.ChangesetDetailPage(data), "Changeset Detail")
}

func (h *Handler) ApplyChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	user := h.User(c)

	if !changeset.CanBeAppliedBy(user) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to apply this changeset")
	}

	err = h.ChangesetService.ApplyChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to apply changeset").WithInternal(err)
	}

	c.Set(constants.ChangesetRemovedKey, true)

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetailPage(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) CommitChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot commit changeset that is not open")
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to commit this changeset")
	}

	err = h.ChangesetService.CommitChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit changeset").WithInternal(err)
	}

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetailPage(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) ReopenChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if !changeset.IsCommitted() {
		return echo.NewHTTPError(http.StatusBadRequest, "Cannot reopen changeset that is not committed")
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to reopen this changeset")
	}

	err = h.ChangesetService.ReopenChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reopen changeset").WithInternal(err)
	}

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetailPage(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}
