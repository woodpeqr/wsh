package warg

type FlagFactory[Ctx any, Val any] func(
	ptr *Ctx, 
	names []string,
	description string,
	accessor func(*Ctx) *Val,
	cmd CommandFunc[Val],
	children ...FlagFactory[any, any],
) *WargFlag[Val]
type CommandFunc[T any] func(args *WargFlag[T]) error

type WargFlag[T any] struct {
	Names       []string
	Description string
	Pointer     *T
	Children    []*WargFlag[any]
	Command     CommandFunc[WargFlag[T]]
	IsSet       bool

	isRoot bool
}

type TestStruct struct {
	testBool bool
	testInt  int
	testObj  []struct {
		testString string
	}
}

func T() {
	var t TestStruct

	Define(
		Flag(&t.testBool, []string{"-b", "--bool"}),
		Flag(&t.testInt, []string{"-i", "--int"}),
		Context(&t.testObj, []string{"-O", "--obj", 
			Flag(&?, []string{"-s", "--string"})}) // how to reference specific item in a slice???
		)
}

func Define(defs ...any) error {
	root := &WargFlag[struct{}]{isRoot: true}
	for _, d := range defs {
		root.Children = append(root.Children, d())
	}
	return nil
}
