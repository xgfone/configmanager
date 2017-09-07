# go-config
An extensible go configuration. The default parsers can parse the CLI arguments and the ini file. You can implement and register your parser, and the configuration engine will call the parser to parse the configuration.


## Principle of Work

1. Start to parse the configuration.
2. Register the CLI options into all the parsers.
3. Register the other options into the parsers except the CLI parser.
4. The manager calls the CLI parser to parse the CLI arguments.
5. The manager calls each other parsers according to the order which are registered.
    1. Call the method `GetKeys()` of the parser to get the keys of all the options that the parser needs.
    2. Get the option values by the keys above from the values that has been parsed.
    3. Call the method `Parse()` of the parser with those option values, and get the parsed result.
    4. Merge the parsed result together.
6. Check whether some required options have neither the parsed value nor the default value.


## Parser

In order to deveplop a new CLI parser, you just need to implement the interface `CliParser`. In one `Config`, there is only one CLI parser. But it can have more than one other parsers, and you just need to implement the interface `Parser`, then add it into `Config` by the method `AddParser()`. See the example above. See [doc](https://godoc.org/github.com/xgfone/go-config).


## Usage
```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	config "github.com/xgfone/go-config"
)

func main() {
	cliParser := config.NewFlagCliParser(filepath.Base(os.Args[0]), flag.ExitOnError)
	iniParser := config.NewSimpleIniParser("config-file")
	conf := config.NewConfig(cliParser).AddParser(iniParser)

	conf.RegisterOpt("", true, config.StrOpt("", "ip", nil, true, "the ip address"))
	conf.RegisterOpt("", true, config.IntOpt("", "port", 80, false, "the port"))
	conf.RegisterOpt("", true, config.StrOpt("", "config-file", nil, false,
		"The path of the ini config file."))
	conf.RegisterOpt("redis", true, config.StrOpt("", "conn", "redis://127.0.0.1:6379/0",
		false, "the redis connection url"))

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

You can also create a new `Config` by the `NewDefault()`, which will use `NewFlagCliParser()` as the CLI parser, add the ini parser `NewSimpleIniParser()` and register the CLI option `config-file`.

The package has created a global default `Config` by `NewDefault()` like doing above, which is `Conf`. You can use it, like the global variable `CONF` in `oslo.config`. For example,
```go
package main

import (
	"fmt"

	config "github.com/xgfone/go-config"
)

var opts = []config.Opt{
	config.StrOpt("", "ip", nil, true, "the ip address"),
	config.IntOpt("", "port", 80, true, "the port"),
}

func main() {
	config.Conf.RegisterOpts("", true, opts)
	config.Conf.Parse([]string{"-ip", "0.0.0.0"}) // You can pass nil

	fmt.Println(config.Conf.String("ip")) // Output: 0.0.0.0
	fmt.Println(config.Conf.Int("port"))  // Output: 80
}
```
