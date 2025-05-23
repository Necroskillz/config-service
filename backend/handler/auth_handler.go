package handler

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/services/core"
)

// @Summary Get current user
// @Description Get the current user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} auth.User
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/user [get]
func (h *Handler) User(c echo.Context) error {
	user := h.CurrentUserAccessor.GetUser(c.Request().Context())

	return c.JSON(http.StatusOK, user)
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type TokensResponse struct {
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Summary Login
// @Description Login to the application
// @Accept json
// @Produce json
// @Param loginRequest body LoginRequest true "Login request"
// @Success 200 {object} TokensResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 422 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var data LoginRequest

	err := c.Bind(&data)
	if err != nil {
		return err
	}

	userId, err := h.UserService.Authenticate(c.Request().Context(), data.Username, data.Password)
	if err != nil {
		if errors.Is(err, core.ErrInvalidPassword) || errors.Is(err, core.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Invalid username or password")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate user").WithInternal(err)
	}

	accessToken, refreshToken, err := auth.GenerateTokens(userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate access and refresh tokens").WithInternal(err)
	}

	return c.JSON(http.StatusOK, TokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Summary Refresh token
// @Description Refresh token
// @Accept json
// @Produce json
// @Param refreshTokenRequest body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} TokensResponse
// @Failure 400 {object} echo.HTTPError
// @Failure 401 {object} echo.HTTPError
// @Failure 500 {object} echo.HTTPError
// @Router /auth/refresh_token [post]
func (h *Handler) RefreshToken(c echo.Context) error {
	var data RefreshTokenRequest

	err := c.Bind(&data)
	if err != nil {
		return err
	}

	claims, err := auth.ParseToken(data.RefreshToken)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, "Invalid refresh token").WithInternal(err)
	}

	accessToken, refreshToken, err := auth.GenerateTokens(claims.UserId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to generate access and refresh tokens").WithInternal(err)
	}

	return c.JSON(http.StatusOK, TokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}
