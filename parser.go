package config

import (
	"flag"
	"fmt"
	"io/ioutil"
	"strings"
	"unicode"
)

// Parser is an parser interface.
type Parser interface {
	// Name returns the name of the parser to identify a parser.
	Name() string

	// If the parser needs some configurations, it can return all those names
	// by the method, and mark whether they are required.
	//
	// For example, the method returns {"ip": true, "port": false},
	// which indicates the configuration must pass the parser the value of the
	// option 'ip' when calling the method Parse, but the value of 'port' is
	// optional. These option values will be acquired from the default group.
	GetKeys() map[string]bool

	// Register the options of all the groups to the parser.
	//
	// The first argument is the name of the default group.
	// The key of the second argument map is the group name.
	//
	// The parser should only parse the value of the registered option and use
	// the method GetName() or GetShort().
	Register(string, map[string][]Opt)

	// Parse the value of the registered options.
	//
	// The arguments is the metadata in order to parse the options. For example,
	// for the redis parser, it's the connection information to connect to the
	// redis server; for the configuration file parser, it's the directory or
	// path of the configuration file, such as `config_file` in the ini file
	// parser `NewSimpleIniParser`.
	//
	// The key of the result map is the group name, and the value of that is
	// the key-value pairs, which the key is the name of the registered option.
	//
	// If a certain option has no value, the parser should not return a default
	// one.
	Parse(map[string]string) (map[string]map[string]string, error)
}

// CliParser is an interface to parse the CLI arguments.
type CliParser interface {
	// Name returns the name of the CLI parser to identify a CLI parser.
	Name() string

	// Register the options to the CLI parser.
	//
	// The first argument is the name of the default group.
	// The key of the second argument map is the group name.
	//
	// The parser should only parse the value of the registered option and use
	// the method GetName() or GetShort().
	Register(string, map[string][]Opt)

	// Parse the value of the registered CLI options.
	//
	// The argument is the CLI arguments, but it may be nil.
	//
	// The key of the result map is the group name, and the value of that is
	// the key-value pairs, which the key is the name of the registered option.
	//
	// If a certain option has no value, the CLI parser should not return a
	// default one.
	Parse([]string) (map[string]map[string]string, []string, error)
}

type flagGroupOpt struct {
	group  string
	option string
}

type flagParser struct {
	flagSet *flag.FlagSet
	groups  map[string]flagGroupOpt
}

// NewFlagCliParser returns a new CLI parser based on flag.FlagSet.
func NewFlagCliParser(appName string, errhandler flag.ErrorHandling) CliParser {
	return flagParser{
		flagSet: flag.NewFlagSet(appName, errhandler),
		groups:  make(map[string]flagGroupOpt, 8),
	}
}

func (f flagParser) Name() string {
	return "flag"
}

func (f flagParser) Register(_default string, opts map[string][]Opt) {
	for group, _opts := range opts {
		for _, opt := range _opts {
			name := opt.GetName()
			if group != _default {
				name = fmt.Sprintf("%s_%s", group, name)
			}
			f.groups[name] = flagGroupOpt{group: group, option: opt.GetName()}

			if opt.IsBool() {
				var _default bool
				if v := opt.GetDefault(); v != nil {
					_default = v.(bool)
				}
				f.flagSet.Bool(name, _default, opt.GetHelp())
			} else {
				f.flagSet.String(name, "", opt.GetHelp())
			}
		}
	}
}

func (f flagParser) Parse(as []string) (opts map[string]map[string]string, args []string, err error) {
	if err = f.flagSet.Parse(as); err != nil {
		return
	}

	args = f.flagSet.Args()
	opts = make(map[string]map[string]string, len(f.groups))
	f.flagSet.Visit(func(fg *flag.Flag) {
		opt := f.groups[fg.Name]
		if group, ok := opts[opt.group]; ok {
			group[opt.option] = fg.Value.String()
		} else {
			opts[opt.group] = map[string]string{opt.option: fg.Value.String()}
		}
	})
	return
}

type iniParser struct {
	sep          string
	optName      string
	defaultGroup string
	opts         map[string]map[string]struct{}
}

// NewSimpleIniParser returns a new ini parser based on the file.
//
// The ini parser supports the line comments starting with "#" or "//".
//
// The key and the value is separated by an equal sign, that's =.
//
// Notice: the options that have not been assigned to a certain group will be
// divided into the default group.
func NewSimpleIniParser(optName string) Parser {
	return &iniParser{
		optName:      optName,
		sep:          "=",
		defaultGroup: "DEFAULT",
		opts:         make(map[string]map[string]struct{}, 8),
	}
}

func (p *iniParser) Name() string {
	return "ini"
}

func (p *iniParser) GetKeys() map[string]bool {
	return map[string]bool{
		p.optName: false,
	}
}

func (p *iniParser) Register(_default string, opts map[string][]Opt) {
	p.defaultGroup = _default
	for group, _opts := range opts {
		g, ok := p.opts[group]
		if !ok {
			g = make(map[string]struct{}, len(_opts))
			p.opts[group] = g
		}
		for _, opt := range _opts {
			g[opt.GetName()] = struct{}{}
		}
	}
}

func (p *iniParser) Parse(conf map[string]string) (opts map[string]map[string]string, err error) {
	filename, ok := conf[p.optName]
	if !ok || len(filename) == 0 {
		return
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	group := make(map[string]string, 8)
	opts = make(map[string]map[string]string, len(p.opts))
	opts[p.defaultGroup] = group

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#" or "//".
		if (line[0] == '#') || (len(line) > 1 && line[0] == '/' && line[1] == '/') {
			continue
		}

		// Start a new group
		if line[0] == '[' && line[len(line)-1] == ']' {
			gname := strings.TrimSpace(line[1 : len(line)-1])
			if gname == "" {
				return nil, fmt.Errorf("the group is empty")
			}
			if group = opts[gname]; group == nil {
				group = make(map[string]string, 4)
				opts[gname] = group
			}
			continue
		}

		n := strings.Index(line, p.sep)
		if n == -1 {
			err = fmt.Errorf("the line misses the separator '%s'", p.sep)
			return
		}

		key := strings.TrimSpace(line[0:n])
		for _, r := range key {
			if !unicode.IsNumber(r) && !unicode.IsLetter(r) {
				err = fmt.Errorf("the key is not an valid identifier")
				return
			}
		}
		group[key] = strings.TrimSpace(line[n+len(p.sep) : len(line)])
	}
	return
}
