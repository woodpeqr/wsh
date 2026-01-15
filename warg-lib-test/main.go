package main

import (
	"fmt"
)

// TODO: Implement library integration test once the library API is available
//
// Current status: The warg project currently only has CLI-specific parsing
// functions exposed (ParseInlineDefinition, ParseJSONDefinitions, etc.).
// These are meant for the CLI tool to parse flag DEFINITIONS, not for
// library users who want to use warg to parse command-line arguments in
// their Go programs.
//
// What's needed:
// 1. API separation - move CLI parsing to internal/cli/
// 2. Library API - create public API for Go programs to:
//    - Register flags via struct tags
//    - Parse os.Args or custom arguments
//    - Extract typed values
//
// Once the library API exists, this test should validate:
// - Importing warg as a Go module
// - Using struct tags to define flags
// - Parsing arguments
// - Extracting typed values
// - Context/hierarchical flag behavior
//
// See TODO.md for more details (issues #1 and #2)

func main() {
	fmt.Println("=== warg Library Integration Test ===")
	fmt.Println()
	fmt.Println("⚠️  TODO: Library API not yet implemented")
	fmt.Println()
	fmt.Println("Current state:")
	fmt.Println("  - CLI-specific parsing functions exist (for parsing flag definitions)")
	fmt.Println("  - Library API for Go programs does not exist yet")
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  1. Separate CLI and library APIs (see TODO.md #1)")
	fmt.Println("  2. Implement library-focused parsing API")
	fmt.Println("  3. Write integration tests for library usage")
	fmt.Println()
	fmt.Println("See warg/TODO.md for full details")
}
