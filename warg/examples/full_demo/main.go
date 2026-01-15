package main

import (
	"fmt"
	"os"
	"time"
	
	"V-Woodpecker-V/wsh/warg/lib"
)

func main() {
	// Example 1: Basic flags with type inference
	basicExample()
	
	// Example 2: Slice support (repeatable flags)
	sliceExample()
	
	// Example 3: Hierarchical contexts
	contextExample()
	
	// Example 4: Custom types
	customTypeExample()
}

func basicExample() {
	fmt.Println("=== Basic Example ===")
	
	// Declare variables with their types
	var verbose bool
	var name string
	var count int
	
	// Create parser - types are automatically inferred!
	// No need to call Bool(), String(), Int() - just Flag()
	parser := lib.New().
		Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
		Flag(&name, []string{"n", "name"}, "User name").
		Flag(&count, []string{"c", "count"}, "Count value")
	
	// Parse arguments
	result := parser.Parse(os.Args[1:])
	
	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
		os.Exit(1)
	}
	
	// Use the values - they're already set!
	fmt.Printf("verbose: %v\n", verbose)
	fmt.Printf("name: %s\n", name)
	fmt.Printf("count: %d\n", count)
	fmt.Println()
}

func sliceExample() {
	fmt.Println("=== Slice Example (Repeatable Flags) ===")
	
	// Declare slice variables
	var tags []string
	var ports []int
	
	// Slices are detected automatically!
	parser := lib.New().
		Flag(&tags, []string{"t", "tag"}, "Add tag (repeatable)").
		Flag(&ports, []string{"p", "port"}, "Add port (repeatable)")
	
	// Supports multiple ways:
	// --tag go --tag cli --tag parser
	// --tag go,cli,parser (comma-separated for strings)
	// --port 8080 --port 3000
	
	result := parser.Parse(os.Args[1:])
	
	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
		os.Exit(1)
	}
	
	fmt.Printf("tags: %v\n", tags)
	fmt.Printf("ports: %v\n", ports)
	fmt.Println()
}

func contextExample() {
	fmt.Println("=== Hierarchical Context Example ===")
	
	type GitConfig struct {
		Commit  bool
		Message string
		Push    bool
	}
	
	type Config struct {
		Verbose bool
		Git     GitConfig
	}
	
	var config Config
	
	// Functional composition with contexts
	parser := lib.New().
		Flag(&config.Verbose, []string{"v", "verbose"}, "Enable verbose").
		Context(&config.Git, []string{"G", "git"}, "Git operations", func(p *lib.Parser, git *GitConfig) *lib.Parser {
			// Build sub-parser for git context
			return p.
				Flag(&git.Commit, []string{"c", "commit"}, "Commit changes").
				Flag(&git.Message, []string{"m", "message"}, "Commit message").
				Flag(&git.Push, []string{"p", "push"}, "Push changes")
		})
	
	// Supports: -v -Gcm "Fix bug" -p
	// Or: --verbose --git --commit --message "Fix bug" --push
	
	result := parser.Parse(os.Args[1:])
	
	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
		os.Exit(1)
	}
	
	fmt.Printf("verbose: %v\n", config.Verbose)
	fmt.Printf("git commit: %v\n", config.Git.Commit)
	fmt.Printf("commit message: %s\n", config.Git.Message)
	fmt.Printf("git push: %v\n", config.Git.Push)
	fmt.Println()
}

func customTypeExample() {
	fmt.Println("=== Custom Type Example ===")
	
	var timeout time.Duration
	
	// time.Duration is automatically supported!
	parser := lib.New().
		Flag(&timeout, []string{"t", "timeout"}, "Operation timeout")
	
	// Supports: --timeout 30s, --timeout 5m, --timeout 1h30m
	
	result := parser.Parse(os.Args[1:])
	
	if len(result.Errors) > 0 {
		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
		os.Exit(1)
	}
	
	fmt.Printf("timeout: %v\n", timeout)
	fmt.Println()
}

// Reusable parser fragments (functional composition)
func commonFlags(p *lib.Parser) *lib.Parser {
	var verbose bool
	var debug bool
	return p.
		Flag(&verbose, []string{"v", "verbose"}, "Verbose output").
		Flag(&debug, []string{"d", "debug"}, "Debug mode")
}

func withCommonFlags() {
	var name string
	
	// Compose reusable fragments
	parser := commonFlags(lib.New()).
		Flag(&name, []string{"n", "name"}, "Name")
	
	parser.Parse(os.Args[1:])
}

