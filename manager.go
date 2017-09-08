package config

import (
	"fmt"
	"os"
	"path/filepath"
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
	IsBool() bool
	Parse(string) (interface{}, error)
}

// Validator is an interface to validate whether the value v is valid.
type Validator interface {
	// Validate whether the value v is valid.
	//
	// Return nil if the value is ok, or an error instead.
	Validate(v interface{}) error
}

// ValidatorOpt is an Opt interface with the validator.
type ValidatorOpt interface {
	Opt
	Validator
}

// ValidatorChainOpt is an Opt interface with more than one validator.
//
// The validators in the chain will be called in turn. The validation is
// considered as failure only if one validator returns an error, that's,
// only all the validators return nil, it's successful.
type ValidatorChainOpt interface {
	Opt

	// Set the validator chain.
	//
	// Notice: this method should return the option itself.
	SetValidators([]Validator) ValidatorChainOpt

	// Return the validator chain.
	GetValidators() []Validator
}

// Config is used to manage the configuration parsers.
type Config struct {
	// If true, it will check whether some options have neither the value
	// nor the default value; or won't check. If the check result is yes,
	// it will return an error.
	//
	// It will enable by default, and you can set it to false to cancel it.
	// But you must set it before calling the method Parse().
	IsRequired bool

	// Args is the rest of the CLI arguments, which are not the options
	// starting with the prefix "-", "--" or others, etc.
	Args []string

	defaultGroup string

	parsed  bool
	cli     CliParser
	parsers []Parser

	groups map[string]OptGroup
}

// NewConfig returns a new Config.
//
// The name of the default group is DEFAULT.
func NewConfig(cli CliParser) *Config {
	return &Config{
		IsRequired:   true,
		defaultGroup: "DEFAULT",

		cli:     cli,
		parsers: make([]Parser, 0, 2),
		groups:  make(map[string]OptGroup, 2),
	}
}

// Audit outputs the internal information to find out the troube.
func (c *Config) Audit() {
	fmt.Printf("%s:\n", filepath.Base(os.Args[0]))
	fmt.Printf("    Args: %v\n", c.Args)
	fmt.Printf("    DefaultGroup: %s\n", c.defaultGroup)
	fmt.Printf("    Cli Parser: %s\n", c.cli.Name())

	// Parsers
	fmt.Printf("    Parsers:")
	for _, parser := range c.parsers {
		fmt.Printf(" %s", parser.Name())
	}
	fmt.Printf("\n")

	// Group
	fmt.Printf("    Group:\n")
	for gname, group := range c.groups {
		fmt.Printf("        %s:\n", gname)

		fmt.Printf("            Opts:")
		for name, opt := range group.opts {
			if opt.isCli {
				fmt.Printf(" %s[CLI]", name)
			} else {
				fmt.Printf(" %s", name)
			}
		}
		fmt.Printf("\n")

		fmt.Printf("            Values:\n")
		for name, value := range group.values {
			fmt.Printf("                %s=%s\n", name, value)
		}
	}

	fmt.Println()
}

// Parse parses the option, including CLI, the config file, or others.
//
// if the arguments is nil, it's equal to os.Args[1:].
//
// After parsing a certain option, it will call the validators of the option
// to validate whether the option value is valid.
func (c *Config) Parse(arguments []string) (err error) {
	if c.parsed {
		return ErrParsed
	}
	c.parsed = true

	if arguments == nil {
		arguments = os.Args[1:]
	}

	// Ensure that the default group exists.
	if _, ok := c.groups[c.defaultGroup]; !ok {
		c.groups[c.defaultGroup] = NewOptGroup(c.defaultGroup)
	}

	// Parse the CLI arguments.
	groupOpts := make(map[string][]Opt, len(c.groups))
	for name, group := range c.groups {
		groupOpts[name] = group.getAllOpts(true)
	}
	if groups, args, err := c.cli.Parse(c.defaultGroup, groupOpts, arguments); err == nil {
		for gname, opts := range groups {
			if group, ok := c.groups[gname]; ok {
				if err = group.setOptions(opts); err != nil {
					return err
				}
			}
		}
		c.Args = args
	} else {
		return err
	}

	// Parse the other options by other parsers.
	for name, group := range c.groups {
		groupOpts[name] = group.getAllOpts(false)
	}
	for _, parser := range c.parsers {
		args, err := c.getValuesByKeys(parser.Name(), parser.GetKeys())
		if err != nil {
			return err
		}

		groups, err := parser.Parse(c.defaultGroup, groupOpts, args)
		if err != nil {
			return err
		}

		for gname, options := range groups {
			if group, ok := c.groups[gname]; ok {
				if err = group.setOptions(options); err != nil {
					return nil
				}
			}
		}
	}

	// Check whether all the groups have parsed all the required options.
	for _, group := range c.groups {
		if err = group.checkRequiredOption(c.IsRequired); err != nil {
			return err
		}
	}

	return
}

