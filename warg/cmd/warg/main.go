package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"V-Woodpecker-V/wsh/warg/flags"
	"V-Woodpecker-V/wsh/warg/internal/cli"
	"V-Woodpecker-V/wsh/warg/internal/parser"
)

func main() {
	// TODO: warg should use itself to parse its own flags
	// For now, use ad-hoc parsing
	
	args := os.Args[1:]
	
	// Check for help flag
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			printUsage()
			return
		}
	}
	
	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("warg - A flexible argument parser")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  warg [definition] -- [args...]")
	fmt.Println("  warg [definition]                # Just output the definition structure")
	fmt.Println()
	fmt.Println("Definition Formats:")
	fmt.Println("  1. Inline flags:  warg -A -n v,verbose -s -d \"Verbose\" -- -v")
	fmt.Println("  2. Heredoc:       warg <<EOF -- -v")
	fmt.Println("                    -v, --verbose Enable verbose")
	fmt.Println("                    EOF")
	fmt.Println("  3. JSON string:   warg '{\"names\":[\"-v\"],\"switch\":true,\"desc\":\"Verbose\"}' -- -v")
	fmt.Println("  4. JSON from stdin: cat flags.json | warg -- -v")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  -h, --help          Show this help message")
	fmt.Println("  -A, --add           Add a flag definition (context flag)")
	fmt.Println("    -n, --name <names>     Comma-separated flag names (without - or --)")
	fmt.Println("    -s, --switch           Flag is a switch (no value)")
	fmt.Println("    -d, --description <text>  Help message for the flag")
	fmt.Println()
	fmt.Println("Heredoc Format:")
	fmt.Println("  -s, --long [value] Help text")
	fmt.Println("  Where:")
	fmt.Println("    - Names can be short (-s) or long (--long) or both")
	fmt.Println("    - [value] is optional; if present, flag takes a value")
	fmt.Println("    - If [value] is omitted, flag is a switch")
	fmt.Println("    - Help text is always the last argument")
	fmt.Println()
	fmt.Println("JSON Format:")
	fmt.Println("  Array of flag objects: [{\"names\":[\"-v\",\"--verbose\"],\"switch\":true,\"desc\":\"...\"}]")
	fmt.Println("  Or wrapped format: {\"flags\":[...]}")
	fmt.Println()
	fmt.Println("The \"--\" separator:")
	fmt.Println("  Everything before \"--\" defines flags, everything after is parsed")
	fmt.Println("  If no \"--\" is given, only the definition structure is output")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Define and parse")
	fmt.Println("  warg -A -n v,verbose -s -d \"Enable verbose output\" -- -v")
	fmt.Println()
	fmt.Println("  # Just show definition")
	fmt.Println("  warg -A -n v,verbose -s -d \"Enable verbose output\"")
	fmt.Println()
	fmt.Println("  # Heredoc with parsing")
	fmt.Println("  warg <<EOF -- -v --name Alice")
	fmt.Println("  -v, --verbose Enable verbose output")
	fmt.Println("  -n, --name [string] User name")
	fmt.Println("  EOF")
}

func run(args []string) error {
	// Split args at "--" separator
	var defArgs []string
	var parseArgs []string
	separatorIdx := -1
	
	for i, arg := range args {
		if arg == "--" {
			separatorIdx = i
			break
		}
	}
	
	if separatorIdx >= 0 {
		defArgs = args[:separatorIdx]
		parseArgs = args[separatorIdx+1:]
	} else {
		defArgs = args
	}
	
	// Determine input type and parse flag definitions
	var defs []flags.FlagDefinition
	var err error
	
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// stdin is piped - parse as heredoc or JSON
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		input := strings.TrimSpace(string(data))
		if input != "" {
			// Try to parse as JSON first (object or array)
			if strings.HasPrefix(input, "{") || strings.HasPrefix(input, "[") {
				defs, err = cli.ParseJSONDefinitions([]byte(input))
				if err != nil {
					return fmt.Errorf("failed to parse JSON: %w", err)
				}
			} else {
				// Parse as heredoc
				defs, err = cli.ParseHeredocDefinition(input)
				if err != nil {
					return fmt.Errorf("failed to parse heredoc: %w", err)
				}
			}
		}
	} else if len(defArgs) > 0 {
		// Check if first arg is a JSON string (object or array)
		firstArg := strings.TrimSpace(defArgs[0])
		if strings.HasPrefix(firstArg, "{") || strings.HasPrefix(firstArg, "[") {
			// Parse as JSON
			defs, err = cli.ParseJSONDefinitions([]byte(firstArg))
			if err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}
		} else {
			// Parse inline -A --add flags
			defs, err = parseAddFlags(defArgs)
			if err != nil {
				return fmt.Errorf("failed to parse flags: %w", err)
			}
		}
	}
	
	if len(defs) == 0 {
		return fmt.Errorf("no flag definitions provided")
	}
	
	// If no args to parse after "--", just output the definitions
	if len(parseArgs) == 0 {
		output, err := json.MarshalIndent(defs, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(output))
		return nil
	}
	
	// Parse the actual arguments using the definitions
	p := parser.NewParser(defs)
	result, err := p.Parse(parseArgs)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	
	// Output the parsed result as JSON
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}
	
	fmt.Println(string(output))
	return nil
}

// parseAddFlags parses -A --add flags with their subflags
func parseAddFlags(args []string) ([]flags.FlagDefinition, error) {
	var defs []flags.FlagDefinition
	
	i := 0
	for i < len(args) {
		arg := args[i]
		
		if arg == "-A" || arg == "--add" {
			// Start of a new flag definition
			i++
			
			var names []string
			var isSwitch bool
			var desc string
			
			// Parse subflags of -A
			for i < len(args) {
				subarg := args[i]
				
				// Check if we've hit the next -A or end
				if subarg == "-A" || subarg == "--add" {
					break
				}
				
				if subarg == "-n" || subarg == "--name" {
					if i+1 >= len(args) {
						return nil, fmt.Errorf("-n/--name requires a value")
					}
					i++
					// Parse comma-separated names and infer prefix
					nameList := strings.Split(args[i], ",")
					for _, name := range nameList {
						name = strings.TrimSpace(name)
						if name != "" {
							// Infer: 1 char = short (-), multiple chars = long (--)
							if len(name) == 1 {
								names = append(names, "-"+name)
							} else {
								names = append(names, "--"+name)
							}
						}
					}
					i++
				} else if subarg == "-s" || subarg == "--switch" {
					isSwitch = true
					i++
				} else if subarg == "-d" || subarg == "--description" {
					if i+1 >= len(args) {
						return nil, fmt.Errorf("-d/--description requires a value")
					}
					i++
					desc = args[i]
					i++
				} else {
					return nil, fmt.Errorf("unknown flag in -A context: %s", subarg)
				}
			}
			
			if len(names) == 0 {
				return nil, fmt.Errorf("-A requires at least -n/--name to be specified")
			}
			
			defs = append(defs, flags.FlagDefinition{
				Names:       names,
				Switch:      isSwitch,
				Description: desc,
				Children:    []flags.FlagDefinition{},
			})
		} else {
			return nil, fmt.Errorf("unexpected argument: %s (expected -A/--add)", arg)
		}
	}
	
	return defs, nil
}
