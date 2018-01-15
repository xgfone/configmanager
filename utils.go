/*
Copyright 2017 xgfone

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
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
		case "t", "T", "1", "on", "On", "ON", "true", "True", "TRUE":
			return true, nil
		case "f", "F", "0", "off", "Off", "OFF", "false", "False", "FALSE", "":
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
		err = fmt.Errorf("unknown type of %T", _v)
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
		err = fmt.Errorf("unknown type of %T", _v)
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
		err = fmt.Errorf("unknown type of %T", _v)
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
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToStringSlice does the best to convert a certain value to []string.
func ToStringSlice(_v interface{}) (v []string, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]string, 0, len(vs))
		for _, s := range vs {
			s = strings.TrimSpace(s)
			if s != "" {
				v = append(v, s)
			}
		}
	case []string:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToIntSlice does the best to convert a certain value to []int.
func ToIntSlice(_v interface{}) (v []int, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]int, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := ToInt64(s)
			if err != nil {
				return nil, err
			}
			v = append(v, int(i))
		}
	case []int:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToInt64Slice does the best to convert a certain value to []int64.
func ToInt64Slice(_v interface{}) (v []int64, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]int64, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := ToInt64(s)
			if err != nil {
				return nil, err
			}
			v = append(v, i)
		}
	case []int64:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToUintSlice does the best to convert a certain value to []uint.
func ToUintSlice(_v interface{}) (v []uint, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]uint, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := ToUint64(s)
			if err != nil {
				return nil, err
			}
			v = append(v, uint(i))
		}
	case []uint:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToUint64Slice does the best to convert a certain value to []uint64.
func ToUint64Slice(_v interface{}) (v []uint64, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]uint64, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := ToUint64(s)
			if err != nil {
				return nil, err
			}
			v = append(v, i)
		}
	case []uint64:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToFloat64Slice does the best to convert a certain value to []float64.
func ToFloat64Slice(_v interface{}) (v []float64, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]float64, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := ToFloat64(s)
			if err != nil {
				return nil, err
			}
			v = append(v, i)
		}
	case []float64:
		v = vv
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}
