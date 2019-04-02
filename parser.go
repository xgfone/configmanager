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
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

// Parser is an parser interface.
type Parser interface {
	// Name returns the name of the parser to identify it.
	Name() string

	// Priority reports the priority of the current parser, which should be
	// a natural number.
	//
	// The smaller the number, the higher the priority. And the higher priority
	// parser can cover the option value set by the lower priority parser.
	//
	// For the cli parser, it maybe return 0 to indicate the highest priority.
	Priority() int

	// Pre is called before parsing the configuration, so it may be used to
	// initialize the parser, such as registering the itself options.
	Pre(*Config) error

	// Parse the value of the registered options.
	//
	// The parser can get any information from the argument, config.
	//
	// When the parser parsed out the option value, it should call
	// config.SetOptValue(), which will set the group option.
	// For the default group, the group name may be "" instead,
	//
	// For the CLI parser, it should get the parsed argument by config.CliArgs(),
	// which is a string slice, not nil, but it maybe have no elements.
	// The CLI parser should not use os.Args[1:] as the parsed CLI arguments.
	// If there are the rest CLI arguments, that's those that does not start
	// with the prefix "-", "--" or others, etc, the CLI parser should call
	// config.SetArgs() to set them.
	//
	// If there is any error, the parser should stop to parse and return it.
	//
	// If a certain option has no value, the parser should not return a default
	// one instead. Also, the parser has no need to convert the value to the
	// corresponding specific type, just string is ok. Because the configuration
	// manager will convert the value to the specific type automatically.
	// Certainly, it's not harmless for the parser to convert the value to
	// the specific type.
	Parse(*Config) error

	// Pre is called before parsing the configuration, so it may be used to
	// clean the parser.
	Post(*Config) error
}

type flagParser struct {
	utoh bool
	fset *flag.FlagSet
}

// NewDefaultFlagCliParser returns a new CLI parser based on flag,
// which is equal to NewFlagCliParser("", 0, underlineToHyphen, flag.CommandLine).
func NewDefaultFlagCliParser(underlineToHyphen ...bool) Parser {
	var u2h bool
	if len(underlineToHyphen) > 0 {
		u2h = underlineToHyphen[0]
	}
	return NewFlagCliParser(flag.CommandLine, u2h)
}

// NewFlagCliParser returns a new CLI parser based on flag.FlagSet.
//
// If flagSet is nil, it will create a default flag.FlagSet, which is equal to
//
//    flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
//
// If underlineToHyphen is true, it will convert the underline to the hyphen.
//
// Notice: when other libraries use the default global flag.FlagSet, that's
// flag.CommandLine, such as github.com/golang/glog, please use flag.CommandLine
// as flag.FlagSet.
func NewFlagCliParser(flagSet *flag.FlagSet, underlineToHyphen bool) Parser {
	if flagSet == nil {
		flagSet = flag.NewFlagSet(filepath.Base(os.Args[0]), flag.ContinueOnError)
	}

	return flagParser{
		fset: flagSet,
		utoh: underlineToHyphen,
	}
}

func (f flagParser) Name() string {
	return "flag"
}

func (f flagParser) Priority() int {
	return 0
}

func (f flagParser) Pre(c *Config) error {
	return nil
}

func (f flagParser) Post(c *Config) error {
	return nil
}

func (f flagParser) Parse(c *Config) (err error) {
	// Convert the option name.
	name2group := make(map[string]string, 8)
	name2opt := make(map[string]string, 8)
	for _, group := range c.Groups() {
		gname := group.FullName()
		for _, opt := range group.CliOpts() {
			name := opt.Name()
			if gname != c.GetDefaultGroupName() {
				name = fmt.Sprintf("%s%s%s", gname, c.GetGroupSeparator(), name)
			}

			if f.utoh {
				name = strings.Replace(name, "_", "-", -1)
			}

			name2group[name] = gname
			name2opt[name] = opt.Name()

			switch opt.Zero().(type) {
			case bool:
				var _default bool
				if v := opt.Default(); v != nil {
					_default = v.(bool)
				}
				f.fset.Bool(name, _default, opt.Help())
			case int, int8, int16, int32, int64:
				var _default int64
				if v := opt.Default(); v != nil {
					_default, _ = ToInt64(v)
				}
				f.fset.Int64(name, _default, opt.Help())
			case uint, uint8, uint16, uint32, uint64:
				var _default uint64
				if v := opt.Default(); v != nil {
					_default, _ = ToUint64(v)
				}
				f.fset.Uint64(name, _default, opt.Help())
			case float32, float64:
				var _default float64
				if v := opt.Default(); v != nil {
					_default, _ = ToFloat64(v)
				}
				f.fset.Float64(name, _default, opt.Help())
			case time.Duration:
				var _default time.Duration
				if v := opt.Default(); v != nil {
					_default = v.(time.Duration)
				}
				f.fset.Duration(name, _default, opt.Help())
			default:
				var _default string
				if v := opt.Default(); v != nil {
					_default = fmt.Sprintf("%v", v)
				}
				f.fset.String(name, _default, opt.Help())
			}
		}
	}

	// Register the version option.
	var _version *bool
	name, version, help := c.GetVersion()
	if name != "" {
		_version = f.fset.Bool(name, false, help)
	}

	// Parse the CLI arguments.
	if err = f.fset.Parse(c.CliArgs()); err != nil {
		return
	}

	if _version != nil && *_version {
		fmt.Println(version)
		os.Exit(0)
	}

	// Acquire the result.
	c.SetArgs(f.fset.Args())
	f.fset.Visit(func(fg *flag.Flag) {
		c.Printf("[%s] Parsing flag '%s'", f.Name(), fg.Name)
		gname := name2group[fg.Name]
		optname := name2opt[fg.Name]
		if gname != "" && optname != "" && fg.Name != name {
			c.SetOptValue(0, gname, optname, fg.Value.String())
		}
	})

	return
}

