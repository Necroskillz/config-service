package handler

import (
	"errors"
	"fmt"
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

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetail(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) CommitChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() {
		echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Cannot commit changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to commit this changeset")
	}

	err = h.ChangesetService.CommitChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit changeset").WithInternal(err)
	}

	c.Set(constants.ChangesetRemovedKey, true)

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetail(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) ReopenChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if h.User(c).ChangesetID != 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "You already have an open changeset")
	}

	if !changeset.IsCommitted() && !changeset.IsStashed() {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Cannot reopen changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to reopen this changeset")
	}

	err = h.ChangesetService.ReopenChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reopen changeset").WithInternal(err)
	}

	c.Set(constants.ChangesetCreatedKey, true)
	h.User(c).ChangesetID = changeset.ID

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetail(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) DiscardChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() && !changeset.IsCommitted() {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Cannot discard changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to discard this changeset")
	}

	err = h.ChangesetService.DiscardChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to discard changeset").WithInternal(err)
	}

	c.Set(constants.ChangesetRemovedKey, true)

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetail(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}

func (h *Handler) StashChangeset(c echo.Context) error {
	changeset, err := h.getChangeset(c)
	if err != nil {
		return err
	}

	if !changeset.IsOpen() {
		echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Cannot stash changeset in state %s", changeset.State))
	}

	if !changeset.BelongsTo(h.User(c).ID) {
		return echo.NewHTTPError(http.StatusForbidden, "You are not allowed to stash this changeset")
	}

	err = h.ChangesetService.StashChangeset(c.Request().Context(), &changeset)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to stash changeset").WithInternal(err)
	}

	c.Set(constants.ChangesetRemovedKey, true)

	return h.RenderPartial(c, http.StatusOK, views.ChangesetDetail(views.ChangesetDetailData{
		Changeset: changeset,
	}))
}
