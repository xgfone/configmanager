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

func ExampleConfig_Observe() {
	conf := NewConfig()
	conf.RegisterCliOpt("test", Str("watchval", "abc", "test watch value"))
	conf.Observe(func(gname, name string, value interface{}) {
		fmt.Printf("group=%s, name=%s, value=%v\n", gname, name, value)
	})

	conf.Parse() // Start the config

	// Set the option vlaue during the program is running.
	conf.SetOptValue(0, "test", "watchval", "123")

	// Output:
	// group=test, name=watchval, value=abc
	// group=test, name=watchval, value=123
}

func ExampleNewEnvVarParser() {
	// Simulate the environment variable.
	os.Setenv("TEST_VAR1", "abc")
	os.Setenv("TEST_GROUP1_GROUP2_VAR2", "123")

	conf := NewConfig().AddParser(NewEnvVarParser("test"))

	opt1 := Str("var1", "", "the environment var 1")
	opt2 := Int("var2", 0, "the environment var 2")

	conf.RegisterOpt("", opt1)
	conf.RegisterOpt("group1.group2", opt2)

	if err := conf.Parse(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("var1=%s\n", conf.String("var1"))
	fmt.Printf("var2=%d\n", conf.Group("group1.group2").Int("var2"))

	// Output:
	// var1=abc
	// var2=123
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

	conf := NewConfig().AddParser(NewFlagCliParser(nil, true))
	conf.RegisterCliOpts("", cliOpts1)
	conf.RegisterCliOpts("cli", cliOpts2)
	conf.RegisterCliOpts("group1.group2", opts)

	cliArgs := []string{
		"--cli.no=0",
		"--required", "required",
		"--group1.group2.opt", "option",
	}
	if err := conf.Parse(cliArgs...); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf.StringD("required", "abc"))
	fmt.Println(conf.Bool("yes"))

	fmt.Println(conf.Group("cli").String("optional"))
	fmt.Println(conf.Group("cli").Bool("no"))

	fmt.Println(conf.Group("group1.group2").StringD("opt", "opt"))
	fmt.Println(conf.Group("group1").Group("group2").StringD("opt", "opt"))

	// Output:
	// required
	// true
	// optional
	// false
	// option
	// option
}

