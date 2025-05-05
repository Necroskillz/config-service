package str

// FromPtr returns the string value of a pointer to a string, or an empty string if the pointer is nil.
func FromPtr(s *string) string {
	if s == nil {
		return ""
	}

	return *s
}

// ToPtr returns a pointer to a string, or nil if the string is empty.
func ToPtr(s string) *string {
	if s == "" {
		return nil
	}

	return &s
}
