package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/service"
)

func ChangesetMiddleware(changesetService *service.ChangesetService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := auth.GetUserFromEchoContext(c)

			if user.IsAuthenticated {
				changeset, err := changesetService.GetOpenChangesetForUser(c.Request().Context(), user.ID)

				if err != nil {
					return err
				}

				changesetID := uint(0)

				if changeset != nil {
					changesetID = changeset.ID
				}

				c.Set(constants.ChangesetIdKey, changesetID)
			}

			return next(c)
		}
	}
}
