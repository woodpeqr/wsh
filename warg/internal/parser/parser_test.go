package parser

import (
	"reflect"
	"strings"
	"testing"

	"V-Woodpecker-V/wsh/warg/flags"
)

func TestParser_Parse_SimpleBoolFlag(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose output",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}
	
	flag := result.Flags[0]
	if !flag.Present {
		t.Errorf("Expected verbose Present=true, got %v", flag.Present)
	}
	if flag.Definition.Names[0] != "-v" {
		t.Errorf("Expected flag name -v, got %v", flag.Definition.Names[0])
	}
}

func TestParser_Parse_StringFlag(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-n", "--name"},
			Switch:      false,
			Description: "User name",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-n", "Alice"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}
	
	flag := result.Flags[0]
	if flag.Value != "Alice" {
		t.Errorf("Expected name=Alice, got %v", flag.Value)
	}
	if flag.Definition.Names[0] != "-n" {
		t.Errorf("Expected flag name -n, got %v", flag.Definition.Names[0])
	}
}

func TestParser_Parse_LongFlag(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-n", "--name"},
			Switch:      false,
			Description: "User name",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"--name", "Bob"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}
	
	flag := result.Flags[0]
	if flag.Value != "Bob" {
		t.Errorf("Expected name=Bob, got %v", flag.Value)
	}
}

func TestParser_Parse_CombinedShortFlags(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-a"},
			Switch:      true,
			Description: "Flag A",
		},
		{
			Names:       []string{"-b"},
			Switch:      true,
			Description: "Flag B",
		},
		{
			Names:       []string{"-c"},
			Switch:      true,
			Description: "Flag C",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-abc"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 3 {
		t.Fatalf("Expected 3 flags, got %d", len(result.Flags))
	}
	
	// Check all flags are present
	foundA, foundB, foundC := false, false, false
	for _, flag := range result.Flags {
		switch flag.Definition.Names[0] {
		case "-a":
			foundA = flag.Present
		case "-b":
			foundB = flag.Present
		case "-c":
			foundC = flag.Present
		}
	}
	
	if !foundA {
		t.Errorf("Expected a=true")
	}
	if !foundB {
		t.Errorf("Expected b=true")
	}
	if !foundC {
		t.Errorf("Expected c=true")
	}
}

func TestParser_Parse_ContextFlag(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit changes",
				},
				{
					Names:       []string{"-m", "--message"},
					Switch:      false,
					Description: "Commit message",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-G", "-c", "-m", "fix bug"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 root flag, got %d", len(result.Flags))
	}
	
	gitFlag := result.Flags[0]
	if !gitFlag.Present {
		t.Errorf("Expected git Present=true, got %v", gitFlag.Present)
	}
	if gitFlag.Definition.Names[0] != "-G" {
		t.Errorf("Expected flag name -G, got %v", gitFlag.Definition.Names[0])
	}
	
	if len(gitFlag.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(gitFlag.Children))
	}
	
	// Check children
	for _, child := range gitFlag.Children {
		switch child.Definition.Names[0] {
		case "-c":
			if !child.Present {
				t.Errorf("Expected commit Present=true, got %v", child.Present)
			}
		case "-m":
			if child.Value != "fix bug" {
				t.Errorf("Expected message='fix bug', got %v", child.Value)
			}
		default:
			t.Errorf("Unexpected child flag: %v", child.Definition.Names[0])
		}
	}
}

func TestParser_Parse_CombinedContextFlags(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit changes",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-Gc"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 root flag, got %d", len(result.Flags))
	}
	
	gitFlag := result.Flags[0]
	if !gitFlag.Present {
		t.Errorf("Expected git Present=true, got %v", gitFlag.Present)
	}
	
	if len(gitFlag.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(gitFlag.Children))
	}
	
	commitFlag := gitFlag.Children[0]
	if !commitFlag.Present {
		t.Errorf("Expected commit Present=true, got %v", commitFlag.Present)
	}
}

