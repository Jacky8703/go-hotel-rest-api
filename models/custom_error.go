package models

// custom error type for validation errors in services
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}
