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
