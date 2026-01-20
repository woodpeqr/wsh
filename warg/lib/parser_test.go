package lib

import (
	"testing"
	"time"
)

func TestParser_BasicTypes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() (*Parser, any)
		args     []string
		expected any
	}{
		{
			name: "bool flag",
			setup: func() (*Parser, any) {
				var verbose bool
				p := New().Flag(&verbose, []string{"v", "verbose"}, "Verbose output")
				return p, &verbose
			},
			args:     []string{"-v"},
			expected: true,
		},
		{
			name: "string flag",
			setup: func() (*Parser, any) {
				var name string
				p := New().Flag(&name, []string{"n", "name"}, "User name")
				return p, &name
			},
			args:     []string{"--name", "Alice"},
			expected: "Alice",
		},
		{
			name: "int flag",
			setup: func() (*Parser, any) {
				var count int
				p := New().Flag(&count, []string{"c", "count"}, "Count")
				return p, &count
			},
			args:     []string{"--count", "42"},
			expected: 42,
		},
		{
			name: "float flag",
			setup: func() (*Parser, any) {
				var rate float64
				p := New().Flag(&rate, []string{"r", "rate"}, "Rate")
				return p, &rate
			},
			args:     []string{"--rate", "3.14"},
			expected: 3.14,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, ptr := tt.setup()
			result := parser.Parse(tt.args)

			if len(result.Errors) > 0 {
				t.Fatalf("Parse error: %v", result.Errors[0])
			}

			// Check the value
			switch v := ptr.(type) {
			case *bool:
				if *v != tt.expected.(bool) {
					t.Errorf("Expected %v, got %v", tt.expected, *v)
				}
			case *string:
				if *v != tt.expected.(string) {
					t.Errorf("Expected %v, got %v", tt.expected, *v)
				}
			case *int:
				if *v != tt.expected.(int) {
					t.Errorf("Expected %v, got %v", tt.expected, *v)
				}
			case *float64:
				if *v != tt.expected.(float64) {
					t.Errorf("Expected %v, got %v", tt.expected, *v)
				}
			}
		})
	}
}

func TestParser_Slices(t *testing.T) {
	t.Run("string slice with repeated flags", func(t *testing.T) {
		var tags []string
		parser := New().Flag(&tags, []string{"t", "tag"}, "Tags")

		result := parser.Parse([]string{"--tag", "go", "--tag", "cli", "-t", "parser"})
		if len(result.Errors) > 0 {
			t.Fatalf("Parse error: %v", result.Errors[0])
		}

		expected := []string{"go", "cli", "parser"}
		if len(tags) != len(expected) {
			t.Fatalf("Expected %d tags, got %d", len(expected), len(tags))
		}
		for i, tag := range tags {
			if tag != expected[i] {
				t.Errorf("Expected tag[%d]=%s, got %s", i, expected[i], tag)
			}
		}
	})

	t.Run("string slice with comma-separated values", func(t *testing.T) {
		var tags []string
		parser := New().Flag(&tags, []string{"t", "tag"}, "Tags")

		result := parser.Parse([]string{"--tag", "go,cli,parser"})
		if len(result.Errors) > 0 {
			t.Fatalf("Parse error: %v", result.Errors[0])
		}

		expected := []string{"go", "cli", "parser"}
		if len(tags) != len(expected) {
			t.Fatalf("Expected %d tags, got %d", len(expected), len(tags))
		}
		for i, tag := range tags {
			if tag != expected[i] {
				t.Errorf("Expected tag[%d]=%s, got %s", i, expected[i], tag)
			}
		}
	})

	t.Run("int slice with repeated flags", func(t *testing.T) {
		var ports []int
		parser := New().Flag(&ports, []string{"p", "port"}, "Ports")

		result := parser.Parse([]string{"--port", "8080", "-p", "3000", "--port", "9000"})
		if len(result.Errors) > 0 {
			t.Fatalf("Parse error: %v", result.Errors[0])
		}

		expected := []int{8080, 3000, 9000}
		if len(ports) != len(expected) {
			t.Fatalf("Expected %d ports, got %d", len(expected), len(ports))
		}
		for i, port := range ports {
			if port != expected[i] {
				t.Errorf("Expected port[%d]=%d, got %d", i, expected[i], port)
			}
		}
	})
}

