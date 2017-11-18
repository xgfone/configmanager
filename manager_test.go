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
	"fmt"
	"os"
)

func ExampleConfig() {
	validators := []Validator{NewStrLenValidator(1, 10)}
	cliOpts1 := []Opt{
		StrOpt("", "required", "", "required").SetValidators(validators),
		BoolOpt("", "yes", true, "test bool option"),
	}

	cliOpts2 := []Opt{
		BoolOpt("", "no", false, "test bool option"),
		StrOpt("", "optional", "optional", "optional"),
	}

	opts := []Opt{
		StrOpt("", "opt", "", "test opt"),
	}

	Conf.RegisterCliOpts("", cliOpts1)
	Conf.RegisterCliOpts("cli", cliOpts2)
	Conf.RegisterOpts("group", opts)

	args := []string{"-cli_no=0", "-required", "required"}
	// args = nil // You can pass nil.
	if err := Conf.Parse(args); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Conf.StringD("required", "abc"))
	fmt.Println(Conf.Bool("yes"))

	fmt.Println(Conf.Group("cli").String("optional"))
	fmt.Println(Conf.Group("cli").Bool("no"))

	fmt.Println(Conf.Group("group").StringD("opt", "opt"))

	// Output:
	// required
	// true
	// optional
	// false
	//
}

func ExampleConfig_RegisterStruct() {
	type Address struct {
		Address []string
	}

	type S struct {
		Name    string  `name:"name" cli:"1" default:"Aaron"`
		Age     int8    `cli:"t" default:"123"`
		Address Address `group:"group" cli:"true"`
		Ignore  string  `name:"-"`
	}

	cli := NewDefaultFlagCliParser()
	env := NewEnvVarParser("test")
	conf := NewConfig(cli).AddParser(env)

	s := S{}
	conf.RegisterStruct("", &s)
	if err := conf.Parse([]string{"-age", "18", "-group_address", "Beijing,Shanghai"}); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Name: %s\n", s.Name)
	fmt.Printf("Age: %d\n", s.Age)
	fmt.Printf("Address: %s\n", s.Address.Address)

	// Output:
	// Name: Aaron
	// Age: 18
	// Address: [Beijing Shanghai]
}

func ExampleNewEnvVarParser() {
	// Simulate the environment variable.
	os.Setenv("TEST_VAR1", "abc")
	os.Setenv("TEST_GROUP_VAR2", "123")

	cli := NewDefaultFlagCliParser()
	env := NewEnvVarParser("test")
	conf := NewConfig(cli).AddParser(env)

	opt1 := Str("var1", "", "the environment var 1")
	opt2 := Int("var2", 0, "the environment var 2")
	conf.RegisterOpt("", opt1)
	conf.RegisterOpt("group", opt2)
	if err := conf.Parse(nil); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("var1=%s\n", conf.String("var1"))
	fmt.Printf("var2=%d\n", conf.Group("group").Int("var2"))

	// Output:
	// var1=abc
	// var2=123
}
