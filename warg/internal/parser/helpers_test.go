package parser

import (
	"testing"

	"V-Woodpecker-V/wsh/warg/flags"
)

func TestParseResult_Find(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose",
		},
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-v", "-Gc"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Find root-level flag
	verboseFlag := result.Find("-v")
	if verboseFlag == nil {
		t.Fatal("Expected to find -v flag")
	}
	if !verboseFlag.Present {
		t.Errorf("Expected verbose Present=true")
	}

	// Find by long name
	verboseFlag = result.Find("--verbose")
	if verboseFlag == nil {
		t.Fatal("Expected to find --verbose flag")
	}

	// Find nested flag
	commitFlag := result.Find("-c")
	if commitFlag == nil {
		t.Fatal("Expected to find -c flag")
	}
	if !commitFlag.Present {
		t.Errorf("Expected commit Present=true")
	}

	// Find non-existent flag
	missingFlag := result.Find("-x")
	if missingFlag != nil {
		t.Errorf("Expected not to find -x flag, but got %v", missingFlag)
	}
}

func TestParseResult_Walk(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-v", "--verbose"},
			Switch:      true,
			Description: "Verbose",
		},
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-v", "-Gc"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Count all flags using Walk
	count := 0
	result.Walk(func(fv *FlagValue) {
		count++
	})

	// Should have 3 flags: -v, -G, -c
	if count != 3 {
		t.Errorf("Expected 3 flags, got %d", count)
	}
}

func TestFlagValue_Find(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit",
				},
				{
					Names:       []string{"-m", "--message"},
					Switch:      false,
					Description: "Message",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-G", "-c", "-m", "test"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	gitFlag := result.Flags[0]

	// Find itself
	found := gitFlag.Find("-G")
	if found != gitFlag {
		t.Error("Expected to find itself")
	}

	// Find child
	commitFlag := gitFlag.Find("-c")
	if commitFlag == nil {
		t.Fatal("Expected to find -c flag")
	}
	if !commitFlag.Present {
		t.Errorf("Expected commit Present=true")
	}

	// Find another child
	messageFlag := gitFlag.Find("--message")
	if messageFlag == nil {
		t.Fatal("Expected to find --message flag")
	}
	if messageFlag.Value != "test" {
		t.Errorf("Expected message='test', got %v", messageFlag.Value)
	}
}

func TestFlagValue_Walk(t *testing.T) {
	defs := []flags.FlagDefinition{
		{
			Names:       []string{"-G", "--git"},
			Switch:      true,
			Description: "Git operations",
			Children: []flags.FlagDefinition{
				{
					Names:       []string{"-c", "--commit"},
					Switch:      true,
					Description: "Commit",
				},
			},
		},
	}

	parser := NewParser(defs)
	result, err := parser.Parse([]string{"-Gc"})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	gitFlag := result.Flags[0]

	// Count flags in git subtree
	count := 0
	gitFlag.Walk(func(fv *FlagValue) {
		count++
	})

	// Should have 2 flags: -G and -c
	if count != 2 {
		t.Errorf("Expected 2 flags in git subtree, got %d", count)
	}
}

func TestFlagValue_IsSwitch(t *testing.T) {
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

	verboseFlag := &FlagValue{Definition: &verboseDef}
	nameFlag := &FlagValue{Definition: &nameDef}

	if !verboseFlag.IsSwitch() {
		t.Error("Expected verbose to be a switch")
	}
	if nameFlag.IsSwitch() {
		t.Error("Expected name not to be a switch")
	}
}

func TestFlagValue_GetBool(t *testing.T) {
	def := flags.FlagDefinition{
		Names:       []string{"-v"},
		Switch:      true,
		Description: "Verbose",
	}

	flag := &FlagValue{
		Definition: &def,
		Present:    true,
	}

	if !flag.GetBool() {
		t.Error("Expected GetBool() to return true")
	}
}

func TestFlagValue_GetString(t *testing.T) {
	def := flags.FlagDefinition{
		Names:       []string{"-n"},
		Switch:      false,
		Description: "Name",
	}

	flag := &FlagValue{
		Definition: &def,
		Value:      "Alice",
	}

	if flag.GetString() != "Alice" {
		t.Errorf("Expected GetString() to return 'Alice', got %v", flag.GetString())
	}
}
