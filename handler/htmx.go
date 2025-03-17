package handler

import "github.com/labstack/echo/v4"

func Redirect(c echo.Context, url string) error {
	c.Response().Header().Set("HX-Redirect", url)
	return nil
}
