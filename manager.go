package configmanager

import (
	"fmt"
)

var (
	// ErrParsed is an error that the config has been parsed.
	ErrParsed = fmt.Errorf("the config manager has been parsed")

	// ErrNotParsed is an error that the config has not been parsed.
	ErrNotParsed = fmt.Errorf("the config manager has not been parsed")
)

// Opt stands for an opt value.
type Opt interface {
	GetName() string
	GetShort() string
	GetHelp() string
	GetDefault() interface{}
	IsRequired() bool
	Parse(string) (interface{}, error)
}

// ConfigManager is used to manage the configuration parsers.
type ConfigManager struct {
	Args []string

	parsed  bool
	cli     CliParser
	parsers []Parser
	opts    []Opt
	cliopts []Opt
	config  map[string]interface{}
}

// NewConfigManager returns a new ConfigMangaer.
func NewConfigManager(cli CliParser) *ConfigManager {
	return &ConfigManager{
		cli:     cli,
		parsers: make([]Parser, 0, 2),
		opts:    make([]Opt, 0),
		cliopts: make([]Opt, 0),
		config:  make(map[string]interface{}),
	}
}

// Parse parses the option, including CLI, the config file, or others.
func (c *ConfigManager) Parse(arguments []string) (err error) {
	if c.parsed {
		return ErrParsed
	}
	c.parsed = true

	// Register the CLI option into the CLI parser.
	c.cli.Register(c.cliopts)

	// Register the option into the other parsers.
	for _, p := range c.parsers {
		p.Register(c.opts)
	}

	// Parse the CLI arguments.
	if err = c.parseCli(arguments); err != nil {
		return
	}

	// Parse the other options by other parsers.
	for _, parser := range c.parsers {
		args, err := c.getValuesByKeys(parser.Name(), parser.GetKeys())
		if err != nil {
			return err
		}

		opts, err := parser.Parse(args)
		if err != nil {
			return err
		}

		for _, opt := range c.opts {
			if value, ok := opts[opt.GetName()]; ok {
				v, err := opt.Parse(value)
				if err != nil {
					return err
				}
				c.config[opt.GetName()] = v
			} else if _default := opt.GetDefault(); _default != nil {
				c.config[opt.GetName()] = _default
			}
		}
	}

	// Check whether some required options neither have the value nor the default value.
	for _, opt := range c.cliopts {
		if _, ok := c.config[opt.GetName()]; !ok && opt.IsRequired() {
			return fmt.Errorf("the option '%s' is required, but has no value", opt.GetName())
		}
	}
	for _, opt := range c.opts {
		if _, ok := c.config[opt.GetName()]; !ok && opt.IsRequired() {
			return fmt.Errorf("the option '%s' is required, but has no value", opt.GetName())
		}
	}

	return
}

func (c *ConfigManager) getValuesByKeys(name string, keys map[string]bool) (args map[string]string, err error) {
	if len(keys) == 0 {
		return
	}

	args = make(map[string]string, len(keys))
	for key, required := range keys {
		v, ok := c.config[key]
		if !ok {
			if !required {
				continue
			}
			err = fmt.Errorf("the option '%s' is missing, which is reqired by the parser '%s'", key, name)
			return
		}
		s, ok := v.(string)
		if !ok {
			err = fmt.Errorf("the type of the option '%s' is not string", key)
			return
		}
		args[key] = s
	}

	return
}

func (c *ConfigManager) parseCli(arguments []string) (err error) {
	opts, args, err := c.cli.Parse(arguments)
	if err != nil {
		return
	}

	// Parse the values of all the options
	for _, opt := range c.cliopts {
		if value, ok := opts[opt.GetName()]; ok {
			v, err := opt.Parse(value)
			if err != nil {
				return err
			}
			c.config[opt.GetName()] = v
		} else if _default := opt.GetDefault(); _default != nil {
			c.config[opt.GetName()] = _default
		}
	}

	c.Args = args
	return
}

// Parsed returns true if has been parsed, or false.
func (c *ConfigManager) Parsed() bool {
	return c.parsed
}

// AddParser adds a named parser.
//
// It will panic if the parser has been added.
func (c *ConfigManager) AddParser(parser Parser) *ConfigManager {
	if c.parsed {
		panic(ErrParsed)
	}

	name := parser.Name()
	for _, p := range c.parsers {
		if p.Name() == name {
			panic(fmt.Errorf("the parser '%s' has been added", name))
		}
	}

	c.parsers = append(c.parsers, parser)
	return c
}

// RegisterCliOpt registers a CLI option, the type of which is string, also
// registers it by RegisterOpt.
//
// It will panic if the option has been registered or is nil.
func (c *ConfigManager) RegisterCliOpt(opt Opt) {
	if c.parsed {
		panic(ErrParsed)
	}
	c.RegisterOpt(opt)

	name := opt.GetName()
	for _, _opt := range c.cliopts {
		if _opt.GetName() == name {
			panic(fmt.Errorf("the option '%s' has been registered", name))
		}
	}

	c.cliopts = append(c.cliopts, opt)
}

