package parser

import (
	"fmt"
	"strings"

	"V-Woodpecker-V/wsh/warg/flags"
)

// FlagValue represents a parsed flag with its value and children
type FlagValue struct {
	Definition *flags.FlagDefinition `json:"definition"`
	Present    bool                  `json:"present"`
	Value      string                `json:"value"`
	Children   []*FlagValue          `json:"children"`
}

// ParseResult holds the parsed flag values
type ParseResult struct {
	Flags []*FlagValue `json:"flags"`
}

// Find searches for a flag by name in the result (searches all levels)
func (r *ParseResult) Find(name string) *FlagValue {
	for _, flag := range r.Flags {
		if found := flag.Find(name); found != nil {
			return found
		}
	}
	return nil
}

// Walk traverses all flags in the tree, calling fn for each
func (r *ParseResult) Walk(fn func(*FlagValue)) {
	for _, flag := range r.Flags {
		flag.Walk(fn)
	}
}

// Find searches for a flag by name (any of its names)
func (fv *FlagValue) Find(name string) *FlagValue {
	// Check if this flag matches
	for _, n := range fv.Definition.Names {
		if n == name {
			return fv
		}
	}
	// Search children
	for _, child := range fv.Children {
		if found := child.Find(name); found != nil {
			return found
		}
	}
	return nil
}

// Walk traverses this flag and all its children, calling fn for each
func (fv *FlagValue) Walk(fn func(*FlagValue)) {
	fn(fv)
	for _, child := range fv.Children {
		child.Walk(fn)
	}
}

// IsSwitch returns true if this is a switch flag
func (fv *FlagValue) IsSwitch() bool {
	return fv.Definition.Switch
}

// GetBool returns the Present value for switch flags
func (fv *FlagValue) GetBool() bool {
	return fv.Present
}

// GetString returns the Value for value flags
func (fv *FlagValue) GetString() string {
	return fv.Value
}

// Parser handles argument parsing with flag definitions
type Parser struct {
	rootContext *Context
}

// NewParser creates a new parser with the given flag definitions
func NewParser(defs []flags.FlagDefinition) *Parser {
	return &Parser{
		rootContext: NewContext(defs, nil),
	}
}

// Parse parses the given arguments according to the flag definitions
func (p *Parser) Parse(args []string) (*ParseResult, error) {
	result := &ParseResult{
		Flags: []*FlagValue{},
	}

	contextStack := []*Context{p.rootContext}
	parentStack := []*FlagValue{nil} // Track parent FlagValue for each context level

	i := 0
	for i < len(args) {
		arg := args[i]

		// Handle -- separator (end of flags)
		if arg == "--" {
			i++
			break
		}

		// Check if it's a flag
		if !strings.HasPrefix(arg, "-") {
			// Not a flag, treat as positional argument
			i++
			continue
		}

		// Handle combined short flags (e.g., -abc)
		if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) > 2 {
			// Split combined short flags
			for j := 1; j < len(arg); j++ {
				shortFlag := "-" + string(arg[j])
				consumed, err := p.processFlag(shortFlag, args, i+1, &contextStack, &parentStack, result)
				if err != nil {
					return nil, err
				}
				// Only advance if we consumed additional arguments
				if consumed > 0 {
					i += consumed
					// For combined flags, we can only consume args after processing all flags
					// This is a limitation - combined flags should not have value flags except the last one
					break
				}
			}
			i++
			continue
		}

		// Handle single flag
		consumed, err := p.processFlag(arg, args, i+1, &contextStack, &parentStack, result)
		if err != nil {
			return nil, err
		}
		i += 1 + consumed
	}

	return result, nil
}

// processFlag processes a single flag and returns the number of additional arguments consumed
func (p *Parser) processFlag(flag string, args []string, nextArgIdx int, contextStack *[]*Context, parentStack *[]*FlagValue, result *ParseResult) (int, error) {
	// Look up the flag in the current context (searches parents too)
	currentContext := (*contextStack)[len(*contextStack)-1]
	def := currentContext.Lookup(flag)
	if def == nil {
		return 0, fmt.Errorf("unknown flag: %s", flag)
	}

	// Determine which context level this flag belongs to
	// and find the appropriate parent FlagValue
	var targetParent *FlagValue
	
	// Search from current context backwards to find where this flag is defined
	for i := len(*contextStack) - 1; i >= 0; i-- {
		ctx := (*contextStack)[i]
		if _, ok := ctx.Flags[flag]; ok {
			if i > 0 {
				targetParent = (*parentStack)[i]
			}
			break
		}
	}

	// Create a FlagValue for this flag
	flagValue := &FlagValue{
		Definition: def,
		Present:    false,
		Value:      "",
		Children:   []*FlagValue{},
	}

	if def.Switch {
		// Switch flag (bool or context)
		flagValue.Present = true
		
		// Add to appropriate parent
		if targetParent != nil {
			targetParent.Children = append(targetParent.Children, flagValue)
		} else {
			result.Flags = append(result.Flags, flagValue)
		}
		
		// If it has children, create a new context
		if len(def.Children) > 0 {
			newContext := NewContext(def.Children, currentContext)
			*contextStack = append(*contextStack, newContext)
			*parentStack = append(*parentStack, flagValue)
		}
		return 0, nil
	} else {
		// Value flag (string)
		if nextArgIdx >= len(args) {
			return 0, fmt.Errorf("flag %s requires a value", flag)
		}
		flagValue.Present = true
		flagValue.Value = args[nextArgIdx]
		
		// Add to appropriate parent
		if targetParent != nil {
			targetParent.Children = append(targetParent.Children, flagValue)
		} else {
			result.Flags = append(result.Flags, flagValue)
		}
		
		return 1, nil
	}
}
