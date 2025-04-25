package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	_ "github.com/necroskillz/config-service/docs"
	"github.com/necroskillz/config-service/handler"
	"github.com/necroskillz/config-service/middleware"
	"github.com/necroskillz/config-service/service"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/dgraph-io/ristretto/v2"
)

type Server struct {
	echo   *echo.Echo
	dbpool *pgxpool.Pool
	cache  *ristretto.Cache[string, any]
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	e := echo.New()
	ctx := context.Background()

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})

	if err != nil {
		return fmt.Errorf("failed to initialize cache: %w", err)
	}

	dbpool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.dbpool = dbpool
	s.echo = e
	s.cache = cache

	queries := db.New(dbpool)

	e.Static("/assets", "views/assets")

	currentUserAccessor := auth.NewCurrentUserAccessor()

	unitOfWorkRunner := db.NewPgxUnitOfWorkRunner(dbpool, queries)
	coreService := service.NewCoreService(queries, currentUserAccessor)
	variationHierarchyService := service.NewVariationHierarchyService(queries, cache)
	variationContextService := service.NewVariationContextService(queries, unitOfWorkRunner, cache)
	validationService := service.NewValidationService(queries, variationContextService, variationHierarchyService, currentUserAccessor, coreService)
	validator := service.NewValidator(validationService)
	changesetService := service.NewChangesetService(queries, variationContextService, unitOfWorkRunner, currentUserAccessor, validator)
	serviceService := service.NewServiceService(queries, unitOfWorkRunner, changesetService, currentUserAccessor, validator)
	userService := service.NewUserService(queries, variationContextService)
	featureService := service.NewFeatureService(unitOfWorkRunner, queries, changesetService, currentUserAccessor, validator, coreService)
	keyService := service.NewKeyService(unitOfWorkRunner, variationContextService, queries, changesetService, currentUserAccessor, validator, coreService)
	valueService := service.NewValueService(unitOfWorkRunner, variationContextService, variationHierarchyService, queries, changesetService, currentUserAccessor, validator, coreService, validationService)

	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins:     []string{os.Getenv("FRONTEND_URL")},
		AllowMethods:     []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}))
	e.Use(echojwt.WithConfig(echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.Claims)
		},
		SigningKey: []byte(os.Getenv("JWT_SECRET")),
		Skipper: func(c echo.Context) bool {
			skippedPaths := []string{
				"/api/auth/login",
				"/api/auth/refresh_token",
				"/swagger",
			}

			for _, path := range skippedPaths {
				if strings.HasPrefix(c.Request().URL.Path, path) {
					return true
				}
			}

			return false
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		},
		ContextKey: "claims",
	}))
	e.Use(middleware.AuthMiddleware(userService, variationHierarchyService, changesetService))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code

			if he.Internal != nil {
				c.Logger().Error(he.Internal)
			}

			c.JSON(code, err)
		} else {
			// should not happen
			c.Logger().Error(err)
			c.JSON(code, err)
		}
	}

	handler := handler.NewHandler(
		serviceService,
		userService,
		featureService,
		keyService,
		changesetService,
		validationService,
		valueService,
		variationHierarchyService,
		currentUserAccessor,
	)
	handler.RegisterRoutes(e)

	return e.Start(fmt.Sprintf(":%s", os.Getenv("PORT")))
}

func (s *Server) Stop(ctx context.Context) error {
	if s.dbpool != nil {
		s.dbpool.Close()
	}

	if s.cache != nil {
		s.cache.Close()
	}

	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	return nil
}
