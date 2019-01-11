# go-config [![Build Status](https://travis-ci.org/xgfone/go-config.svg?branch=master)](https://travis-ci.org/xgfone/go-config) [![GoDoc](https://godoc.org/github.com/xgfone/go-config?status.svg)](http://godoc.org/github.com/xgfone/go-config) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square)](https://raw.githubusercontent.com/xgfone/go-config/master/LICENSE)
An extensible go configuration. The default parsers can parse the CLI arguments and the ini file. You can implement and register your parser, and the configuration engine will call the parser to parse the configuration.

The inspiration is from [oslo.config](https://github.com/openstack/oslo.config), which is a `OpenStack` library for config.

The current version is `v9`. See [DOC](https://godoc.org/github.com/xgfone/go-config).

The supported Go version: `1.x`.


## Principle of Work

1. Create a `Config` engine.
2. (Optional) Add the CLI and non-CLI parsers into the `Config`.
3. Register the common options into `Config`.
3. Call the method `Parse()` to parse the options.
    1. Start to parse the configuration.
    2. Call the CLI parser if exists, and the CLI parser parses the CLI arguments.
    3. Call each other parsers according to the order that they are registered.
        1. Call the method `Parse()` of the parser.
        2. The parser parses the options and sets values.
    4. Check whether some required options have neither the parsed value nor the default value.

**Notice:** when setting the parsed value, it will calling the validators to validate it if setting the validators for the option.


## Parser

In order to deveplop a new parser, you just need to implement the interface `Parser`. But `Config` distinguishes the CLI parser and the common parser, which have the same interface `Parser`. But `Config` must have no more than one CLI parser set by `ResetCLIParser()` and maybe have many common parsers added by `AddParser()`. See the example above.


## Usage
```go
package main

import (
    "flag"
    "fmt"
    "os"

    config "github.com/xgfone/go-config"
)

func main() {
    cliParser := config.NewDefaultFlagCliParser(true)
    iniParser := config.NewSimpleIniParser("config-file")
    conf := config.NewConfig(cliParser).AddParser(iniParser)

    ipOpt := config.StrOpt("", "ip", "", "the ip address").SetValidators(NewIPValidator())
    conf.RegisterCliOpt("", ipOpt)
    conf.RegisterCliOpt("", config.IntOpt("", "port", 80, "the port"))
    conf.RegisterCliOpt("", config.StrOpt("", "config-file", "", "The path of the ini config file."))
    conf.RegisterCliOpt("redis", config.StrOpt("", "conn", "redis://127.0.0.1:6379/0", "the redis connection url"))
    conf.SetAddVersion("1.0.0") // Print the version and exit when giving the CLI option version.

    if err := conf.Parse(); err != nil {
        conf.Audit() // View the internal information.
        fmt.Println(err)
        return
    }

    fmt.Println(conf.String("ip"))
    fmt.Println(conf.Int("port"))
    fmt.Println(conf.Group("redis").String("conn"))
    fmt.Println(conf.Args)

    // Execute:
    //     PROGRAM -ip 0.0.0.0 aa bb cc
    //
    // Output:
    //     0.0.0.0
    //     80
    //     [aa bb cc]
}
```

You can also create a new `Config` by the `NewDefault()`, which will use `NewDefaultFlagCliParser(true)` as the CLI parser, add the ini parser `NewSimpleIniParser()` and register the CLI option `config-file`, which you change it by modifying the value of the variable `IniParserOptName`. Notice: `NewDefault()` does not add the environment variable parser, and you need to add it by hand, such as `NewDefault().AddParser(NewEnvVarParser(""))`.

The package has created a global default `Config` created by `NewDefault()` like doing above, which is `Conf`. You can use it, like the global variable `CONF` in `oslo.config`. For example,
```go
package main

import (
    "fmt"

    config "github.com/xgfone/go-config"
)

var opts = []config.Opt{
    config.Str("ip", "", "the ip address").AddValidators(NewIPValidator()),
    config.Int("port", 80, "the port").AddValidators(NewPortValidator()),
}

func main() {
    config.Conf.RegisterCliOpts("", opts)
    config.Conf.Parse("-ip", "0.0.0.0") // You can pass nil

    fmt.Println(config.Conf.String("ip")) // Output: 0.0.0.0
    fmt.Println(config.Conf.Int("port"))  // Output: 80
}
```

You also register a struct then use it.
```go
package main

import (
    "fmt"
    config "github.com/xgfone/go-config"
)

func main() {
    type Address struct {
		Address []string `default:""`
    }

    type S struct {
        Name    string  `name:"name" cli:"1" default:"Aaron" help:"The user name"`
        Age     int8    `cli:"t" default:"123"`
        Addr1   Address `group:"group" cli:"true"`
        Addr2   Address `cli:"true"`
        Addr3   Address
        Ignore  string  `name:"-"`
    }

    args := []string{"--age", "18", "--group-address", "abc,def", "--addr2-address", "xyz"}

    s := S{}
    config.Conf.RegisterStruct("", &s)
    if err := config.Conf.Parse(args...); err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("Name: %s\n", s.Name)
    fmt.Printf("Age: %d\n", s.Age)
    fmt.Printf("Address: %v\n", s.Addr1.Address)
    fmt.Printf("Address: %v\n", s.Addr2.Address)
    fmt.Printf("Address: %v\n", s.Addr3.Address)
    // Output:
    // Name: Aaron
    // Age: 18
    // Address: [abc def]
    // Address: [xyz]
    // Address: []

    // Or
    fmt.Printf("Name: %s\n", config.Conf.String("name"))
    fmt.Printf("Age: %d\n", config.Conf.Int8("age"))
    fmt.Printf("Address: %v\n", config.Conf.Group("group").Strings("address"))
    fmt.Printf("Address: %v\n", config.Conf.Group("addr2").Strings("address"))
    fmt.Printf("Address: %v\n", config.Conf.Group("addr3").Strings("address"))
    // Output:
    // Name: Aaron
    // Age: 18
    // Address: [abc def]
    // Address: [xyz]
    // Address: []
}
```
