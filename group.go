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
	"reflect"
	"strings"
	"sync"
	"time"
)

// DefaultGroupName is the name of the default group.
const DefaultGroupName = "DEFAULT"

type option struct {
	opt   Opt
	isCli bool
}

// OptGroup is the group of the option.
type OptGroup struct {
	conf *Config
	lock sync.RWMutex

	fname  string
	name   string
	opts   map[string]option
	values map[string]interface{}
	fields map[string]reflect.Value
	defers [][2]interface{}
	ignore []string
}

// NewOptGroup returns a new OptGroup.
func newOptGroup(name, fullName string, conf *Config) *OptGroup {
	if name == "" || fullName == "" {
		panic(fmt.Errorf("the group name is empty"))
	}

	if conf == nil {
		panic(fmt.Errorf("Config is nil"))
	}

	return &OptGroup{
		conf:  conf,
		name:  name,
		fname: fullName,
		lock:  sync.RWMutex{},

		opts:   make(map[string]option, 8),
		values: make(map[string]interface{}, 8),
		fields: make(map[string]reflect.Value),
		defers: make([][2]interface{}, 0, 16),
		ignore: make([]string, 0, 4),
	}
}

//////////////////////////////////////////////////////////////////////////////
/// MetaData

// Name returns the name of the current group.
func (g *OptGroup) Name() string {
	return g.name
}

// FullName returns the full name of the current group.
func (g *OptGroup) FullName() string {
	return g.fname
}

// AllOpts returns all the registered options, including the CLI options.
func (g *OptGroup) AllOpts() []Opt {
	opts := make([]Opt, 0, len(g.opts))
	for _, opt := range g.opts {
		opts = append(opts, opt.opt)
	}
	return opts
}

// Opts returns all the registered options, except the CLI options.
func (g *OptGroup) Opts() []Opt {
	opts := make([]Opt, 0, len(g.opts))
	for _, opt := range g.opts {
		if !opt.isCli {
			opts = append(opts, opt.opt)
		}
	}
	return opts
}

// CliOpts returns all the registered CLI options, except the non-CLI options.
func (g *OptGroup) CliOpts() []Opt {
	opts := make([]Opt, 0, len(g.opts))
	for _, opt := range g.opts {
		if opt.isCli {
			opts = append(opts, opt.opt)
		}
	}
	return opts
}

// HasOpt reports whether the group contains the option named 'name'.
func (g *OptGroup) HasOpt(name string) bool {
	_, ok := g.opts[name]
	return ok
}

// HasGroup reports whether the group contains the sub-group named 'name'.
func (g *OptGroup) HasGroup(name string) bool {
	if name == "" {
		return false
	}
	return g.conf.HasGroup(g.conf.mergeGroupName(g.name, name))
}

// Group returns the sub-group by the name.
func (g *OptGroup) Group(name string) *OptGroup {
	return g.conf.Group(g.conf.mergeGroupName(g.name, name))
}

// NewGroup news and returns a sub-group named group.
func (g *OptGroup) NewGroup(name string) *OptGroup {
	return g.NewGroup(g.conf.mergeGroupName(g.name, name))
}

///////////////////////////////////////////////////////////////////////////////
/// Set the option value.

func (g *OptGroup) parseOptValue(name string, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	opt, ok := g.opts[name]
	if !ok {
		return nil, fmt.Errorf("no the option '%s' in the group '%s'", name, g.name)
	}

	var err error
	if value, err = opt.opt.Parse(value); err != nil {
		return nil, err
	}

	// The option has a validator.
	if v, ok := opt.opt.(Validator); ok {
		if err = v.Validate(g.name, name, value); err != nil {
			return nil, err
		}
	}

	// The option has a validator chain.
	if vc, ok := opt.opt.(ValidatorChainOpt); ok {
		vs := vc.GetValidators()
		if len(vs) > 0 {
			for _, v := range vs {
				if err = v.Validate(g.name, name, value); err != nil {
					return nil, err
				}
			}
		}
	}

	return value, nil
}

func (g *OptGroup) _setOptValue(name string, value interface{}) {
	func() {
		g.lock.Lock()
		defer g.lock.Unlock()

		g.values[name] = value
		if field, ok := g.fields[name]; ok {
			field.Set(reflect.ValueOf(value))
		}
	}()

	g.conf.debug("Set [%s]:[%s] to [%v]", g.name, name, value)
	if g.conf.watch != nil {
		g.conf.watch(g.name, name, value)
	}
}

