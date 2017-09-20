# go-config
An extensible go configuration. The default parsers can parse the CLI arguments and the ini file. You can implement and register your parser, and the configuration engine will call the parser to parse the configuration.

The inspiration is from [oslo.config](https://github.com/openstack/oslo.config), which is a `OpenStack` library for config.

**NOTICE: The API has been stable.**

## Principle of Work

1. Create a `Config` engine.
2. Register the CLI or common options into `Config`.
3. Call the method `Parse()` to parse the options.
    1. Start to parse the configuration.
    2. Get all the registered CLI options.
    3. Call the CLI parser with the CLI optons and arguments to parse.
    4. Get all the registered common options.
    5. Call each other parsers according to the order that they are registered.
        1. Call the method `GetKeys()` of the parser to get the keys of all the configurations that the parser needs.
        2. Get the values of the keys above from the default group that has been parsed.
        3. Call the method `Parse()` of the parser with the registered options and the configurations, and get the parsed result.
        4. Merge the parsed result together. Notice: before merging, it will call the validators of the option to validate the value of the option, if have.
    6. Check whether some required options have neither the parsed value nor the default value.


## Parser

In order to deveplop a new CLI parser, you just need to implement the interface `CliParser`. In one `Config`, there is only one CLI parser. But it can have more than one other parsers, and you just need to implement the interface `Parser`, then add it into `Config` by the method `AddParser()`. See the example above. See [DOC](https://godoc.org/github.com/xgfone/go-config).


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
    cliParser := config.NewFlagCliParser("", flag.ExitOnError)
    iniParser := config.NewSimpleIniParser("config-file")
    conf := config.NewConfig(cliParser).AddParser(iniParser)

    validators := []Validator{NewIPValidator()}
    ipOpt := config.StrOpt("", "ip", "", "the ip address").SetValidators(validators)
    conf.RegisterCliOpt("", ipOpt)
    conf.RegisterCliOpt("", config.IntOpt("", "port", 80, "the port"))
    conf.RegisterCliOpt("", config.StrOpt("", "config-file", "", "The path of the ini config file."))
    conf.RegisterCliOpt("redis", config.StrOpt("", "conn", "redis://127.0.0.1:6379/0", "the redis connection url"))

    if err := conf.Parse(nil); err != nil {
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

You can also create a new `Config` by the `NewDefault()`, which will use `NewFlagCliParser()` as the CLI parser, add the ini parser `NewSimpleIniParser()` and register the CLI option `config-file`, which you change it by modifying the value of the variable `IniParserOptName`. Notice: `NewDefault()` does not add the environment variable parser, and you need to add it by hand, such as `NewDefault().AddParser(NewEnvVarParser(""))`.

The package has created a global default `Config` created by `NewDefault()` like doing above, which is `Conf`. You can use it, like the global variable `CONF` in `oslo.config`. For example,
```go
package main

import (
    "fmt"

    config "github.com/xgfone/go-config"
)

var ipValidators = []Validator{NewIPValidator()}
var portValidators = []Validator{NewPortValidator()}

var opts = []config.Opt{
    config.Str("ip", "", "the ip address").SetValidators(ipValidators),
    config.Int("port", 80, "the port").SetValidators(portValidators),
}

func main() {
    config.Conf.RegisterCliOpts("", opts)
    config.Conf.Parse([]string{"-ip", "0.0.0.0"}) // You can pass nil

    fmt.Println(config.Conf.String("ip")) // Output: 0.0.0.0
    fmt.Println(config.Conf.Int("port"))  // Output: 80
}
```

You also register a struct then use it.
```go
func main() {
    type S struct {
        Name    string `name:"name" cli:"1" default:"Aaron" help:"The user name"`
        Age     int8   `cli:"t" default:"123"`
        Address string `cli:"true"`
        Ignore  string `name:"-"`
    }

    s := S{}
    Conf.RegisterStruct("", &s)
    if err := Conf.Parse([]string{"-age", "18", "-address", "China"}); err != nil {
        fmt.Println(err)
        return
    }

    fmt.Printf("Name: %s\n", s.Name)
    fmt.Printf("Age: %d\n", s.Age)
    fmt.Printf("Address: %s\n", s.Address)

    // Or
    fmt.Printf("Name: %s\n", Conf.String("name"))
    fmt.Printf("Age: %d\n", Conf.Int8("age"))
    fmt.Printf("Address: %s\n", Conf.String("Address"))
}
```