func TestParser_MultipleFlags(t *testing.T) {
	var verbose bool
	var name string
	var count int

	parser := New().
		Flag(&verbose, []string{"v", "verbose"}, "Verbose").
		Flag(&name, []string{"n", "name"}, "Name").
		Flag(&count, []string{"c", "count"}, "Count")

	result := parser.Parse([]string{"-v", "--name", "Bob", "--count", "5"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if !verbose {
		t.Error("Expected verbose to be true")
	}
	if name != "Bob" {
		t.Errorf("Expected name=Bob, got %s", name)
	}
	if count != 5 {
		t.Errorf("Expected count=5, got %d", count)
	}
}

func TestParser_Context(t *testing.T) {
	type GitConfig struct {
		Commit  bool
		Message string
	}
	
	type Config struct {
		Verbose bool
		Git     GitConfig
	}
	
	var config Config

	parser := New().
		Flag(&config.Verbose, []string{"v", "verbose"}, "Verbose").
		Context(&config.Git, []string{"G", "git"}, "Git operations", func(p *Parser, git *GitConfig) *Parser {
			return p.
				Flag(&git.Commit, []string{"c", "commit"}, "Commit").
				Flag(&git.Message, []string{"m", "message"}, "Message")
		})

	result := parser.Parse([]string{"-v", "-G", "-c", "-m", "Fix bug"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if !config.Verbose {
		t.Error("Expected verbose to be true")
	}
	if !config.Git.Commit {
		t.Error("Expected gitCommit to be true")
	}
	if config.Git.Message != "Fix bug" {
		t.Errorf("Expected commitMsg='Fix bug', got '%s'", config.Git.Message)
	}
}

func TestParser_RepeatableSliceContext(t *testing.T) {
	type AddFlagDef struct {
		Names       string
		Description string
		IsSwitch    bool
	}
	
	var addFlags []AddFlagDef

	parser := New().
		Context(&addFlags, []string{"A", "add"}, "Add flag definition", 
			func(p *Parser, def *AddFlagDef) *Parser {
				return p.
					Flag(&def.Names, []string{"n", "name"}, "Flag names").
					Flag(&def.Description, []string{"d", "description"}, "Description").
					Flag(&def.IsSwitch, []string{"s", "switch"}, "Switch flag")
			})

	result := parser.Parse([]string{
		"-A", "-n", "v,verbose", "-d", "Verbose output", "-s",
		"-A", "-n", "n,name", "-d", "User name",
	})
	
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if len(addFlags) != 2 {
		t.Fatalf("Expected 2 flag definitions, got %d", len(addFlags))
	}

	// Check first definition
	if addFlags[0].Names != "v,verbose" {
		t.Errorf("Expected first names='v,verbose', got '%s'", addFlags[0].Names)
	}
	if addFlags[0].Description != "Verbose output" {
		t.Errorf("Expected first description='Verbose output', got '%s'", addFlags[0].Description)
	}
	if !addFlags[0].IsSwitch {
		t.Error("Expected first IsSwitch to be true")
	}

	// Check second definition
	if addFlags[1].Names != "n,name" {
		t.Errorf("Expected second names='n,name', got '%s'", addFlags[1].Names)
	}
	if addFlags[1].Description != "User name" {
		t.Errorf("Expected second description='User name', got '%s'", addFlags[1].Description)
	}
	if addFlags[1].IsSwitch {
		t.Error("Expected second IsSwitch to be false")
	}
}

func TestParser_CombinedShortFlags(t *testing.T) {
	var verbose bool
	var debug bool
	var interactive bool

	parser := New().
		Flag(&verbose, []string{"v"}, "Verbose").
		Flag(&debug, []string{"d"}, "Debug").
		Flag(&interactive, []string{"i"}, "Interactive")

	result := parser.Parse([]string{"-vdi"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if !verbose {
		t.Error("Expected verbose to be true")
	}
	if !debug {
		t.Error("Expected debug to be true")
	}
	if !interactive {
		t.Error("Expected interactive to be true")
	}
}

func TestParser_CustomTypes(t *testing.T) {
	var timeout time.Duration

	parser := New().Flag(&timeout, []string{"t", "timeout"}, "Timeout")

	result := parser.Parse([]string{"--timeout", "5m30s"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	expected := 5*time.Minute + 30*time.Second
	if timeout != expected {
		t.Errorf("Expected timeout=%v, got %v", expected, timeout)
	}
}

func TestParser_Immutability(t *testing.T) {
	// Test that creating new parsers doesn't modify the original
	p1 := New()
	var flag1 bool
	p2 := p1.Flag(&flag1, []string{"a"}, "Flag A")

	var flag2 bool
	p3 := p2.Flag(&flag2, []string{"b"}, "Flag B")

	// p1 should have 0 definitions
	if len(p1.definitions) != 0 {
		t.Errorf("p1 should have 0 definitions, got %d", len(p1.definitions))
	}

	// p2 should have 1 definition
	if len(p2.definitions) != 1 {
		t.Errorf("p2 should have 1 definition, got %d", len(p2.definitions))
	}

	// p3 should have 2 definitions
	if len(p3.definitions) != 2 {
		t.Errorf("p3 should have 2 definitions, got %d", len(p3.definitions))
	}
}

func TestParser_ErrorHandling(t *testing.T) {
	t.Run("unknown flag", func(t *testing.T) {
		var name string
		parser := New().Flag(&name, []string{"n"}, "Name")

		result := parser.Parse([]string{"--unknown"})
		if len(result.Errors) == 0 {
			t.Error("Expected error for unknown flag")
		}
	})

	t.Run("missing value", func(t *testing.T) {
		var name string
		parser := New().Flag(&name, []string{"n"}, "Name")

		result := parser.Parse([]string{"-n"})
		if len(result.Errors) == 0 {
			t.Error("Expected error for missing value")
		}
	})

	t.Run("invalid int", func(t *testing.T) {
		var count int
		parser := New().Flag(&count, []string{"c"}, "Count")

		result := parser.Parse([]string{"-c", "not-a-number"})
		if len(result.Errors) == 0 {
			t.Error("Expected error for invalid int")
		}
	})
}

func TestParser_EmptyArgs(t *testing.T) {
	var verbose bool
	parser := New().Flag(&verbose, []string{"v"}, "Verbose")

	result := parser.Parse([]string{})
	if len(result.Errors) > 0 {
		t.Fatalf("Unexpected error: %v", result.Errors[0])
	}

	if verbose {
		t.Error("Expected verbose to be false with no args")
	}
}

func TestParser_UnsignedInts(t *testing.T) {
	var port uint16
	parser := New().Flag(&port, []string{"p", "port"}, "Port number")

	result := parser.Parse([]string{"--port", "8080"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if port != 8080 {
		t.Errorf("Expected port=8080, got %d", port)
	}
}

func TestParser_LongNames(t *testing.T) {
	var name string
	parser := New().Flag(&name, []string{"user-name"}, "User name")

	result := parser.Parse([]string{"--user-name", "Alice"})
	if len(result.Errors) > 0 {
		t.Fatalf("Parse error: %v", result.Errors[0])
	}

	if name != "Alice" {
		t.Errorf("Expected name=Alice, got %s", name)
	}
}
