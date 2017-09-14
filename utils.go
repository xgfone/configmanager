package config

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
)

func inMap(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// IsZero returns true if the value is ZERO, or false.
func IsZero(v interface{}) bool {
	ok, _ := template.IsTrue(v)
	return !ok
}

// bool2Int converts the bool to int64.
func bool2Int(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// ToBool does the best to convert a certain value to bool
//
// For "t", "T", "1", "true", "True", "TRUE", it's true.
// For "f", "F", "0", "false", "False", "FALSE", it's false.
func ToBool(v interface{}) (bool, error) {
	switch v.(type) {
	case string:
		_v := v.(string)
		switch _v {
		case "t", "T", "1", "true", "True", "TRUE":
			return true, nil
		case "f", "F", "0", "false", "False", "FALSE", "":
			return false, nil
		default:
			return false, fmt.Errorf("unrecognized bool string: %s", _v)
		}
	}
	return !IsZero(v), nil
}

// ToInt64 does the best to convert a certain value to int64.
func ToInt64(_v interface{}) (v int64, err error) {
	switch _v.(type) {
	case complex64, complex128:
		v = int64(real(reflect.ValueOf(_v).Complex()))
	case bool:
		v = int64(bool2Int(_v.(bool)))
	case int, int8, int16, int32, int64:
		v = reflect.ValueOf(_v).Int()
	case uint, uint8, uint16, uint32, uint64:
		v = int64(reflect.ValueOf(_v).Uint())
	case float32, float64:
		v = int64(reflect.ValueOf(_v).Float())
	case string:
		return strconv.ParseInt(_v.(string), 10, 64)
	default:
		err = fmt.Errorf("unknown type of %t", _v)
	}
	return
}

// ToUint64 does the best to convert a certain value to uint64.
func ToUint64(_v interface{}) (v uint64, err error) {
	switch _v.(type) {
	case complex64, complex128:
		v = uint64(real(reflect.ValueOf(_v).Complex()))
	case bool:
		v = uint64(bool2Int(_v.(bool)))
	case int, int8, int16, int32, int64:
		v = reflect.ValueOf(_v).Uint()
	case uint, uint8, uint16, uint32, uint64:
		v = uint64(reflect.ValueOf(_v).Uint())
	case float32, float64:
		v = uint64(reflect.ValueOf(_v).Float())
	case string:
		return strconv.ParseUint(_v.(string), 10, 64)
	default:
		err = fmt.Errorf("unknown type of %t", _v)
	}
	return
}

// ToFloat64 does the best to convert a certain value to float64.
func ToFloat64(_v interface{}) (v float64, err error) {
	switch _v.(type) {
	case complex64, complex128:
		v = float64(real(reflect.ValueOf(_v).Complex()))
	case bool:
		v = float64(bool2Int(_v.(bool)))
	case int, int8, int16, int32, int64:
		v = float64(reflect.ValueOf(_v).Int())
	case uint, uint8, uint16, uint32, uint64:
		v = float64(reflect.ValueOf(_v).Uint())
	case float32, float64:
		v = reflect.ValueOf(_v).Float()
	case string:
		return strconv.ParseFloat(_v.(string), 64)
	default:
		err = fmt.Errorf("unknown type of %t", _v)
	}
	return
}

// ToString does the best to convert a certain value to string.
func ToString(_v interface{}) (v string, err error) {
	switch _v.(type) {
	case string:
		v = _v.(string)
	case []byte:
		v = string(_v.([]byte))
	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32,
		uint64:
		v = fmt.Sprintf("%d", _v)
	case float32, float64:
		v = fmt.Sprintf("%f", _v)
	default:
		err = fmt.Errorf("unknown type of %t", _v)
	}
	return
}
