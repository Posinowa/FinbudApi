package apperror

import (
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
	}
}

// ValidationError represents a validation error detail
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse represents validation errors
type ValidationErrorResponse struct {
	Error  string            `json:"error"`
	Errors []ValidationError `json:"errors"`
}

// NewValidationErrorResponse creates a validation error response from binding errors
func NewValidationErrorResponse(err error) ValidationErrorResponse {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getValidationMessage(e),
			})
		}
	} else {
		errors = append(errors, ValidationError{
			Field:   "unknown",
			Message: err.Error(),
		})
	}

	return ValidationErrorResponse{
		Error:  "validation_error",
		Errors: errors,
	}
}

func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "gt":
		return "Must be greater than " + e.Param()
	case "oneof":
		return "Must be one of: " + e.Param()
	case "uuid":
		return "Must be a valid UUID"
	default:
		return "Invalid value"
	}
}