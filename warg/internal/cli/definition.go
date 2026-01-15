package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"V-Woodpecker-V/wsh/warg/flags"
)

// ParseInlineDefinition parses inline flag definitions in the format:
// "n,name;string;User name" or "G.c,commit;bool;Commit changes"
// Note: This is kept for backward compatibility with JSON parsing
func ParseInlineDefinition(def string) (*flags.FlagDefinition, string, error) {
	parts := strings.Split(def, ";")
	if len(parts) < 3 {
		return nil, "", fmt.Errorf("invalid definition format, expected 'names;type;description': %s", def)
	}

	namesStr := strings.TrimSpace(parts[0])
	typeStr := strings.TrimSpace(parts[1])
	desc := strings.TrimSpace(parts[2])

	// Parse parent context (e.g., "G.c" -> parent="G", names="c")
	var parent string
	if idx := strings.Index(namesStr, "."); idx != -1 {
		parent = namesStr[:idx]
		namesStr = namesStr[idx+1:]
	}

	// Parse names (e.g., "n,name" -> ["-n", "--name"])
	nameList := strings.Split(namesStr, ",")
	names := make([]string, 0, len(nameList))
	for _, name := range nameList {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		// Add appropriate prefix if not present
		if !strings.HasPrefix(name, "-") {
			if len(name) == 1 {
				names = append(names, "-"+name)
			} else {
				names = append(names, "--"+name)
			}
		} else {
			names = append(names, name)
		}
	}

	if len(names) == 0 {
		return nil, "", fmt.Errorf("no valid flag names found in: %s", def)
	}

	// Parse type - convert old type system to new switch system
	isSwitch := (typeStr == "bool" || typeStr == "context")

	return &flags.FlagDefinition{
		Names:       names,
		Switch:      isSwitch,
		Description: desc,
		Children:    []flags.FlagDefinition{},
	}, parent, nil
}

// ParseInlineDefinitions parses multiple inline definitions and builds hierarchy
func ParseInlineDefinitions(defs []string) ([]flags.FlagDefinition, error) {
	rootFlags := make(map[string]*flags.FlagDefinition)
	childFlags := make(map[string][]flags.FlagDefinition)

	// First pass: parse all definitions
	for _, def := range defs {
		flagDef, parent, err := ParseInlineDefinition(def)
		if err != nil {
			return nil, err
		}

		if parent == "" {
			// Root level flag
			for _, name := range flagDef.Names {
				rootFlags[name] = flagDef
			}
		} else {
			// Child flag
			if childFlags[parent] == nil {
				childFlags[parent] = []flags.FlagDefinition{}
			}
			childFlags[parent] = append(childFlags[parent], *flagDef)
		}
	}

	// Second pass: attach children to parents
	for parentName, children := range childFlags {
		found := false
		for _, rootFlag := range rootFlags {
			for _, name := range rootFlag.Names {
				nameWithoutPrefix := strings.TrimLeft(name, "-")
				if nameWithoutPrefix == parentName {
					rootFlag.Children = append(rootFlag.Children, children...)
					// Context flags (with children) are always switches
					rootFlag.Switch = true
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("parent flag '%s' not found for child flags", parentName)
		}
	}

	// Convert map to slice
	result := make([]flags.FlagDefinition, 0, len(rootFlags))
	seen := make(map[*flags.FlagDefinition]bool)
	for _, flag := range rootFlags {
		if !seen[flag] {
			result = append(result, *flag)
			seen[flag] = true
		}
	}

	return result, nil
}

// ParseJSONDefinitions parses JSON flag definitions
// Supports both {"flags": [...]} and direct array [...] formats
func ParseJSONDefinitions(jsonData []byte) ([]flags.FlagDefinition, error) {
	// Try direct array format first
	var directDefs []flags.FlagDefinition
	if err := json.Unmarshal(jsonData, &directDefs); err == nil {
		return directDefs, nil
	}
	
	// Try wrapped format {"flags": [...]}
	var wrappedDefs struct {
		Flags []flags.FlagDefinition `json:"flags"`
	}
	if err := json.Unmarshal(jsonData, &wrappedDefs); err != nil {
		return nil, fmt.Errorf("failed to parse JSON definitions: %w", err)
	}
	return wrappedDefs.Flags, nil
}

// ParseHeredocDefinition parses heredoc-style DSL definitions
// Format: -s, --long [value] Help text
// Where [value] is optional - if present, flag takes a value; if absent, it's a switch
func ParseHeredocDefinition(input string) ([]flags.FlagDefinition, error) {
	lines := strings.Split(input, "\n")
	var result []flags.FlagDefinition
	var stack []*flags.FlagDefinition
	var indentLevels []int

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Calculate indentation level
		indent := 0
		for _, ch := range line {
			if ch == ' ' {
				indent++
			} else if ch == '\t' {
				indent += 2
			} else {
				break
			}
		}

		// Parse line: "-n, --name [value] Help text" or "-v, --verbose Help text"
		line = strings.TrimSpace(line)
		
		// Find where description starts (after the last ] or after the last flag name)
		var names []string
		var isSwitch = true // default to switch
		var desc string
		
		// Parse tokens
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}
		
		i := 0
		// Collect flag names (start with - or --)
		for i < len(tokens) && (strings.HasPrefix(tokens[i], "-") || tokens[i] == ",") {
			token := tokens[i]
			if token != "," {
				// Remove trailing comma if present
				token = strings.TrimSuffix(token, ",")
				if strings.HasPrefix(token, "-") {
					names = append(names, token)
				}
			}
			i++
		}
		
		// Check for [value] token
		if i < len(tokens) && strings.HasPrefix(tokens[i], "[") {
			// This flag takes a value
			isSwitch = false
			i++
		}
		
		// Rest is description
		if i < len(tokens) {
			desc = strings.Join(tokens[i:], " ")
		}
		
		if len(names) == 0 {
			return nil, fmt.Errorf("no flag names found in line: %s", line)
		}

		flagDef := &flags.FlagDefinition{
			Names:       names,
			Switch:      isSwitch,
			Description: desc,
			Children:    []flags.FlagDefinition{},
		}

		// Determine where to add this flag based on indentation
		if indent == 0 {
			// Root level
			result = append(result, *flagDef)
			stack = []*flags.FlagDefinition{&result[len(result)-1]}
			indentLevels = []int{0}
		} else {
			// Find parent based on indentation
			for len(indentLevels) > 0 && indentLevels[len(indentLevels)-1] >= indent {
				stack = stack[:len(stack)-1]
				indentLevels = indentLevels[:len(indentLevels)-1]
			}

			if len(stack) == 0 {
				return nil, fmt.Errorf("invalid indentation at line: %s", line)
			}

			parent := stack[len(stack)-1]
			parent.Children = append(parent.Children, *flagDef)
			// Context flags (with children) are always switches
			parent.Switch = true
			stack = append(stack, &parent.Children[len(parent.Children)-1])
			indentLevels = append(indentLevels, indent)
		}
	}

	return result, nil
}
