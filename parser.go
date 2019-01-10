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
	"unicode"
)

// Parser is an parser interface.
//
// If the parser implementation needs to register some options into the config
// manager, it should get the instance of the config manager then register them
// when creating the parser instance, because the config manager does not allow
// anyone to register the option.
//
//    conf := NewConfig(cliParser)
//    parser := NewXxxParser(conf) // Register the options into conf.
//    conf.AddParser(parser)
//    conf.Parse(nil)
type Parser interface {
	// Name returns the name of the parser to identify it.
	Name() string

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
	Parse(config *Config) error
}

type flagParser struct {
	flagSet    *flag.FlagSet
	name       string
	errhandler flag.ErrorHandling

	underlineToHyphen bool
}

// NewDefaultFlagCliParser returns a new CLI parser based on flag.
//
// The parser will use flag.CommandLine to parse the CLI arguments.
//
// If underlineToHyphen is true, it will convert the underline to the hyphen.
func NewDefaultFlagCliParser(underlineToHyphen ...bool) Parser {
	var u2h bool
	if len(underlineToHyphen) > 0 {
		u2h = underlineToHyphen[0]
	}

	return flagParser{
		flagSet: flag.CommandLine,

		underlineToHyphen: u2h,
	}
}

// NewFlagCliParser returns a new CLI parser based on flag.FlagSet.
//
// The arguments is the same as that of flag.NewFlagSet(), but if the name is
// "", it will be filepath.Base(os.Args[0]).
//
// If underlineToHyphen is true, it will convert the underline to the hyphen.
//
// When other libraries use the default global flag.FlagSet, that's
// flag.CommandLine, such as github.com/golang/glog, please use
// NewDefaultFlagCliParser(), not this function.
func NewFlagCliParser(appName string, errhandler flag.ErrorHandling,
	underlineToHyphen ...bool) Parser {

	if appName == "" {
		appName = filepath.Base(os.Args[0])
	}

	var u2h bool
	if len(underlineToHyphen) > 0 {
		u2h = underlineToHyphen[0]
	}

	return flagParser{
		name:       appName,
		errhandler: errhandler,

		underlineToHyphen: u2h,
	}
}

func (f flagParser) Name() string {
	return "flag"
}

func (f flagParser) Parse(c *Config) (err error) {
	// Register the options into flag.FlagSet.
	flagSet := f.flagSet
	if flagSet == nil {
		flagSet = flag.NewFlagSet(f.name, f.errhandler)
	}

	// Convert the option name.
	name2group := make(map[string]string, 8)
	name2opt := make(map[string]string, 8)
	for gname, group := range c.Groups() {
		for name, opt := range group.CliOpts() {
			if gname != c.GetDefaultGroupName() {
				name = fmt.Sprintf("%s_%s", gname, name)
			}

			if f.underlineToHyphen {
				name = strings.Replace(name, "_", "-", -1)
			}

			name2group[name] = gname
			name2opt[name] = opt.Name()

			if opt.IsBool() {
				var _default bool
				if v := opt.Default(); v != nil {
					_default = v.(bool)
				}
				flagSet.Bool(name, _default, opt.Help())
			} else {
				_default := ""
				if opt.Default() != nil {
					_default = fmt.Sprintf("%v", opt.Default())
				}
				flagSet.String(name, _default, opt.Help())
			}
		}
	}

	// Register the version option.
	name, version, help := c.GetVersion()
	_version := flagSet.Bool(name, false, help)

	// Parse the CLI arguments.
	if err = flagSet.Parse(c.CliArgs()); err != nil {
		return
	}

	if *_version {
		fmt.Println(version)
		os.Exit(0)
	}

	// Acquire the result.
	c.SetArgs(flagSet.Args())
	flagSet.Visit(func(fg *flag.Flag) {
		gname := name2group[fg.Name]
		optname := name2opt[fg.Name]
		if gname != "" && optname != "" && fg.Name != name {
			c.SetOptValue(gname, optname, fg.Value.String())
		}
	})

	return
}

