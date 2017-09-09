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

var strNotEmptyV = notEmptyStrValidator{}

// NewStrNotEmptyValidator returns a validator to validate that the value must
// not be an empty string.
func NewStrNotEmptyValidator() Validator {
	return strNotEmptyV
}

type strArrayValidator struct {
	sArray []string
}

func (a strArrayValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	for _, v := range a.sArray {
		if s == v {
			return nil
		}
	}
	return fmt.Errorf("the value %s is invalid", s)
}

// NewStrArrayValidator returns a validator to validate that the value is in
// the array.
func NewStrArrayValidator(array []string) Validator {
	return strArrayValidator{array}
}

type regexpValidator struct {
	re *regexp.Regexp
}

func (r regexpValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	if r.re.MatchString(s) {
		return nil
	}
	return fmt.Errorf("the value '%s' doesn't match the regexp '%s'", s, r.re)
}

// NewRegexpValidator returns a validator to validate whether the value match
// the regular expression.
//
// If the regular expression can't be parsed, it will panic.
//
// This validator will call the method regexp.MatchString(s) when validating.
// So it is equal to regexp.MustCompile(pattern).MatchString(s).
func NewRegexpValidator(pattern string) Validator {
	return regexpValidator{regexp.MustCompile(pattern)}
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

var urlV = urlValidator{}

// NewURLValidator returns a validator to validate whether a url is valid.
func NewURLValidator() Validator {
	return urlV
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

var ipV = ipValidator{}

// NewIPValidator returns a validator to validate whether an ip is valid.
func NewIPValidator() Validator {
	return ipV
}

type integerRangeValidator struct {
	min int64
	max int64
}

func (r integerRangeValidator) Validate(v interface{}) error {
	i, err := toInt64(v)
	if err != nil {
		return err
	}
	if r.min > i || i > r.max {
		return fmt.Errorf("the value %d is not between %d and %d",
			i, r.min, r.max)
	}
	return nil
}

// NewIntegerRangeValidator returns a validator to validate whether the integer
// value is between the min and the max.
func NewIntegerRangeValidator(min, max int64) Validator {
	return integerRangeValidator{min: min, max: max}
}

type floatRangeValidator struct {
	min float64
	max float64
}

func (r floatRangeValidator) Validate(v interface{}) error {
	f, err := toFloat64(v)
	if err != nil {
		return err
	}
	if r.min > f || f > r.max {
		return fmt.Errorf("the value %f is not between %f and %f",
			f, r.min, r.max)
	}
	return nil
}

// NewFloatRangeValidator returns a validator to validate whether the float
// value is between the min and the max.
func NewFloatRangeValidator(min, max float64) Validator {
	return floatRangeValidator{min: min, max: max}
}

var portValidator = NewIntegerRangeValidator(0, 65535)

// NewPortValidator returns a validator to validate whether a port is between
// 0 and 65535.
func NewPortValidator() Validator {
	return portValidator
}

type emailValidator struct{}

func (e emailValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	_, err = mail.ParseAddress(s)
	return err
}

var emailV = emailValidator{}

// NewEmailValidator returns a validator to validate whether an email is valid.
func NewEmailValidator() Validator {
	return emailV
}

type addressValidator struct{}

func (a addressValidator) Validate(v interface{}) error {
	s, err := toString(v)
	if err != nil {
		return err
	}
	_, _, err = net.SplitHostPort(s)
	return err
}

var addressV = addressValidator{}

// NewAddressValidator returns a validator to validate whether an address is
// like "host:port", "host%zone:port", "[host]:port" or "[host%zone]:port".
//
// This validator uses net.SplitHostPort() to validate it.
func NewAddressValidator() Validator {
	return addressV
}
