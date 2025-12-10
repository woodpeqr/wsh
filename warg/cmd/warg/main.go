package main

import (
	"fmt"
	"os"
)

type CliFlags struct {
	Add struct {
		Names   []string `warg:"-n,--names; names of the flag"`
		Desc    string   `warg:"-d,--description; description of the flag"`
		IsValue bool     `warg:"-v,--value; does the flag require a value"`
	} `warg:"-A,--add; add a new flag"`
}

type Flag struct {
	Names []string
	Help  string
}

type WFlag interface {
	Flag

	Set(s string) error
	Get() string
	IsSet() bool
}

type LibFlag[T any] struct {
	Flag	
	valuePtr *T
}

func (f *LibFlag[T]) Set(s string) error {
	panic("not implemented")
}

func (f *LibFlag[T]) Get() string {
	panic("not implemented")
}

func (f *LibFlag[T]) IsSet() bool {
	return f.Get() != ""
}

type CfgFlag struct {
	Flag
	value string
	isSwitch bool
}


func main() {
	fmt.Println(os.Args)
}
