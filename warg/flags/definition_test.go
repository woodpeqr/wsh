package flags

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFlagDefinitionJSON(t *testing.T) {
	def := FlagDefinition{
		Names:       []string{"-n", "--name"},
		Switch:      false,
		Description: "User name",
		Children: []FlagDefinition{
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
	var result FlagDefinition
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !reflect.DeepEqual(def, result) {
		t.Errorf("JSON roundtrip failed: got %+v, want %+v", result, def)
	}
}
