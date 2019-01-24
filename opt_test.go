package config

import (
	"testing"
	"time"
)

func TestOpt(t *testing.T) {
	if newBaseOpt("", "bool", nil, "", boolType).Zero().(bool) {
		t.Fail()
	}
	if newBaseOpt("", "int", nil, "", intType).Zero().(int) != 0 {
		t.Fail()
	}
	if newBaseOpt("", "float", nil, "", float32Type).Zero().(float32) != 0 {
		t.Fail()
	}
	if newBaseOpt("", "string", nil, "", stringType).Zero().(string) != "" {
		t.Fail()
	}
	if newBaseOpt("", "duration", nil, "", durationType).Zero().(time.Duration) != 0 {
		t.Fail()
	}
	if newBaseOpt("", "time", nil, "", timeType).Zero().(time.Time) != *new(time.Time) {
		t.Fail()
	}
	if len(newBaseOpt("", "durations", nil, "", durationsType).Zero().([]time.Duration)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "times", nil, "", timesType).Zero().([]time.Time)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "strings", nil, "", stringsType).Zero().([]string)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "ints", nil, "", intsType).Zero().([]int)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "int64s", nil, "", int64sType).Zero().([]int64)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "uints", nil, "", uintsType).Zero().([]uint)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "uint64s", nil, "", uint64sType).Zero().([]uint64)) != 0 {
		t.Fail()
	}
	if len(newBaseOpt("", "float64s", nil, "", float64sType).Zero().([]float64)) != 0 {
		t.Fail()
	}
}
