package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/services/changeset"
	_ "github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/util/ptr"
)

// @Summary Get changesets
// @Description Get changesets
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param pageSize query int false "Page Size" default(20)
// @Param authorId query uint false "Author ID"
// @Param approvable query bool false "Approvable"
// @Success 200 {object} core.PaginatedResult[changeset.ChangesetItemDto]
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets [get]
func (h *Handler) Changesets(c echo.Context) error {
	page := 1
	pageSize := 20
	var authorID uint
	var approvable bool

	err := echo.QueryParamsBinder(c).
		Int("page", &page).
		Int("pageSize", &pageSize).
		Uint("authorId", &authorID).
		Bool("approvable", &approvable).
		BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	changesets, err := h.ChangesetService.GetChangesets(c.Request().Context(), changeset.Filter{
		AuthorID:   ptr.To(authorID, ptr.NilIfZero()),
		Approvable: approvable,
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, changesets)
}

// @Summary Get a changeset
// @Description Get a changeset by ID
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 200 {object} changeset.ChangesetDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id} [get]
func (h *Handler) Changeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	changeset, err := h.ChangesetService.GetChangeset(c.Request().Context(), changesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, changeset)
}

type OptionalCommentRequest struct {
	Comment *string `json:"comment"`
}

// @Summary Apply a changeset
// @Description Apply a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/apply [put]
func (h *Handler) ApplyChangeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var request OptionalCommentRequest
	err = c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.ApplyChangeset(c.Request().Context(), changesetID, request.Comment)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Commit a changeset
// @Description Commit a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/commit [put]
func (h *Handler) CommitChangeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var request OptionalCommentRequest
	err = c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.CommitChangeset(c.Request().Context(), changesetID, request.Comment)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Reopen a changeset
// @Description Reopen a changeset by ID
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/reopen [put]
func (h *Handler) ReopenChangeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.ReopenChangeset(c.Request().Context(), changesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Discard a changeset
// @Description Discard a changeset by ID
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id} [delete]
func (h *Handler) DiscardChangeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.DiscardChangeset(c.Request().Context(), changesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Discard a change
// @Description Discard a change by ID
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Param change_id path uint true "Change ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/changes/{change_id} [delete]
func (h *Handler) DiscardChange(c echo.Context) error {
	var changesetID uint
	var changeID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).MustUint("change_id", &changeID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.DiscardChange(c.Request().Context(), changesetID, changeID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

// @Summary Stash a changeset
// @Description Stash a changeset by ID
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/stash [put]
func (h *Handler) StashChangeset(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.StashChangeset(c.Request().Context(), changesetID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type AddCommentRequest struct {
	Comment string `json:"comment" validate:"required"`
}

// @Summary Add a comment to a changeset
// @Description Add a comment to a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Param comment body AddCommentRequest true "Comment"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/comment [post]
func (h *Handler) AddComment(c echo.Context) error {
	var changesetID uint
	err := echo.PathParamsBinder(c).MustUint("changeset_id", &changesetID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var request AddCommentRequest
	err = c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.ChangesetService.AddComment(c.Request().Context(), changesetID, request.Comment)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}

type ChangesetInfoResponse struct {
	ID              uint `json:"id" validate:"required"`
	NumberOfChanges int  `json:"numberOfChanges" validate:"required"`
}

// @Summary Get the current changeset info
// @Description Get the current changeset info
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ChangesetInfoResponse
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/current [get]
func (h *Handler) GetCurrentChangesetInfo(c echo.Context) error {
	user := h.CurrentUserAccessor.GetUser(c.Request().Context())

	count, err := h.ChangesetService.GetChangesetChangesCount(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, ChangesetInfoResponse{
		ID:              user.ChangesetID,
		NumberOfChanges: count,
	})
}

type ApprovableChangesetCountResponse struct {
	Count int `json:"count" validate:"required"`
}

// @Summary Get the number of approvable changesets
// @Description Get the number of approvable changesets
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ApprovableChangesetCountResponse
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/approvable-count [get]
func (h *Handler) GetApprovableChangesetCount(c echo.Context) error {
	count, err := h.ChangesetService.GetApprovableChangesetCount(c.Request().Context())
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, ApprovableChangesetCountResponse{
		Count: count,
	})
}
