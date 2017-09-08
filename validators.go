package config

import (
	"fmt"
)

var (
	errStrEmtpy = fmt.Errorf("the string is empty")
	errStrType  = fmt.Errorf("the type of the value is not string")
)

type strLenValidator struct {
	min int
	max int
}

// NewStrLenValidator returns a validator to validate that the length of the
// string must be between min and max.
func NewStrLenValidator(min, max int) Validator {
	return strLenValidator{min: min, max: max}
}

func (sv strLenValidator) Validate(v interface{}) error {
	if v == nil {
		return errStrEmtpy
	}

	s, ok := v.(string)
	if !ok {
		return errStrType
	}

	_len := len(s)
	if _len > sv.max || _len < sv.min {
		return fmt.Errorf("the length of %s is %d, not between %d and %d",
			s, _len, sv.min, sv.max)
	}
	return nil
}

type notEmptyStrValidator struct{}

func (e notEmptyStrValidator) Validate(v interface{}) error {
	if v != nil {
		if s, ok := v.(string); ok {
			if len(s) > 0 {
				return nil
			}
			return errStrEmtpy
		}
		return errStrType
	}
	return errStrEmtpy
}

// NewStrNotEmptyValidator returns a validator to validate that the value must
// not be an empty string.
func NewStrNotEmptyValidator() Validator {
	return notEmptyStrValidator{}
}
