package warg

type FlagFactory[T any] func(ptr *T) *WargFlag[T]
type FlagConstructor[T any] func(ptr *T, names []string) *WargFlag[any]

type WargFlag[T any] struct {
	Names       []string
	Description string
	Value       any
	Children    []*WargFlag[any]
}

type TestStruct struct {
	testBool bool
	testInt  int
	testObj  []struct {
		testString string
	}
}

func T() {
	var testArgs TestStruct
	New(&testArgs, func (t *TestStruct) *WargFlag[TestStruct]{
		Flag(&t.testBool, []string{"b"})
		Flag(&t.testInt, []string{"i"})
	})
}

func New[T any](ptr *T, flags FlagFactory[T]) []*WargFlag[T] {
	return []*WargFlag[T]{}
}

func Flag[T any](ptr *T, names []string) FlagFactory[T] {
	return func(ctx *T) *WargFlag[T] {
		return &WargFlag[T]{
			Names:       names,
			Description: "a simple flag", //TODO: Fix
			Value:       ptr,
			Children:    make([]*WargFlag[any], 0),
		}
	}
}

func Context(ptr *bool, names []string, children FlagFactory[any]) FlagFactory[bool] {
	return func(ctx *bool) *WargFlag[bool] {
		return &WargFlag[bool]{
			Names:       names,
			Description: "a context flag",
			Value:       ptr,
		}
	}
}
