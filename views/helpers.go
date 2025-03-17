package views

import (
	"context"

	"github.com/a-h/templ"
	"github.com/necroskillz/config-service/auth"
	"github.com/necroskillz/config-service/constants"
)

func User(ctx context.Context) *auth.User {
	return auth.GetUserFromContext(ctx)
}

func ValidationErrorClass(err string) templ.KeyValue[string, bool] {
	return templ.KV("validation-error", err != "")
}

func CurrentChangesetID(ctx context.Context) uint {
	changesetID, ok := ctx.Value(constants.ChangesetIdContextKey).(uint)
	if !ok {
		return 0
	}

	return changesetID
}

type ViewData struct {
	Error            string
	ValidationErrors map[string]string
}

type ViewDataSetter interface {
	SetError(err string)
	SetValidationErrors(errors map[string]string)
}

func (v *ViewData) SetError(err string) {
	v.Error = err
}

func (v *ViewData) SetValidationErrors(errors map[string]string) {
	v.ValidationErrors = errors
}