func ExampleConfig_RegisterStruct() {
	type MySQL struct {
		Conn       string `help:"the connection to mysql server"`
		MaxConnNum int    `name:"maxconn" default:"3" help:"the maximum number of connections"`
	}

	type DB struct {
		MySQL MySQL
	}

	type DBWrapper struct {
		DB1 DB `group:".db111"` // Add the prefix . to reset the parent group.
		DB2 DB `group:"db222"`
		DB3 DB
	}

	type Config struct {
		Addr  string `default:":80" help:"the address to listen to"`
		File  string `default:"" group:"log" help:"the log file path"`
		Level string `default:"debug" group:"log" help:"the log level, such as debug, info, etc"`

		Ignore bool `name:"-" default:"true"`

		DB1 DB
		DB2 DB `name:"db02"`
		DB3 DB `group:"db03"`
		DB4 DB `name:"db04" group:"db004"`

		DB5 DBWrapper `group:"db"`
	}

	// Set the debug to output the process that handles the configuration.
	// Conf.SetDebug(true)

	var config Config
	Conf.RegisterCliStruct("", &config) // We use RegisterCliStruct instead of RegisterStruct

	// Only for test
	cliArgs := []string{
		"--addr", "0.0.0.0:80",
		"--log.file", "/var/log/test.log",
		"--db1.mysql.conn", "user:pass@tcp(localhost:3306)/db1",
		"--db02.mysql.conn", "user:pass@tcp(localhost:3306)/db2",
		"--db03.mysql.conn", "user:pass@tcp(localhost:3306)/db3",
		"--db004.mysql.conn", "user:pass@tcp(localhost:3306)/db4",
		"--db111.mysql.conn", "user:pass@tcp(localhost:3306)/db5-1",
		"--db.db222.mysql.conn", "user:pass@tcp(localhost:3306)/db5-2",
		"--db.db3.mysql.conn", "user:pass@tcp(localhost:3306)/db5-3",
	}

	if err := Conf.Parse(cliArgs...); err != nil {
		fmt.Println(err)
		return
	}

	// Get the configuration by the struct.
	fmt.Printf("------ Struct ------\n")
	fmt.Printf("Addr: %s\n", config.Addr)
	fmt.Printf("File: %s\n", config.File)
	fmt.Printf("Level: %s\n", config.Level)
	fmt.Printf("Ignore: %v\n", config.Ignore)
	fmt.Printf("DB1.MySQL.Conn: %s\n", config.DB1.MySQL.Conn)
	fmt.Printf("DB1.MySQL.MaxConnNum: %d\n", config.DB1.MySQL.MaxConnNum)
	fmt.Printf("DB2.MySQL.Conn: %s\n", config.DB2.MySQL.Conn)
	fmt.Printf("DB2.MySQL.MaxConnNum: %d\n", config.DB2.MySQL.MaxConnNum)
	fmt.Printf("DB3.MySQL.Conn: %s\n", config.DB3.MySQL.Conn)
	fmt.Printf("DB3.MySQL.MaxConnNum: %d\n", config.DB3.MySQL.MaxConnNum)
	fmt.Printf("DB4.MySQL.Conn: %s\n", config.DB4.MySQL.Conn)
	fmt.Printf("DB4.MySQL.MaxConnNum: %d\n", config.DB4.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB1.MySQL.Conn: %s\n", config.DB5.DB1.MySQL.Conn)
	fmt.Printf("DB5.DB1.MySQL.MaxConnNum: %d\n", config.DB5.DB1.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB2.MySQL.Conn: %s\n", config.DB5.DB2.MySQL.Conn)
	fmt.Printf("DB5.DB2.MySQL.MaxConnNum: %d\n", config.DB5.DB2.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB3.MySQL.Conn: %s\n", config.DB5.DB3.MySQL.Conn)
	fmt.Printf("DB5.DB3.MySQL.MaxConnNum: %d\n", config.DB5.DB3.MySQL.MaxConnNum)

	// Get the configuration by the Config.
	fmt.Printf("\n------ Config ------\n")
	fmt.Printf("Addr: %s\n", Conf.String("addr"))
	fmt.Printf("File: %s\n", Conf.Group("log").String("file"))
	fmt.Printf("Level: %s\n", Conf.Group("log").String("level"))
	fmt.Printf("Ignore: %v\n", Conf.BoolD("ignore", true))
	fmt.Printf("DB1.MySQL.Conn: %s\n", Conf.Group("db1").Group("mysql").String("conn"))
	fmt.Printf("DB1.MySQL.MaxConnNum: %d\n", Conf.Group("db1.mysql").Int("maxconn"))
	fmt.Printf("DB2.MySQL.Conn: %s\n", Conf.Group("db02.mysql").String("conn"))
	fmt.Printf("DB2.MySQL.MaxConnNum: %d\n", Conf.Group("db02").Group("mysql").Int("maxconn"))
	fmt.Printf("DB3.MySQL.Conn: %s\n", Conf.Group("db03.mysql").String("conn"))
	fmt.Printf("DB3.MySQL.MaxConnNum: %d\n", Conf.Group("db03").Group("mysql").Int("maxconn"))
	fmt.Printf("DB4.MySQL.Conn: %s\n", Conf.Group("db004").Group("mysql").String("conn"))
	fmt.Printf("DB4.MySQL.MaxConnNum: %d\n", Conf.Group("db004.mysql").Int("maxconn"))
	fmt.Printf("DB5.DB1.MySQL.Conn: %s\n", Conf.Group("db111").Group("mysql").String("conn"))
	fmt.Printf("DB5.DB1.MySQL.MaxConnNum: %d\n", Conf.Group("db111").Group("mysql").Int("maxconn"))
	fmt.Printf("DB5.DB2.MySQL.Conn: %s\n", Conf.Group("db").Group("db222").Group("mysql").String("conn"))
	fmt.Printf("DB5.DB2.MySQL.MaxConnNum: %d\n", Conf.Group("db.db222").Group("mysql").Int("maxconn"))
	fmt.Printf("DB5.DB3.MySQL.Conn: %s\n", Conf.Group("db").Group("db3.mysql").String("conn"))
	fmt.Printf("DB5.DB3.MySQL.MaxConnNum: %d\n", Conf.Group("db.db3.mysql").Int("maxconn"))

	// Print the group tree for debug.
	fmt.Printf("\n------ Debug ------\n")
	Conf.PrintGroupTree()

	// Unordered output:
	// ------ Struct ------
	// Addr: 0.0.0.0:80
	// File: /var/log/test.log
	// Level: debug
	// Ignore: false
	// DB1.MySQL.Conn: user:pass@tcp(localhost:3306)/db1
	// DB1.MySQL.MaxConnNum: 3
	// DB2.MySQL.Conn: user:pass@tcp(localhost:3306)/db2
	// DB2.MySQL.MaxConnNum: 3
	// DB3.MySQL.Conn: user:pass@tcp(localhost:3306)/db3
	// DB3.MySQL.MaxConnNum: 3
	// DB4.MySQL.Conn: user:pass@tcp(localhost:3306)/db4
	// DB4.MySQL.MaxConnNum: 3
	// DB5.DB1.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-1
	// DB5.DB1.MySQL.MaxConnNum: 3
	// DB5.DB2.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-2
	// DB5.DB2.MySQL.MaxConnNum: 3
	// DB5.DB3.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-3
	// DB5.DB3.MySQL.MaxConnNum: 3
	//
	// ------ Config ------
	// Addr: 0.0.0.0:80
	// File: /var/log/test.log
	// Level: debug
	// Ignore: true
	// DB1.MySQL.Conn: user:pass@tcp(localhost:3306)/db1
	// DB1.MySQL.MaxConnNum: 3
	// DB2.MySQL.Conn: user:pass@tcp(localhost:3306)/db2
	// DB2.MySQL.MaxConnNum: 3
	// DB3.MySQL.Conn: user:pass@tcp(localhost:3306)/db3
	// DB3.MySQL.MaxConnNum: 3
	// DB4.MySQL.Conn: user:pass@tcp(localhost:3306)/db4
	// DB4.MySQL.MaxConnNum: 3
	// DB5.DB1.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-1
	// DB5.DB1.MySQL.MaxConnNum: 3
	// DB5.DB2.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-2
	// DB5.DB2.MySQL.MaxConnNum: 3
	// DB5.DB3.MySQL.Conn: user:pass@tcp(localhost:3306)/db5-3
	// DB5.DB3.MySQL.MaxConnNum: 3
	//
	// ------ Debug ------
	// |-->[DEFAULT]
	// |   |--> addr
	// |   |--> config-file
	// |-->[log]
	// |   |--> file
	// |   |--> level
	// |-->[db]
	// |   |-->[db222]
	// |   |   |-->[mysql]
	// |   |   |   |--> conn
	// |   |   |   |--> maxconn
	// |   |-->[db3]
	// |   |   |-->[mysql]
	// |   |   |   |--> conn
	// |   |   |   |--> maxconn
	// |-->[db1]
	// |   |-->[mysql]
	// |   |   |--> conn
	// |   |   |--> maxconn
	// |-->[db02]
	// |   |-->[mysql]
	// |   |   |--> conn
	// |   |   |--> maxconn
	// |-->[db03]
	// |   |-->[mysql]
	// |   |   |--> conn
	// |   |   |--> maxconn
	// |-->[db004]
	// |   |-->[mysql]
	// |   |   |--> conn
	// |   |   |--> maxconn
	// |-->[db111]
	// |   |-->[mysql]
	// |   |   |--> conn
	// |   |   |--> maxconn
}
