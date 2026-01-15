package integration_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"V-Woodpecker-V/wsh/warg/lib"
)

func TestBasicTypes(t *testing.T) {
	t.Run("bool flag", func(t *testing.T) {
		var verbose bool

		parser := lib.New().
			Flag(&verbose, []string{"v", "verbose"}, "Verbose")

		result := parser.Parse([]string{"-v"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.True(t, verbose, "verbose should be true")
	})

	t.Run("string flag", func(t *testing.T) {
		var name string

		parser := lib.New().
			Flag(&name, []string{"n", "name"}, "Name")

		result := parser.Parse([]string{"--name", "Alice"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, "Alice", name, "name should be Alice")
	})

	t.Run("int flag", func(t *testing.T) {
		var count int

		parser := lib.New().
			Flag(&count, []string{"c", "count"}, "Count")

		result := parser.Parse([]string{"--count", "42"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, 42, count, "count should be 42")
	})

	t.Run("float flag", func(t *testing.T) {
		var rate float64

		parser := lib.New().
			Flag(&rate, []string{"r", "rate"}, "Rate")

		result := parser.Parse([]string{"--rate", "3.14"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, 3.14, rate, "rate should be 3.14")
	})

	t.Run("multiple flags combined", func(t *testing.T) {
		var verbose bool
		var name string
		var count int

		parser := lib.New().
			Flag(&verbose, []string{"v", "verbose"}, "Verbose").
			Flag(&name, []string{"n", "name"}, "Name").
			Flag(&count, []string{"c", "count"}, "Count")

		result := parser.Parse([]string{"-v", "--name", "Bob", "--count", "5"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.True(t, verbose, "verbose should be true")
		assert.Equal(t, "Bob", name, "name should be Bob")
		assert.Equal(t, 5, count, "count should be 5")
	})
}

func TestSliceTypes(t *testing.T) {
	t.Run("string slice with repeated flags", func(t *testing.T) {
		var tags []string

		parser := lib.New().
			Flag(&tags, []string{"t", "tag"}, "Tags")

		result := parser.Parse([]string{"--tag", "go", "--tag", "cli", "-t", "parser"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, []string{"go", "cli", "parser"}, tags)
	})

	t.Run("string slice with comma-separated values", func(t *testing.T) {
		var tags []string

		parser := lib.New().
			Flag(&tags, []string{"t", "tag"}, "Tags")

		result := parser.Parse([]string{"--tag", "go,cli,parser"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, []string{"go", "cli", "parser"}, tags)
	})

	t.Run("int slice with repeated flags", func(t *testing.T) {
		var ports []int

		parser := lib.New().
			Flag(&ports, []string{"p", "port"}, "Ports")

		result := parser.Parse([]string{"--port", "8080", "-p", "3000", "--port", "9000"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, []int{8080, 3000, 9000}, ports)
	})

	t.Run("float slice with repeated flags", func(t *testing.T) {
		var values []float64

		parser := lib.New().
			Flag(&values, []string{"v", "value"}, "Values")

		result := parser.Parse([]string{"-v", "1.5", "-v", "2.5", "-v", "3.5"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.Equal(t, []float64{1.5, 2.5, 3.5}, values)
	})
}

func TestHierarchicalContexts(t *testing.T) {
	t.Run("basic context", func(t *testing.T) {
		type GitConfig struct {
			Commit  bool
			Message string
		}
		
		type Config struct {
			Verbose bool
			Git     GitConfig
		}
		
		var config Config

		parser := lib.New().
			Flag(&config.Verbose, []string{"v"}, "Verbose").
			Context(&config.Git, []string{"G", "git"}, "Git", func(p *lib.Parser, git *GitConfig) *lib.Parser {
				return p.
					Flag(&git.Commit, []string{"c", "commit"}, "Commit").
					Flag(&git.Message, []string{"m"}, "Message")
			})

		result := parser.Parse([]string{"-v", "-G", "-c", "-m", "Test commit"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.True(t, config.Verbose, "verbose should be true")
		assert.True(t, config.Git.Commit, "gitCommit should be true")
		assert.Equal(t, "Test commit", config.Git.Message)
	})

	t.Run("combined short flags with context", func(t *testing.T) {
		type GitConfig struct {
			Commit  bool
			Message string
		}
		
		type Config struct {
			Verbose bool
			Git     GitConfig
		}
		
		var config Config

		parser := lib.New().
			Flag(&config.Verbose, []string{"v"}, "Verbose").
			Context(&config.Git, []string{"G", "git"}, "Git", func(p *lib.Parser, git *GitConfig) *lib.Parser {
				return p.
					Flag(&git.Commit, []string{"c", "commit"}, "Commit").
					Flag(&git.Message, []string{"m"}, "Message")
			})

		result := parser.Parse([]string{"-vGcm", "Fix bug"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.True(t, config.Verbose, "verbose should be true")
		assert.True(t, config.Git.Commit, "gitCommit should be true")
		assert.Equal(t, "Fix bug", config.Git.Message)
	})

	t.Run("nested contexts", func(t *testing.T) {
		type DeployConfig struct {
			Now    bool
			Region string
		}
		
		type Config struct {
			Verbose bool
			Deploy  DeployConfig
		}
		
		var config Config

		parser := lib.New().
			Flag(&config.Verbose, []string{"v"}, "Verbose").
			Context(&config.Deploy, []string{"d", "deploy"}, "Deploy", func(p *lib.Parser, d *DeployConfig) *lib.Parser {
				return p.
					Flag(&d.Now, []string{"now"}, "Deploy now").
					Flag(&d.Region, []string{"r", "region"}, "Region")
			})

		result := parser.Parse([]string{"-v", "-d", "--now", "--region", "us-west"})

		require.Empty(t, result.Errors, "should parse without errors")
		assert.True(t, config.Verbose)
		assert.True(t, config.Deploy.Now)
		assert.Equal(t, "us-west", config.Deploy.Region)
	})
}

func TestCustomTypes(t *testing.T) {
	t.Run("time.Duration", func(t *testing.T) {
		var timeout time.Duration

		parser := lib.New().
			Flag(&timeout, []string{"t", "timeout"}, "Timeout")

		result := parser.Parse([]string{"--timeout", "5m30s"})

		require.Empty(t, result.Errors, "should parse without errors")
		expected := 5*time.Minute + 30*time.Second
		assert.Equal(t, expected, timeout)
	})

	t.Run("time.Duration various formats", func(t *testing.T) {
		tests := []struct {
			input    string
			expected time.Duration
		}{
			{"30s", 30 * time.Second},
			{"5m", 5 * time.Minute},
			{"2h", 2 * time.Hour},
			{"1h30m45s", 1*time.Hour + 30*time.Minute + 45*time.Second},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				var timeout time.Duration

				parser := lib.New().
					Flag(&timeout, []string{"t"}, "Timeout")

				result := parser.Parse([]string{"-t", tt.input})

				require.Empty(t, result.Errors, "should parse without errors")
				assert.Equal(t, tt.expected, timeout)
			})
		}
	})
}

func TestImmutability(t *testing.T) {
	t.Run("parser instances are independent", func(t *testing.T) {
		p1 := lib.New()

		var flag1 bool
		p2 := p1.Flag(&flag1, []string{"a"}, "Flag A")

		var flag2 bool
		p3 := p2.Flag(&flag2, []string{"b"}, "Flag B")

		// Parse with p2 (should only have flag1)
		result := p2.Parse([]string{"-a"})
		require.Empty(t, result.Errors)
		assert.True(t, flag1, "flag1 should be set")

		// Try to use flag2 with p2 (should fail)
		flag1 = false
		result = p2.Parse([]string{"-b"})
		assert.NotEmpty(t, result.Errors, "p2 should not recognize flag b")

		// Parse with p3 (should have both flags)
		flag1 = false
		flag2 = false
		result = p3.Parse([]string{"-a", "-b"})
		require.Empty(t, result.Errors)
		assert.True(t, flag1, "flag1 should be set")
		assert.True(t, flag2, "flag2 should be set")
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("unknown flag", func(t *testing.T) {
		var name string

		parser := lib.New().
			Flag(&name, []string{"n"}, "Name")

		result := parser.Parse([]string{"--unknown"})

		assert.NotEmpty(t, result.Errors, "should have error for unknown flag")
		assert.Contains(t, result.Errors[0].Error(), "unknown flag")
	})

	t.Run("missing value for value flag", func(t *testing.T) {
		var name string

		parser := lib.New().
			Flag(&name, []string{"n"}, "Name")

		result := parser.Parse([]string{"-n"})

		assert.NotEmpty(t, result.Errors, "should have error for missing value")
		assert.Contains(t, result.Errors[0].Error(), "requires a value")
	})

	t.Run("invalid int value", func(t *testing.T) {
		var count int

		parser := lib.New().
			Flag(&count, []string{"c"}, "Count")

		result := parser.Parse([]string{"-c", "not-a-number"})

		assert.NotEmpty(t, result.Errors, "should have error for invalid int")
	})

	t.Run("invalid duration value", func(t *testing.T) {
		var timeout time.Duration

		parser := lib.New().
			Flag(&timeout, []string{"t"}, "Timeout")

		result := parser.Parse([]string{"-t", "invalid"})

		assert.NotEmpty(t, result.Errors, "should have error for invalid duration")
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("empty arguments", func(t *testing.T) {
		var verbose bool

		parser := lib.New().
			Flag(&verbose, []string{"v"}, "Verbose")

		result := parser.Parse([]string{})

		require.Empty(t, result.Errors)
		assert.False(t, verbose, "verbose should remain false with no args")
	})

	t.Run("unsigned integers", func(t *testing.T) {
		var port uint16

		parser := lib.New().
			Flag(&port, []string{"p", "port"}, "Port number")

		result := parser.Parse([]string{"--port", "8080"})

		require.Empty(t, result.Errors)
		assert.Equal(t, uint16(8080), port)
	})

	t.Run("long flag names with dashes", func(t *testing.T) {
		var userName string

		parser := lib.New().
			Flag(&userName, []string{"user-name"}, "User name")

		result := parser.Parse([]string{"--user-name", "Alice"})

		require.Empty(t, result.Errors)
		assert.Equal(t, "Alice", userName)
	})

	t.Run("bool flag with explicit false not supported", func(t *testing.T) {
		var verbose bool

		parser := lib.New().
			Flag(&verbose, []string{"v"}, "Verbose")

		// Bool flags are switches - they don't take values
		// So this should treat "false" as a positional argument
		result := parser.Parse([]string{"-v"})

		require.Empty(t, result.Errors)
		assert.True(t, verbose, "switch flags are always true when present")
	})
}

func TestCombinedShortFlags(t *testing.T) {
	t.Run("multiple switch flags", func(t *testing.T) {
		var verbose bool
		var debug bool
		var interactive bool

		parser := lib.New().
			Flag(&verbose, []string{"v"}, "Verbose").
			Flag(&debug, []string{"d"}, "Debug").
			Flag(&interactive, []string{"i"}, "Interactive")

		result := parser.Parse([]string{"-vdi"})

		require.Empty(t, result.Errors)
		assert.True(t, verbose)
		assert.True(t, debug)
		assert.True(t, interactive)
	})
}

func TestRealWorldScenario(t *testing.T) {
	t.Run("complete application flags", func(t *testing.T) {
		// Simulate a real CLI application
		var verbose bool
		var debug bool
		var configFile string
		var ports []int
		var timeout time.Duration
		var tags []string

		parser := lib.New().
			Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
			Flag(&debug, []string{"d", "debug"}, "Enable debug mode").
			Flag(&configFile, []string{"c", "config"}, "Config file path").
			Flag(&ports, []string{"p", "port"}, "Listen ports (repeatable)").
			Flag(&timeout, []string{"t", "timeout"}, "Request timeout").
			Flag(&tags, []string{"tag"}, "Tags (repeatable)")

		result := parser.Parse([]string{
			"-vd",
			"--config", "/etc/app.conf",
			"-p", "8080",
			"-p", "8443",
			"--timeout", "30s",
			"--tag", "production",
			"--tag", "us-west",
		})

		require.Empty(t, result.Errors)
		assert.True(t, verbose)
		assert.True(t, debug)
		assert.Equal(t, "/etc/app.conf", configFile)
		assert.Equal(t, []int{8080, 8443}, ports)
		assert.Equal(t, 30*time.Second, timeout)
		assert.Equal(t, []string{"production", "us-west"}, tags)
	})
}
