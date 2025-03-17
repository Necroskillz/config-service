package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/necroskillz/config-service/handler"
	"github.com/necroskillz/config-service/middleware"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/repository"
	"github.com/necroskillz/config-service/service"
	error_views "github.com/necroskillz/config-service/views/error"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Server struct {
	db   *gorm.DB
	echo *echo.Echo
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Start() error {
	e := echo.New()

	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	db, err := initDB()
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	s.db = db
	s.echo = e

	e.Static("/assets", "views/assets")

	serviceRepository := repository.NewServiceRepository(db)
	serviceVersionRepository := repository.NewServiceVersionRepository(db)
	variationPropertyRepository := repository.NewVariationPropertyRepository(db)
	serviceTypeRepository := repository.NewServiceTypeRepository(db)
	serviceVersionFeatureVersionRepository := repository.NewServiceVersionFeatureVersionRepository(db)
	changesetRepository := repository.NewChangesetRepository(db)
	changesetChangeRepository := repository.NewChangesetChangeRepository(db)
	featureRepository := repository.NewFeatureRepository(db)
	featureVersionRepository := repository.NewFeatureVersionRepository(db)
	userRepository := repository.NewUserRepository(db)
	keyRepository := repository.NewKeyRepository(db)
	valueTypeRepository := repository.NewValueTypeRepository(db)
	variationValueRepository := repository.NewVariationValueRepository(db)

	unitOfWorkCreator := repository.NewGormUnitOfWorkCreator(db)

	serviceService := service.NewServiceService(unitOfWorkCreator, serviceRepository, serviceVersionRepository, serviceTypeRepository)
	userService := service.NewUserService(userRepository)
	changesetService := service.NewChangesetService(changesetRepository, changesetChangeRepository)
	variationHierarchyService := service.NewVariationHierarchyService(variationPropertyRepository)
	featureService := service.NewFeatureService(unitOfWorkCreator, featureRepository, featureVersionRepository, serviceVersionFeatureVersionRepository, changesetService)
	keyService := service.NewKeyService(unitOfWorkCreator, keyRepository, valueTypeRepository, changesetService, variationValueRepository)
	validationService := service.NewValidationService(keyRepository)
	valueService := service.NewValueService(unitOfWorkCreator, variationValueRepository)

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
	customValidator := NewCustomValidator(validator, serviceService, featureService)

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
	if s.db != nil {
		dbInstance, err := s.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get database instance: %w", err)
		}

		err = dbInstance.Close()
		if err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
	}

	return s.echo.Shutdown(ctx)
}

func initDB() (*gorm.DB, error) {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbPort := os.Getenv("DB_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		}),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(
		&model.Service{},
		&model.ServiceVersion{},
		&model.Changeset{},
		&model.Feature{},
		&model.FeatureVersion{},
		&model.ValueType{},
		&model.Key{},
		&model.VariationValue{},
		&model.VariationProperty{},
		&model.VariationPropertyValue{},
		&model.User{},
		&model.ChangesetChange{},
		&model.FeatureVersionServiceVersion{},
		&model.UserPermission{},
		&model.ServiceType{},
		&model.ServiceTypeVariationProperty{},
	)

	if err != nil {
		return nil, err
	}

	return db, nil
}
