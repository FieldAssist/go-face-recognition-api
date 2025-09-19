package validation

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Init initializes the global validator instance
func Init() {
	if validate == nil {
		validate = validator.New(validator.WithRequiredStructEnabled())
	}
}

// ValidateStruct validates a struct using the global validator
func ValidateStruct(v interface{}) error {
	if validate == nil {
		Init()
	}
	if err := validate.Struct(v); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return nil
}