type iniParser struct {
	sep  string
	opt  string
	prio int
	init func(*Config) error
}

// NewSimpleIniParser returns a INI parser with the priority 100, which registers
// the option, optName, before parsing the option.
func NewSimpleIniParser(optName string) Parser {
	return NewIniParser(100, optName, func(c *Config) error {
		c.RegisterCliOpt("", Str(optName, "", "The path of the INI config file."))
		return nil
	})
}

// NewIniParser returns a new ini parser based on the file.
//
// The first argument is used to customized the priority.
//
// The second argument is the option name which the parser needs. It will be
// registered, and parsed before this parser runs.
//
// The third argument sets the Init function.
//
// The ini parser supports the line comments starting with "#", "//" or ";".
// The key and the value is separated by an equal sign, that's =. The key must
// be in one of ., :, _, -, number and letter. If giving fmtKey, it can convert
// the key in the ini file to the new one.
//
// If the value ends with "\", it will continue the next line. The lines will
// be joined by "\n" together.
//
// Notice: the options that have not been assigned to a certain group will be
// divided into the default group.
func NewIniParser(priority int, optName string, init func(*Config) error) Parser {
	return iniParser{prio: priority, opt: optName, sep: "=", init: init}
}

func (p iniParser) Name() string {
	return "ini"
}

func (p iniParser) Priority() int {
	return p.prio
}

func (p iniParser) Pre(c *Config) error {
	if p.init != nil {
		return p.init(c)
	}
	return nil
}

func (p iniParser) Post(c *Config) error {
	return nil
}

