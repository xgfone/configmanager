package configmanager

import (
	"flag"
	"os"
	"path/filepath"
)

// PropertyParserOptName is the option name to indicate the path of the property
// config file.
var PropertyParserOptName = "config_file"

// Conf is the global config manager.
var Conf = NewDefault()

// NewDefault returns a new default config manager.
func NewDefault() *ConfigManager {
	cli := NewFlagCliParser(filepath.Base(os.Args[0]), flag.ExitOnError)
	prop := NewSimplePropertyParser(PropertyParserOptName)
	conf := NewConfigManager(cli).AddParser(prop)
	conf.RegisterCliOpt(NewStrOpt("", PropertyParserOptName, nil, false,
		"The path of the property config file."))
	return conf
}
