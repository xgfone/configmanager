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

// Package config is an extensible go configuration manager.
//
// The default parsers can parse the CLI arguments and the ini file. You can
// implement and register your parser, and the configuration engine will call
// the parser to parse the configuration.
//
// The inspiration is from [oslo.config](https://github.com/openstack/oslo.config),
// which is a OpenStack library for config.
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

// Config is used to manage the configuration parsers.
type Config struct {
	isRequired bool
	isDebug    bool

	versionName string
	versionNum  string

	defaultGroupName string

	cli     CliParser
	parsers []Parser

	args   []string
	parsed bool
	groups map[string]OptGroup
}

// NewConfig returns a new Config.
//
// The name of the default group is DEFAULT.
func NewConfig(cli CliParser) *Config {
	return &Config{
		defaultGroupName: DefaultGroupName,

		cli:     cli,
		parsers: make([]Parser, 0, 2),
		groups:  make(map[string]OptGroup, 2),
	}
}

// AddVersion adds the CLI version option.
//
// When parsing CLI arguments, if giving the option, print the version and exit.
func (c *Config) AddVersion(name, version string) {
	if name == "" || version == "" {
		panic(fmt.Errorf("name or version must not be empty"))
	}
	c.RegisterCliOpt("", Bool(name, false, "Print the version and exit."))
	c.versionName = name
	c.versionNum = version
}

// Parse parses the option, including CLI, the config file, or others.
//
// if the arguments is nil, it's equal to os.Args[1:].
//
// After parsing a certain option, it will call the validators of the option
// to validate whether the option value is valid.
//
// If parsed, it will panic when calling it.
func (c *Config) Parse(args []string) (err error) {
	c.checkIsParsed(true)
	c.parsed = true

	if args == nil {
		args = os.Args[1:]
	}

	// Ensure that the default group exists.
	c.getGroupByName(c.defaultGroupName)

	var optErr error
	setGroupOption := func(gname, name string, value interface{}) {
		optErr = c.getGroupByName(gname).setOptValue(name, value)
	}

	// Parse the CLI arguments.
	if err = c.cli.Parse(c, setGroupOption, c.setArgs, args); err != nil {
		return fmt.Errorf("The CLI parser failed: %s", err)
	}
	if optErr != nil {
		return fmt.Errorf("The CLI parser failed: %s", optErr)
	}

	// Check the version option.
	if c.versionName != "" && c.BoolD(c.versionName, false) {
		fmt.Println(c.versionNum)
		os.Exit(0)
	}

	// Parse the other options by other parsers.
	for _, parser := range c.parsers {
		if err = parser.Parse(c, setGroupOption); err != nil {
			return fmt.Errorf("The %s parser failed: %s", parser.Name(), err)
		}
		if optErr != nil {
			return fmt.Errorf("The %s parser failed: %s", parser.Name(), optErr)
		}
	}

	// Check whether all the groups have parsed all the required options.
	for _, group := range c.groups {
		if err = group.checkRequiredOption(); err != nil {
			return err
		}
	}

	return
}

func (c *Config) setArgs(args []string) {
	c.args = args
}

