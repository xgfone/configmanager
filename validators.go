package config

import (
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
)

var (
	errNil       = fmt.Errorf("the value is nil")
	errStrEmtpy  = fmt.Errorf("the string is empty")
	errStrType   = fmt.Errorf("the value is not string type")
	errIntType   = fmt.Errorf("the value is not an integer type")
	errFloatType = fmt.Errorf("the value is not an float type")
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

func toInt64(v interface{}) (int64, error) {
	if v == nil {
		return 0, errNil
	}

	switch v.(type) {
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(v).Int(), nil
	case uint, uint8, uint16, uint32, uint64:
		return int64(reflect.ValueOf(v).Uint()), nil
	default:
		return 0, errIntType
	}
}

func toFloat64(v interface{}) (float64, error) {
	if v == nil {
		return 0, errNil
	}

	switch v.(type) {
	case float32, float64:
		return reflect.ValueOf(v).Float(), nil
	default:
		return 0, errFloatType
	}
}

// ValidatorFunc is a wrapper of a function validator.
type ValidatorFunc func(v interface{}) error

// Validate implements the method Validate of the interface Validator.
func (f ValidatorFunc) Validate(v interface{}) error {
	return f(v)
}

// NewStrLenValidator returns a validator to validate that the length of the
// string must be between min and max.
func NewStrLenValidator(min, max int) Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}

		_len := len(s)
		if _len > max || _len < min {
			return fmt.Errorf("the length of %s is %d, not between %d and %d",
				s, _len, min, max)
		}
		return nil
	})
}

// NewStrNotEmptyValidator returns a validator to validate that the value must
// not be an empty string.
func NewStrNotEmptyValidator() Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}

		if len(s) == 0 {
			return errStrEmtpy
		}
		return nil
	})
}

// NewStrArrayValidator returns a validator to validate that the value is in
// the array.
func NewStrArrayValidator(array []string) Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		for _, v := range array {
			if s == v {
				return nil
			}
		}
		return fmt.Errorf("the value %s is not in %v", s, array)
	})
}

// NewRegexpValidator returns a validator to validate whether the value match
// the regular expression.
//
// This validator uses regexp.MatchString(pattern, s) to validate it.
func NewRegexpValidator(pattern string) Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}

		if ok, err := regexp.MatchString(pattern, s); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("'%s' doesn't match the value '%s'", s, pattern)
		}
		return nil
	})
}

// NewURLValidator returns a validator to validate whether a url is valid.
func NewURLValidator() Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		_, err = url.Parse(s)
		return err
	})
}

// NewIPValidator returns a validator to validate whether an ip is valid.
func NewIPValidator() Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		if net.ParseIP(s) == nil {
			return fmt.Errorf("the value is not a valid ip")
		}
		return nil
	})
}

// NewIntegerRangeValidator returns a validator to validate whether the integer
// value is between the min and the max.
//
// This validator can be used to validate the value of the type int, int8,
// int16, int32, int64, uint, uint8, uint16, uint32, uint64.
func NewIntegerRangeValidator(min, max int64) Validator {
	return ValidatorFunc(func(v interface{}) error {
		i, err := toInt64(v)
		if err != nil {
			return err
		}
		if min > i || i > max {
			return fmt.Errorf("the value %d is not between %d and %d",
				i, min, max)
		}
		return nil
	})
}

// NewFloatRangeValidator returns a validator to validate whether the float
// value is between the min and the max.
//
// This validator can be used to validate the value of the type float32 and
// float64.
func NewFloatRangeValidator(min, max float64) Validator {
	return ValidatorFunc(func(v interface{}) error {
		f, err := toFloat64(v)
		if err != nil {
			return err
		}
		if min > f || f > max {
			return fmt.Errorf("the value %f is not between %f and %f",
				f, min, max)
		}
		return nil
	})
}

// NewPortValidator returns a validator to validate whether a port is between
// 0 and 65535.
func NewPortValidator() Validator {
	return NewIntegerRangeValidator(0, 65535)
}

// NewEmailValidator returns a validator to validate whether an email is valid.
func NewEmailValidator() Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		_, err = mail.ParseAddress(s)
		return err
	})
}

// NewAddressValidator returns a validator to validate whether an address is
// like "host:port", "host%zone:port", "[host]:port" or "[host%zone]:port".
//
// This validator uses net.SplitHostPort() to validate it.
func NewAddressValidator() Validator {
	return ValidatorFunc(func(v interface{}) error {
		s, err := toString(v)
		if err != nil {
			return err
		}
		_, _, err = net.SplitHostPort(s)
		return err
	})
}
