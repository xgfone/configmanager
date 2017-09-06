package configmanager

// StrOpt is a string option
type StrOpt struct {
	Name     string
	Help     string
	Short    string
	Required bool
	Default  interface{}
}

var _ Opt = StrOpt{}

// NewStrOpt return a new string option.
//
// Notice: the type of the default value must be string or nil.
// If no default, it's nil.
func NewStrOpt(short, name string, _default interface{}, required bool, help string) StrOpt {
	if _default != nil {
		_default = _default.(string)
	}

	return StrOpt{
		Short:    short,
		Name:     name,
		Help:     help,
		Required: required,
		Default:  _default,
	}
}

// GetName returns the name of the option.
func (o StrOpt) GetName() string {
	return o.Name
}

// GetShort returns the shorthand name of the option.
func (o StrOpt) GetShort() string {
	return o.Short
}

// IsRequired returns ture if the option must have the value, or false.
func (o StrOpt) IsRequired() bool {
	return o.Required
}

// GetDefault returns the default value of the option.
func (o StrOpt) GetDefault() interface{} {
	if o.Default != nil {
		return o.Default.(string)
	}
	return nil
}

// GetHelp returns the help doc of the option.
func (o StrOpt) GetHelp() string {
	return o.Help
}

// Parse parses the value of the option to string.
func (o StrOpt) Parse(data string) (interface{}, error) {
	return ToString(data)
}

// IntOpt is a int option
type IntOpt struct {
	Name     string
	Help     string
	Short    string
	Required bool
	Default  interface{}
}

var _ Opt = IntOpt{}

// NewIntOpt return a new int option.
//
// Notice: the type of the default value must be int or nil.
// If no default, it's nil.
func NewIntOpt(short, name string, _default interface{}, required bool, help string) IntOpt {
	if _default != nil {
		_default = _default.(int)
	}

	return IntOpt{
		Short:    short,
		Name:     name,
		Help:     help,
		Required: required,
		Default:  _default,
	}
}

// GetName returns the name of the option.
func (o IntOpt) GetName() string {
	return o.Name
}

// GetShort returns the shorthand name of the option.
func (o IntOpt) GetShort() string {
	return o.Short
}

// IsRequired returns ture if the option must have the value, or false.
func (o IntOpt) IsRequired() bool {
	return o.Required
}

// GetDefault returns the default value of the option.
func (o IntOpt) GetDefault() interface{} {
	if o.Default != nil {
		return o.Default.(int)
	}
	return nil
}

// GetHelp returns the help doc of the option.
func (o IntOpt) GetHelp() string {
	return o.Help
}

// Parse parses the value of the option to int.
func (o IntOpt) Parse(data string) (interface{}, error) {
	v, err := ToInt64(data)
	if err != nil {
		return nil, err
	}
	return int(v), nil
}