func (p iniParser) Parse(c *Config) error {
	// Read the content of the config file.
	filename := c.StringD(p.opt, "")
	if filename == "" {
		return nil
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the config file.
	gname := c.GetDefaultGroupName()
	lines := strings.Split(string(data), "\n")
	for index, maxIndex := 0, len(lines); index < maxIndex; {
		line := strings.TrimSpace(lines[index])
		index++

		c.Printf("[%s] Parsing %dth line: '%s'", p.Name(), index, line)

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#", ";" or "//".
		if (line[0] == '#') || (line[0] == ';') ||
			(len(line) > 1 && line[0] == '/' && line[1] == '/') {
			continue
		}

		// Start a new group
		if line[0] == '[' && line[len(line)-1] == ']' {
			gname = strings.TrimSpace(line[1 : len(line)-1])
			if gname == "" {
				return fmt.Errorf("the group is empty")
			}
			continue
		}

		n := strings.Index(line, p.sep)
		if n == -1 {
			return fmt.Errorf("the %dth line misses the separator '%s'", index, p.sep)
		}

		key := strings.TrimSpace(line[0:n])
		for _, r := range key {
			if r != '_' && r != '-' && !unicode.IsNumber(r) && !unicode.IsLetter(r) {
				return fmt.Errorf("invalid identifier key '%s'", key)
			}
		}
		value := strings.TrimSpace(line[n+len(p.sep) : len(line)])

		// The continuation line
		if value != "" && value[len(value)-1] == '\\' {
			vs := []string{strings.TrimSpace(strings.TrimRight(value, "\\"))}
			for index < maxIndex {
				value = strings.TrimSpace(lines[index])
				vs = append(vs, strings.TrimSpace(strings.TrimRight(value, "\\")))
				index++
				c.Printf("[%s] Parsing %dth line: '%s'", p.Name(), index, value)
				if value == "" || value[len(value)-1] != '\\' {
					break
				}
			}
			value = strings.TrimSpace(strings.Join(vs, "\n"))
		}

		if err = c.SetOptValue(p.prio, gname, key, value); err != nil {
			return err
		}
	}

	return nil
}

type envVarParser struct {
	prefix string
}

// NewEnvVarParser returns a new environment variable parser.
//
// For the environment variable name, it's the format "PREFIX_GROUP_OPTION".
// If the prefix is empty, it's "GROUP_OPTION". For the default group, it's
// "PREFIX_OPTION". When the prefix is empty and the group is the default,
// it's "OPTION". "GROUP" is the group name, and "OPTION" is the option name.
//
// Notice: the prefix, the group name and the option name will be converted to
// the upper, and the group separator will be converted to "_".
func NewEnvVarParser(prefix string) Parser {
	return envVarParser{prefix: prefix}
}

func (e envVarParser) Name() string {
	return "env"
}

func (e envVarParser) Priority() int {
	return 10
}

func (e envVarParser) Pre(c *Config) error {
	return nil
}

func (e envVarParser) Post(c *Config) error {
	return nil
}

func (e envVarParser) Parse(c *Config) (err error) {
	// Initialize the prefix
	prefix := e.prefix
	if prefix != "" {
		prefix += "_"
	}

	// Convert the option to the variable name
	env2opts := make(map[string][]string, len(c.Groups())*8)
	for _, group := range c.Groups() {
		gname := ""
		if group.Name() != c.GetDefaultGroupName() {
			gname = strings.Replace(group.FullName(), c.GetGroupSeparator(), "_", -1) + "_"
		}
		for _, opt := range group.AllOpts() {
			e := fmt.Sprintf("%s%s%s", prefix, gname, opt.Name())
			env2opts[strings.ToUpper(e)] = []string{group.Name(), opt.Name()}
		}
	}

	// Get the option value from the environment variable.
	envs := os.Environ()
	for _, env := range envs {
		c.Printf("[%s] Parsing Env '%s'", e.Name(), env)
		items := strings.SplitN(env, "=", 2)
		if len(items) == 2 {
			if info, ok := env2opts[items[0]]; ok {
				if err = c.SetOptValue(10, info[0], info[1], items[1]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

type propertyParser struct {
	sep  string
	opt  string
	prio int
	init func(*Config) error
}

// NewSimplePropertyParser returns a INI parser with the priority 100,
// which registers the option, optName, before parsing the option.
func NewSimplePropertyParser(optName string) Parser {
	return NewPropertyParser(100, optName, func(c *Config) error {
		c.RegisterCliOpt("", Str(optName, "", "The path of the property config file."))
		return nil
	})
}

// NewPropertyParser returns a new property parser based on the file.
//
// The first argument is used to customized the priority.
//
// The second argument is the option name which the parser needs. It will be
// registered, and parsed before this parser runs.
//
// The third argument sets the Init function.
//
// The ini parser supports the line comments starting with "#", "//" or ";".
// The key and the value is separated by an equal sign, that's =. The key must
// be in one of ., :, _, -, number and letter. If giving fmtKey, it can convert
// the key in the ini file to the new one.
//
// If the value ends with "\", it will continue the next line. The lines will
// be joined by "\n" together.
//
// Notice: the options that have not been assigned to a certain group will be
// divided into the default group.
func NewPropertyParser(priority int, optName string, init func(*Config) error) Parser {
	return propertyParser{prio: priority, opt: optName, sep: "=", init: init}
}

func (p propertyParser) Name() string {
	return "property"
}

func (p propertyParser) Priority() int {
	return p.prio
}

func (p propertyParser) Pre(c *Config) error {
	if p.init != nil {
		return p.init(c)
	}
	return nil
}

func (p propertyParser) Post(c *Config) error {
	return nil
}

func (p propertyParser) Parse(c *Config) error {
	// Read the content of the config file.
	filename := c.StringD(p.opt, "")
	if filename == "" {
		return nil
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Parse the config file.
	lines := strings.Split(string(data), "\n")
	for index, maxIndex := 0, len(lines); index < maxIndex; {
		line := strings.TrimSpace(lines[index])
		index++

		c.Printf("[%s] Parsing %dth line: '%s'", p.Name(), index, line)

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#", ";" or "//".
		if (line[0] == '#') || (line[0] == ';') ||
			(len(line) > 1 && line[0] == '/' && line[1] == '/') {
			continue
		}

		ss := strings.SplitN(line, p.sep, 2)
		if len(ss) != 2 {
			return fmt.Errorf("the %dth line misses the separator '%s'", index, p.sep)
		}

		key := strings.TrimSpace(ss[0])
		value := strings.TrimSpace(ss[1])
		if value != "" {
			for index < maxIndex && value[len(value)-1] == '\\' {
				value = strings.TrimRight(value, "\\") + strings.TrimSpace(lines[index])
				index++
				c.Printf("[%s] Parsing %dth line: '%s'", p.Name(), index, lines[index])
			}
		}

		ss = strings.Split(key, c.GetGroupSeparator())
		switch _len := len(ss) - 1; _len {
		case 0:
			err = c.SetOptValue(p.prio, "", key, value)
		default:
			err = c.SetOptValue(p.prio, strings.Join(ss[:_len], c.GetGroupSeparator()), ss[_len], value)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
