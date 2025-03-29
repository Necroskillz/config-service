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

func Primary() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-primary text-white")
		return options
	}
}

func Secondary() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-secondary text-black")
		return options
	}
}

func Danger() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-danger text-white")
		return options
	}
}

func Success() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-success text-white")
		return options
	}
}

func Warning() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-warning text-white")
		return options
	}
}

func Info() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-info text-white")
		return options
	}
}

func Ghost() ElementOption {
	return func(options *ElementOptions) *ElementOptions {
		options.Classes = append(options.Classes, "bg-ghost text-white")
		return options
	}
}