func (g *OptGroup) setOptValue(name string, value interface{}) (err error) {
	if value, err = g.parseOptValue(name, value); err == nil {
		g._setOptValue(name, value)
	}
	return
}

func (g *OptGroup) _deferSetOptValue() {
	var cache [][2]interface{}

	func() {
		g.lock.Lock()
		defer g.lock.Unlock()

		_len := len(g.defers)
		cache = make([][2]interface{}, _len)
		for i := _len - 1; i >= 0; i-- {
			v := g.defers[i]
			cache[_len-i-1] = v

			name, value := v[0].(string), v[1]
			g.values[name] = value
			if field, ok := g.fields[name]; ok {
				field.Set(reflect.ValueOf(value))
			}
		}

		// Clear
		g.defers = g.defers[:0]
	}()

	for _, v := range cache {
		name, value := v[0].(string), v[1]
		g.conf.debug("Defer to set [%s]:[%s] to [%v]", g.name, name, value)
		if g.conf.watch != nil {
			g.conf.watch(g.name, name, value)
		}
	}
}

func (g *OptGroup) deferSetOptValue(name string, value interface{}) (err error) {
	for _, s := range g.ignore {
		if s == name {
			return g.setOptValue(name, value)
		}
	}

	if value, err = g.parseOptValue(name, value); err == nil && !IsZero(value) {
		g.lock.Lock()
		g.defers = append(g.defers, [2]interface{}{name, value})
		g.lock.Unlock()
	}
	return
}

func (g *OptGroup) addIgnoredDeferOption(name string) {
	g.ignore = append(g.ignore, name)
}

