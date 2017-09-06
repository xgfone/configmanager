package configmanager

import "fmt"

func ExampleConfigManager() {
	Conf.RegisterCliOpt(NewStrOpt("", "required", nil, false, "required"))
	Conf.RegisterCliOpt(NewStrOpt("", "optional", "optional", false, "optional"))
	Conf.RegisterCliOpt(NewIntOpt("", "int1", nil, false, "required int"))
	Conf.RegisterCliOpt(NewIntOpt("", "int2", 789, false, "optional int"))

	if err := Conf.Parse(nil); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Conf.StringD("required", "abc"))
	fmt.Println(Conf.String("optional"))
	fmt.Println(Conf.IntD("int1", 123))
	fmt.Println(Conf.Int("int2"))

	// Output:
	// abc
	// optional
	// 123
	// 789
}
