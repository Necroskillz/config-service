package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/service"
	views "github.com/necroskillz/config-service/views/auth"
)

func (h *Handler) LoginPage(c echo.Context) error {
	data := views.LoginData{}

	return h.RenderPage(c, http.StatusOK, views.Login(data), "Login")
}

func (h *Handler) Login(c echo.Context) error {
	var data views.LoginData

	valid, err := h.BindAndValidate(c, &data)

	if err != nil {
		return err
	}

	if !valid {
		return h.RenderPartial(c, http.StatusUnprocessableEntity, views.Login(data))
	}

	userId, err := h.UserService.Authenticate(c.Request().Context(), data.Username, data.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPassword) || errors.Is(err, service.ErrRecordNotFound) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Invalid username or password")
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate user").WithInternal(err)
	}

	err = auth.SaveAuthToSession(c, userId)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Unable to start an authenticated session").WithInternal(err)
	}

	return Redirect(c, "/")
}

func (h *Handler) Logout(c echo.Context) error {
	err := auth.ClearAuthFromSession(c)
	if err != nil {
		c.Logger().Error(fmt.Errorf("failed to clear auth from session: %w", err))

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to clear auth from session")
	}

	return Redirect(c, "/")
}
