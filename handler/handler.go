package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/a-h/templ"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/model"
	"github.com/necroskillz/config-service/service"
	"github.com/necroskillz/config-service/views"
	"github.com/necroskillz/config-service/views/layouts"
)

type Handler struct {
	ServiceService            *service.ServiceService
	UserService               *service.UserService
	FeatureService            *service.FeatureService
	KeyService                *service.KeyService
	ChangesetService          *service.ChangesetService
	ValidationService         *service.ValidationService
	ValueService              *service.ValueService
	VariationHierarchyService *service.VariationHierarchyService
	Translator                ut.Translator
}

func NewHandler(
	serviceService *service.ServiceService,
	userService *service.UserService,
	featureService *service.FeatureService,
	keyService *service.KeyService,
	changesetService *service.ChangesetService,
	validationService *service.ValidationService,
	valueService *service.ValueService,
	variationHierarchyService *service.VariationHierarchyService,
	translator ut.Translator,
) *Handler {
	return &Handler{
		ServiceService:            serviceService,
		UserService:               userService,
		FeatureService:            featureService,
		KeyService:                keyService,
		ChangesetService:          changesetService,
		ValidationService:         validationService,
		ValueService:              valueService,
		Translator:                translator,
		VariationHierarchyService: variationHierarchyService,
	}
}

func (h *Handler) User(c echo.Context) *auth.User {
	return auth.GetUserFromEchoContext(c)
}

func (h *Handler) EnsureChangesetID(c echo.Context) (uint, error) {
	changesetId := h.User(c).ChangesetID
	if changesetId == 0 {
		changeset, err := h.ChangesetService.CreateChangesetForUser(c.Request().Context(), h.User(c).ID)
		if err != nil {
			return 0, echo.NewHTTPError(http.StatusInternalServerError, "Failed to create changeset").WithInternal(err)
		}

		changesetId = changeset.ID
	}

	return changesetId, nil
}

func (h *Handler) ViewContext(c echo.Context) context.Context {
	ctx := context.WithValue(c.Request().Context(), constants.UserContextKey, h.User(c))

	return ctx
}

func (h *Handler) Render(c echo.Context, statusCode int, renderFn func(ctx context.Context, w io.Writer) error) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	ctx := h.ViewContext(c)

	err := renderFn(ctx, buf)

	if err != nil {
		return err
	}

	return c.HTML(statusCode, buf.String())
}

func (h *Handler) RenderPage(c echo.Context, statusCode int, component templ.Component, title string) error {
	return h.Render(c, statusCode, func(ctx context.Context, w io.Writer) error {
		return layouts.WithBase(component, title, true).Render(ctx, w)
	})
}

func (h *Handler) RenderPartial(c echo.Context, statusCode int, component templ.Component) error {
	return h.Render(c, statusCode, func(ctx context.Context, w io.Writer) error {
		return component.Render(ctx, w)
	})
}

func (h *Handler) GetValidationErrors(err validator.ValidationErrors) map[string]string {
	errors := make(map[string]string, len(err))

	for _, err := range err {
		errors[err.Field()] = err.Translate(h.Translator)
	}

	return errors
}

func (h *Handler) BindAndValidate(c echo.Context, data views.ViewDataSetter, errorCollectors ...func() []error) (bool, error) {
	err := c.Bind(data)
	if err != nil {
		data.SetError("Failed to process form data")
		return false, echo.NewHTTPError(http.StatusBadRequest, "Failed to process form data").WithInternal(err)
	}

	err = c.Validate(data)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			data.SetError("Failed to validate form data")
			return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate form data").WithInternal(err)
		}

		data.SetValidationErrors(h.GetValidationErrors(validationErrors))
		return false, nil
	}

	if len(errorCollectors) > 0 {
		for _, errorCollector := range errorCollectors {
			serviceErrors := errorCollector()
			if len(serviceErrors) > 0 {
				return h.ApplyServiceValidationErrors(c, data, serviceErrors...)
			}
		}
	}

	return true, nil
}

type ServiceErrorCollector struct {
	errors []error
}

func (h *Handler) CollectServiceErrors(fn func(sec *ServiceErrorCollector)) func() []error {
	collector := &ServiceErrorCollector{}

	return func() []error {
		fn(collector)
		return collector.errors
	}
}

func (h *ServiceErrorCollector) Collect(err error) {
	if err != nil {
		h.errors = append(h.errors, err)
	}
}

func (h *Handler) ApplyServiceValidationErrors(c echo.Context, data views.ViewDataSetter, serviceErrors ...error) (bool, error) {
	validationErrors := make(map[string]string)
	errorMessages := make([]string, 0)
	valid := true

	for _, err := range serviceErrors {
		var validationError *service.ValidationError
		if errors.As(err, &validationError) {
			if validationError.Field != "" {
				validationErrors[validationError.Field] = validationError.Message
			} else {
				errorMessages = append(errorMessages, validationError.Message)
			}
		} else {
			return false, echo.NewHTTPError(http.StatusInternalServerError, "Failed to validate form data").WithInternal(err)
		}
	}

	if len(errorMessages) > 0 {
		data.SetError(strings.Join(errorMessages, "\n"))
		valid = false
	}

	if len(validationErrors) > 0 {
		data.SetValidationErrors(validationErrors)
		valid = false
	}

	return valid, nil
}

func (h *Handler) LoadBasicData(c echo.Context, chain ...any) error {
	if len(chain) == 0 {
		panic("no data to load")
	}

	serviceVersion, ok := chain[0].(*model.ServiceVersion)
	if !ok {
		panic("first argument must be a pointer to a model.ServiceVersion")
	}

	err := LoadEntity(c, serviceVersion, "service_version_id", h.ServiceService.GetServiceVersion)
	if err != nil {
		return err
	}

	if len(chain) > 1 {
		featureVersion, ok := chain[1].(*model.FeatureVersion)
		if !ok {
			panic("second argument must be a pointer to a model.FeatureVersion")
		}

		err = LoadEntity(c, featureVersion, "feature_version_id", h.FeatureService.GetFeatureVersion)
		if err != nil {
			return err
		}

		// TODO: check if feature version is linked to service version
	}

	if len(chain) > 2 {
		key, ok := chain[2].(*model.Key)
		if !ok {
			panic("third argument must be a pointer to a model.Key")
		}

		err = LoadEntity(c, key, "key_id", h.KeyService.GetKey)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadEntity[T any](c echo.Context, entity *T, paramName string, loader func(ctx context.Context, id uint) (*T, error)) error {
	var id uint
	err := echo.PathParamsBinder(c).MustUint(paramName, &id).BindError()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid path parameters (%s)", paramName)
	}

	loadedEntity, err := loader(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get %s with ID %d", reflect.TypeOf(entity).String(), id).WithInternal(err)
	}

	if loadedEntity == nil {
		return echo.NewHTTPError(http.StatusNotFound, "%s with ID %d not found", reflect.TypeOf(entity).String(), id)
	}

	*entity = *loadedEntity

	return nil
}
