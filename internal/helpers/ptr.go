package helpers

// Ptr creates a pointer to any given type
func Ptr[T any](v T) *T { return &v }
