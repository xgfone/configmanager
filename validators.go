package config

import (
	"fmt"
	"net"
	"net/url"
)

var (
	errNil      = fmt.Errorf("the value is nil")
	errStrEmtpy = fmt.Errorf("the string is empty")
	errStrType  = fmt.Errorf("the type of the value is not string")
	errIntType  = fmt.Errorf("the type of the value is not int")
)

func toString(v interface{}) (string, error) {
	if v == nil {
		return "", errNil
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return "", errStrType
}

func toInt(v interface{}) (int, error) {
	if v == nil {
		return 0, errNil
	}
	if i, ok := v.(int); ok {
		return i, nil
	}
	return 0, errIntType
}

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
	s, err := toString(v)
	if err != nil {
		return err
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
	s, err := toString(v)
	if err != nil {
		return err
	}

	if len(s) == 0 {
		return errStrEmtpy
	}
	return nil
}

// NewStrNotEmptyValidator returns a validator to validate that the value must
// not be an empty string.
func NewStrNotEmptyValidator() Validator {
	return notEmptyStrValidator{}
}

type urlValidator struct{}

func (u urlValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	_, err = url.Parse(s)
	return err
}

// NewURLValidator returns a validator to validate whether a url is valid.
func NewURLValidator() Validator {
	return urlValidator{}
}

type ipValidator struct{}

func (i ipValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	if net.ParseIP(s) == nil {
		return fmt.Errorf("the value is not a valid ip")
	}
	return nil
}

// NewIPValidator returns a validator to validate whether an ip is valid.
func NewIPValidator() Validator {
	return ipValidator{}
}
