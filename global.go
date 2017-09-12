package config

// IniParserOptName is the name of the option to indicate the path of
// the ini config file.
var IniParserOptName = "config-file"

// Conf is the global config manager.
var Conf = NewDefault()

// NewDefault returns a new default config manager.
func NewDefault() *Config {
	cli := NewDefaultFlagCliParser()
	ini := NewSimpleIniParser(IniParserOptName)
	conf := NewConfig(cli).AddParser(ini)
	conf.RegisterCliOpt("", StrOpt("", IniParserOptName, "",
		"The path of the ini config file."))
	return conf
}
