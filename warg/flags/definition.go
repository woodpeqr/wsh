package flags

// FlagDefinition defines a flag's structure and metadata
type FlagDefinition struct {
	Names       []string         `json:"names"`
	Switch      bool             `json:"switch"`
	Description string           `json:"desc"`
	Children    []FlagDefinition `json:"children,omitempty"`
}
