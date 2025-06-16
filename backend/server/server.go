package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	"github.com/joho/godotenv"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	_ "github.com/necroskillz/config-service/docs"
	"github.com/necroskillz/config-service/handler"
	"github.com/necroskillz/config-service/middleware"
	"github.com/necroskillz/config-service/services"
	"github.com/necroskillz/config-service/util/logging"
	slogecho "github.com/samber/slog-echo"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Server struct {
	echo   *echo.Echo
	dbpool *pgxpool.Pool
	cache  *ristretto.Cache[string, any]
}

type PgxTraceLogger struct {
}

func (t *PgxTraceLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	fmt.Println("pgx", msg, data)
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

	connectionString := os.Getenv("DATABASE_URL")

	if err := db.Migrate(connectionString); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	logger := logging.ConfigureSlog("Config Service HTTP/server")

	if os.Getenv("PGX_TRACE") != "" {
		config.ConnConfig.Tracer = &tracelog.TraceLog{
			LogLevel: tracelog.LogLevelTrace,
			Logger:   &PgxTraceLogger{},
		}
	}

	dbpool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	s.dbpool = dbpool
	s.echo = e
	s.cache = cache

	svc := services.InitializeServices(dbpool, cache)

	e.Use(slogecho.NewWithFilters(logger,
		slogecho.IgnoreStatus(http.StatusUnauthorized, http.StatusConflict),
	))
	e.Use(echoMiddleware.Recover())
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
			url := c.Request().URL.Path

			if !strings.HasPrefix(url, "/api/") {
				return true
			}

			skippedPaths := []string{
				"/api/auth/login",
				"/api/auth/refresh_token",
				"/api/configuration",
			}

			for _, path := range skippedPaths {
				if strings.HasPrefix(url, path) {
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
	e.Use(middleware.AuthMiddleware(svc.AuthService, svc.VariationHierarchyService, svc.ChangesetService))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	handler := handler.NewHandler(
		svc.ServiceService,
		svc.AuthService,
		svc.FeatureService,
		svc.KeyService,
		svc.ChangesetService,
		svc.ValidationService,
		svc.ValueService,
		svc.VariationHierarchyService,
		svc.CurrentUserAccessor,
		svc.ValueTypeService,
		svc.VariationPropertyService,
		svc.ServiceTypeService,
		svc.ConfigurationService,
		svc.MembershipService,
	)
	handler.RegisterRoutes(e)

	return e.Start(fmt.Sprintf(":%s", os.Getenv("PORT")))
}

func (s *Server) Stop(ctx context.Context) error {
	if s.echo != nil {
		if err := s.echo.Shutdown(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				return fmt.Errorf("failed to shutdown server: %w", err)
			}
		}
	}

	if s.dbpool != nil {
		s.dbpool.Close()
	}

	if s.cache != nil {
		s.cache.Close()
	}

	return nil
}
