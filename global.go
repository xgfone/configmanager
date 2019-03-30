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

// IniParserOptName is the name of the option to indicate the path of
// the ini config file.
var IniParserOptName = "config-file"

// Conf is the global config manager.
var Conf = NewDefault()

// NewDefault returns a new default config manager.
//
// The default config manager does not add the environment variable parser.
// You need to add it by hand, such as NewDefault().AddParser(NewEnvVarParser("")).
func NewDefault() *Config {
	cli := NewDefaultFlagCliParser(true)
	ini := NewSimpleIniParser(IniParserOptName)
	conf := NewConfig().SetCliParser(cli).AddParser(ini)
	conf.AddIgnoredDeferOption("", IniParserOptName)

	return conf
}
