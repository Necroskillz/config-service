package auth

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/constants"
)

type AuthState struct {
	IsAuthenticated bool
	UserId          uint
}

type Claims struct {
	UserId uint `json:"user_id"`
	jwt.RegisteredClaims
}

func GenerateTokens(userId uint) (string, string, error) {
	accessTokenExpiresAt := time.Now().Add(time.Minute * 15)
	refreshTokenExpiresAt := time.Now().Add(time.Hour * 24 * 30)

	accessToken, err := generateToken(userId, accessTokenExpiresAt, []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateToken(userId, refreshTokenExpiresAt, []byte(os.Getenv("JWT_REFRESH_SECRET")))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func generateToken(userId uint, expiresAt time.Time, secret []byte) (string, error) {
	claims := &Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

var ErrInvalidToken = errors.New("invalid refresh token")

func ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv("JWT_REFRESH_SECRET")), nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

var ErrUserNotAuthenticated = errors.New("unable to get claims because user is not authenticated")

func GetClaims(c echo.Context) (*Claims, error) {
	u, ok := c.Get("claims").(*jwt.Token)
	if !ok {
		return nil, ErrUserNotAuthenticated
	}

	claims := u.Claims.(*Claims)

	return claims, nil
}

func StoreUserInContext(c echo.Context, user *User) {
	c.Set(constants.UserKey, user)
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

type CurrentUserAccessor struct {
}

func NewCurrentUserAccessor() *CurrentUserAccessor {
	return &CurrentUserAccessor{}
}

func (c *CurrentUserAccessor) GetUser(ctx context.Context) *User {
	return GetUserFromContext(ctx)
}
