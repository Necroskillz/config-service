package handler

import (
	"fmt"

	"github.com/necroskillz/config-service/views/components"
)

func MakeSelectOptions[T any](items []T, fn func(item T) (uint, string)) []components.SelectOption {
	options := []components.SelectOption{}
	for _, item := range items {
		value, text := fn(item)
		options = append(options, components.SelectOption{Value: fmt.Sprintf("%d", value), Text: text})
	}
	return options
}
