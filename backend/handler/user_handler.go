package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/necroskillz/config-service/services/core"
	"github.com/necroskillz/config-service/services/membership"
)

// @Summary Get users
// @Description Get list of users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param name query string false "Name"
// @Success 200 {object} core.PaginatedResult[membership.UserDto]
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users [get]
func (h *Handler) Users(c echo.Context) error {
	var limit, offset int
	var name string

	binder := echo.QueryParamsBinder(c)
	err := binder.MustInt("limit", &limit).
		MustInt("offset", &offset).
		String("name", &name).
		BindError()

	if err != nil {
		return ToHTTPError(err)
	}

	users, err := h.UserService.GetUsers(c.Request().Context(), membership.UsersFilter{
		Limit:  limit,
		Offset: offset,
		Name:   name,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, users)
}

type GetUserResponse struct {
	ID                  uint   `json:"id"`
	Username            string `json:"username"`
	GlobalAdministrator bool   `json:"globalAdministrator"`
}

// @Summary Get a user
// @Description Get a user by ID
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Success 200 {object} membership.UserDto
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users/{user_id} [get]
func (h *Handler) GetUser(c echo.Context) error {
	var userID uint
	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	user, err := h.UserService.Get(c.Request().Context(), userID)
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, membership.UserDto{
		ID:                  user.ID,
		Username:            user.Username,
		GlobalAdministrator: user.GlobalAdministrator,
	})
}

type CreateUserRequest struct {
	Username            string `json:"username" validate:"required"`
	Password            string `json:"password" validate:"required"`
	GlobalAdministrator bool   `json:"globalAdministrator" validate:"required"`
}

// @Summary Create a user
// @Description Create a new user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body CreateUserRequest true "User"
// @Success 200 {object} CreateResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users [post]
func (h *Handler) CreateUser(c echo.Context) error {
	var request CreateUserRequest
	err := c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	userID, err := h.UserService.CreateUser(c.Request().Context(), membership.CreateUserParams{
		Username:            request.Username,
		Password:            request.Password,
		GlobalAdministrator: request.GlobalAdministrator,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.JSON(http.StatusOK, NewCreateResponse(userID))
}

type UpdateUserRequest struct {
	GlobalAdministrator bool `json:"globalAdministrator" validate:"required"`
}

// @Summary Update a user
// @Description Update a user by ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id path uint true "User ID"
// @Param user body UpdateUserRequest true "User"
// @Success 204
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 403 {object} echo.HTTPError
// @Failure 404 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users/{user_id} [put]
func (h *Handler) UpdateUser(c echo.Context) error {
	var userID uint
	err := echo.PathParamsBinder(c).MustUint("user_id", &userID).BindError()
	if err != nil {
		return ToHTTPError(err)
	}

	var request UpdateUserRequest
	err = c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	err = h.UserService.UpdateUser(c.Request().Context(), userID, membership.UpdateUserParams{
		GlobalAdministrator: request.GlobalAdministrator,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
