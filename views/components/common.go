package components

import (
	"github.com/a-h/templ"
)

type ElementOptions struct {
	Attributes templ.Attributes
	Classes    templ.CSSClasses
}

type ElementOption func(options *ElementOptions) *ElementOptions

func NewElementOptions(options []ElementOption) *ElementOptions {
	opts := &ElementOptions{
		Attributes: templ.Attributes{},
		Classes:    templ.CSSClasses{},
	}
	for _, option := range options {
		option(opts)
	}
	return opts
}

func WithAttribute(name, value string) ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Attributes[name] = value
		return options
	}
}

func WithClass(class any) ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, class)
		return options
	}
}
