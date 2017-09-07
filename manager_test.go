package config

import "fmt"

func ExampleConfig() {
	cliOpts1 := []Opt{
		StrOpt("", "required", nil, "required"),
		IntOpt("", "int1", nil, "required int"),
		BoolOpt("", "no", nil, "test bool option"),
	}

	cliOpts2 := []Opt{
		IntOpt("", "int2", 789, "optional int"),
		BoolOpt("", "yes", nil, "test bool option"),
		StrOpt("", "optional", "optional", "optional"),
	}

	opts := []Opt{
		StrOpt("", "test1", "test1", "test2"),
	}

	Conf.RegisterCliOpts("", cliOpts1)
	Conf.RegisterCliOpts("cli", cliOpts2)
	Conf.RegisterOpts("group", opts)

	args := []string{"-cli_yes"}
	// args = nil // You can pass nil to get the arguments from the command line.
	if err := Conf.Parse(args); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Conf.StringD("required", "abc"))
	fmt.Println(Conf.IntD("int1", 123))
	fmt.Println(Conf.BoolD("no", true))

	fmt.Println(Conf.Group("cli").String("optional"))
	fmt.Println(Conf.Group("cli").Int("int2"))
	fmt.Println(Conf.Group("cli").Bool("yes"))

	fmt.Println(Conf.Group("group").String("test1"))

	// Output:
	// abc
	// 123
	// true
	// optional
	// 789
	// true
	// test1
}
