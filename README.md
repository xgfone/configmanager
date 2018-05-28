# go-config
An extensible go configuration. The default parsers can parse the CLI arguments and the ini file. You can implement and register your parser, and the configuration engine will call the parser to parse the configuration.

The inspiration is from [oslo.config](https://github.com/openstack/oslo.config), which is a `OpenStack` library for config.

The current version is `v3`, which has changed the interface of the configuration manager and the parser plugin, but the interfaces that the user uses have not been changed. They should be better friendly. See [DOC](https://godoc.org/github.com/xgfone/go-config).

The biggest difference between `v2` to `v1` is to remove the method `GetKeys()` from the parser interface `Parser`. But you can use the branch [v1](https://github.com/xgfone/go-config/tree/v1). For the user, there is no effect. So you have no need to modify any code, and just the parser plugin should be modified.

## Principle of Work

1. Create a `Config` engine.
2. (Optional) Add the parsers into the `Config`.
3. Register the CLI or common options into `Config`.
3. Call the method `Parse()` to parse the options.
    1. Start to parse the configuration.
    2. Call the CLI parser with the CLI arguments to parse.
    3. The CLI parser parses CLI options and sets values and the rest arguments.
    4. Call each other parsers according to the order that they are registered.
        1. Call the method `Parse()` of the parser.
        2. The parser parses the options and sets values.
    5. Check whether some required options have neither the parsed value nor the default value.

**Notice:** when setting the parsed value, it will calling the validators to validate it if setting the validators for the option.


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
    conf.SetAddVersion("1.0.0") // Print the version and exit when giving the CLI option version.

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

var opts = []config.Opt{
    config.Str("ip", "", "the ip address").AddValidators(NewIPValidator()),
    config.Int("port", 80, "the port").AddValidators(NewPortValidator()),
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
