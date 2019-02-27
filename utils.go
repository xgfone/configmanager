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
	"strconv"
	"strings"
	"time"
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

// bool2Int64 converts the bool to int64.
func bool2Int64(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// ToBool does the best to convert any certain value to bool.
//
// When the value is string, for "t", "T", "1", "on", "On", "ON", "true",
// "True", "TRUE", it's true, for "f", "F", "0", "off", "Off", "OFF", "false",
// "False", "FALSE", "", it's false.
//
// For other types, if the value is ZERO of the type, it's false. Or it's true.
func ToBool(v interface{}) (bool, error) {
	switch _v := v.(type) {
	case nil:
		return false, nil
	case bool:
		return _v, nil
	case string:
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

// ToInt64 does the best to convert any certain value to int64.
func ToInt64(_v interface{}) (v int64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = bool2Int64(t)
	case string:
		v, err = strconv.ParseInt(t, 10, 64)
	case int:
		v = int64(t)
	case int8:
		v = int64(t)
	case int16:
		v = int64(t)
	case int32:
		v = int64(t)
	case int64:
		v = t
	case uint:
		v = int64(t)
	case uint8:
		v = int64(t)
	case uint16:
		v = int64(t)
	case uint32:
		v = int64(t)
	case uint64:
		v = int64(t)
	case float32:
		v = int64(t)
	case float64:
		v = int64(t)
	case complex64:
		v = int64(real(t))
	case complex128:
		v = int64(real(t))
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToUint64 does the best to convert any certain value to uint64.
func ToUint64(_v interface{}) (v uint64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = uint64(bool2Int64(t))
	case string:
		v, err = strconv.ParseUint(t, 10, 64)
	case int:
		v = uint64(t)
	case int8:
		v = uint64(t)
	case int16:
		v = uint64(t)
	case int32:
		v = uint64(t)
	case int64:
		v = uint64(t)
	case uint:
		v = uint64(t)
	case uint8:
		v = uint64(t)
	case uint16:
		v = uint64(t)
	case uint32:
		v = uint64(t)
	case uint64:
		v = t
	case float32:
		v = uint64(t)
	case float64:
		v = uint64(t)
	case complex64:
		v = uint64(real(t))
	case complex128:
		v = uint64(real(t))
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToFloat64 does the best to convert any certain value to float64.
func ToFloat64(_v interface{}) (v float64, err error) {
	switch t := _v.(type) {
	case nil:
	case bool:
		v = float64(bool2Int64(t))
	case string:
		v, err = strconv.ParseFloat(t, 64)
	case int:
		v = float64(t)
	case int8:
		v = float64(t)
	case int16:
		v = float64(t)
	case int32:
		v = float64(t)
	case int64:
		v = float64(t)
	case uint:
		v = float64(t)
	case uint8:
		v = float64(t)
	case uint16:
		v = float64(t)
	case uint32:
		v = float64(t)
	case uint64:
		v = float64(t)
	case float32:
		v = float64(t)
	case float64:
		v = t
	case complex64:
		v = float64(real(t))
	case complex128:
		v = real(t)
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToString does the best to convert any certain value to string.
//
// For time.Time, it will use time.RFC3339Nano to format it.
func ToString(_v interface{}) (v string, err error) {
	switch t := _v.(type) {
	case nil:
	case string:
		v = t
	case []byte:
		v = string(t)
	case bool:
		if t {
			v = "true"
		} else {
			v = "false"
		}
	case int:
		v = strconv.FormatInt(int64(t), 10)
	case int8:
		v = strconv.FormatInt(int64(t), 10)
	case int16:
		v = strconv.FormatInt(int64(t), 10)
	case int32:
		v = strconv.FormatInt(int64(t), 10)
	case int64:
		v = strconv.FormatInt(t, 10)
	case uint:
		v = strconv.FormatUint(uint64(t), 10)
	case uint8:
		v = strconv.FormatUint(uint64(t), 10)
	case uint16:
		v = strconv.FormatUint(uint64(t), 10)
	case uint32:
		v = strconv.FormatUint(uint64(t), 10)
	case uint64:
		v = strconv.FormatUint(t, 10)
	case float32:
		v = strconv.FormatFloat(float64(t), 'f', -1, 32)
	case float64:
		v = strconv.FormatFloat(t, 'f', -1, 64)
	case error:
		v = t.Error()
	case time.Time:
		v = t.Format(time.RFC3339Nano)
	case fmt.Stringer:
		v = t.String()
	default:
		err = fmt.Errorf("unknown type of %T", _v)
	}
	return
}

// ToStringSlice does the best to convert a certain value to []string.
//
// If the value is string, they are separated by the comma.
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
//
// If the value is string, they are separated by the comma.
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
//
// If the value is string, they are separated by the comma.
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
//
// If the value is string, they are separated by the comma.
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
//
// If the value is string, they are separated by the comma.
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
//
// If the value is string, they are separated by the comma.
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

// ToTimes does the best to convert a certain value to []time.Time.
//
// If the value is string, they are separated by the comma and the each value
// is parsed by the format, layout.
func ToTimes(layout string, _v interface{}) (v []time.Time, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]time.Time, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := time.Parse(layout, s)
			if err != nil {
				return nil, err
			}
			v = append(v, i)
		}
	case []time.Time:
		v = vv
	default:
		err = fmt.Errorf("unknown type of '%T'", _v)
	}
	return
}

// ToDurations does the best to convert a certain value to []time.Duration.
//
// If the value is string, they are separated by the comma and the each value
// is parsed by time.ParseDuration().
func ToDurations(_v interface{}) (v []time.Duration, err error) {
	switch vv := _v.(type) {
	case string:
		vs := strings.Split(vv, ",")
		v = make([]time.Duration, 0, len(vs))
		for _, s := range vs {
			if s = strings.TrimSpace(s); s == "" {
				continue
			}

			i, err := time.ParseDuration(s)
			if err != nil {
				return nil, err
			}
			v = append(v, i)
		}
	case []time.Duration:
		v = vv
	default:
		err = fmt.Errorf("unknown type of '%T'", _v)
	}
	return
}
