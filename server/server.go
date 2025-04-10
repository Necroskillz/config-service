package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/a-h/templ"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/db"
	"github.com/necroskillz/config-service/handler"
	"github.com/necroskillz/config-service/middleware"
	"github.com/necroskillz/config-service/service"
	error_views "github.com/necroskillz/config-service/views/error"

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
	variationContextService := service.NewVariationContextService(queries, unitOfWorkRunner, cache)
	serviceService := service.NewServiceService(queries, unitOfWorkRunner)
	userService := service.NewUserService(queries, variationContextService)
	changesetService := service.NewChangesetService(queries, variationContextService, unitOfWorkRunner, currentUserAccessor)
	variationHierarchyService := service.NewVariationHierarchyService(queries, cache)
	featureService := service.NewFeatureService(unitOfWorkRunner, queries)
	keyService := service.NewKeyService(unitOfWorkRunner, variationContextService, queries)
	validationService := service.NewValidationService(queries, variationContextService)
	valueService := service.NewValueService(unitOfWorkRunner, variationContextService, queries)

	e.Use(session.Middleware(sessions.NewCookieStore([]byte(os.Getenv("SESSION_SECRET")))))
	e.Use(middleware.AuthMiddleware(userService, variationHierarchyService, changesetService))
	e.Use(echoMiddleware.Logger())

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code

			if he.Internal != nil {
				c.Logger().Error(he.Internal)
			}

			if c.Request().Header.Get("HX-Request") == "true" {
				buf := templ.GetBuffer()
				defer templ.ReleaseBuffer(buf)

				c.Response().Header().Set("HX-Retarget", "#error-message")
				c.Response().Header().Set("HX-Reswap", "innerHTML")
				error_views.ErrorMessage(he.Message.(string)).Render(c.Request().Context(), buf)
				c.HTML(code, buf.String())
			} else {
				// TODO: error pages, return html or json based on accept header
				c.HTML(code, "<h1>Error Page</h1>")
			}
		} else {
			// should not happen
			c.Logger().Error(err)
			c.JSON(code, err)
		}
	}

	en := en.New()
	uni := ut.New(en)
	trans, _ := uni.GetTranslator("en")

	validator := validator.New()
	en_translations.RegisterDefaultTranslations(validator, trans)
	customValidator := NewCustomValidator(validator, validationService)

	err = customValidator.RegisterCustomValidation(trans)
	if err != nil {
		return fmt.Errorf("failed to register custom validation: %w", err)
	}

	e.Validator = customValidator

	handler := handler.NewHandler(
		serviceService,
		userService,
		featureService,
		keyService,
		changesetService,
		validationService,
		valueService,
		variationHierarchyService,
		trans,
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
