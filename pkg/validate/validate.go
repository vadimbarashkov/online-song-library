package validate

import (
	"time"

	"github.com/go-playground/validator/v10"
)

func ReleaseDateValidation(fl validator.FieldLevel) bool {
	date := fl.Field().String()
	_, err := time.Parse("02.01.2006", date)
	return err == nil
}
