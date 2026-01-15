package parser

import (
	"encoding/json"
	"strings"
	"testing"
	
	"V-Woodpecker-V/wsh/warg/flags"
)

func TestFormatOutput_JSON(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose",
		},
	}
	
	result := &ParseResult{
		Flags: []*FlagValue{
			{
				Definition: &defs[0],
				Present:    true,
				Value:      "",
				Children:   []*FlagValue{},
			},
		},
	}

	output, err := FormatOutput(result, OutputJSON)
	if err != nil {
		t.Fatalf("FormatOutput() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed ParseResult
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if len(parsed.Flags) != 1 {
		t.Fatalf("Expected 1 flag in parsed result, got %d", len(parsed.Flags))
	}
	
	if !parsed.Flags[0].Present {
		t.Errorf("Expected verbose Present=true, got %v", parsed.Flags[0].Present)
	}
}

func TestFormatOutput_Env(t *testing.T) {
	verboseDef := flags.FlagDefinition{
		Names:       []string{"-v", "--verbose"},
		Switch:      true,
		Description: "Verbose",
	}
	nameDef := flags.FlagDefinition{
		Names:       []string{"-n", "--name"},
		Switch:      false,
		Description: "Name",
	}
	
	result := &ParseResult{
		Flags: []*FlagValue{
			{
				Definition: &verboseDef,
				Present:    true,
				Value:      "",
				Children:   []*FlagValue{},
			},
			{
				Definition: &nameDef,
				Present:    false,
				Value:      "Alice",
				Children:   []*FlagValue{},
			},
		},
	}

	output, err := FormatOutput(result, OutputEnv)
	if err != nil {
		t.Fatalf("FormatOutput() error = %v", err)
	}

	// Check that output contains environment variable assignments
	if !strings.Contains(output, "WARG_V=true") {
		t.Errorf("Expected output to contain WARG_V=true, got: %s", output)
	}
	if !strings.Contains(output, "WARG_N=Alice") {
		t.Errorf("Expected output to contain WARG_N=Alice, got: %s", output)
	}
}

func TestFormatOutput_Eval(t *testing.T) {
	verboseDef := flags.FlagDefinition{
		Names:       []string{"-v", "--verbose"},
		Switch:      true,
		Description: "Verbose",
	}
	nameDef := flags.FlagDefinition{
		Names:       []string{"-n", "--name"},
		Switch:      false,
		Description: "Name",
	}
	
	result := &ParseResult{
		Flags: []*FlagValue{
			{
				Definition: &verboseDef,
				Present:    true,
				Value:      "",
				Children:   []*FlagValue{},
			},
			{
				Definition: &nameDef,
				Present:    false,
				Value:      "Alice",
				Children:   []*FlagValue{},
			},
		},
	}

	output, err := FormatOutput(result, OutputEval)
	if err != nil {
		t.Fatalf("FormatOutput() error = %v", err)
	}

	// Check that output contains export statements
	if !strings.Contains(output, "export WARG_V=") {
		t.Errorf("Expected output to contain export WARG_V=, got: %s", output)
	}
	if !strings.Contains(output, "export WARG_N=") {
		t.Errorf("Expected output to contain export WARG_N=, got: %s", output)
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"with'quote", "'with'\\''quote'"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := shellEscape(tt.input)
			if result != tt.expected {
				t.Errorf("shellEscape(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		input    interface{}
		expected string
	}{
		{true, "true"},
		{false, "false"},
		{"hello", "hello"},
		{42, "42"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatValue(tt.input)
			if result != tt.expected {
				t.Errorf("formatValue(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
