package warg

import (
	"encoding/json"
	"fmt"
)

type FlagFactory[T any] func() FlagDef[T]
type rootFlag struct{}

type Setter[T any] interface {
	Set(T)
}

type flagSetter[T any] struct {
	ptr *T
}

func PointsTo[T any](ptr *T) flagSetter[T] {
	return flagSetter[T]{
		ptr: ptr,
	}
}

func (s flagSetter[T]) Set(val T) {
	*s.ptr = val
}

type fieldSetter[T any] struct {
	acc func() *T
}

func TakesFrom[T any](acc func() *T) fieldSetter[T] {
	return fieldSetter[T]{
		acc: acc,
	}
}

func (s fieldSetter[T]) Set(val T) {
	*s.acc() = val
}

type CommandFunc func(args any) error

type FlagDef[T any] struct {
	Names       []string       `json:"names"`
	Description string         `json:"description"`
	Children    []FlagDef[any] `json:"children"`
	Command     CommandFunc    `json:"-"`
	IsSet       bool           `json:"-"`
	Setter      Setter[T]      `json:"-"`

	isRoot bool
}

type TestStruct struct {
	TestBool bool
	TestInt  int
	TestObj  []TestObject
}

type TestObject struct {
	TestString string
}

func T() {
	var t TestStruct

	flags := NewParser("Some general help",
		WithFlag([]string{"-b", "--bool"}, "this is a bool", PointsTo(&t.TestBool)),
		WithFlag([]string{"-i", "--integer"}, "this is an integer", PointsTo(&t.TestInt)),
		WithContext([]string{"-o", "--object"}, "this is an object", PointsTo(&t.TestObj)),
	)

	j, _ := json.MarshalIndent(flags, "", "  ")
	fmt.Printf("%v\n", string(j))
}

func NewParser(description string, defs ...FlagFactory[any]) error {
	WithContext([]string{}, description, PointsTo(&struct{}{}), defs...)
	return nil
}

func WithFlag[T any](names []string, description string, setter flagSetter[T]) FlagFactory[T] {
	return func() FlagDef[T] {
		return FlagDef[T]{
			Names:       names,
			Description: description,
			Children:    []FlagDef[any]{},
			Command:     nil,
			Setter:      setter,
		}
	}
}

func WithContext[T any](names []string, description string, setter Setter[T], children ...FlagFactory[any]) FlagFactory[T] {
	return func() FlagDef[T] {
		childDefs := []FlagDef[any]{}
		for _, def := range children {
			childDefs = append(childDefs, def())
		}

		return FlagDef[T]{
			Names:       names,
			Description: description,
			Children:    childDefs,
			Command:     nil,
			Setter:      setter,
		}
	}
}

func WithSubCommand[T any](names []string, description string, setter Setter[T], cmd CommandFunc, children ...FlagFactory[any]) FlagFactory[T] {
	return func() FlagDef[T] {
		childDefs := []FlagDef[any]{}
		for _, def := range children {
			childDefs = append(childDefs, def())
		}

		return FlagDef[T]{
			Names:       names,
			Description: description,
			Children:    childDefs,
			Command:     cmd,
			Setter:      setter,
		}
	}
}

func WithField[T any](names []string, description string, setter Setter[T]) FlagFactory[T] {
	return func() FlagDef[T] {
		return FlagDef[T]{
			Names:       names,
			Description: description,
			Children:    []FlagDef[any]{},
			Command:     nil,
			Setter:      setter,
		}
	}
}
