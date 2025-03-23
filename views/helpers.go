package views

import (
	"context"

	"github.com/a-h/templ"
	"github.com/necroskillz/config-service/auth"
)

func User(ctx context.Context) *auth.User {
	return auth.GetUserFromContext(ctx)
}

func ValidationErrorClass(err string) templ.KeyValue[string, bool] {
	return templ.KV("validation-error", err != "")
}

type ViewData struct {
	Error            string
	ValidationErrors map[string]string
}

type ViewDataSetter interface {
	SetValidationErrors(errors map[string]string)
}

func (v *ViewData) SetValidationErrors(errors map[string]string) {
	v.ValidationErrors = errors
}
