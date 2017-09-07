package config

import "fmt"

func ExampleConfig() {
	Conf.RegisterCliOpt(NewStrOpt("", "required", nil, false, "required"))
	Conf.RegisterCliOpt(NewStrOpt("", "optional", "optional", false, "optional"))
	Conf.RegisterCliOpt(NewIntOpt("", "int1", nil, false, "required int"))
	Conf.RegisterCliOpt(NewIntOpt("", "int2", 789, false, "optional int"))
	Conf.RegisterCliOpt(NewBoolOpt("", "yes", nil, false, "test bool option"))
	Conf.RegisterCliOpt(NewBoolOpt("", "no", nil, false, "test bool option"))

	args := []string{"-yes"}
	if err := Conf.Parse(args); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Conf.BoolD("yes", true))
	fmt.Println(Conf.BoolD("no", false))
	fmt.Println(Conf.StringD("required", "abc"))
	fmt.Println(Conf.String("optional"))
	fmt.Println(Conf.IntD("int1", 123))
	fmt.Println(Conf.Int("int2"))

	// Output:
	// true
	// false
	// abc
	// optional
	// 123
	// 789
}
