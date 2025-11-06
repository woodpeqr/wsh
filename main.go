package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	cmdFlag := flag.String("c", "", "Execute command and exit")
	flag.Parse()

	shell, err := NewShell()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wsh: %v\n", err)
		os.Exit(1)
	}

	exitCode := shell.Run(*cmdFlag, flag.Args())
	os.Exit(exitCode)
}
