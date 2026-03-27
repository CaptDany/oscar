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
	return level.Validate() == nil
}
