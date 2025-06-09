package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/services/changeset"
	"github.com/necroskillz/config-service/services/membership"
	"github.com/necroskillz/config-service/services/variation"
)

func AuthMiddleware(authService *membership.AuthService, variationHierarchyService *variation.HierarchyService, changesetService *changeset.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.HasPrefix(c.Request().URL.Path, "/assets/") {
				return next(c)
			}

			claims, err := auth.GetClaims(c)
			if err != nil {
				if errors.Is(err, auth.ErrUserNotAuthenticated) {
					auth.StoreUserInContext(c, auth.AnonymousUser())
					return next(c)
				}

				return err
			}

			user, err := authService.GetUser(c.Request().Context(), claims.UserId)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user").WithInternal(err)
			}

			variationHierarchy, err := variationHierarchyService.GetVariationHierarchy(c.Request().Context())
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get variation hierarchy").WithInternal(err)
			}

			changesetId, err := changesetService.GetOpenChangesetIDForUser(c.Request().Context(), user.ID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get open changeset for user").WithInternal(err)
			}

			userBuilder := auth.NewUserBuilder(variationHierarchy)
			userBuilder.WithBasicInfo(user.ID, user.Username, user.GlobalAdministrator)
			userBuilder.WithChangesetID(changesetId)

			for _, permission := range user.Permissions {
				userBuilder.WithPermission(permission.ServiceID, permission.FeatureID, permission.KeyID, permission.Variation, permission.Permission)
			}

			auth.StoreUserInContext(c, userBuilder.User())
			c.SetRequest(c.Request().WithContext(context.WithValue(c.Request().Context(), constants.UserContextKey, userBuilder.User())))

			return next(c)
		}
	}
}
