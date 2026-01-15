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
	"V-Woodpecker-V/wsh/warg/lib"
)

func main() {
	args := os.Args[1:]
	
	// Parse help flag using lib
	var helpFlag bool
	helpParser := lib.New().
		Flag(&helpFlag, []string{"h", "help"}, "Show help message")
	
	// Parse just to extract help flag - ignore errors as we'll parse properly later
	helpParser.Parse(args)
	
	if helpFlag {
		printUsage()
		return
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
		// Check if first arg looks like inline -A flags
		firstArg := strings.TrimSpace(defArgs[0])
		if firstArg == "-A" || firstArg == "--add" {
			// Parse inline -A flags using lib
			defs, err = parseAddFlagsWithLib(defArgs)
			if err != nil {
				return fmt.Errorf("failed to parse inline flags: %w", err)
			}
		} else if strings.HasPrefix(firstArg, "{") || strings.HasPrefix(firstArg, "[") {
			// Parse as JSON
			defs, err = cli.ParseJSONDefinitions([]byte(firstArg))
			if err != nil {
				return fmt.Errorf("failed to parse JSON: %w", err)
			}
		} else {
			// Try parsing as inline flags (fallback)
			defs, err = parseAddFlagsWithLib(defArgs)
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

// AddFlagDef represents a single flag definition from -A context
type AddFlagDef struct {
	Names       string
	Switch      bool
	Description string
}

// parseAddFlagsWithLib parses -A --add flags using the lib package
func parseAddFlagsWithLib(args []string) ([]flags.FlagDefinition, error) {
	var addFlags []AddFlagDef
	
	parser := lib.New().
		Context(&addFlags, []string{"A", "add"}, "Add flag definition", 
			func(p *lib.Parser, def *AddFlagDef) *lib.Parser {
				return p.
					Flag(&def.Names, []string{"n", "name"}, "Flag names").
					Flag(&def.Switch, []string{"s", "switch"}, "Switch flag").
					Flag(&def.Description, []string{"d", "description"}, "Description")
			})
	
	result := parser.Parse(args)
	if len(result.Errors) > 0 {
		return nil, result.Errors[0]
	}
	
	// Convert to FlagDefinition
	var defs []flags.FlagDefinition
	for _, addFlag := range addFlags {
		if addFlag.Names == "" {
			return nil, fmt.Errorf("-A requires -n/--name to be specified")
		}
		
		// Parse comma-separated names and infer prefix
		nameList := strings.Split(addFlag.Names, ",")
		var names []string
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
		
		if len(names) == 0 {
			return nil, fmt.Errorf("-A requires at least one flag name")
		}
		
		defs = append(defs, flags.FlagDefinition{
			Names:       names,
			Switch:      addFlag.Switch,
			Description: addFlag.Description,
			Children:    []flags.FlagDefinition{},
		})
	}
	
	return defs, nil
}
