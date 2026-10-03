package shared

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type ValidationErrorResponse struct {
	Title      string            `json:"title"`
	Details    string            `json:"details"`
	Validation map[string]string `json:"validation"`
}

func ValidationResponse(
	err error,
	message func(validator.FieldError) string,
) ValidationErrorResponse {
	var validationErrors validator.ValidationErrors
	validation := make(map[string]string)

	if errors.As(err, &validationErrors) {
		for _, fieldError := range validationErrors {
			validation[fieldError.Field()] = message(fieldError)
		}
	}

	return ValidationErrorResponse{
		Title:      "The payload is invalid",
		Details:    "The payload has one or more validation errors, please fix them and try again.",
		Validation: validation,
	}
}
