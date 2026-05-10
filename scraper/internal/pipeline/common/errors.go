package common

import "fmt"

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func Required(field string) error {
	return ValidationError{Message: fmt.Sprintf("missing required field: %s", field)}
}