type iniParser struct {
	sep     string
	optName string
	fmtKey  func(string) string
}

// NewSimpleIniParser returns a new ini parser based on the file.
//
// The argument is the option name which the parser needs. It should be
// registered, and parsed before this parser runs.
//
// The ini parser supports the line comments starting with "#", "//" or ";".
// The key and the value is separated by an equal sign, that's =. The key must
// be in one of _, -, number and letter. If giving fmtKey, it can convert
// the key in the ini file to the new one.
//
// If the value ends with "\", it will continue the next line. The lines will
// be joined by "\n" together.
//
// Notice: the options that have not been assigned to a certain group will be
// divided into the default group.
func NewSimpleIniParser(optName string, fmtKey ...func(string) string) Parser {
	f := func(key string) string { return key }
	if len(fmtKey) > 0 && fmtKey[0] != nil {
		f = fmtKey[0]
	}
	return iniParser{optName: optName, sep: "=", fmtKey: f}
}

func (p iniParser) Name() string {
	return "ini"
}

// func (p iniParser) Parse(_default string, opts map[string][]Opt,
// 	conf map[string]interface{}) (results map[string]map[string]interface{},
// 	err error) {
func (p iniParser) Parse(c *Config) error {
	// Read the content of the config file.
	filename := c.Group("").StringD(p.optName, "")
	if filename == "" {
		return nil
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Convert the format of the optons.
	options := make(map[string]map[string]struct{}, len(c.Groups()))
	for gname, group := range c.Groups() {
		opts := group.AllOpts()
		g, ok := options[gname]
		if !ok {
			g = make(map[string]struct{}, len(opts))
			options[gname] = g
		}
		for _, opt := range opts {
			g[opt.Name()] = struct{}{}
		}
	}

	// Parse the config file.
	gname := c.GetDefaultGroupName()
	lines := strings.Split(string(data), "\n")
	for index, maxIndex := 0, len(lines); index < maxIndex; {
		line := strings.TrimSpace(lines[index])
		index++

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#" or "//".
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
			return fmt.Errorf("the line misses the separator %s", p.sep)
		}

		key := strings.TrimSpace(line[0:n])
		for _, r := range key {
			if r != '_' && r != '-' && !unicode.IsNumber(r) && !unicode.IsLetter(r) {
				return fmt.Errorf("valid identifier key '%s'", key)
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
				if value == "" || value[len(value)-1] != '\\' {
					break
				}
			}
			value = strings.TrimSpace(strings.Join(vs, "\n"))
		}

		if newkey := p.fmtKey(key); newkey != "" {
			key = newkey
		} else {
			panic(fmt.Errorf("convert the key '%s' to ''", key))
		}

		c.SetOptValue(gname, key, value)
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
// the upper.
func NewEnvVarParser(prefix string) Parser {
	return envVarParser{prefix: prefix}
}

func (e envVarParser) Name() string {
	return "env"
}

// func (e envVarParser) Parse(_default string, opts map[string][]Opt,
// 	conf map[string]interface{}) (results map[string]map[string]interface{},
// 	err error) {
func (e envVarParser) Parse(c *Config) error {
	// Initialize the prefix
	prefix := e.prefix
	if prefix != "" {
		prefix += "_"
	}

	// Convert the option to the variable name
	env2opts := make(map[string][]string, len(c.Groups())*8)
	for gname, group := range c.Groups() {
		_gname := ""
		if gname != c.GetDefaultGroupName() {
			_gname = gname + "_"
		}
		for name := range group.AllOpts() {
			e := fmt.Sprintf("%s%s%s", prefix, _gname, name)
			env2opts[strings.ToUpper(e)] = []string{gname, name}
		}
	}

	// Get the option value from the environment variable.
	envs := os.Environ()
	for _, env := range envs {
		items := strings.SplitN(env, "=", 2)
		if len(items) == 2 {
			if info, ok := env2opts[items[0]]; ok {
				c.SetOptValue(info[0], info[1], items[1])
			}
		}
	}

	return nil
}
