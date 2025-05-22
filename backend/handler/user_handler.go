package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/service"
)

type UsersRequest struct {
	Limit  int `query:"limit" validate:"required"`
	Offset int `query:"offset" validate:"required"`
}

// @Summary Get users
// @Description Get list of users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} service.PaginatedResult[service.UserDto]
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /users [get]
func (h *Handler) Users(c echo.Context) error {
	var request UsersRequest
	err := c.Bind(&request)
	if err != nil {
		return ToHTTPError(err)
	}

	users, err := h.UserService.GetUsers(c.Request().Context(), service.UsersFilter{
		Limit:  request.Limit,
		Offset: request.Offset,
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
// @Success 200 {object} service.User
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

	return c.JSON(http.StatusOK, service.UserDto{
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

	userID, err := h.UserService.CreateUser(c.Request().Context(), service.CreateUserParams{
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

	err = h.UserService.UpdateUser(c.Request().Context(), userID, service.UpdateUserParams{
		GlobalAdministrator: request.GlobalAdministrator,
	})
	if err != nil {
		return ToHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
