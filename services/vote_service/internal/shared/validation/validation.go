package validation

import (
	"github.com/go-playground/validator/v10"
)

type StructValidator struct {
    Validating *validator.Validate
}
func (s *StructValidator) Validate(i any) error {
    return s.Validating.Struct(i)
}