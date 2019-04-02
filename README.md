# go-config [![Build Status](https://travis-ci.org/xgfone/go-config.svg?branch=master)](https://travis-ci.org/xgfone/go-config) [![GoDoc](https://godoc.org/github.com/xgfone/go-config?status.svg)](http://godoc.org/github.com/xgfone/go-config) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-config/master/LICENSE)
An extensible go configuration. The default parsers can parse the CLI and ENV arguments and the ini and property file. You can implement and register your parser, and the configuration engine will call the parser to parse the configuration.

The inspiration is from [oslo.config](https://github.com/openstack/oslo.config), which is a `OpenStack` library for config.

The current version is `v11`. See [DOC](https://godoc.org/github.com/xgfone/go-config).

The supported Go version: `1.x`.


## Goal

1. A atomic key-value configuration center based on group as the core function.
2. A set of the plugin based on parser as the auxiliary  function.
3. Change the configuration dynamically during running and watch it.


## Principle of Work

1. New a `Config` engine.
2. (Optional) Set the CLI parser or add non-CLI parsers into the `Config`.
3. Register the configuration options into `Config`.
3. Call the method `Parse()` to parse the configurations.
    1. Call the method `Pre()` of the parsers in turn to initialize them.
    2. Call the method `Parse()` of the parsers in turn to parse the options.
    3. Call the method `Post()` of the parsers in turn to clean them.
    4. Assign the default value to the unresolved option if it has a default.
    5. Check whether some required options have not been unresolved.

**Notice:** when setting the parsed value, it will calling the validators to validate it if setting the validators for the option.


## Parser

In order to deveplop a new parser, you just need to implement the interface `Parser`. But `Config` does not distinguish the CLI parser and the common parser, which have the same interface `Parser`. You can add them by calling `AddParser()`. See the example below.

**Notice:** the priority of the CLI parser should be higher than that of other parsers.


## Read and Modify the value from `Config`

It's thread-safe for the application to read the configuration value from the `Config`, but you must not modify it.

If you want to the value of a certain configuration, you should call the method `SetOptValue(priority int, groupName, optName, newOptValue)`. For the default group, `groupName` may be `""`. If the setting fails, it will return an error. Moreover, `SetOptValue` is thread-safe. During the running, therefore, you can get and set the configuration value between goroutines dynamically.

For the modifiable type, such as slice or map, in order to modify them, you should clone them firstly, then modify the cloned value and call `SetOptValue` with the cloned one.

For modify the value of the option, you can use the priority `0`, which is the highest priority and can be cover any other value.


## Observe the changed configuration

You can use the method `Observe(callback func(groupName, optName string, optValue interface{}))` to monitor what the configuration is modified to: when a certain configuration is modified, the callback function will be called.

Notice: the callback should finish as soon as possible because the callback is called synchronously at when the configuration is modified.


## Usage
```go
package main

import (
	"fmt"

	config "github.com/xgfone/go-config"
)

func main() {
	cliParser := config.NewDefaultFlagCliParser(true)
	iniParser := config.NewSimpleIniParser("config-file")
	conf := config.NewConfig().AddParser(cliParser, iniParser)

	ipOpt := config.StrOpt("", "ip", "", "the ip address").SetValidators(config.NewIPValidator())
	conf.RegisterCliOpt("", ipOpt)
	conf.RegisterCliOpt("", config.IntOpt("", "port", 80, "the port"))
	conf.RegisterCliOpt("", config.StrOpt("", "config-file", "", "The path of the ini config file."))
	conf.RegisterCliOpt("redis", config.StrOpt("", "conn", "redis://127.0.0.1:6379/0", "the redis connection url"))
	conf.SetVersion("1.0.0") // Print the version and exit when giving the CLI option version.

	if err := conf.Parse(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf.String("ip"))
	fmt.Println(conf.Int("port"))
	fmt.Println(conf.Group("redis").String("conn"))
	fmt.Println(conf.Args())

	// Execute:
	//     PROGRAM -ip 0.0.0.0 aa bb cc
	//
	// Output:
	//     0.0.0.0
	//     80
	//     redis://127.0.0.1:6379/0
	//     [aa bb cc]
}
```

You can also create a new `Config` by the `NewDefault()`, which will use `NewDefaultFlagCliParser(true)` as the CLI parser, add the ini parser `NewSimpleIniParser()` and register the CLI option `config-file`, which you change it by modifying the value of the variable `IniParserOptName`. Notice: `NewDefault()` does not add the environment variable parser, and you need to add it by hand, such as `NewDefault().AddParser(NewEnvVarParser(""))`.

The package has created a global default `Config`, `Conf`, created by `NewDefault()` like doing above. You can use it, like the global variable `CONF` in `oslo.config`. For example,
```go
package main

import (
	"fmt"

	config "github.com/xgfone/go-config"
)

var opts = []config.Opt{
	config.Str("ip", "", "the ip address").AddValidators(config.NewIPValidator()),
	config.Int("port", 80, "the port").AddValidators(config.NewPortValidator()),
}

func main() {
	config.Conf.RegisterCliOpts("", opts)
	config.Conf.Parse("-ip", "0.0.0.0") // You can pass nil

	fmt.Println(config.Conf.String("ip")) // Output: 0.0.0.0
	fmt.Println(config.Conf.Int("port"))  // Output: 80
}
```

You can watch the change of the configuration option.
```go
package main

import (
	"fmt"

	config "github.com/xgfone/go-config"
)

func main() {
	conf := config.NewConfig()

	conf.RegisterCliOpt("test", config.Str("watchval", "abc", "test watch value"))
	conf.Observe(func(gname, name string, value interface{}) {
		fmt.Printf("[Observer] group=%s, name=%s, value=%v\n", gname, name, value)
	})

	conf.Parse()

	// Set the option vlaue during the program is running.
	conf.SetOptValue(0, "test", "watchval", "123")

	// Output:
	// [Observer] group=test, name=watchval, value=abc
	// [Observer] group=test, name=watchval, value=123
}
```


You also register a struct then use it.
```go
package main

import (
	"fmt"

	config "github.com/xgfone/go-config"
)

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

func main() {
	var conf Config
	config.Conf.RegisterCliStruct("", &conf)

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

	if err := config.Conf.Parse(cliArgs...); err != nil {
		fmt.Println(err)
		return
	}

	// Get the configuration by the struct.
	fmt.Printf("------ Struct ------\n")
	fmt.Printf("Addr: %s\n", conf.Addr)
	fmt.Printf("File: %s\n", conf.File)
	fmt.Printf("Level: %s\n", conf.Level)
	fmt.Printf("Ignore: %v\n", conf.Ignore)
	fmt.Printf("DB1.MySQL.Conn: %s\n", conf.DB1.MySQL.Conn)
	fmt.Printf("DB1.MySQL.MaxConnNum: %d\n", conf.DB1.MySQL.MaxConnNum)
	fmt.Printf("DB2.MySQL.Conn: %s\n", conf.DB2.MySQL.Conn)
	fmt.Printf("DB2.MySQL.MaxConnNum: %d\n", conf.DB2.MySQL.MaxConnNum)
	fmt.Printf("DB3.MySQL.Conn: %s\n", conf.DB3.MySQL.Conn)
	fmt.Printf("DB3.MySQL.MaxConnNum: %d\n", conf.DB3.MySQL.MaxConnNum)
	fmt.Printf("DB4.MySQL.Conn: %s\n", conf.DB4.MySQL.Conn)
	fmt.Printf("DB4.MySQL.MaxConnNum: %d\n", conf.DB4.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB1.MySQL.Conn: %s\n", conf.DB5.DB1.MySQL.Conn)
	fmt.Printf("DB5.DB1.MySQL.MaxConnNum: %d\n", conf.DB5.DB1.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB2.MySQL.Conn: %s\n", conf.DB5.DB2.MySQL.Conn)
	fmt.Printf("DB5.DB2.MySQL.MaxConnNum: %d\n", conf.DB5.DB2.MySQL.MaxConnNum)
	fmt.Printf("DB5.DB3.MySQL.Conn: %s\n", conf.DB5.DB3.MySQL.Conn)
	fmt.Printf("DB5.DB3.MySQL.MaxConnNum: %d\n", conf.DB5.DB3.MySQL.MaxConnNum)

	// Get the configuration by the Config.
	fmt.Printf("\n------ Config ------\n")
	fmt.Printf("Addr: %s\n", config.Conf.String("addr"))
	fmt.Printf("File: %s\n", config.Conf.Group("log").String("file"))
	fmt.Printf("Level: %s\n", config.Conf.Group("log").String("level"))
	fmt.Printf("Ignore: %v\n", config.Conf.BoolD("ignore", true))
	fmt.Printf("DB1.MySQL.Conn: %s\n", config.Conf.Group("db1").Group("mysql").String("conn"))
	fmt.Printf("DB1.MySQL.MaxConnNum: %d\n", config.Conf.Group("db1.mysql").Int("maxconn"))
	fmt.Printf("DB2.MySQL.Conn: %s\n", config.Conf.Group("db02.mysql").String("conn"))
	fmt.Printf("DB2.MySQL.MaxConnNum: %d\n", config.Conf.Group("db02").Group("mysql").Int("maxconn"))
	fmt.Printf("DB3.MySQL.Conn: %s\n", config.Conf.Group("db03.mysql").String("conn"))
	fmt.Printf("DB3.MySQL.MaxConnNum: %d\n", config.Conf.Group("db03").Group("mysql").Int("maxconn"))
	fmt.Printf("DB4.MySQL.Conn: %s\n", config.Conf.Group("db004").Group("mysql").String("conn"))
	fmt.Printf("DB4.MySQL.MaxConnNum: %d\n", config.Conf.Group("db004.mysql").Int("maxconn"))
	fmt.Printf("DB5.DB1.MySQL.Conn: %s\n", config.Conf.Group("db111").Group("mysql").String("conn"))
	fmt.Printf("DB5.DB1.MySQL.MaxConnNum: %d\n", config.Conf.Group("db111").Group("mysql").Int("maxconn"))
	fmt.Printf("DB5.DB2.MySQL.Conn: %s\n", config.Conf.Group("db").Group("db222").Group("mysql").String("conn"))
	fmt.Printf("DB5.DB2.MySQL.MaxConnNum: %d\n", config.Conf.Group("db.db222").Group("mysql").Int("maxconn"))
	fmt.Printf("DB5.DB3.MySQL.Conn: %s\n", config.Conf.Group("db").Group("db3.mysql").String("conn"))
	fmt.Printf("DB5.DB3.MySQL.MaxConnNum: %d\n", config.Conf.Group("db.db3.mysql").Int("maxconn"))

	// Print the group tree for debug.
	fmt.Printf("\n------ Debug ------\n")
	Conf.PrintGroupTree()

	// Output:
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
```
