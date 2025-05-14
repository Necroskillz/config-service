package ptr

type PtrToOptions struct {
	NilIfZero bool
}

type PtrToOptionFunc func(*PtrToOptions)

func WithNilIfZero() PtrToOptionFunc {
	return func(o *PtrToOptions) {
		o.NilIfZero = true
	}
}

func To[T comparable](v T, options ...PtrToOptionFunc) *T {
	opt := PtrToOptions{}
	for _, option := range options {
		option(&opt)
	}

	if opt.NilIfZero && v == *new(T) {
		return nil
	}

	return &v
}

func From[T any](v *T) T {
	if v == nil {
		var zero T
		return zero
	}

	return *v
}
