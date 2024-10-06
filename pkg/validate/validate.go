package validate

import (
	"time"

	"github.com/go-playground/validator/v10"
)

// ReleaseDateValidation checks if the release date is in the correct format.
// This function is designed to be used as a custom validation function
// with the go-playground validator library.
func ReleaseDateValidation(fl validator.FieldLevel) bool {
	date := fl.Field().String()
	_, err := time.Parse("02.01.2006", date)
	return err == nil
}
