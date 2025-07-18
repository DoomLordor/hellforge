package helpers

// ValueToPtr returns a pointer to a value
func ValueToPtr[T any](v T) *T {
	return &v
}

// PtrToValue returns a value from a pointer
func PtrToValue[T any](v *T) (result T) {
	if v == nil {
		return result
	}

	return *v
}

// PtrToValuePtr returns a pointer from a pointer value
func PtrToValuePtr[T any](v *T) (result *T) {
	if v == nil {
		var def T
		return &def
	}

	return v
}

// ValueToPtrWithoutDefault returns a pointer to a value only when value not default
func ValueToPtrWithoutDefault[K comparable](v K) *K {
	var def K
	if v == def {
		return nil
	}

	return &v
}
