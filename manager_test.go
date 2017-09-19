package config

import "fmt"

func ExampleConfig() {
	validators := []Validator{NewStrLenValidator(1, 10)}
	cliOpts1 := []Opt{
		StrOpt("", "required", "", "required").SetValidators(validators),
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

	// We don't ask that all the options must have a value or the default value.
	// Conf.IsRequired = false

	args := []string{"-cli_no=0", "-required", "required"}
	// args = nil // You can pass nil.
	if err := Conf.Parse(args); err != nil {
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
		Address string
	}

	type S struct {
		Name    string  `name:"name" cli:"1" default:"Aaron"`
		Age     int8    `cli:"t" default:"123"`
		Address Address `group:"group" cli:"true"`
		Ignore  string  `name:"-"`
	}

	s := S{}
	Conf.RegisterStruct("", &s)
	if err := Conf.Parse([]string{"-age", "18", "-group_address", "China"}); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Name: %s\n", s.Name)
	fmt.Printf("Age: %d\n", s.Age)
	fmt.Printf("Address: %s\n", s.Address.Address)

	// Output:
	// Name: Aaron
	// Age: 18
	// Address: China
}
