# configmanager
An extensible go configuration manager. The default parsers can parse the CLI arguments and the property file. You can implement and register your parser, and the manager engine will call the parser to parse the config.

### Notice
At present, the ConfigManager does not support the section like [ini](https://github.com/go-ini/ini) or the group in [oslo.log](https://github.com/openstack/oslo.log) developed by OpenStack. The function will be added later.
