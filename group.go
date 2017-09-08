package config

import "fmt"

type option struct {
	opt   Opt
	isCli bool
}

// OptGroup is the group of the option.
type OptGroup struct {
	name   string
	opts   map[string]option
	values map[string]interface{}
}

// NewOptGroup returns a new OptGroup.
func NewOptGroup(name string) OptGroup {
	if name == "" {
		panic(fmt.Errorf("the group is empty"))
	}

	return OptGroup{
		name:   name,
		opts:   make(map[string]option, 8),
		values: make(map[string]interface{}, 8),
	}
}

func (g OptGroup) getAllOpts(cli bool) []Opt {
	opts := make([]Opt, 0, len(g.opts))
	for _, opt := range g.opts {
		if cli && !opt.isCli {
			continue
		}
		opts = append(opts, opt.opt)
	}
	return opts
}

func (g OptGroup) setOptValue(name string, value interface{}, notEmpty bool) (err error) {
	if notEmpty {
		if value == nil {
			return fmt.Errorf("the value of %s in the group %s is nil", name, g.name)
		} else if IsZero(value) {
			return fmt.Errorf("the value of %s in the group %s is ZERO", name, g.name)
		}
	}

	opt := g.opts[name].opt

	// The option has a validator.
	if v, ok := opt.(Validator); ok {
		if err = v.Validate(value); err != nil {
			return
		}
	}

	// The option has a validator chain.
	if vc, ok := opt.(ValidatorChainOpt); ok {
		vs := vc.GetValidators()
		if len(vs) > 0 {
			for _, v := range vs {
				if err = v.Validate(value); err != nil {
					return
				}
			}
		}
	}

	g.values[name] = value
	return
}

func (g OptGroup) setOptions(options map[string]string, notEmpty bool) error {
	for name, opt := range g.opts {
		if value, ok := options[name]; ok {
			v, err := opt.opt.Parse(value)
			if err != nil {
				return err
			}
			if err := g.setOptValue(name, v, notEmpty); err != nil {
				return err
			}
		} else if _default := opt.opt.GetDefault(); _default != nil {
			if err := g.setOptValue(name, _default, notEmpty); err != nil {
				return err
			}
		}
	}
	return nil
}

// Check whether the required option has no value or a ZORE value.
func (g OptGroup) checkRequiredOption(notEmpty bool) (err error) {
	for name, opt := range g.opts {
		if _, ok := g.values[name]; !ok {
			if v := opt.opt.GetDefault(); v != nil {
				if err = g.setOptValue(name, v, notEmpty); err != nil {
					return
				}
				continue
			}

			if notEmpty {
				return fmt.Errorf("the option %s in the group %s has no value",
					name, g.name)
			}
		}
	}
	return nil
}

// registerOpt registers the option into the group.
//
// The first argument cli indicates whether the option is as the CLI option, too.
func (g OptGroup) registerOpt(cli bool, opt Opt) {
	if _, ok := g.opts[opt.GetName()]; ok {
		panic(fmt.Errorf("the option %s has been registered into the group %s",
			opt.GetName(), g.name))
	}
	g.opts[opt.GetName()] = option{isCli: cli, opt: opt}
}

// registerOpts registers many options into the group once.
//
// The first argument cli indicates whether the option is as the CLI option, too.
func (g OptGroup) registerOpts(cli bool, opts []Opt) {
	for _, opt := range opts {
		g.registerOpt(cli, opt)
	}
}

// Value returns the value of the option.
//
// Return nil if the option does not exist.
func (g OptGroup) Value(name string) interface{} {
	return g.values[name]
}

func (g OptGroup) getValue(name string, _type optType) (interface{}, error) {
	opt := g.Value(name)
	if opt == nil {
		return nil, fmt.Errorf("the group %s has no option %s", name, g.name)
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
	default:
		return nil, fmt.Errorf("don't support the type %s", _type)
	}
	return nil, fmt.Errorf("the type of the option %s in the group %s is not %s",
		name, g.name, _type)
}

// BoolE returns the option value, the type of which is bool.
//
// Return an error if no the option or the type of the option isn't bool.
func (g OptGroup) BoolE(name string) (bool, error) {
	v, err := g.getValue(name, boolType)
	if err != nil {
		return false, err
	}
	return v.(bool), nil
}

// BoolD is the same as BoolE, but returns the default if there is an error.
func (g OptGroup) BoolD(name string, _default bool) bool {
	if value, err := g.BoolE(name); err == nil {
		return value
	}
	return _default
}

