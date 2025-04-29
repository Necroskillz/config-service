package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// @Summary Get a changeset
// @Description Get a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 200 {object} service.ChangesetDto
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

	return c.NoContent(http.StatusOK)
}

// @Summary Commit a changeset
// @Description Commit a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
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

	return c.NoContent(http.StatusOK)
}

// @Summary Reopen a changeset
// @Description Reopen a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
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

	return c.NoContent(http.StatusOK)
}

// @Summary Discard a changeset
// @Description Discard a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /changesets/{changeset_id}/discard [delete]
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

	return c.NoContent(http.StatusOK)
}

// @Summary Stash a changeset
// @Description Stash a changeset by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param changeset_id path uint true "Changeset ID"
// @Success 204
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

	return c.NoContent(http.StatusOK)
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

	return c.NoContent(http.StatusOK)
}

type ChangesetInfoResponse struct {
	ID              uint `json:"id" validate:"required"`
	NumberOfChanges int  `json:"numberOfChanges" validate:"required"`
}

// @Summary Get the current changeset info
// @Description Get the current changeset info
// @Accept json
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
