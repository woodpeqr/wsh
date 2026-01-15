package main

import (
	"fmt"
	"os"
	"time"
	
	"V-Woodpecker-V/wsh/warg/lib"
)

func main() {
	// Declare variables
	var verbose bool
	var name string
	var count int
	var tags []string
	var ports []int
	var timeout time.Duration
	
	// Create parser with all flags
	parser := lib.New().
		Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
		Flag(&name, []string{"n", "name"}, "User name").
		Flag(&count, []string{"c", "count"}, "Count value").
		Flag(&tags, []string{"t", "tag"}, "Tags (repeatable)").
		Flag(&ports, []string{"p", "port"}, "Ports (repeatable)").
		Flag(&timeout, []string{"timeout"}, "Operation timeout")
	
	// Parse arguments
	result := parser.Parse(os.Args[1:])
	
	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
		os.Exit(1)
	}
	
	// Display results
	fmt.Println("=== Parsed Values ===")
	fmt.Printf("verbose: %v\n", verbose)
	fmt.Printf("name: %s\n", name)
	fmt.Printf("count: %d\n", count)
	fmt.Printf("tags: %v\n", tags)
	fmt.Printf("ports: %v\n", ports)
	fmt.Printf("timeout: %v\n", timeout)
}