// Audit outputs the internal information to find out the troube.
//
// If not parsed, it will panic when calling it.
func (c *Config) Audit() {
	c.checkIsParsed(false)

	fmt.Printf("%s:\n", filepath.Base(os.Args[0]))
	fmt.Printf("    Args: %v\n", c.args)
	fmt.Printf("    DefaultGroup: %s\n", c.defaultGroupName)
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

func (c *Config) debug(format string, args ...interface{}) {
	if c.isDebug {
		fmt.Printf(format+"\n", args...)
	}
}

// SetDebug sets the debug to true.
//
// If setting, when registering the option, it'll output the verbose information.
// You should set it before registering the option.
//
// If parsed, it will panic when calling it.
func (c *Config) SetDebug() {
	c.checkIsParsed(true)
	c.isDebug = true
}

// IsDebug returns whether the config manager is on the debug mode.
func (c *Config) IsDebug() bool {
	return c.isDebug
}

// SetRequired asks that all the registered options have a value.
//
// Notice: the nil value is not considered that there is a value, but the ZERO
// value is that.
//
// If parsed, it will panic when calling it.
func (c *Config) SetRequired() {
	c.checkIsParsed(true)
	c.isRequired = true
}

// SetDefaultGroupName resets the name of the default group.
//
// If you want to modify it, you must do it before registering any options.
//
// If parsed, it will panic when calling it.
func (c *Config) SetDefaultGroupName(name string) {
	c.checkIsParsed(true)
	c.defaultGroupName = name
}

// GetDefaultGroupName returns the name of the default group.
func (c *Config) GetDefaultGroupName() string {
	return c.defaultGroupName
}

// Args returns the rest of the CLI arguments, which are not the options
// starting with the prefix "-", "--" or others, etc.
//
// Notice: you should not modify the returned string slice result.
//
// If not parsed, it will panic when calling it.
func (c *Config) Args() []string {
	c.checkIsParsed(false)
	return c.args
}

func (c *Config) checkIsParsed(p bool) {
	if p && c.parsed {
		panic(ErrParsed)
	}
	if !p && !c.parsed {
		panic(ErrNotParsed)
	}
}

// Parsed returns true if has been parsed, or false.
func (c *Config) Parsed() bool {
	return c.parsed
}

// AddParser adds a named parser.
//
// You can add many parsers, which are sequential, that's, the arguments needed
// by the next parser can be acquired from the results parsed by the previous
// parser.
//
// Notice: The parser having the same name has only been registered once. Or it
// will panic..
//
// If parsed, it will panic when calling it.
func (c *Config) AddParser(parser Parser) *Config {
	c.checkIsParsed(true)

	name := parser.Name()
	for _, p := range c.parsers {
		if p.Name() == name {
			panic(fmt.Errorf("the parser %s has been added", name))
		}
	}

	c.parsers = append(c.parsers, parser)
	return c
}

// RegisterStruct registers the field name of the struct as options into the
// group "group".
//
// If the group name is "", it's regarded as the default group. And the struct
// must be a pointer to a struct variable, or it will panic.
//
// If parsed, it will panic when calling it.
//
// The tag of the field supports "name", "short", "default", "help", which are
// equal to the name, the short name, the default, the help of the option.
// If you want to ignore a certain field, just set the tag "name" to "-",
// such as `name:"-"`. The field also contains the tag "cli", whose value maybe
// "1", "t", "T", "true", "True", "TRUE", and which represents the option is
// also registered into the CLI parser. Moreover, you can use the tag "group"
// to reset the group name, that's, the group of the field with the tag "group"
// is different to the group of the whole struct. If the value of the tag
// "group" is empty, the default group will be used in preference.
//
// Notice: If having no the tag "name", the name of the option is the lower-case
// of the field name.
//
// Notice: The struct supports the nested struct, but not the pointer field.
//
// Notice: The struct doesn't support the validator. You maybe choose others,
// such as github.com/asaskevich/govalidator.
//
// NOTICE: ALL THE TAGS ARE OPTIONAL.
func (c *Config) RegisterStruct(group string, s interface{}) {
	c.checkIsParsed(true)
	c.getGroupByName(group).registerStruct(s)
}

// RegisterCliOpt registers the option into the group.
//
// It registers the option to not only all the common parsers but also the CLI
// parser.
//
// If the group name is "", it's regarded as the default group.
//
// If parsed, it will panic when calling it.
func (c *Config) RegisterCliOpt(group string, opt Opt) {
	c.registerOpt(group, true, opt)
}

// RegisterCliOpts registers the options into the group.
//
// It registers the options to not only all the common parsers but also the CLI
// parser.
//
// If the group name is "", it's regarded as the default group.
//
// If parsed, it will panic when calling it.
func (c *Config) RegisterCliOpts(group string, opts []Opt) {
	for _, opt := range opts {
		c.RegisterCliOpt(group, opt)
	}
}

// RegisterOpt registers the option into the group.
//
// It only registers the option to all the common parsers, not the CLI parser.
//
// If the group name is "", it's regarded as the default group.
//
// If parsed, it will panic when calling it.
func (c *Config) RegisterOpt(group string, opt Opt) {
	c.registerOpt(group, false, opt)
}

// RegisterOpts registers the options into the group.
//
// It only registers the options to all the common parsers, not the CLI parser.
//
// If the group name is "", it's regarded as the default group.
//
// If parsed, it will panic when calling it.
func (c *Config) RegisterOpts(group string, opts []Opt) {
	for _, opt := range opts {
		c.RegisterOpt(group, opt)
	}
}

// registerOpt registers the option into the group.
//
// If the group name is "", it's regarded as the default group.
//
// The first argument, cli, indicates whether the option is as the CLI option,
// too.
//
// If parsed, it will panic when calling it.
func (c *Config) registerOpt(group string, cli bool, opt Opt) {
	c.checkIsParsed(true)
	c.getGroupByName(group).registerOpt(cli, opt)
}

// Groups returns all the groups, the key of which is the group name.
//
// If not parsed, it will panic when calling it.
//
// Notice: you should not modify the returned map result.
func (c *Config) Groups() map[string]OptGroup {
	c.checkIsParsed(false)
	return c.groups
}

func (c *Config) getGroupName(name string) string {
	if name == "" {
		return c.defaultGroupName
	}
	return name
}

func (c *Config) getGroupByName(name string) OptGroup {
	name = c.getGroupName(name)
	g, ok := c.groups[name]
	if !ok {
		g = NewOptGroup(name, c)
		c.groups[name] = g
	}
	return g
}

// Group returns the OptGroup named group.
//
// Return the default group if the group name is "".
//
// The group must exist, or panic.
//
// If not parsed, it will panic when calling it.
func (c *Config) Group(group string) OptGroup {
	c.checkIsParsed(false)
	group = c.getGroupName(group)
	if g, ok := c.groups[group]; ok {
		return g
	}
	panic(fmt.Errorf("have no group %s", group))
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

// StringsE is equal to c.Group("").StringsE(name).
func (c *Config) StringsE(name string) ([]string, error) {
	return c.Group("").StringsE(name)
}

// StringsD is equal to c.Group("").StringsD(name, _default).
func (c *Config) StringsD(name string, _default []string) []string {
	return c.Group("").StringsD(name, _default)
}

// Strings is equal to c.Group("").Strings(name).
func (c *Config) Strings(name string) []string {
	return c.Group("").Strings(name)
}

// IntsE is equal to c.Group("").IntsE(name).
func (c *Config) IntsE(name string) ([]int, error) {
	return c.Group("").IntsE(name)
}

// IntsD is equal to c.Group("").IntsD(name, _default).
func (c *Config) IntsD(name string, _default []int) []int {
	return c.Group("").IntsD(name, _default)
}

// Ints is equal to c.Group("").Ints(name).
func (c *Config) Ints(name string) []int {
	return c.Group("").Ints(name)
}

// Int64sE is equal to c.Group("").Int64sE(name).
func (c *Config) Int64sE(name string) ([]int64, error) {
	return c.Group("").Int64sE(name)
}

// Int64sD is equal to c.Group("").Int64sD(name, _default).
func (c *Config) Int64sD(name string, _default []int64) []int64 {
	return c.Group("").Int64sD(name, _default)
}

// Int64s is equal to c.Group("").Int64s(name).
func (c *Config) Int64s(name string) []int64 {
	return c.Group("").Int64s(name)
}

// UintsE is equal to c.Group("").UintsE(name).
func (c *Config) UintsE(name string) ([]uint, error) {
	return c.Group("").UintsE(name)
}

// UintsD is equal to c.Group("").UintsD(name, _default).
func (c *Config) UintsD(name string, _default []uint) []uint {
	return c.Group("").UintsD(name, _default)
}

// Uints is equal to c.Group("").Uints(name).
func (c *Config) Uints(name string) []uint {
	return c.Group("").Uints(name)
}

// Uint64sE is equal to c.Group("").Uint64sE(name).
func (c *Config) Uint64sE(name string) ([]uint64, error) {
	return c.Group("").Uint64sE(name)
}

// Uint64sD is equal to c.Group("").Uint64sD(name, _default).
func (c *Config) Uint64sD(name string, _default []uint64) []uint64 {
	return c.Group("").Uint64sD(name, _default)
}

// Uint64s is equal to c.Group("").Uint64s(name).
func (c *Config) Uint64s(name string) []uint64 {
	return c.Group("").Uint64s(name)
}

// Float64sE is equal to c.Group("").Float64sE(name).
func (c *Config) Float64sE(name string) ([]float64, error) {
	return c.Group("").Float64sE(name)
}

// Float64sD is equal to c.Group("").Float64sD(name, _default).
func (c *Config) Float64sD(name string, _default []float64) []float64 {
	return c.Group("").Float64sD(name, _default)
}

// Float64s is equal to c.Group("").Float64s(name).
func (c *Config) Float64s(name string) []float64 {
	return c.Group("").Float64s(name)
}
