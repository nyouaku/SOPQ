// ===== FILE: backend/pkg/validator/validator.go =====
package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate melakukan validasi struct menggunakan tag validator
func Validate(data interface{}) error {
	return validate.Struct(data)
}

// FormatValidationError memformat validation error menjadi pesan yang user-friendly
func FormatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var messages []string
		for _, e := range validationErrors {
			messages = append(messages, fmt.Sprintf("Field '%s' gagal validasi '%s'", e.Field(), e.Tag()))
		}
		return strings.Join(messages, ", ")
	}
	return err.Error()
}
// ===== END FILE =====
