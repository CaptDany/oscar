package validator

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()

	validate.RegisterValidation("alphanumdash", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		match, _ := regexp.MatchString(`^[a-zA-Z0-9-_]+$`, value)
		return match
	})

	validate.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		if value == "" {
			return true
		}
		match, _ := regexp.MatchString(`^\+?[1-9]\d{1,14}$`, value)
		return match
	})

	validate.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		match, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]{0,62}$`, strings.ToLower(value))
		return match
	})
}

func ValidateStruct(s interface{}) error {
	return validate.Struct(s)
}

func ValidateField(level validator.FieldLevel) bool {
	return true
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	// Handle Echo's validation errors
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := e.Field()
			message := getValidationMessage(e.Tag())
			errors = append(errors, ValidationError{
				Field:   field,
				Message: message,
			})
		}
		return errors
	}

	// Handle generic errors
	errors = append(errors, ValidationError{
		Field:   "unknown",
		Message: err.Error(),
	})
	return errors
}

func getValidationMessage(tag string) string {
	switch tag {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	default:
		return "Invalid value"
	}
}