func (c *Config) getValuesByKeys(name string, keys map[string]bool) (
	args map[string]string, err error) {
	if len(keys) == 0 {
		return
	}

	group := c.Group(c.defaultGroup)
	args = make(map[string]string, len(keys))
	for key, required := range keys {
		if v := group.Value(key); v != nil {
			if s, ok := v.(string); ok {
				args[key] = s
				continue
			}
			err = fmt.Errorf("the type of the option '%s' in the default group is not string",
				key)
			return
		}
		if !required {
			continue
		}
		err = fmt.Errorf("the option '%s' is missing, which is reqired by the parser '%s'",
			key, name)
		return
	}

	return
}

// Parsed returns true if has been parsed, or false.
func (c *Config) Parsed() bool {
	return c.parsed
}

// AddParser adds a named parser.
//
// It will panic if the parser has been added.
func (c *Config) AddParser(parser Parser) *Config {
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

// RegisterCliOpt registers the option into the group of `name`.
//
// It registers the option into not only all the common parsers but also
// the CLI parser.
//
// If the group name is "", it's regarded as the default group.
func (c *Config) RegisterCliOpt(group string, opt Opt) {
	c.registerOpt(group, true, opt)
}

// RegisterCliOpts registers the options into the group of `name`.
//
// It registers the options into not only all the common parsers but also
// the CLI parser.
//
// If the group name is "", it's regarded as the default group.
func (c *Config) RegisterCliOpts(group string, opts []Opt) {
	for _, opt := range opts {
		c.RegisterCliOpt(group, opt)
	}
}

// RegisterOpt registers the option into the group of `name`.
//
// It only registers the option into all the common parsers, not the CLI parser.
//
// If the group name is "", it's regarded as the default group.
func (c *Config) RegisterOpt(group string, opt Opt) {
	c.registerOpt(group, false, opt)
}

// RegisterOpts registers the options into the group of `name`.
//
// It only registers the options into all the common parsers, not the CLI parser.
//
// If the group name is "", it's regarded as the default group.
func (c *Config) RegisterOpts(group string, opts []Opt) {
	for _, opt := range opts {
		c.RegisterOpt(group, opt)
	}
}

// registerOpt registers the option into the group of `name`.
//
// If the group name is "", it's regarded as the default group.
//
// The first argument cli indicates whether the option is as the CLI option, too.
func (c *Config) registerOpt(group string, cli bool, opt Opt) {
	if c.parsed {
		panic(ErrParsed)
	}

	group = c.getGroupName(group)
	g, ok := c.groups[group]
	if !ok {
		g = NewOptGroup(group)
		c.groups[group] = g
	}

	g.registerOpt(cli, opt)
}

func (c *Config) getGroupName(name string) string {
	if name == "" {
		return c.defaultGroup
	}
	return name
}

// Group returns the OptGroup named group.
//
// Return the default group if the group name is "".
//
// The group must exist, or panic.
func (c *Config) Group(group string) OptGroup {
	if !c.parsed {
		panic(ErrNotParsed)
	}

	group = c.getGroupName(group)
	if g, ok := c.groups[group]; ok {
		return g
	}
	panic(fmt.Errorf("have no the group '%s'", group))
}

// SetDefaultGroup resets the name of the default group.
//
// If you want to modify it, you must do it before registering any options.
func (c *Config) SetDefaultGroup(name string) {
	if c.parsed {
		panic(ErrParsed)
	}
	c.defaultGroup = name
}

// Value is equal to c.Group("").Value(name).
func (c *Config) Value(name string) interface{} {
	return c.Group("").Value(name)
}

// BoolE is equal to c.Group("").BoolE(name).
func (c *Config) BoolE(name string) (bool, error) {
	return c.Group("").BoolE(name)
}

// BoolD is equal to c.Group("").BoolD(name, _default).
func (c *Config) BoolD(name string, _default bool) bool {
	return c.Group("").BoolD(name, _default)
}

// Bool is equal to c.Group("").Bool(name).
func (c *Config) Bool(name string) bool {
	return c.Group("").Bool(name)
}

// StringE is equal to c.Group("").StringE(name).
func (c *Config) StringE(name string) (string, error) {
	return c.Group("").StringE(name)
}

// StringD is equal to c.Group("").StringD(name, _default).
func (c *Config) StringD(name, _default string) string {
	return c.Group("").StringD(name, _default)
}

// String is equal to c.Group("").String(name).
func (c *Config) String(name string) string {
	return c.Group("").String(name)
}

// IntE is equal to c.Group("").IntE(name).
func (c *Config) IntE(name string) (int, error) {
	return c.Group("").IntE(name)
}

// IntD is equal to c.Group("").IntD(name, _default).
func (c *Config) IntD(name string, _default int) int {
	return c.Group("").IntD(name, _default)
}

// Int is equal to c.Group("").Int(name).
func (c *Config) Int(name string) int {
	return c.Group("").Int(name)
}

// Int8E is equal to c.Group("").Int8E(name).
func (c *Config) Int8E(name string) (int8, error) {
	return c.Group("").Int8E(name)
}

// Int8D is equal to c.Group("").Int8D(name, _default).
func (c *Config) Int8D(name string, _default int8) int8 {
	return c.Group("").Int8D(name, _default)
}

// Int8 is equal to c.Group("").Int8(name).
func (c *Config) Int8(name string) int8 {
	return c.Group("").Int8(name)
}

// Int16E is equal to c.Group("").Int16E(name).
func (c *Config) Int16E(name string) (int16, error) {
	return c.Group("").Int16E(name)
}

// Int16D is equal to c.Group("").Int16D(name, _default).
func (c *Config) Int16D(name string, _default int16) int16 {
	return c.Group("").Int16D(name, _default)
}

// Int16 is equal to c.Group("").Int16(name).
func (c *Config) Int16(name string) int16 {
	return c.Group("").Int16(name)
}

// Int32E is equal to c.Group("").Int32E(name).
func (c *Config) Int32E(name string) (int32, error) {
	return c.Group("").Int32E(name)
}

// Int32D is equal to c.Group("").Int32D(name, _default).
func (c *Config) Int32D(name string, _default int32) int32 {
	return c.Group("").Int32D(name, _default)
}

// Int32 is equal to c.Group("").Int32(name).
func (c *Config) Int32(name string) int32 {
	return c.Group("").Int32(name)
}

// Int64E is equal to c.Group("").Int64E(name).
func (c *Config) Int64E(name string) (int64, error) {
	return c.Group("").Int64E(name)
}

// Int64D is equal to c.Group("").Int64D(name, _default).
func (c *Config) Int64D(name string, _default int64) int64 {
	return c.Group("").Int64D(name, _default)
}

// Int64 is equal to c.Group("").Int64(name).
func (c *Config) Int64(name string) int64 {
	return c.Group("").Int64(name)
}

// UintE is equal to c.Group("").UintE(name).
func (c *Config) UintE(name string) (uint, error) {
	return c.Group("").UintE(name)
}

// UintD is equal to c.Group("").UintD(name, _default).
func (c *Config) UintD(name string, _default uint) uint {
	return c.Group("").UintD(name, _default)
}

// Uint is equal to c.Group("").Uint(name).
func (c *Config) Uint(name string) uint {
	return c.Group("").Uint(name)
}

// Uint8E is equal to c.Group("").Uint8E(name).
func (c *Config) Uint8E(name string) (uint8, error) {
	return c.Group("").Uint8E(name)
}

// Uint8D is equal to c.Group("").Uint8D(name, _default).
func (c *Config) Uint8D(name string, _default uint8) uint8 {
	return c.Group("").Uint8D(name, _default)
}

// Uint8 is equal to c.Group("").Uint8(name).
func (c *Config) Uint8(name string) uint8 {
	return c.Group("").Uint8(name)
}

// Uint16E is equal to c.Group("").Uint16E(name).
func (c *Config) Uint16E(name string) (uint16, error) {
	return c.Group("").Uint16E(name)
}

// Uint16D is equal to c.Group("").Uint16D(name, _default).
func (c *Config) Uint16D(name string, _default uint16) uint16 {
	return c.Group("").Uint16D(name, _default)
}

// Uint16 is equal to c.Group("").Uint16(name).
func (c *Config) Uint16(name string) uint16 {
	return c.Group("").Uint16(name)
}

// Uint32E is equal to c.Group("").Uint32E(name).
func (c *Config) Uint32E(name string) (uint32, error) {
	return c.Group("").Uint32E(name)
}

// Uint32D is equal to c.Group("").Uint32D(name, _default).
func (c *Config) Uint32D(name string, _default uint32) uint32 {
	return c.Group("").Uint32D(name, _default)
}

// Uint32 is equal to c.Group("").Uint32(name).
func (c *Config) Uint32(name string) uint32 {
	return c.Group("").Uint32(name)
}

// Uint64E is equal to c.Group("").Uint64E(name).
func (c *Config) Uint64E(name string) (uint64, error) {
	return c.Group("").Uint64E(name)
}

// Uint64D is equal to c.Group("").Uint64D(name, _default).
func (c *Config) Uint64D(name string, _default uint64) uint64 {
	return c.Group("").Uint64D(name, _default)
}

// Uint64 is equal to c.Group("").Uint64(name).
func (c *Config) Uint64(name string) uint64 {
	return c.Group("").Uint64(name)
}

// Float32E is equal to c.Group("").Float32E(name).
func (c *Config) Float32E(name string) (float32, error) {
	return c.Group("").Float32E(name)
}

// Float32D is equal to c.Group("").Float32D(name, _default).
func (c *Config) Float32D(name string, _default float32) float32 {
	return c.Group("").Float32D(name, _default)
}

// Float32 is equal to c.Group("").Float32(name).
func (c *Config) Float32(name string) float32 {
	return c.Group("").Float32(name)
}

// Float64E is equal to c.Group("").Float64E(name).
func (c *Config) Float64E(name string) (float64, error) {
	return c.Group("").Float64E(name)
}

// Float64D is equal to c.Group("").Float64D(name, _default).
func (c *Config) Float64D(name string, _default float64) float64 {
	return c.Group("").Float64D(name, _default)
}

// Float64 is equal to c.Group("").Float64(name).
func (c *Config) Float64(name string) float64 {
	return c.Group("").Float64(name)
}