func TestParser_Parse_ContextResolution(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose output",
		},
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit changes",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-G", "-v", "-c"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 2 {
		t.Fatalf("Expected 2 root flags, got %d", len(result.Flags))
	}
	
	// Find the flags
	var gitFlag, verboseFlag *FlagValue
	for _, flag := range result.Flags {
		switch flag.Definition.Names[0] {
		case "-G":
			gitFlag = flag
		case "-v":
			verboseFlag = flag
		}
	}
	
	if gitFlag == nil {
		t.Fatal("Git flag not found")
	}
	if verboseFlag == nil {
		t.Fatal("Verbose flag not found")
	}
	
	if !gitFlag.Present {
		t.Errorf("Expected git Present=true")
	}
	if !verboseFlag.Present {
		t.Errorf("Expected verbose Present=true")
	}
	
	if len(gitFlag.Children) != 1 {
		t.Fatalf("Expected 1 child for git, got %d", len(gitFlag.Children))
	}
	
	commitFlag := gitFlag.Children[0]
	if !commitFlag.Present {
		t.Errorf("Expected commit Present=true")
	}
}

func TestParser_Parse_UnknownFlag(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose output",
		},
	}

	parser := NewParser(defs)
	_, err := parser.Parse([]string{"-x"})
	if err == nil {
		t.Error("Expected error for unknown flag, got nil")
	}
}

func TestParser_Parse_MissingValue(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-n", "--name"},
			Switch:      false,
			Description: "User name",
		},
	}

	parser := NewParser(defs)
	_, err := parser.Parse([]string{"-n"})
	if err == nil {
		t.Error("Expected error for missing value, got nil")
	}
}

func TestParser_Parse_MultipleFlags(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose output",
		},
		{
			Names:       []string{"-n", "--name"},
			Switch:      false,
			Description: "User name",
		},
		{
			Names:       []string{"-d", "--debug"},
			Switch:      true,
			Description: "Debug mode",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-v", "-n", "Alice", "-d"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(result.Flags) != 3 {
		t.Fatalf("Expected 3 flags, got %d", len(result.Flags))
	}
	
	// Build a flat map to check values (similar to old behavior)
	flat := make(map[string]any)
	for _, flag := range result.Flags {
		key := strings.TrimLeft(flag.Definition.Names[0], "-")
		if flag.Definition.Switch {
			flat[key] = flag.Present
		} else {
			flat[key] = flag.Value
		}
	}

	expected := map[string]any{
		"v": true,
		"n": "Alice",
		"d": true,
	}

	if !reflect.DeepEqual(flat, expected) {
		t.Errorf("Expected %v, got %v", expected, flat)
	}
}

func TestParser_Parse_DoubleDashSeparator(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose output",
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-v", "--", "-n"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should only parse -v, not -n after --
	if len(result.Flags) != 1 {
		t.Fatalf("Expected 1 flag, got %d", len(result.Flags))
	}
	
	if result.Flags[0].Definition.Names[0] != "-v" {
		t.Errorf("Expected -v flag, got %v", result.Flags[0].Definition.Names[0])
	}
	
	if !result.Flags[0].Present {
		t.Errorf("Expected verbose Present=true")
	}
}

func TestContext_Lookup(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose",
		},
	}

	ctx := NewContext(defs, nil)

	// Test short flag lookup
	def := ctx.Lookup("-v")
	if def == nil {
		t.Error("Expected to find -v flag")
	}

	// Test long flag lookup
	def = ctx.Lookup("--verbose")
	if def == nil {
		t.Error("Expected to find --verbose flag")
	}

	// Test non-existent flag
	def = ctx.Lookup("-x")
	if def != nil {
		t.Error("Should not find non-existent flag")
	}
}

func TestContext_ParentLookup(t *testing.T) {
	parentDefs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose",
		},
	}

	childDefs := []flags.FlagDefinition{
		{
			Names:       []string{"-c", "--commit"},
			Switch:      true,
			Description: "Commit",
		},
	}

	parent := NewContext(parentDefs, nil)
	child := NewContext(childDefs, parent)

	// Child should find its own flag
	def := child.Lookup("-c")
	if def == nil {
		t.Error("Expected to find -c in child context")
	}

	// Child should find parent's flag
	def = child.Lookup("-v")
	if def == nil {
		t.Error("Expected to find -v from parent context")
	}
}
