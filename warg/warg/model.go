package warg

import (
	"encoding/json"
	"fmt"
)

var errTypeMismatch = fmt.Errorf("type mismatch")

type FlagFactory func() *FlagDef
type FlagOpt func(*FlagDef)

type Setter interface {
	Set(any) error
}

type flagSetter[T any] struct {
	ptr *T
}

func PointsTo[T any](ptr *T) flagSetter[T] {
	return flagSetter[T]{
		ptr: ptr,
	}
}

func (s flagSetter[T]) Set(val any) error {
	if v, ok := val.(T); ok {
		*s.ptr = v
		return nil
	}
	return errTypeMismatch
}

type fieldSetter[T any] struct {
	acc func() *T
}

func TakesFrom[T any](acc func() *T) fieldSetter[T] {
	return fieldSetter[T]{
		acc: acc,
	}
}

func (s fieldSetter[T]) Set(val any) error {
	if v, ok := val.(T); ok {
		*s.acc() = v
		return nil
	}
	return errTypeMismatch
}

type CommandFunc func(args any) error

type FlagDef struct {
	Names       []string    `json:"names"`
	Description string      `json:"description"`
	Children    []*FlagDef  `json:"children"`
	Command     CommandFunc `json:"-"`
	IsSet       bool        `json:"-"`
	Setter      Setter      `json:"-"`

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

	flags := Define("Some general help",
		Flag([]string{"-b", "--bool"}, "this is a bool", PointsTo(&t.TestBool)),
		Flag([]string{"-i", "--integer"}, "this is an integer", PointsTo(&t.TestInt)),
		Context([]string{"-o", "--object"}, "this is an object", PointsTo(&t.TestObj),
			Flag([]string{"-s", "--string"}, "this is a string", TakesFrom(func() *T {}))), // TODO: THis bitch
	)

	j, _ := json.MarshalIndent(flags, "", "  ")
	fmt.Printf("%v\n", string(j))
}

func Define(description string, defs ...FlagFactory) error {
	Context([]string{}, description, PointsTo(&struct{}{}), defs...)
	return nil
}

func Flag(names []string, description string, setter Setter, opts ...FlagOpt) FlagFactory {
	return func() *FlagDef {
		f := &FlagDef{
			Names:       names,
			Description: description,
			Children:    []*FlagDef{},
			Command:     nil,
			Setter:      setter,
		}
		for _, opt := range opts {
			opt(f)
		}
		return f
	}
}

func Context(names []string, description string, setter Setter, children ...FlagFactory) FlagFactory {
	return Flag(
		names,
		description,
		setter,
		WithSubFlags(children...))
}

func SubCommand(names []string, description string, setter Setter, cmd CommandFunc, children ...FlagFactory) FlagFactory {
	return Flag(
		names,
		description,
		setter,
		WithCommand(cmd),
		WithSubFlags(children...),
	)
}

func WithSubFlags(flags ...FlagFactory) FlagOpt {
	return func(f *FlagDef) {
		childDefs := []*FlagDef{}
		for _, def := range flags {
			childDefs = append(childDefs, def())
		}
		f.Children = childDefs
	}
}

func WithCommand(cmd CommandFunc) FlagOpt {
	return func(f *FlagDef) {
		f.Command = cmd
	}
}
