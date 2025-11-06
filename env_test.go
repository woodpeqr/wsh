package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestEnvironment_GetCurrent(t *testing.T) {
	env := NewEnvironment()
	current := env.GetCurrent()

	// Check that we have some environment variables
	if len(current) == 0 {
		t.Error("Expected some environment variables, got none")
	}

	// Check that PATH exists (should exist in all environments)
	if _, exists := current["PATH"]; !exists {
		t.Error("Expected PATH to exist in environment")
	}
}

func TestEnvironment_ParseEnvLine(t *testing.T) {
	env := NewEnvironment()

	tests := []struct {
		name      string
		line      string
		wantKey   string
		wantValue string
		wantOk    bool
	}{
		{
			name:      "simple variable",
			line:      "FOO=bar",
			wantKey:   "FOO",
			wantValue: "bar",
			wantOk:    true,
		},
		{
			name:      "variable with equals in value",
			line:      "URL=http://example.com?a=1&b=2",
			wantKey:   "URL",
			wantValue: "http://example.com?a=1&b=2",
			wantOk:    true,
		},
		{
			name:      "empty value",
			line:      "EMPTY=",
			wantKey:   "EMPTY",
			wantValue: "",
			wantOk:    true,
		},
		{
			name:      "no equals sign",
			line:      "INVALID",
			wantKey:   "",
			wantValue: "",
			wantOk:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, value, ok := env.parseEnvLine(tt.line)
			if ok != tt.wantOk {
				t.Errorf("parseEnvLine() ok = %v, want %v", ok, tt.wantOk)
			}
			if key != tt.wantKey {
				t.Errorf("parseEnvLine() key = %v, want %v", key, tt.wantKey)
			}
			if value != tt.wantValue {
				t.Errorf("parseEnvLine() value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestEnvironment_BuildExportLine(t *testing.T) {
	env := NewEnvironment()

	tests := []struct {
		name  string
		key   string
		value string
		want  string
	}{
		{
			name:  "simple export",
			key:   "FOO",
			value: "bar",
			want:  "export FOO='bar'\n",
		},
		{
			name:  "value with single quote",
			key:   "MESSAGE",
			value: "it's working",
			want:  "export MESSAGE='it'\"'\"'s working'\n",
		},
		{
			name:  "empty value",
			key:   "EMPTY",
			value: "",
			want:  "export EMPTY=''\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.buildExportLine(tt.key, tt.value)
			if got != tt.want {
				t.Errorf("buildExportLine() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnvironment_BuildExportScript(t *testing.T) {
	env := NewEnvironment()

	current := map[string]string{
		"EXISTING": "old_value",
		"KEEP":     "same_value",
	}

	new := map[string]string{
		"EXISTING": "new_value",
		"KEEP":     "same_value",
		"NEW":      "added",
	}

	script := env.BuildExportScript(current, new)

	// Should export EXISTING (changed) and NEW (added)
	// Should not export KEEP (unchanged)
	if !strings.Contains(script, "export EXISTING='new_value'") {
		t.Error("Expected script to export changed EXISTING variable")
	}
	if !strings.Contains(script, "export NEW='added'") {
		t.Error("Expected script to export new NEW variable")
	}
	if strings.Contains(script, "export KEEP") {
		t.Error("Expected script to NOT export unchanged KEEP variable")
	}
}

func TestEnvironment_Merge(t *testing.T) {
	env := NewEnvironment()

	envChan := make(chan map[string]string, 3)
	envChan <- map[string]string{"A": "1", "B": "2"}
	envChan <- map[string]string{"C": "3", "A": "overwrite"}
	envChan <- map[string]string{"D": "4"}
	close(envChan)

	merged := env.Merge(envChan)

	if len(merged) != 4 {
		t.Errorf("Expected 4 variables, got %d", len(merged))
	}

	// A should be overwritten
	if merged["A"] != "overwrite" && merged["A"] != "1" {
		t.Errorf("Expected A to be either '1' or 'overwrite', got %q", merged["A"])
	}

	if merged["B"] != "2" || merged["C"] != "3" || merged["D"] != "4" {
		t.Errorf("Expected B=2, C=3, D=4, got B=%q, C=%q, D=%q", merged["B"], merged["C"], merged["D"])
	}
}

func TestEnvironment_ParseNullDelimited(t *testing.T) {
	env := NewEnvironment()

	// Create null-delimited environment output
	data := "FOO=bar\x00BAZ=qux\x00URL=http://example.com\x00"
	reader := bytes.NewBufferString(data)

	result, err := env.parseNullDelimited(reader)
	if err != nil {
		t.Fatalf("parseNullDelimited() error = %v", err)
	}

	expected := map[string]string{
		"FOO": "bar",
		"BAZ": "qux",
		"URL": "http://example.com",
	}

	if len(result) != len(expected) {
		t.Errorf("Expected %d variables, got %d", len(expected), len(result))
	}

	for key, value := range expected {
		if result[key] != value {
			t.Errorf("Expected %s=%s, got %s=%s", key, value, key, result[key])
		}
	}
}

func TestScanNullTerminated(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		atEOF   bool
		wantAdv int
		wantTok string
	}{
		{
			name:    "single token",
			data:    []byte("FOO=bar\x00"),
			atEOF:   false,
			wantAdv: 8,
			wantTok: "FOO=bar",
		},
		{
			name:    "empty at EOF",
			data:    []byte{},
			atEOF:   true,
			wantAdv: 0,
			wantTok: "",
		},
		{
			name:    "no null at EOF",
			data:    []byte("FOO=bar"),
			atEOF:   true,
			wantAdv: 7,
			wantTok: "FOO=bar",
		},
		{
			name:    "incomplete without EOF",
			data:    []byte("FOO=bar"),
			atEOF:   false,
			wantAdv: 0,
			wantTok: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			advance, token, err := scanNullTerminated(tt.data, tt.atEOF)
			if err != nil {
				t.Errorf("scanNullTerminated() error = %v", err)
			}
			if advance != tt.wantAdv {
				t.Errorf("scanNullTerminated() advance = %v, want %v", advance, tt.wantAdv)
			}
			if string(token) != tt.wantTok {
				t.Errorf("scanNullTerminated() token = %q, want %q", string(token), tt.wantTok)
			}
		})
	}
}

// Integration test: ExecuteAndCapture
func TestEnvironment_ExecuteAndCapture(t *testing.T) {
	// Skip if zsh not available
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	env := NewEnvironment()

	// Create a temporary script
	tmpfile, err := os.CreateTemp("", "test_script_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	script := `#!/bin/zsh
export TEST_VAR="test_value"
export ANOTHER_VAR="another_value"
`
	if _, err := tmpfile.Write([]byte(script)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	result, err := env.ExecuteAndCapture(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("ExecuteAndCapture() error = %v", err)
	}

	if result["TEST_VAR"] != "test_value" {
		t.Errorf("Expected TEST_VAR=test_value, got %s", result["TEST_VAR"])
	}

	if result["ANOTHER_VAR"] != "another_value" {
		t.Errorf("Expected ANOTHER_VAR=another_value, got %s", result["ANOTHER_VAR"])
	}
}