// Bool is the same as BoolE, but panic if there is an error.
func (g OptGroup) Bool(name string) bool {
	value, err := g.BoolE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// StringE returns the option value, the type of which is string.
//
// Return an error if no the option or the type of the option isn't string.
func (g OptGroup) StringE(name string) (string, error) {
	v, err := g.getValue(name, stringType)
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// StringD is the same as StringE, but returns the default if there is an error.
func (g OptGroup) StringD(name, _default string) string {
	if value, err := g.StringE(name); err == nil {
		return value
	}
	return _default
}

// String is the same as StringE, but panic if there is an error.
func (g OptGroup) String(name string) string {
	value, err := g.StringE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// IntE returns the option value, the type of which is int.
//
// Return an error if no the option or the type of the option isn't int.
func (g OptGroup) IntE(name string) (int, error) {
	v, err := g.getValue(name, intType)
	if err != nil {
		return 0, err
	}
	return v.(int), nil
}

// IntD is the same as IntE, but returns the default if there is an error.
func (g OptGroup) IntD(name string, _default int) int {
	if value, err := g.IntE(name); err == nil {
		return value
	}
	return _default
}

// Int is the same as IntE, but panic if there is an error.
func (g OptGroup) Int(name string) int {
	value, err := g.IntE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int8E returns the option value, the type of which is int8.
//
// Return an error if no the option or the type of the option isn't int8.
func (g OptGroup) Int8E(name string) (int8, error) {
	v, err := g.getValue(name, int8Type)
	if err != nil {
		return 0, err
	}
	return v.(int8), nil
}

// Int8D is the same as Int8E, but returns the default if there is an error.
func (g OptGroup) Int8D(name string, _default int8) int8 {
	if value, err := g.Int8E(name); err == nil {
		return value
	}
	return _default
}

// Int8 is the same as Int8E, but panic if there is an error.
func (g OptGroup) Int8(name string) int8 {
	value, err := g.Int8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int16E returns the option value, the type of which is int16.
//
// Return an error if no the option or the type of the option isn't int16.
func (g OptGroup) Int16E(name string) (int16, error) {
	v, err := g.getValue(name, int16Type)
	if err != nil {
		return 0, err
	}
	return v.(int16), nil
}

// Int16D is the same as Int16E, but returns the default if there is an error.
func (g OptGroup) Int16D(name string, _default int16) int16 {
	if value, err := g.Int16E(name); err == nil {
		return value
	}
	return _default
}

// Int16 is the same as Int16E, but panic if there is an error.
func (g OptGroup) Int16(name string) int16 {
	value, err := g.Int16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int32E returns the option value, the type of which is int32.
//
// Return an error if no the option or the type of the option isn't int32.
func (g OptGroup) Int32E(name string) (int32, error) {
	v, err := g.getValue(name, int32Type)
	if err != nil {
		return 0, err
	}
	return v.(int32), nil
}

// Int32D is the same as Int32E, but returns the default if there is an error.
func (g OptGroup) Int32D(name string, _default int32) int32 {
	if value, err := g.Int32E(name); err == nil {
		return value
	}
	return _default
}

// Int32 is the same as Int32E, but panic if there is an error.
func (g OptGroup) Int32(name string) int32 {
	value, err := g.Int32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int64E returns the option value, the type of which is int64.
//
// Return an error if no the option or the type of the option isn't int64.
func (g OptGroup) Int64E(name string) (int64, error) {
	v, err := g.getValue(name, int64Type)
	if err != nil {
		return 0, err
	}
	return v.(int64), nil
}

// Int64D is the same as Int64E, but returns the default if there is an error.
func (g OptGroup) Int64D(name string, _default int64) int64 {
	if value, err := g.Int64E(name); err == nil {
		return value
	}
	return _default
}

// Int64 is the same as Int64E, but panic if there is an error.
func (g OptGroup) Int64(name string) int64 {
	value, err := g.Int64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// UintE returns the option value, the type of which is uint.
//
// Return an error if no the option or the type of the option isn't uint.
func (g OptGroup) UintE(name string) (uint, error) {
	v, err := g.getValue(name, uintType)
	if err != nil {
		return 0, err
	}
	return v.(uint), nil
}

// UintD is the same as UintE, but returns the default if there is an error.
func (g OptGroup) UintD(name string, _default uint) uint {
	if value, err := g.UintE(name); err == nil {
		return value
	}
	return _default
}

// Uint is the same as UintE, but panic if there is an error.
func (g OptGroup) Uint(name string) uint {
	value, err := g.UintE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint8E returns the option value, the type of which is uint8.
//
// Return an error if no the option or the type of the option isn't uint8.
func (g OptGroup) Uint8E(name string) (uint8, error) {
	v, err := g.getValue(name, uint8Type)
	if err != nil {
		return 0, err
	}
	return v.(uint8), nil
}

// Uint8D is the same as Uint8E, but returns the default if there is an error.
func (g OptGroup) Uint8D(name string, _default uint8) uint8 {
	if value, err := g.Uint8E(name); err == nil {
		return value
	}
	return _default
}

// Uint8 is the same as Uint8E, but panic if there is an error.
func (g OptGroup) Uint8(name string) uint8 {
	value, err := g.Uint8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint16E returns the option value, the type of which is uint16.
//
// Return an error if no the option or the type of the option isn't uint16.
func (g OptGroup) Uint16E(name string) (uint16, error) {
	v, err := g.getValue(name, uint16Type)
	if err != nil {
		return 0, err
	}
	return v.(uint16), nil
}

// Uint16D is the same as Uint16E, but returns the default if there is an error.
func (g OptGroup) Uint16D(name string, _default uint16) uint16 {
	if value, err := g.Uint16E(name); err == nil {
		return value
	}
	return _default
}

// Uint16 is the same as Uint16E, but panic if there is an error.
func (g OptGroup) Uint16(name string) uint16 {
	value, err := g.Uint16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint32E returns the option value, the type of which is uint32.
//
// Return an error if no the option or the type of the option isn't uint32.
func (g OptGroup) Uint32E(name string) (uint32, error) {
	v, err := g.getValue(name, uint32Type)
	if err != nil {
		return 0, err
	}
	return v.(uint32), nil
}

// Uint32D is the same as Uint32E, but returns the default if there is an error.
func (g OptGroup) Uint32D(name string, _default uint32) uint32 {
	if value, err := g.Uint32E(name); err == nil {
		return value
	}
	return _default
}

// Uint32 is the same as Uint32E, but panic if there is an error.
func (g OptGroup) Uint32(name string) uint32 {
	value, err := g.Uint32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint64E returns the option value, the type of which is uint64.
//
// Return an error if no the option or the type of the option isn't uint64.
func (g OptGroup) Uint64E(name string) (uint64, error) {
	v, err := g.getValue(name, uint64Type)
	if err != nil {
		return 0, err
	}
	return v.(uint64), nil
}

// Uint64D is the same as Uint64E, but returns the default if there is an error.
func (g OptGroup) Uint64D(name string, _default uint64) uint64 {
	if value, err := g.Uint64E(name); err == nil {
		return value
	}
	return _default
}

// Uint64 is the same as Uint64E, but panic if there is an error.
func (g OptGroup) Uint64(name string) uint64 {
	value, err := g.Uint64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float32E returns the option value, the type of which is float32.
//
// Return an error if no the option or the type of the option isn't float32.
func (g OptGroup) Float32E(name string) (float32, error) {
	v, err := g.getValue(name, float32Type)
	if err != nil {
		return 0, err
	}
	return v.(float32), nil
}

// Float32D is the same as Float32E, but returns the default if there is an error.
func (g OptGroup) Float32D(name string, _default float32) float32 {
	if value, err := g.Float32E(name); err == nil {
		return value
	}
	return _default
}

// Float32 is the same as Float32E, but panic if there is an error.
func (g OptGroup) Float32(name string) float32 {
	value, err := g.Float32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float64E returns the option value, the type of which is float64.
//
// Return an error if no the option or the type of the option isn't float64.
func (g OptGroup) Float64E(name string) (float64, error) {
	v, err := g.getValue(name, float64Type)
	if err != nil {
		return 0, err
	}
	return v.(float64), nil
}

// Float64D is the same as Float64E, but returns the default if there is an error.
func (g OptGroup) Float64D(name string, _default float64) float64 {
	if value, err := g.Float64E(name); err == nil {
		return value
	}
	return _default
}

// Float64 is the same as Float64E, but panic if there is an error.
func (g OptGroup) Float64(name string) float64 {
	value, err := g.Float64E(name)
	if err != nil {
		panic(err)
	}
	return value
}
