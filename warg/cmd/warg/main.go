package main

import (
	"V-Woodpecker-V/wsh/warg/flags"
	"os"
)

func main() {
	addFlag := &flags.WFlag{
		Short: "A",
		Long:  "add",
		Help:  "add a new flag",
	}
	addFlag.Children = []*flags.WFlag{
		{
			Short:         "s",
			Long:          "short",
			Help:          "short version of a flag",
			Parent:        addFlag,
			ValueRequired: true,
		},
		{
			Short:         "l",
			Long:          "long",
			Help:          "long version of a flag",
			Parent:        addFlag,
			ValueRequired: true,
		},
		{
			Short:         "h",
			Long:          "help",
			Help:          "help message of a flag",
			Parent:        addFlag,
			ValueRequired: true,
		},
		{
			Short:                 "p",
			Long:                  "parent",
			Help:                  "which flag to put it under",
			Parent:                addFlag,
			NonEmptyValueRequired: true,
		},
		{
			Short:  "v",
			Long:   "value",
			Help:   "this flag requires a value",
			Parent: addFlag,
		},
		{
			Short:  "V",
			Long:   "non_empty_value",
			Help:   "this flag requires a value that is not empty",
			Parent: addFlag,
		},
	}
	flags.AddFlag(addFlag)
	flags.ParseArgs(os.Args[1:])
	flags.DebugPrintFlags()
}
