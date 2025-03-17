package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/a-h/templ"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
	"github.com/necroskillz/config-service/service"
	"github.com/necroskillz/config-service/views"
	"github.com/necroskillz/config-service/views/layouts"
)

type Handler struct {
	ServiceService    *service.ServiceService
	UserService       *service.UserService
	FeatureService    *service.FeatureService
	KeyService        *service.KeyService
	ChangesetService  *service.ChangesetService
	ValidationService *service.ValidationService
	ValueService      *service.ValueService
	Translator        ut.Translator
}

func NewHandler(
	serviceService *service.ServiceService,
	userService *service.UserService,
	featureService *service.FeatureService,
	keyService *service.KeyService,
	changesetService *service.ChangesetService,
	validationService *service.ValidationService,
	valueService *service.ValueService,
	translator ut.Translator,
) *Handler {
	return &Handler{
		ServiceService:    serviceService,
		UserService:       userService,
		FeatureService:    featureService,
		KeyService:        keyService,
		ChangesetService:  changesetService,
		ValidationService: validationService,
		ValueService:      valueService,
		Translator:        translator,
	}
}

func (h *Handler) User(c echo.Context) *auth.User {
	return auth.GetUserFromEchoContext(c)
}

func (h *Handler) CurrentChangesetID(c echo.Context) uint {
	changesetId, ok := c.Get(constants.ChangesetIdKey).(uint)
	if !ok {
		return 0
	}

	return changesetId
}

func (h *Handler) EnsureChangesetID(c echo.Context) (uint, error) {
	changesetId := h.CurrentChangesetID(c)
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
	ctx = context.WithValue(ctx, constants.ChangesetIdContextKey, h.CurrentChangesetID(c))

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
