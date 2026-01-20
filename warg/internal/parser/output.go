package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// OutputFormat specifies how to format the parse results
type OutputFormat string

const (
	OutputJSON   OutputFormat = "json"
	OutputEnv    OutputFormat = "env"
	OutputEval   OutputFormat = "eval"
)

// FormatOutput formats the parse result according to the specified format
func FormatOutput(result *ParseResult, format OutputFormat) (string, error) {
	switch format {
	case OutputJSON:
		return formatJSON(result)
	case OutputEnv:
		return formatEnv(result)
	case OutputEval:
		return formatEval(result)
	default:
		return "", fmt.Errorf("unknown output format: %s", format)
	}
}

// formatJSON outputs the result as JSON
func formatJSON(result *ParseResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return string(data), nil
}

// formatEnv outputs the result as environment variable assignments
func formatEnv(result *ParseResult) (string, error) {
	var lines []string
	flatMap := flattenResult(result)
	for key, value := range flatMap {
		// Convert key to uppercase and prefix with WARG_
		envKey := "WARG_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		envValue := formatValue(value)
		lines = append(lines, fmt.Sprintf("%s=%s", envKey, envValue))
	}
	return strings.Join(lines, "\n"), nil
}

// formatEval outputs the result in shell eval format
func formatEval(result *ParseResult) (string, error) {
	var lines []string
	flatMap := flattenResult(result)
	for key, value := range flatMap {
		// Convert key to uppercase and prefix with WARG_
		envKey := "WARG_" + strings.ToUpper(strings.ReplaceAll(key, "-", "_"))
		envValue := formatValue(value)
		// Use export for shell compatibility
		lines = append(lines, fmt.Sprintf("export %s=%s", envKey, shellEscape(envValue)))
	}
	return strings.Join(lines, "\n"), nil
}

// flattenResult converts the tree structure to a flat map for env/eval output
func flattenResult(result *ParseResult) map[string]any {
	flat := make(map[string]any)
	for _, flag := range result.Flags {
		flattenFlagValue(flag, flat)
	}
	return flat
}

// flattenFlagValue recursively flattens a flag value and its children
func flattenFlagValue(flag *FlagValue, flat map[string]any) {
	// Get a canonical name (use the first name, stripped of dashes)
	canonicalName := flag.Definition.Names[0]
	key := strings.TrimLeft(canonicalName, "-")
	
	if flag.Definition.Switch {
		flat[key] = flag.Present
	} else {
		flat[key] = flag.Value
	}
	
	// Flatten children
	for _, child := range flag.Children {
		flattenFlagValue(child, flat)
	}
}

// formatValue converts a value to string representation
func formatValue(value any) string {
	switch v := value.(type) {
	case bool:
		if v {
			return "true"
		}
		return "false"
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// shellEscape escapes a value for safe use in shell
func shellEscape(value string) string {
	// Simple escaping - wrap in single quotes and escape single quotes
	if strings.Contains(value, "'") {
		value = strings.ReplaceAll(value, "'", "'\\''")
	}
	return "'" + value + "'"
}
