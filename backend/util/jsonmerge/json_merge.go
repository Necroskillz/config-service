package jsonmerge

import (
	"maps"
)

func Merge(a, b any) any {
	switch a := a.(type) {
	case map[string]any:
		res := make(map[string]any)
		switch b := b.(type) {
		case map[string]any:
			maps.Copy(res, a)

			for k, v := range b {
				res[k] = Merge(a[k], v)
			}
		default:
			return b
		}

		return res
	default:
		return b
	}
}
