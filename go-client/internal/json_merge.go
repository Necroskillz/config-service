package internal

func MergeObjects(a, b any) any {
	switch a := a.(type) {
	case map[string]any:
		switch b := b.(type) {
		case map[string]any:
			for k, v := range b {
				a[k] = MergeObjects(a[k], v)
			}
		default:
			return b
		}

		return a
	default:
		return b
	}
}
