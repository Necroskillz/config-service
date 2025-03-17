package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
)

type AuthState struct {
	IsAuthenticated bool
	UserId          uint
}

func SaveAuthToSession(c echo.Context, userId uint) error {
	sess, err := session.Get(constants.UserSessionKey, c)
	if err != nil {
		return fmt.Errorf("failed to create a session: %v", err.Error())
	}

	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 30,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}

	sess.Values = map[any]any{
		constants.UserIdKey:        userId,
		constants.AuthenticatedKey: true,
	}

	sess.Save(c.Request(), c.Response())

	return nil
}

func GetAuthStateFromSession(c echo.Context) (*AuthState, error) {
	sess, err := session.Get(constants.UserSessionKey, c)
	if err != nil {
		return nil, fmt.Errorf("failed to get a session: %v", err.Error())
	}

	authState := &AuthState{
		IsAuthenticated: false,
	}

	isAuthenticated, ok := sess.Values[constants.AuthenticatedKey]
	if !ok || !isAuthenticated.(bool) {
		return authState, nil
	}

	authState.IsAuthenticated = true

	userId, ok := sess.Values[constants.UserIdKey]
	if !ok {
		return nil, fmt.Errorf("user ID not found in session")
	}

	authState.UserId = userId.(uint)

	return authState, nil
}

func ClearAuthFromSession(c echo.Context) error {
	sess, err := session.Get(constants.UserSessionKey, c)
	if err != nil {
		return fmt.Errorf("failed to get a session: %v", err.Error())
	}

	sess.Values = map[any]any{
		constants.AuthenticatedKey: false,
	}

	sess.Save(c.Request(), c.Response())

	return nil
}

func StoreUserInContext(c echo.Context, user *model.User, parentsProvider VariationPropertyValueParentsProvider) {
	authUser := NewUser(user, parentsProvider)

	c.Set(constants.UserKey, authUser)
}

func GetUserFromEchoContext(c echo.Context) *User {
	user, ok := c.Get(constants.UserKey).(*User)

	if !ok {
		user = AnonymousUser()
		c.Set(constants.UserKey, user)
	}

	return user
}

func GetUserFromContext(c context.Context) *User {
	user, ok := c.Value(constants.UserContextKey).(*User)

	if !ok {
		panic("attempted to get user from context but it is not set")
	}

	return user
}
