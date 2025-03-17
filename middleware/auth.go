package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/service"
)

func AuthMiddleware(userService *service.UserService, variationHierarchyService *service.VariationHierarchyService, changesetService *service.ChangesetService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authState, err := auth.GetAuthStateFromSession(c)
			if err != nil {
				return err
			}

			if authState.IsAuthenticated {
				user, err := userService.Get(c.Request().Context(), authState.UserId)
				if err != nil {
					return err
				}

				variationHierarchy, err := variationHierarchyService.GetVariationHierarchy(c.Request().Context())
				if err != nil {
					return err
				}

				changeset, err := changesetService.GetOpenChangesetForUser(c.Request().Context(), user.ID)
				if err != nil {
					return err
				}

				auth.StoreUserInContext(c, user, changeset, variationHierarchy)
			}

			return next(c)
		}
	}
}
