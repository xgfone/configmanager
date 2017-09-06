package configmanager

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
	// which indicates the configuration manager must pass the parser the value
	// of the option 'ip' when calling the method Parse, but the value of 'port'
	// is optional.
	GetKeys() map[string]bool

	// Register the options to the parser.
	//
	// The parser should only parse the value of the registered option and use
	// the method GetName() or GetShort().
	Register([]Opt)

	// Parse the value of the registered options.
	//
	// The arguments is the metadata in order to parse the options. For example,
	// for the redis parser, it's the connection information to connect to the
	// redis server; for the configuration file parser, it's the directory or
	// path of the configuration file, such as `config_file` in the property
	// file parser `NewSimplePropertyParser`.
	//
	// The result is the key-value pairs, which the key is the name of the
	// registered option.
	//
	// If a certain option has no value, the parser should not return a default
	// one.
	Parse(map[string]string) (map[string]string, error)
}

// CliParser is an interface to parse the CLI arguments.
type CliParser interface {
	// Name returns the name of the CLI parser to identify a CLI parser.
	Name() string

	// Register the options to the CLI parser.
	//
	// The parser should only parse the value of the registered option and use
	// the method GetName() or GetShort().
	Register([]Opt)

	// Parse the value of the registered CLI options.
	//
	// The arguments is the CLI arguments, but it may be nil.
	//
	// The result is the key-value pairs, which the key is the name of the
	// registered option.
	//
	// If a certain option has no value, the CLI parser should not return a
	// default one.
	Parse([]string) (map[string]string, []string, error)
}

type flagParser struct {
	flagSet *flag.FlagSet
}

// NewFlagCliParser returns a new CLI parser based on flag.FlagSet.
func NewFlagCliParser(appName string, errhandler flag.ErrorHandling) CliParser {
	return flagParser{
		flagSet: flag.NewFlagSet(appName, errhandler),
	}
}

func (f flagParser) Name() string {
	return "flag"
}

func (f flagParser) Register(opts []Opt) {
	for _, opt := range opts {
		f.flagSet.String(opt.GetName(), "", opt.GetHelp())
	}
}

func (f flagParser) Parse(as []string) (opts map[string]string, args []string, err error) {
	if err = f.flagSet.Parse(as); err != nil {
		return
	}

	args = f.flagSet.Args()
	opts = make(map[string]string, f.flagSet.NFlag())
	f.flagSet.Visit(func(fg *flag.Flag) {
		opts[fg.Name] = fg.Value.String()
	})
	return
}

type propertyParser struct {
	sep     string
	optName string
	opts    map[string]struct{}
}

// NewSimplePropertyParser returns a new property parser based on the file.
//
// The property parser supports the line comments starting with "#" or "//".
//
// The key and the value is separated by an equal sign, that's =.
func NewSimplePropertyParser(optName string) Parser {
	return propertyParser{
		optName: optName,

		sep:  "=",
		opts: make(map[string]struct{}, 8),
	}
}

func (p propertyParser) Name() string {
	return "property"
}

func (p propertyParser) GetKeys() map[string]bool {
	return map[string]bool{
		p.optName: false,
	}
}

func (p propertyParser) Register(opts []Opt) {
	for _, opt := range opts {
		p.opts[opt.GetName()] = struct{}{}
	}
}

func (p propertyParser) Parse(conf map[string]string) (opts map[string]string, err error) {
	filename, ok := conf[p.optName]
	if !ok || len(filename) == 0 {
		return
	}
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}

	lines := strings.Split(string(data), "\n")
	opts = make(map[string]string, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore the empty line.
		if len(line) == 0 {
			continue
		}

		// Ignore the line comments starting with "#" or "//".
		if (len(line) > 0 && line[0] == '#') ||
			(len(line) > 1 && line[0] == '/' && line[1] == '/') {
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
		opts[key] = strings.TrimSpace(line[n+len(p.sep) : len(line)])
	}
	return
}
