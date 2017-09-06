# configmanager
An extensible go configuration manager. The default parsers can parse the CLI arguments and the property file. You can implement and register your parser, and the manager engine will call the parser to parse the config.

## Principle of Work

1. Start to parse the configuration.
2. Register the CLI options into all the parsers.
3. Register the other options into the parsers except the CLI parser.
4. The manager calls the CLI parser to parse the CLI arguments.
5. The manager calls each other parsers according to the order which are registered.
    1. Call `GetKeys()` to get the keys of all the options that the parser needs.
    2. Get the option values by the keys above from the values that has been parsed.
    3. Call the method `Parse` of the parser with those option values, and get the parsed result.
    4. Merge the parsed result together.
6. Check whether some required options neither have the value nor the default value.

## Usage
```go
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xgfone/configmanager"
)

func main() {
	cliParser := configmanager.NewFlagCliParser(filepath.Base(os.Args[0]), flag.ExitOnError)
	propertyParser := configmanager.NewSimplePropertyParser("config_file")
	conf := configmanager.NewConfigManager(cliParser).AddParser(propertyParser)

	conf.RegisterCliOpt(configmanager.NewStrOpt("", "ip", nil, true, "the ip address"))
	conf.RegisterCliOpt(configmanager.NewIntOpt("", "port", 80, false, "the port"))
	conf.RegisterOpt(configmanager.NewStrOpt("", "redis", "redis://127.0.0.1:6379/0",
		false, "the redis connection url"))
	conf.RegisterCliOpt(configmanager.NewStrOpt("", "config_file", nil, false,
		"The path of the config file."))

	if err := conf.Parse(nil); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(conf.String("ip"))
	fmt.Println(conf.Int("port"))
	fmt.Println(conf.String("redis"))
	fmt.Println(conf.Args)
}
```

You can also create a new `ConfigManager` by the `NewDefault()`, which will use `NewFlagCliParser()` as the CLI parser, add the property parser `NewSimplePropertyParser()` and register the CLI option `config_file`.

The package has created a global default `ConfigManager` by `NewDefault()` like doing above, which is `Conf`. You can use it, like the global variable `CONF` in `oslo.config`.

## Parser

In order to deveplop a CLI parser, you just need to implement the interface `CliParser`. In a `ConfigManager`, there is only one CLI parser. But it can have more than one other parsers, and you just need to implement the interface `Parser`, then add it into `ConfigManager` by the method `AddParser()`. See the example above. See [doc](https://godoc.org/github.com/xgfone/configmanager).

## Notice
At present, the ConfigManager does not support the section like [ini](https://github.com/go-ini/ini) or the group in [oslo.config](https://github.com/openstack/oslo.config) developed by OpenStack. The function will be added later.