// RegisterCliOpts registers lots of options once.
func (c *ConfigManager) RegisterCliOpts(opts []Opt) {
	for _, opt := range opts {
		c.RegisterCliOpt(opt)
	}
}

// RegisterOpt registers a option, the type of which is string.
//
// It will panic if the option has been registered or is nil.
func (c *ConfigManager) RegisterOpt(opt Opt) {
	if c.parsed {
		panic(ErrParsed)
	}

	name := opt.GetName()
	for _, _opt := range c.opts {
		if _opt.GetName() == name {
			panic(fmt.Errorf("the option '%s' has been registered", name))
		}
	}

	c.opts = append(c.opts, opt)
}

// RegisterOpts registers lots of options once.
func (c *ConfigManager) RegisterOpts(opts []Opt) {
	for _, opt := range opts {
		c.RegisterOpt(opt)
	}
}

// Value returns the option value named name.
//
// If no the option, return nil.
func (c *ConfigManager) Value(name string) interface{} {
	if !c.parsed {
		panic(ErrNotParsed)
	}
	return c.config[name]
}

func (c *ConfigManager) getValue(name string, _type optType) (interface{}, error) {
	if !c.parsed {
		return nil, ErrNotParsed
	}

	opt := c.Value(name)
	if opt == nil {
		return nil, fmt.Errorf("no option '%s'", name)
	}

	switch _type {
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
	return nil, fmt.Errorf("the type of the option '%s' is not %s", name, _type)
}

// StringE returns the option value, the type of which is string.
//
// Return an error if no the option or the type of the option isn't string.
func (c *ConfigManager) StringE(name string) (string, error) {
	v, err := c.getValue(name, stringType)
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

// StringD is the same as StringE, but returns the default if there is an error.
func (c *ConfigManager) StringD(name, _default string) string {
	if value, err := c.StringE(name); err == nil {
		return value
	}
	return _default
}

// String is the same as StringE, but panic if there is an error.
func (c *ConfigManager) String(name string) string {
	value, err := c.StringE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// IntE returns the option value, the type of which is int.
//
// Return an error if no the option or the type of the option isn't int.
func (c *ConfigManager) IntE(name string) (int, error) {
	v, err := c.getValue(name, intType)
	if err != nil {
		return 0, err
	}
	return v.(int), nil
}

// IntD is the same as IntE, but returns the default if there is an error.
func (c *ConfigManager) IntD(name string, _default int) int {
	if value, err := c.IntE(name); err == nil {
		return value
	}
	return _default
}

// Int is the same as IntE, but panic if there is an error.
func (c *ConfigManager) Int(name string) int {
	value, err := c.IntE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int8E returns the option value, the type of which is int8.
//
// Return an error if no the option or the type of the option isn't int8.
func (c *ConfigManager) Int8E(name string) (int8, error) {
	v, err := c.getValue(name, int8Type)
	if err != nil {
		return 0, err
	}
	return v.(int8), nil
}

// Int8D is the same as Int8E, but returns the default if there is an error.
func (c *ConfigManager) Int8D(name string, _default int8) int8 {
	if value, err := c.Int8E(name); err == nil {
		return value
	}
	return _default
}

// Int8 is the same as Int8E, but panic if there is an error.
func (c *ConfigManager) Int8(name string) int8 {
	value, err := c.Int8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int16E returns the option value, the type of which is int16.
//
// Return an error if no the option or the type of the option isn't int16.
func (c *ConfigManager) Int16E(name string) (int16, error) {
	v, err := c.getValue(name, int16Type)
	if err != nil {
		return 0, err
	}
	return v.(int16), nil
}

// Int16D is the same as Int16E, but returns the default if there is an error.
func (c *ConfigManager) Int16D(name string, _default int16) int16 {
	if value, err := c.Int16E(name); err == nil {
		return value
	}
	return _default
}

// Int16 is the same as Int16E, but panic if there is an error.
func (c *ConfigManager) Int16(name string) int16 {
	value, err := c.Int16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int32E returns the option value, the type of which is int32.
//
// Return an error if no the option or the type of the option isn't int32.
func (c *ConfigManager) Int32E(name string) (int32, error) {
	v, err := c.getValue(name, int32Type)
	if err != nil {
		return 0, err
	}
	return v.(int32), nil
}

// Int32D is the same as Int32E, but returns the default if there is an error.
func (c *ConfigManager) Int32D(name string, _default int32) int32 {
	if value, err := c.Int32E(name); err == nil {
		return value
	}
	return _default
}

// Int32 is the same as Int32E, but panic if there is an error.
func (c *ConfigManager) Int32(name string) int32 {
	value, err := c.Int32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Int64E returns the option value, the type of which is int64.
//
// Return an error if no the option or the type of the option isn't int64.
func (c *ConfigManager) Int64E(name string) (int64, error) {
	v, err := c.getValue(name, int64Type)
	if err != nil {
		return 0, err
	}
	return v.(int64), nil
}

// Int64D is the same as Int64E, but returns the default if there is an error.
func (c *ConfigManager) Int64D(name string, _default int64) int64 {
	if value, err := c.Int64E(name); err == nil {
		return value
	}
	return _default
}

// Int64 is the same as Int64E, but panic if there is an error.
func (c *ConfigManager) Int64(name string) int64 {
	value, err := c.Int64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// UintE returns the option value, the type of which is uint.
//
// Return an error if no the option or the type of the option isn't uint.
func (c *ConfigManager) UintE(name string) (uint, error) {
	v, err := c.getValue(name, uintType)
	if err != nil {
		return 0, err
	}
	return v.(uint), nil
}

// UintD is the same as UintE, but returns the default if there is an error.
func (c *ConfigManager) UintD(name string, _default uint) uint {
	if value, err := c.UintE(name); err == nil {
		return value
	}
	return _default
}

// Uint is the same as UintE, but panic if there is an error.
func (c *ConfigManager) Uint(name string) uint {
	value, err := c.UintE(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint8E returns the option value, the type of which is uint8.
//
// Return an error if no the option or the type of the option isn't uint8.
func (c *ConfigManager) Uint8E(name string) (uint8, error) {
	v, err := c.getValue(name, uint8Type)
	if err != nil {
		return 0, err
	}
	return v.(uint8), nil
}

// Uint8D is the same as Uint8E, but returns the default if there is an error.
func (c *ConfigManager) Uint8D(name string, _default uint8) uint8 {
	if value, err := c.Uint8E(name); err == nil {
		return value
	}
	return _default
}

// Uint8 is the same as Uint8E, but panic if there is an error.
func (c *ConfigManager) Uint8(name string) uint8 {
	value, err := c.Uint8E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint16E returns the option value, the type of which is uint16.
//
// Return an error if no the option or the type of the option isn't uint16.
func (c *ConfigManager) Uint16E(name string) (uint16, error) {
	v, err := c.getValue(name, uint16Type)
	if err != nil {
		return 0, err
	}
	return v.(uint16), nil
}

// Uint16D is the same as Uint16E, but returns the default if there is an error.
func (c *ConfigManager) Uint16D(name string, _default uint16) uint16 {
	if value, err := c.Uint16E(name); err == nil {
		return value
	}
	return _default
}

// Uint16 is the same as Uint16E, but panic if there is an error.
func (c *ConfigManager) Uint16(name string) uint16 {
	value, err := c.Uint16E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint32E returns the option value, the type of which is uint32.
//
// Return an error if no the option or the type of the option isn't uint32.
func (c *ConfigManager) Uint32E(name string) (uint32, error) {
	v, err := c.getValue(name, uint32Type)
	if err != nil {
		return 0, err
	}
	return v.(uint32), nil
}

// Uint32D is the same as Uint32E, but returns the default if there is an error.
func (c *ConfigManager) Uint32D(name string, _default uint32) uint32 {
	if value, err := c.Uint32E(name); err == nil {
		return value
	}
	return _default
}

// Uint32 is the same as Uint32E, but panic if there is an error.
func (c *ConfigManager) Uint32(name string) uint32 {
	value, err := c.Uint32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Uint64E returns the option value, the type of which is uint64.
//
// Return an error if no the option or the type of the option isn't uint64.
func (c *ConfigManager) Uint64E(name string) (uint64, error) {
	v, err := c.getValue(name, uint64Type)
	if err != nil {
		return 0, err
	}
	return v.(uint64), nil
}

// Uint64D is the same as Uint64E, but returns the default if there is an error.
func (c *ConfigManager) Uint64D(name string, _default uint64) uint64 {
	if value, err := c.Uint64E(name); err == nil {
		return value
	}
	return _default
}

// Uint64 is the same as Uint64E, but panic if there is an error.
func (c *ConfigManager) Uint64(name string) uint64 {
	value, err := c.Uint64E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float32E returns the option value, the type of which is float32.
//
// Return an error if no the option or the type of the option isn't float32.
func (c *ConfigManager) Float32E(name string) (float32, error) {
	v, err := c.getValue(name, float32Type)
	if err != nil {
		return 0, err
	}
	return v.(float32), nil
}

// Float32D is the same as Float32E, but returns the default if there is an error.
func (c *ConfigManager) Float32D(name string, _default float32) float32 {
	if value, err := c.Float32E(name); err == nil {
		return value
	}
	return _default
}

// Float32 is the same as Float32E, but panic if there is an error.
func (c *ConfigManager) Float32(name string) float32 {
	value, err := c.Float32E(name)
	if err != nil {
		panic(err)
	}
	return value
}

// Float64E returns the option value, the type of which is float64.
//
// Return an error if no the option or the type of the option isn't float64.
func (c *ConfigManager) Float64E(name string) (float64, error) {
	v, err := c.getValue(name, float64Type)
	if err != nil {
		return 0, err
	}
	return v.(float64), nil
}

// Float64D is the same as Float64E, but returns the default if there is an error.
func (c *ConfigManager) Float64D(name string, _default float64) float64 {
	if value, err := c.Float64E(name); err == nil {
		return value
	}
	return _default
}

// Float64 is the same as Float64E, but panic if there is an error.
func (c *ConfigManager) Float64(name string) float64 {
	value, err := c.Float64E(name)
	if err != nil {
		panic(err)
	}
	return value
}