// Check whether the required option has no value or a ZORE value.
func (g *OptGroup) checkRequiredOption() (err error) {
	for name, opt := range g.opts {
		if _, ok := g.values[name]; !ok {
			if v := opt.opt.Default(); v != nil {
				if err = g.setOptValue(name, v); err != nil {
					return
				}
				continue
			}

			if g.conf.isZero {
				if v := opt.opt.Zero(); v != nil {
					if err = g.setOptValue(name, opt.opt.Zero()); err != nil {
						return
					}
					continue
				}
			}

			if g.conf.isRequired {
				return fmt.Errorf("the option '%s' in the group '%s' has no value",
					name, g.name)
			}
		}
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
/// Register Options

func (g *OptGroup) registerStruct(s interface{}, cli bool) {
	sv := reflect.ValueOf(s)
	if sv.IsNil() || !sv.IsValid() {
		panic(fmt.Errorf("the struct is invalid or can't be set"))
	}

	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}

	if sv.Kind() != reflect.Struct {
		panic(fmt.Errorf("the struct is not a struct"))
	}

	g.registerStructByValue(g.name, sv, cli)
}

func (g *OptGroup) registerStructByValue(parent string, sv reflect.Value, cli bool) {
	if sv.Kind() == reflect.Ptr {
		sv = sv.Elem()
	}
	st := sv.Type()

	// Register the field as the option
	num := sv.NumField()
	for i := 0; i < num; i++ {
		field := st.Field(i)
		fieldV := sv.Field(i)

		// Check whether the field can be set.
		if !fieldV.CanSet() {
			continue
		}

		name := strings.ToLower(field.Name)
		tagname := strings.TrimSpace(field.Tag.Get("name"))
		if tagname == "-" {
			continue
		} else if tagname != "" {
			name = tagname
		}

		isCli := cli
		if _cli := strings.TrimSpace(field.Tag.Get("cli")); _cli != "" {
			switch _cli {
			case "1", "t", "T", "on", "On", "ON", "true", "True", "TRUE":
				isCli = true
			case "0", "f", "F", "off", "Off", "OFF", "false", "False", "FALSE":
				isCli = false
			default:
				panic(fmt.Errorf("no support '%s' for cli", field.Tag.Get("cli")))
			}
		}

		gname := g.name
		taggroup, resetgroup := field.Tag.Lookup("group")
		if resetgroup {
			taggroup = strings.TrimSpace(taggroup)
			gname = taggroup
		}

		// Check whether the field is the struct.
		if t := field.Type.Kind(); t == reflect.Struct {
			if _, ok := fieldV.Interface().(time.Time); !ok {
				parentGroup := g.conf.mergeGroupName(parent, name)
				if resetgroup {
					if strings.Contains(taggroup, g.conf.groupSep) {
						parentGroup = strings.Trim(taggroup, g.conf.groupSep)
					} else if taggroup == "" {
						parentGroup = g.conf.groupName // Default Group
					} else {
						parentGroup = g.conf.mergeGroupName(parent, taggroup)
					}
				}

				g.conf.getGroupByName(parentGroup, true).registerStructByValue(parentGroup, fieldV, isCli)
				continue
			}
		}

		_type := getOptType(fieldV)
		if _type == int64Type {
			if _, ok := fieldV.Interface().(time.Duration); ok {
				_type = durationType
			}
		}

		// Get the short name from the tag "short"
		short := strings.TrimSpace(field.Tag.Get("short"))

		// Get the help doc from the tag "help"
		help := strings.TrimSpace(field.Tag.Get("help"))

		// Get the default value from the tag "default"
		var err error
		var _default interface{}
		if v, ok := field.Tag.Lookup("default"); ok {
			if _default, err = parseOpt(strings.TrimSpace(v), _type); err != nil {
				panic(fmt.Errorf("can't parse the default in the field %s: %s",
					field.Name, err))
			}
		}

		opt := newBaseOpt(short, name, _default, help, _type)
		group := g.conf.getGroupByName(gname, true)
		group.registerOpt(isCli, opt)
		group.fields[name] = fieldV
	}
}

// registerOpt registers the option into the group.
//
// The first argument, cli, indicates whether the option is as the CLI option,
// too.
func (g *OptGroup) registerOpt(cli bool, opt Opt) {
	if opt == nil {
		return
	}

	// g.conf.debug("+++Register group=%s, name=%s, cli=%t", g.name, opt.Name(), cli)
	if _, ok := g.opts[opt.Name()]; ok {
		// panic(fmt.Errorf("the option '%s' has been registered into the group '%s'",
		// 	opt.Name(), g.name))
	}
	g.opts[opt.Name()] = option{isCli: cli, opt: opt}
	g.conf.debug("Register group=%s, name=%s, cli=%t", g.name, opt.Name(), cli)
}

///////////////////////////////////////////////////////////////////////////////
/// Get the value from the current group.

// Value returns the value of the option.
//
// Return nil if the option does not exist.
func (g *OptGroup) Value(name string) (v interface{}) {
	g.lock.RLock()
	v = g.values[name]
	g.lock.RUnlock()
	return v
}

// V is the short for g.Value(name).
func (g *OptGroup) V(name string) interface{} {
	return g.Value(name)
}

func (g *OptGroup) getValue(name string, _type optType) (interface{}, error) {
	opt := g.Value(name)
	if opt == nil {
		return nil, fmt.Errorf("the group '%s' has no option '%s'", g.name, name)
	}

	switch _type {
	case boolType:
		if v, ok := opt.(bool); ok {
			return v, nil
		}
	case stringType:
		if v, ok := opt.(string); ok {
			return v, nil
		}
	case intType:
		if v, ok := opt.(int); ok {
			return v, nil
		}
	case int8Type:
		if v, ok := opt.(int8); ok {
			return v, nil
		}
	case int16Type:
		if v, ok := opt.(int16); ok {
			return v, nil
		}
	case int32Type:
		if v, ok := opt.(int32); ok {
			return v, nil
		}
	case int64Type:
		if v, ok := opt.(int64); ok {
			return v, nil
		}
	case uintType:
		if v, ok := opt.(uint); ok {
			return v, nil
		}
	case uint8Type:
		if v, ok := opt.(uint8); ok {
			return v, nil
		}
	case uint16Type:
		if v, ok := opt.(uint16); ok {
			return v, nil
		}
	case uint32Type:
		if v, ok := opt.(uint32); ok {
			return v, nil
		}
	case uint64Type:
		if v, ok := opt.(uint64); ok {
			return v, nil
		}
	case float32Type:
		if v, ok := opt.(float32); ok {
			return v, nil
		}
	case float64Type:
		if v, ok := opt.(float64); ok {
			return v, nil
		}
	case durationType:
		if v, ok := opt.(time.Duration); ok {
			return v, nil
		}
	case timeType:
		if v, ok := opt.(time.Time); ok {
			return v, nil
		}
	case stringsType:
		if v, ok := opt.([]string); ok {
			return v, nil
		}
	case intsType:
		if v, ok := opt.([]int); ok {
			return v, nil
		}
	case int64sType:
		if v, ok := opt.([]int64); ok {
			return v, nil
		}
	case uintsType:
		if v, ok := opt.([]uint); ok {
			return v, nil
		}
	case uint64sType:
		if v, ok := opt.([]uint64); ok {
			return v, nil
		}
	case float64sType:
		if v, ok := opt.([]float64); ok {
			return v, nil
		}
	case durationsType:
		if v, ok := opt.([]time.Duration); ok {
			return v, nil
		}
	case timesType:
		if v, ok := opt.([]time.Time); ok {
			return v, nil
		}
	default:
		return nil, fmt.Errorf("don't support the type '%s'", _type)
	}
	return nil, fmt.Errorf("the option '%s' in the group '%s' is not the type '%s'",
		name, g.name, _type)
}

// BoolE returns the option value, the type of which is bool.
//
// Return an error if no the option or the type of the option isn't bool.
func (g *OptGroup) BoolE(name string) (bool, error) {
	v, err := g.getValue(name, boolType)
	if err != nil {
		return false, err
	}
	return v.(bool), nil
}

// BoolD is the same as BoolE, but returns the default if there is an error.
func (g *OptGroup) BoolD(name string, _default bool) bool {
	if value, err := g.BoolE(name); err == nil {
		return value
	}
	return _default
}

// Bool is the same as BoolE, but panic if there is an error.
func (g *OptGroup) Bool(name string) bool {
	value, err := g.BoolE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// StringE returns the option value, the type of which is string.
//
// Return an error if no the option or the type of the option isn't string.
func (g *OptGroup) StringE(name string) (string, error) {
	v, err := g.getValue(name, stringType)
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// StringD is the same as StringE, but returns the default if there is an error.
func (g *OptGroup) StringD(name, _default string) string {
	if value, err := g.StringE(name); err == nil {
		return value
	}
	return _default
}

// String is the same as StringE, but panic if there is an error.
func (g *OptGroup) String(name string) string {
	value, err := g.StringE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// IntE returns the option value, the type of which is int.
//
// Return an error if no the option or the type of the option isn't int.
func (g *OptGroup) IntE(name string) (int, error) {
	v, err := g.getValue(name, intType)
	if err != nil {
		return 0, err
	}
	return v.(int), nil
}

// IntD is the same as IntE, but returns the default if there is an error.
func (g *OptGroup) IntD(name string, _default int) int {
	if value, err := g.IntE(name); err == nil {
		return value
	}
	return _default
}

// Int is the same as IntE, but panic if there is an error.
func (g *OptGroup) Int(name string) int {
	value, err := g.IntE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int8E returns the option value, the type of which is int8.
//
// Return an error if no the option or the type of the option isn't int8.
func (g *OptGroup) Int8E(name string) (int8, error) {
	v, err := g.getValue(name, int8Type)
	if err != nil {
		return 0, err
	}
	return v.(int8), nil
}

// Int8D is the same as Int8E, but returns the default if there is an error.
func (g *OptGroup) Int8D(name string, _default int8) int8 {
	if value, err := g.Int8E(name); err == nil {
		return value
	}
	return _default
}

// Int8 is the same as Int8E, but panic if there is an error.
func (g *OptGroup) Int8(name string) int8 {
	value, err := g.Int8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int16E returns the option value, the type of which is int16.
//
// Return an error if no the option or the type of the option isn't int16.
func (g *OptGroup) Int16E(name string) (int16, error) {
	v, err := g.getValue(name, int16Type)
	if err != nil {
		return 0, err
	}
	return v.(int16), nil
}

// Int16D is the same as Int16E, but returns the default if there is an error.
func (g *OptGroup) Int16D(name string, _default int16) int16 {
	if value, err := g.Int16E(name); err == nil {
		return value
	}
	return _default
}

// Int16 is the same as Int16E, but panic if there is an error.
func (g *OptGroup) Int16(name string) int16 {
	value, err := g.Int16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int32E returns the option value, the type of which is int32.
//
// Return an error if no the option or the type of the option isn't int32.
func (g *OptGroup) Int32E(name string) (int32, error) {
	v, err := g.getValue(name, int32Type)
	if err != nil {
		return 0, err
	}
	return v.(int32), nil
}

// Int32D is the same as Int32E, but returns the default if there is an error.
func (g *OptGroup) Int32D(name string, _default int32) int32 {
	if value, err := g.Int32E(name); err == nil {
		return value
	}
	return _default
}

// Int32 is the same as Int32E, but panic if there is an error.
func (g *OptGroup) Int32(name string) int32 {
	value, err := g.Int32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int64E returns the option value, the type of which is int64.
//
// Return an error if no the option or the type of the option isn't int64.
func (g *OptGroup) Int64E(name string) (int64, error) {
	v, err := g.getValue(name, int64Type)
	if err != nil {
		return 0, err
	}
	return v.(int64), nil
}

// Int64D is the same as Int64E, but returns the default if there is an error.
func (g *OptGroup) Int64D(name string, _default int64) int64 {
	if value, err := g.Int64E(name); err == nil {
		return value
	}
	return _default
}

// Int64 is the same as Int64E, but panic if there is an error.
func (g *OptGroup) Int64(name string) int64 {
	value, err := g.Int64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// UintE returns the option value, the type of which is uint.
//
// Return an error if no the option or the type of the option isn't uint.
func (g *OptGroup) UintE(name string) (uint, error) {
	v, err := g.getValue(name, uintType)
	if err != nil {
		return 0, err
	}
	return v.(uint), nil
}

// UintD is the same as UintE, but returns the default if there is an error.
func (g *OptGroup) UintD(name string, _default uint) uint {
	if value, err := g.UintE(name); err == nil {
		return value
	}
	return _default
}

// Uint is the same as UintE, but panic if there is an error.
func (g *OptGroup) Uint(name string) uint {
	value, err := g.UintE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint8E returns the option value, the type of which is uint8.
//
// Return an error if no the option or the type of the option isn't uint8.
func (g *OptGroup) Uint8E(name string) (uint8, error) {
	v, err := g.getValue(name, uint8Type)
	if err != nil {
		return 0, err
	}
	return v.(uint8), nil
}

// Uint8D is the same as Uint8E, but returns the default if there is an error.
func (g *OptGroup) Uint8D(name string, _default uint8) uint8 {
	if value, err := g.Uint8E(name); err == nil {
		return value
	}
	return _default
}

// Uint8 is the same as Uint8E, but panic if there is an error.
func (g *OptGroup) Uint8(name string) uint8 {
	value, err := g.Uint8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint16E returns the option value, the type of which is uint16.
//
// Return an error if no the option or the type of the option isn't uint16.
func (g *OptGroup) Uint16E(name string) (uint16, error) {
	v, err := g.getValue(name, uint16Type)
	if err != nil {
		return 0, err
	}
	return v.(uint16), nil
}

// Uint16D is the same as Uint16E, but returns the default if there is an error.
func (g *OptGroup) Uint16D(name string, _default uint16) uint16 {
	if value, err := g.Uint16E(name); err == nil {
		return value
	}
	return _default
}

// Uint16 is the same as Uint16E, but panic if there is an error.
func (g *OptGroup) Uint16(name string) uint16 {
	value, err := g.Uint16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint32E returns the option value, the type of which is uint32.
//
// Return an error if no the option or the type of the option isn't uint32.
func (g *OptGroup) Uint32E(name string) (uint32, error) {
	v, err := g.getValue(name, uint32Type)
	if err != nil {
		return 0, err
	}
	return v.(uint32), nil
}

// Uint32D is the same as Uint32E, but returns the default if there is an error.
func (g *OptGroup) Uint32D(name string, _default uint32) uint32 {
	if value, err := g.Uint32E(name); err == nil {
		return value
	}
	return _default
}

// Uint32 is the same as Uint32E, but panic if there is an error.
func (g *OptGroup) Uint32(name string) uint32 {
	value, err := g.Uint32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint64E returns the option value, the type of which is uint64.
//
// Return an error if no the option or the type of the option isn't uint64.
func (g *OptGroup) Uint64E(name string) (uint64, error) {
	v, err := g.getValue(name, uint64Type)
	if err != nil {
		return 0, err
	}
	return v.(uint64), nil
}

// Uint64D is the same as Uint64E, but returns the default if there is an error.
func (g *OptGroup) Uint64D(name string, _default uint64) uint64 {
	if value, err := g.Uint64E(name); err == nil {
		return value
	}
	return _default
}

// Uint64 is the same as Uint64E, but panic if there is an error.
func (g *OptGroup) Uint64(name string) uint64 {
	value, err := g.Uint64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float32E returns the option value, the type of which is float32.
//
// Return an error if no the option or the type of the option isn't float32.
func (g *OptGroup) Float32E(name string) (float32, error) {
	v, err := g.getValue(name, float32Type)
	if err != nil {
		return 0, err
	}
	return v.(float32), nil
}

// Float32D is the same as Float32E, but returns the default value if there is
// an error.
func (g *OptGroup) Float32D(name string, _default float32) float32 {
	if value, err := g.Float32E(name); err == nil {
		return value
	}
	return _default
}

// Float32 is the same as Float32E, but panic if there is an error.
func (g *OptGroup) Float32(name string) float32 {
	value, err := g.Float32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float64E returns the option value, the type of which is float64.
//
// Return an error if no the option or the type of the option isn't float64.
func (g *OptGroup) Float64E(name string) (float64, error) {
	v, err := g.getValue(name, float64Type)
	if err != nil {
		return 0, err
	}
	return v.(float64), nil
}

// Float64D is the same as Float64E, but returns the default value if there is
// an error.
func (g *OptGroup) Float64D(name string, _default float64) float64 {
	if value, err := g.Float64E(name); err == nil {
		return value
	}
	return _default
}

// Float64 is the same as Float64E, but panic if there is an error.
func (g *OptGroup) Float64(name string) float64 {
	value, err := g.Float64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// DurationE returns the option value, the type of which is time.Duration.
//
// Return an error if no the option or the type of the option isn't time.Duration.
func (g *OptGroup) DurationE(name string) (time.Duration, error) {
	v, err := g.getValue(name, durationType)
	if err != nil {
		return 0, err
	}
	return v.(time.Duration), nil
}

// DurationD is the same as DurationE, but returns the default value if there is
// an error.
func (g *OptGroup) DurationD(name string, _default time.Duration) time.Duration {
	if value, err := g.DurationE(name); err == nil {
		return value
	}
	return _default
}

// Duration is the same as DurationE, but panic if there is an error.
func (g *OptGroup) Duration(name string) time.Duration {
	value, err := g.DurationE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// TimeE returns the option value, the type of which is time.Time.
//
// Return an error if no the option or the type of the option isn't time.Time.
func (g *OptGroup) TimeE(name string) (time.Time, error) {
	v, err := g.getValue(name, timeType)
	if err != nil {
		return time.Time{}, err
	}
	return v.(time.Time), nil
}

// TimeD is the same as TimeE, but returns the default value if there is
// an error.
func (g *OptGroup) TimeD(name string, _default time.Time) time.Time {
	if value, err := g.TimeE(name); err == nil {
		return value
	}
	return _default
}

// Time is the same as TimeE, but panic if there is an error.
func (g *OptGroup) Time(name string) time.Time {
	value, err := g.TimeE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// StringsE returns the option value, the type of which is []string.
//
// Return an error if no the option or the type of the option isn't []string.
func (g *OptGroup) StringsE(name string) ([]string, error) {
	v, err := g.getValue(name, stringsType)
	if err != nil {
		return nil, err
	}
	return v.([]string), nil
}

// StringsD is the same as StringsE, but returns the default value if there is
// an error.
func (g *OptGroup) StringsD(name string, _default []string) []string {
	if value, err := g.StringsE(name); err == nil {
		return value
	}
	return _default
}

// Strings is the same as StringsE, but panic if there is an error.
func (g *OptGroup) Strings(name string) []string {
	value, err := g.StringsE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// IntsE returns the option value, the type of which is []int.
//
// Return an error if no the option or the type of the option isn't []int.
func (g *OptGroup) IntsE(name string) ([]int, error) {
	v, err := g.getValue(name, intsType)
	if err != nil {
		return nil, err
	}
	return v.([]int), nil
}

// IntsD is the same as IntsE, but returns the default value if there is
// an error.
func (g *OptGroup) IntsD(name string, _default []int) []int {
	if value, err := g.IntsE(name); err == nil {
		return value
	}
	return _default
}

// Ints is the same as IntsE, but panic if there is an error.
func (g *OptGroup) Ints(name string) []int {
	value, err := g.IntsE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int64sE returns the option value, the type of which is []int64.
//
// Return an error if no the option or the type of the option isn't []int64.
func (g *OptGroup) Int64sE(name string) ([]int64, error) {
	v, err := g.getValue(name, int64sType)
	if err != nil {
		return nil, err
	}
	return v.([]int64), nil
}

// Int64sD is the same as Int64sE, but returns the default value if there is
// an error.
func (g *OptGroup) Int64sD(name string, _default []int64) []int64 {
	if value, err := g.Int64sE(name); err == nil {
		return value
	}
	return _default
}

// Int64s is the same as Int64s, but panic if there is an error.
func (g *OptGroup) Int64s(name string) []int64 {
	value, err := g.Int64sE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// UintsE returns the option value, the type of which is []uint.
//
// Return an error if no the option or the type of the option isn't []uint.
func (g *OptGroup) UintsE(name string) ([]uint, error) {
	v, err := g.getValue(name, uintsType)
	if err != nil {
		return nil, err
	}
	return v.([]uint), nil
}

// UintsD is the same as UintsE, but returns the default value if there is
// an error.
func (g *OptGroup) UintsD(name string, _default []uint) []uint {
	if value, err := g.UintsE(name); err == nil {
		return value
	}
	return _default
}

// Uints is the same as UintsE, but panic if there is an error.
func (g *OptGroup) Uints(name string) []uint {
	value, err := g.UintsE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint64sE returns the option value, the type of which is []uint64.
//
// Return an error if no the option or the type of the option isn't []uint64.
func (g *OptGroup) Uint64sE(name string) ([]uint64, error) {
	v, err := g.getValue(name, uint64sType)
	if err != nil {
		return nil, err
	}
	return v.([]uint64), nil
}

// Uint64sD is the same as Uint64sE, but returns the default value if there is
// an error.
func (g *OptGroup) Uint64sD(name string, _default []uint64) []uint64 {
	if value, err := g.Uint64sE(name); err == nil {
		return value
	}
	return _default
}

// Uint64s is the same as Uint64sE, but panic if there is an error.
func (g *OptGroup) Uint64s(name string) []uint64 {
	value, err := g.Uint64sE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float64sE returns the option value, the type of which is []float64.
//
// Return an error if no the option or the type of the option isn't []float64.
func (g *OptGroup) Float64sE(name string) ([]float64, error) {
	v, err := g.getValue(name, float64sType)
	if err != nil {
		return nil, err
	}
	return v.([]float64), nil
}

// Float64sD is the same as Float64sE, but returns the default value if there is
// an error.
func (g *OptGroup) Float64sD(name string, _default []float64) []float64 {
	if value, err := g.Float64sE(name); err == nil {
		return value
	}
	return _default
}

// Float64s is the same as Float64sE, but panic if there is an error.
func (g *OptGroup) Float64s(name string) []float64 {
	value, err := g.Float64sE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// DurationsE returns the option value, the type of which is []time.Duration.
//
// Return an error if no the option or the type of the option isn't []time.Duration.
func (g *OptGroup) DurationsE(name string) ([]time.Duration, error) {
	v, err := g.getValue(name, durationsType)
	if err != nil {
		return nil, err
	}
	return v.([]time.Duration), nil
}

// DurationsD is the same as DurationsE, but returns the default value if there is
// an error.
func (g *OptGroup) DurationsD(name string, _default []time.Duration) []time.Duration {
	if value, err := g.DurationsE(name); err == nil {
		return value
	}
	return _default
}

// Durations is the same as DurationsE, but panic if there is an error.
func (g *OptGroup) Durations(name string) []time.Duration {
	value, err := g.DurationsE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// TimesE returns the option value, the type of which is []time.Time.
//
// Return an error if no the option or the type of the option isn't []time.Time.
func (g *OptGroup) TimesE(name string) ([]time.Time, error) {
	v, err := g.getValue(name, timesType)
	if err != nil {
		return nil, err
	}
	return v.([]time.Time), nil
}

// TimesD is the same as TimesE, but returns the default value if there is
// an error.
func (g *OptGroup) TimesD(name string, _default []time.Time) []time.Time {
	if value, err := g.TimesE(name); err == nil {
		return value
	}
	return _default
}

// Times is the same as TimesE, but panic if there is an error.
func (g *OptGroup) Times(name string) []time.Time {
	value, err := g.TimesE(name)
	if err != nil {
		panic(err)
	}
	return value
}
