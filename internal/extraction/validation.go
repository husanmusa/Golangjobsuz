package extraction

import (
	"github.com/go-playground/validator/v10"
)

func defaultValidator() Validator {
	return validator.New()
}
