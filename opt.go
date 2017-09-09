package config

import "fmt"

type optType int

func (ot optType) String() string {
	return optTypeMap[ot]
}

const (
	noneType optType = iota
	boolType
	stringType
	intType
	int8Type
	int16Type
	int32Type
	int64Type
	uintType
	uint8Type
	uint16Type
	uint32Type
	uint64Type
	float32Type
	float64Type

	stringsType
	intsType
	int64sType
	uintsType
	uint64sType
)

var optTypeMap = map[optType]string{
	noneType:    "none",
	boolType:    "bool",
	stringType:  "string",
	intType:     "int",
	int8Type:    "int8",
	int16Type:   "int16",
	int32Type:   "int32",
	int64Type:   "int64",
	uintType:    "uint",
	uint8Type:   "uint8",
	uint16Type:  "uint16",
	uint32Type:  "uint32",
	uint64Type:  "uint64",
	float32Type: "float32",
	float64Type: "float64",

	stringsType: "[]string",
	intsType:    "[]int",
	int64sType:  "[]int64",
	uintsType:   "[]uint",
	uint64sType: "[]uint64",
}

type baseOpt struct {
	name     string
	help     string
	short    string
	_default interface{}

	_type      optType
	validators []Validator
}

var _ ValidatorChainOpt = baseOpt{}

func newBaseOpt(short, name string, _default interface{}, help string,
	optType optType) baseOpt {
	o := baseOpt{
		short:    short,
		name:     name,
		help:     help,
		_default: _default,
		_type:    optType,
	}
	o.Default()
	return o
}

// SetValidators sets the validator chain
func (o baseOpt) SetValidators(vs []Validator) ValidatorChainOpt {
	o.validators = vs
	return o
}

// GetValidators returns the validator chain
func (o baseOpt) GetValidators() []Validator {
	return o.validators
}

// GetName returns the name of the option.
func (o baseOpt) Name() string {
	return o.name
}

// GetShort returns the shorthand name of the option.
func (o baseOpt) Short() string {
	return o.short
}

func (o baseOpt) IsBool() bool {
	if o._type == boolType {
		return true
	}
	return false
}

// GetHelp returns the help doc of the option.
func (o baseOpt) Help() string {
	return o.help
}

// GetDefault returns the default value of the option.
func (o baseOpt) Default() interface{} {
	if o._default == nil {
		return nil
	}

	switch o._type {
	case boolType:
		return o._default.(bool)
	case stringType:
		return o._default.(string)
	case intType:
		return o._default.(int)
	case int8Type:
		return o._default.(int8)
	case int16Type:
		return o._default.(int16)
	case int32Type:
		return o._default.(int32)
	case int64Type:
		return o._default.(int64)
	case uintType:
		return o._default.(uint)
	case uint8Type:
		return o._default.(uint8)
	case uint16Type:
		return o._default.(uint16)
	case uint32Type:
		return o._default.(uint32)
	case uint64Type:
		return o._default.(uint64)
	case float32Type:
		return o._default.(float32)
	case float64Type:
		return o._default.(float64)
	default:
		panic(fmt.Errorf("don't support the type %s", o._type))
	}
}

// Parse parses the value of the option to a certain type.
func (o baseOpt) Parse(data interface{}) (v interface{}, err error) {
	switch o._type {
	case boolType:
		return ToBool(data)
	case stringType:
		return ToString(data)
	case intType, int8Type, int16Type, int32Type, int64Type:
		v, err = ToInt64(data)
	case uintType, uint8Type, uint16Type, uint32Type, uint64Type:
		v, err = ToUint64(data)
	case float32Type, float64Type:
		v, err = ToFloat64(data)
	default:
		panic(fmt.Errorf("don't support the type %s", o._type))
	}

	if err != nil {
		return
	}

	switch o._type {
	// case uint64Type:
	// case int64Type:
	// case float64Type:
	case intType:
		v = int(v.(int64))
	case int8Type:
		v = int8(v.(int64))
	case int16Type:
		v = int16(v.(int64))
	case int32Type:
		v = int32(v.(int64))
	case uintType:
		v = uint(v.(uint64))
	case uint8Type:
		v = uint8(v.(uint64))
	case uint16Type:
		v = uint16(v.(uint64))
	case uint32Type:
		v = uint32(v.(uint64))
	case float32Type:
		v = float32(v.(float64))
	}
	return
}

// BoolOpt return a new bool option.
func BoolOpt(short, name string, _default bool, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, boolType)
}

// StrOpt return a new string option.
func StrOpt(short, name string, _default string, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, stringType)
}

// IntOpt return a new int option.
func IntOpt(short, name string, _default int, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, intType)
}

// Int8Opt return a new int8 option.
func Int8Opt(short, name string, _default int8, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, int8Type)
}

// Int16Opt return a new int16 option.
func Int16Opt(short, name string, _default int16, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, int16Type)
}

// Int32Opt return a new int32 option.
func Int32Opt(short, name string, _default int32, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, int32Type)
}

// Int64Opt return a new int64 option.
func Int64Opt(short, name string, _default int64, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, int64Type)
}

// UintOpt return a new uint option.
func UintOpt(short, name string, _default uint, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, uintType)
}

// Uint8Opt return a new uint8 option.
func Uint8Opt(short, name string, _default uint8, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, uint8Type)
}

// Uint16Opt return a new uint16 option.
func Uint16Opt(short, name string, _default uint16, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, uint16Type)
}

// Uint32Opt return a new uint32 option.
func Uint32Opt(short, name string, _default uint32, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, uint32Type)
}

// Uint64Opt return a new uint64 option.
func Uint64Opt(short, name string, _default uint64, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, uint64Type)
}

// Float32Opt return a new float32 option.
func Float32Opt(short, name string, _default float32, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, float32Type)
}

// Float64Opt return a new float64 option.
func Float64Opt(short, name string, _default float64, help string) ValidatorChainOpt {
	return newBaseOpt(short, name, _default, help, float64Type)
}
