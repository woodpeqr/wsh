package cli

import (
	"encoding/json"
	"reflect"
	"testing"

	"V-Woodpecker-V/wsh/warg/flags"
)

func TestParseInlineDefinition(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantParent  string
		wantNames   []string
		wantSwitch  bool
		wantDesc    string
		wantErr     bool
	}{
		{
			name:       "simple bool flag",
			input:      "v,verbose;bool;Verbose output",
			wantParent: "",
			wantNames:  []string{"-v", "--verbose"},
			wantSwitch: true,
			wantDesc:   "Verbose output",
			wantErr:    false,
		},
		{
			name:       "string flag with single name",
			input:      "name;string;User name",
			wantParent: "",
			wantNames:  []string{"--name"},
			wantSwitch: false,
			wantDesc:   "User name",
			wantErr:    false,
		},
		{
			name:       "context flag",
			input:      "G,git;context;Git operations",
			wantParent: "",
			wantNames:  []string{"-G", "--git"},
			wantSwitch: true,
			wantDesc:   "Git operations",
			wantErr:    false,
		},
		{
			name:       "child flag with parent",
			input:      "G.c,commit;bool;Commit changes",
			wantParent: "G",
			wantNames:  []string{"-c", "--commit"},
			wantSwitch: true,
			wantDesc:   "Commit changes",
			wantErr:    false,
		},
		{
			name:    "invalid format - missing parts",
			input:   "v,verbose;bool",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, parent, err := ParseInlineDefinition(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInlineDefinition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if parent != tt.wantParent {
				t.Errorf("ParseInlineDefinition() parent = %v, want %v", parent, tt.wantParent)
			}
			if !reflect.DeepEqual(got.Names, tt.wantNames) {
				t.Errorf("ParseInlineDefinition() names = %v, want %v", got.Names, tt.wantNames)
			}
			if got.Switch != tt.wantSwitch {
				t.Errorf("ParseInlineDefinition() switch = %v, want %v", got.Switch, tt.wantSwitch)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("ParseInlineDefinition() desc = %v, want %v", got.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseInlineDefinitions(t *testing.T) {
	defs := []string{
		"n,name;string;User name",
		"v,verbose;bool;Verbose output",
		"G,git;context;Git operations",
		"G.c,commit;bool;Commit changes",
		"G.m,message;string;Commit message",
	}

	result, err := ParseInlineDefinitions(defs)
	if err != nil {
		t.Fatalf("ParseInlineDefinitions() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("ParseInlineDefinitions() returned %d flags, want 3", len(result))
	}

	// Find git flag and verify children
	var gitFlag *flags.FlagDefinition
	for i := range result {
		for _, name := range result[i].Names {
			if name == "-G" || name == "--git" {
				gitFlag = &result[i]
				break
			}
		}
	}

	if gitFlag == nil {
		t.Fatal("git flag not found")
	}

	if len(gitFlag.Children) != 2 {
		t.Errorf("git flag has %d children, want 2", len(gitFlag.Children))
	}
	
	// Context flags with children should be switches
	if !gitFlag.Switch {
		t.Errorf("git flag with children should have Switch=true")
	}
}

func TestParseJSONDefinitions(t *testing.T) {
	jsonData := `{
		"flags": [
			{
				"names": ["-n", "--name"],
				"switch": false,
				"desc": "User name"
			},
			{
				"names": ["-G", "--git"],
				"switch": true,
				"desc": "Git operations",
				"children": [
					{
						"names": ["-c", "--commit"],
						"switch": true,
						"desc": "Commit changes"
					}
				]
			}
		]
	}`

	result, err := ParseJSONDefinitions([]byte(jsonData))
	if err != nil {
		t.Fatalf("ParseJSONDefinitions() error = %v", err)
	}

	if len(result) != 2 {
		t.Errorf("ParseJSONDefinitions() returned %d flags, want 2", len(result))
	}

	// Verify git flag has children
	var gitFlag *flags.FlagDefinition
	for i := range result {
		if result[i].Switch && len(result[i].Children) > 0 {
			gitFlag = &result[i]
			break
		}
	}

	if gitFlag == nil {
		t.Fatal("context flag not found")
	}

	if len(gitFlag.Children) != 1 {
		t.Errorf("context flag has %d children, want 1", len(gitFlag.Children))
	}
}

func TestParseHeredocDefinition(t *testing.T) {
	input := `-n, --name [value] User name
-v, --verbose Verbose output
-G, --git Git operations
  -c, --commit Commit changes
  -m, --message [string] Commit message`

	result, err := ParseHeredocDefinition(input)
	if err != nil {
		t.Fatalf("ParseHeredocDefinition() error = %v", err)
	}

	if len(result) != 3 {
		t.Errorf("ParseHeredocDefinition() returned %d flags, want 3", len(result))
	}

	// Find git flag and verify children
	var gitFlag *flags.FlagDefinition
	for i := range result {
		if result[i].Switch && len(result[i].Children) > 0 {
			gitFlag = &result[i]
			break
		}
	}

	if gitFlag == nil {
		t.Fatal("context flag not found")
	}

	if len(gitFlag.Children) != 2 {
		t.Errorf("context flag has %d children, want 2", len(gitFlag.Children))
	}
	
	// Parent with children should be switch
	if !gitFlag.Switch {
		t.Errorf("git flag with children should have Switch=true")
	}
}

func TestFlagDefinitionJSON(t *testing.T) {
	def := flags.FlagDefinition{
		Names:       []string{"-n", "--name"},
		Switch:      false,
		Description: "User name",
		Children: []flags.FlagDefinition{
			{
				Names:       []string{"-c", "--child"},
				Switch:      true,
				Description: "Child flag",
			},
		},
	}

	// Test marshaling
	data, err := json.Marshal(def)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Test unmarshaling
	var result flags.FlagDefinition
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(def, result) {
		t.Errorf("JSON roundtrip failed: got %+v, want %+v", result, def)
	}
}
