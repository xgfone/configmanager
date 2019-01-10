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
	"os"
	"testing"
)

type testStruct struct {
	Age int `cli:"1"`
}

func (s *testStruct) Validate() error {
	s.Age = s.Age + 1
	return nil
}

func TestStructValidate(t *testing.T) {
	cli := NewFlagCliParser(os.Args[0], flag.ExitOnError, true)
	env := NewEnvVarParser("test")
	conf := NewConfig(cli).AddParser(env)

	s := testStruct{}
	conf.RegisterStruct("", &s)
	if err := conf.Parse("-age", "2"); err != nil {
		t.Error(err)
	}

	if s.Age != 3 {
		t.Errorf("Age should be 3, but is %d", s.Age)
	}
}

func ExampleConfig() {
	cliOpts1 := []Opt{
		StrOpt("", "required", "", "required").SetValidators(NewStrLenValidator(1, 10)),
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
	if err := Conf.Parse("--cli-no=0", "--required", "required"); err != nil {
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

	type Sub struct {
		Parent string `default:""`
	}

	type S struct {
		Name    string  `name:"name" cli:"1" default:"Aaron"`
		Age     int8    `cli:"t" default:"123"`
		Address Address `group:"group" cli:"true"`
		Ignore  string  `name:"-"`

		Sub1 Sub `group:"sub1" cli:"true"`
		Sub2 Sub `cli:"true"`
		Sub3 Sub `cli:"true"`
	}

	cli := NewFlagCliParser(os.Args[0], flag.ExitOnError, true)
	env := NewEnvVarParser("test")
	conf := NewConfig(cli).AddParser(env)

	s := S{}
	conf.RegisterStruct("", &s)
	if err := conf.Parse("-age", "18", "--group-address", "Beijing,Shanghai", "--sub1-parent", "abc", "--sub2-parent", "xyz"); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Name: %s\n", s.Name)
	fmt.Printf("Age: %d\n", s.Age)
	fmt.Printf("Address: %s\n", s.Address.Address)
	fmt.Printf("Parent1: %s\n", s.Sub1.Parent)
	fmt.Printf("Parent2: %s\n", s.Sub2.Parent)
	fmt.Printf("Parent3: %s\n", s.Sub3.Parent)

	// Output:
	// Name: Aaron
	// Age: 18
	// Address: [Beijing Shanghai]
	// Parent1: abc
	// Parent2: xyz
	// Parent3:
}

func ExampleNewEnvVarParser() {
	// Simulate the environment variable.
	os.Setenv("TEST_VAR1", "abc")
	os.Setenv("TEST_GROUP_VAR2", "123")

	cli := NewFlagCliParser(os.Args[0], flag.ExitOnError, true)
	env := NewEnvVarParser("test")
	conf := NewConfig(cli).AddParser(env)

	opt1 := Str("var1", "", "the environment var 1")
	opt2 := Int("var2", 0, "the environment var 2")
	conf.RegisterOpt("", opt1)
	conf.RegisterOpt("group", opt2)
	if err := conf.Parse([]string{}...); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("var1=%s\n", conf.String("var1"))
	fmt.Printf("var2=%d\n", conf.Group("group").Int("var2"))

	// Output:
	// var1=abc
	// var2=123
}

func ExampleConfig_Watch() {
	opt := Str("watchval", "abc", "test watch value")
	cli := NewFlagCliParser(os.Args[0], flag.ExitOnError, true)
	conf := NewConfig(cli)
	conf.RegisterCliOpt("test", opt)

	conf.Watch(func(gname, name string, value interface{}) {
		fmt.Printf("group=%s, name=%s, value=%v\n", gname, name, value)
	})

	if err := conf.Parse([]string{}...); err != nil {
		fmt.Println(err)
		return
	}

	// Output:
	// group=test, name=watchval, value=abc
}
