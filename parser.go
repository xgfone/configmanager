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
	// The parser can get any information from the first argument.
	//
	// When the parser parsed out the option value, it should call the second
	// argument, which is a function to set the group option. The function has
	// three arguments, that's, the group name, the option name, and the option
	// value. For the default group, the group name may be "" instead,
	//
	// If there is any error, the parser should stop to parse and return it.
	//
	// If a certain option has no value, the parser should not return a default
	// one instead.
	Parse(c *Config, setOptionValue func(string, string, interface{})) error
}

// CliParser is an interface to parse the CLI arguments.
type CliParser interface {
	// Name returns the name of the CLI parser to identify a CLI parser.
	Name() string

	// Parse the value of the registered CLI options.
	//
	// The parser can get any information from the first argument.
	//
	// When the parser parsed out the option value, it should call the second
	// argument, which is a function to set the group option. The function has
	// three arguments, that's, the group name, the option name, and the option
	// value. For the default group, the group name may be "" instead,
	//
	// If there are the rest CLI arguments, that's those that does not start
	// with the prefix "-", "--" or others, etc, the parser should call the
	// second arguments, which is a function to set the rest arguments.
	//
	// The last argument, arguments, is the CLI arguments to be parsed, which
	// is a string slice, not nil, but it maybe have no elements. The parser
	// implementor should not use os.Args[1:] when it's empty, because it has
	// been confirmed.
	//
	// If there is any error, the parser should stop to parse and return it.
	//
	// If a certain option has no value, the parser should not return a default
	// one instead.
	Parse(c *Config, setOptionValue func(string, string, interface{}),
		setArgs func([]string), arguments []string) error
}

type flagParser struct {
	flagSet    *flag.FlagSet
	name       string
	errhandler flag.ErrorHandling
}

// NewDefaultFlagCliParser returns a new CLI parser based on flag.
//
// The parser will use flag.CommandLine to parse the CLI arguments.
func NewDefaultFlagCliParser() CliParser {
	return flagParser{
		flagSet: flag.CommandLine,
	}
}

// NewFlagCliParser returns a new CLI parser based on flag.FlagSet.
//
// The arguments is the same as that of flag.NewFlagSet(), but if the name is
// "", it will be filepath.Base(os.Args[0]).
//
// When other libraries use the default global flag.FlagSet, that's
// flag.CommandLine, such as github.com/golang/glog, please use
// NewDefaultFlagCliParser(), not this function.
func NewFlagCliParser(appName string, errhandler flag.ErrorHandling) CliParser {
	if appName == "" {
		appName = filepath.Base(os.Args[0])
	}
	return flagParser{
		name:       appName,
		errhandler: errhandler,
	}
}

func (f flagParser) Name() string {
	return "flag"
}

func (f flagParser) Parse(c *Config, set1 func(string, string, interface{}),
	set2 func([]string), arguments []string) (err error) {
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

	// Parse the CLI arguments.
	if err = flagSet.Parse(arguments); err != nil {
		return
	}

	// Acquire the result.
	set2(flagSet.Args())
	flagSet.Visit(func(fg *flag.Flag) {
		set1(name2group[fg.Name], name2opt[fg.Name], fg.Value.String())
	})

	return
}

type iniParser struct {
	sep     string
	optName string
}

// NewSimpleIniParser returns a new ini parser based on the file.
//
// The argument is the option name which the parser needs. It should be
// registered, and parsed before this parser runs.
//
// The ini parser supports the line comments starting with "#" or "//".
// The key and the value is separated by an equal sign, that's =.
//
// Notice: the options that have not been assigned to a certain group will be
// divided into the default group.
func NewSimpleIniParser(optName string) Parser {
	return iniParser{optName: optName, sep: "="}
}

func (p iniParser) Name() string {
	return "ini"
}

// func (p iniParser) Parse(_default string, opts map[string][]Opt,
// 	conf map[string]interface{}) (results map[string]map[string]interface{},
// 	err error) {
func (p iniParser) Parse(c *Config, set func(string, string, interface{})) error {
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
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#" or "//".
		if (line[0] == '#') ||
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
			if !unicode.IsNumber(r) && !unicode.IsLetter(r) {
				return fmt.Errorf("the key is not an valid identifier")
			}
		}
		set(gname, key, strings.TrimSpace(line[n+len(p.sep):len(line)]))
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
func (e envVarParser) Parse(c *Config, set func(string, string, interface{})) error {
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
				set(info[0], info[1], items[1])
			}
		}
	}

	return nil
}
