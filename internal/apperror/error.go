package apperror

import (
	"github.com/go-playground/validator/v10"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// NewErrorResponse creates a new error response
func NewErrorResponse(err string, message string) ErrorResponse {
	return ErrorResponse{
		Error:   err,
		Message: message,
	}
}

// ValidationError represents a single validation error detail
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// NewValidationErrorResponse creates a validation error response from binding errors
func NewValidationErrorResponse(err error) ErrorResponse {
	var details []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			details = append(details, ValidationError{
				Field:   e.Field(),
				Message: getValidationMessage(e),
			})
		}
	} else {
		details = append(details, ValidationError{
			Field:   "unknown",
			Message: err.Error(),
		})
	}

	return ErrorResponse{
		Error:   "validation_error",
		Message: "Validasyon hatası",
		Details: details,
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
