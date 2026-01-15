// Package lib provides a functional, type-inferred flag parsing library for Go programs.
//
// This package offers a clean API for parsing command-line flags with automatic type
// detection, support for slices, hierarchical contexts, and functional composition.
//
// Basic usage:
//
//	var verbose bool
//	var name string
//	parser := lib.New().
//		Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
//		Flag(&name, []string{"n", "name"}, "User name")
//	result := parser.Parse(os.Args[1:])
//
// The Parser is immutable - each method returns a new instance, allowing for
// functional composition and reusable parser fragments.
//
// Supported types include all basic Go types (bool, string, int, float, etc.),
// slices for repeatable flags, time.Duration, and any type implementing
// encoding.TextUnmarshaler.
package lib

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"V-Woodpecker-V/wsh/warg/flags"
	"V-Woodpecker-V/wsh/warg/internal/parser"
)

// Parser is an immutable flag parser builder.
// Each method returns a new Parser instance, enabling functional composition.
type Parser struct {
	definitions    []flags.FlagDefinition
	setters        []setter
	sliceContexts  map[string]*sliceContextInfo
}

// setter represents a function that sets a value on a user variable
type setter struct {
	names []string
	fn    func(string) error
}

// sliceContextInfo tracks information for repeatable slice contexts
type sliceContextInfo struct {
	slicePtr     interface{}
	elemType     reflect.Type
	builder      interface{}
	childSetters []setter
	currentIndex int
}

// ParseResult contains the result of parsing arguments.
// Check the Errors slice to see if parsing succeeded.
type ParseResult struct {
	Errors []error
}

// New creates a new Parser instance with no flags defined.
// Use Flag() and Context() methods to build up the parser.
func New() *Parser {
	return &Parser{
		definitions:   []flags.FlagDefinition{},
		setters:       []setter{},
		sliceContexts: make(map[string]*sliceContextInfo),
	}
}

// Flag registers a flag with automatic type inference.
//
// The ptr parameter must be a pointer to a variable. The type is automatically
// detected and the appropriate parser is created. When Parse() is called,
// the variable will be updated with the parsed value.
//
// The names parameter contains flag names without dashes. Single-character names
// become short flags (-v), multi-character names become long flags (--verbose).
//
// Supported types:
//   - bool: switch flags (no value required)
//   - string, int, uint, float: value flags
//   - []string, []int, etc.: repeatable flags
//   - time.Duration: special handling
//   - encoding.TextUnmarshaler: custom types
//
// Example:
//
//	var verbose bool
//	var name string
//	var tags []string
//	parser := lib.New().
//		Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
//		Flag(&name, []string{"n", "name"}, "User name").
//		Flag(&tags, []string{"t", "tag"}, "Tags (repeatable)")
//
// Returns a new Parser instance (this method does not modify the receiver).
func (p *Parser) Flag(ptr interface{}, names []string, description string) *Parser {
	// Detect type and create appropriate setter
	setterFn, isSwitch, err := createSetter(ptr)
	if err != nil {
		panic(fmt.Sprintf("Flag: %v", err))
	}

	// Normalize flag names by adding dashes
	normalizedNames := make([]string, len(names))
	for i, name := range names {
		normalizedNames[i] = normalizeFlagName(name)
	}

	// Create new parser with added definition
	newDefs := make([]flags.FlagDefinition, len(p.definitions)+1)
	copy(newDefs, p.definitions)
	newDefs[len(p.definitions)] = flags.FlagDefinition{
		Names:       normalizedNames,
		Switch:      isSwitch,
		Description: description,
		Children:    []flags.FlagDefinition{},
	}

	// Create new setters slice
	newSetters := make([]setter, len(p.setters)+1)
	copy(newSetters, p.setters)
	newSetters[len(p.setters)] = setter{
		names: normalizedNames,
		fn:    setterFn,
	}

	// Copy slice contexts
	newSliceContexts := make(map[string]*sliceContextInfo)
	for k, v := range p.sliceContexts {
		newSliceContexts[k] = v
	}

	return &Parser{
		definitions:   newDefs,
		setters:       newSetters,
		sliceContexts: newSliceContexts,
	}
}

// Context creates a hierarchical context with child flags.
//
// Contexts allow you to group related flags together. The context itself acts
// as a switch flag, and when present, enables its child flags.
//
// The ptr parameter can be either:
//   - A pointer to a slice (*[]T) for repeatable contexts (each context invocation appends a new item)
//   - A pointer to a struct (*T) for single contexts
//
// The builder function receives a new Parser and the parent pointer, and should return
// the Parser with child flags added.
//
// Example (single context):
//
//	type GitConfig struct {
//		Commit  bool
//		Message string
//	}
//	var git GitConfig
//	parser := lib.New().
//		Context(&git, []string{"G", "git"}, "Git operations", 
//			func(p *Parser, git *GitConfig) *Parser {
//				return p.
//					Flag(&git.Commit, []string{"c", "commit"}, "Commit").
//					Flag(&git.Message, []string{"m"}, "Message")
//			})
//
// Example (repeatable context):
//
//	type AddFlagDef struct {
//		Names       string
//		Description string
//	}
//	var addFlags []AddFlagDef
//	parser := lib.New().
//		Context(&addFlags, []string{"A", "add"}, "Add flag",
//			func(p *Parser, def *AddFlagDef) *Parser {
//				return p.
//					Flag(&def.Names, []string{"n"}, "Names").
//					Flag(&def.Description, []string{"d"}, "Description")
//			})
//
// Returns a new Parser instance (this method does not modify the receiver).
func (p *Parser) Context(ptr interface{}, names []string, description string, builder interface{}) *Parser {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Context: expected pointer, got %T", ptr))
	}

	elem := v.Elem()
	
	// Normalize flag names by adding dashes
	normalizedNames := make([]string, len(names))
	for i, name := range names {
		normalizedNames[i] = normalizeFlagName(name)
	}

	// Detect if this is a slice context or single context
	if elem.Kind() == reflect.Slice {
		return p.contextSlice(ptr, normalizedNames, description, builder)
	}
	return p.contextSingle(ptr, normalizedNames, description, builder)
}

// contextSingle handles non-repeatable struct contexts
func (p *Parser) contextSingle(ptr interface{}, normalizedNames []string, description string, builder interface{}) *Parser {
	// The builder should have signature: func(*Parser, *T) *Parser
	builderVal := reflect.ValueOf(builder)
	if builderVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Context: builder must be a function, got %T", builder))
	}

	// Call the builder with New() parser and the struct pointer
	results := builderVal.Call([]reflect.Value{
		reflect.ValueOf(New()),
		reflect.ValueOf(ptr),
	})
	
	if len(results) != 1 {
		panic("Context: builder must return exactly one value (*Parser)")
	}
	
	subParser, ok := results[0].Interface().(*Parser)
	if !ok {
		panic("Context: builder must return *Parser")
	}

	// Create new parser with added context definition
	newDefs := make([]flags.FlagDefinition, len(p.definitions)+1)
	copy(newDefs, p.definitions)
	newDefs[len(p.definitions)] = flags.FlagDefinition{
		Names:       normalizedNames,
		Switch:      true,
		Description: description,
		Children:    subParser.definitions,
	}

	// Merge setters - context setters are for child flags
	newSetters := make([]setter, len(p.setters)+len(subParser.setters))
	copy(newSetters, p.setters)
	copy(newSetters[len(p.setters):], subParser.setters)

	// Copy slice contexts
	newSliceContexts := make(map[string]*sliceContextInfo)
	for k, v := range p.sliceContexts {
		newSliceContexts[k] = v
	}

	return &Parser{
		definitions:   newDefs,
		setters:       newSetters,
		sliceContexts: newSliceContexts,
	}
}

// contextSlice handles repeatable slice contexts
func (p *Parser) contextSlice(slicePtr interface{}, normalizedNames []string, description string, builder interface{}) *Parser {
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	elemType := sliceVal.Type().Elem()
	
	// The builder should have signature: func(*Parser, *T) *Parser where T is the slice element type
	builderVal := reflect.ValueOf(builder)
	if builderVal.Kind() != reflect.Func {
		panic(fmt.Sprintf("Context: builder must be a function, got %T", builder))
	}

	// Create a temporary instance to build the child definitions
	tempElem := reflect.New(elemType)
	results := builderVal.Call([]reflect.Value{
		reflect.ValueOf(New()),
		tempElem,
	})
	
	if len(results) != 1 {
		panic("Context: builder must return exactly one value (*Parser)")
	}
	
	subParser, ok := results[0].Interface().(*Parser)
	if !ok {
		panic("Context: builder must return *Parser")
	}

	// Store slice context info for later use during parsing
	contextKey := normalizedNames[0]
	sliceCtxInfo := &sliceContextInfo{
		slicePtr:     slicePtr,
		elemType:     elemType,
		builder:      builder,
		childSetters: subParser.setters,
		currentIndex: -1,
	}

	// Create new parser with added context definition
	newDefs := make([]flags.FlagDefinition, len(p.definitions)+1)
	copy(newDefs, p.definitions)
	newDefs[len(p.definitions)] = flags.FlagDefinition{
		Names:       normalizedNames,
		Switch:      true,
		Description: description,
		Children:    subParser.definitions,
	}

	// Don't add child setters directly - we'll handle them specially during parsing
	newSetters := make([]setter, len(p.setters))
	copy(newSetters, p.setters)

	// Copy and add slice contexts
	newSliceContexts := make(map[string]*sliceContextInfo)
	for k, v := range p.sliceContexts {
		newSliceContexts[k] = v
	}
	newSliceContexts[contextKey] = sliceCtxInfo

	return &Parser{
		definitions:   newDefs,
		setters:       newSetters,
		sliceContexts: newSliceContexts,
	}
}

// Parse parses the given arguments according to the registered flags.
//
// Arguments are typically os.Args[1:], but can be any slice of strings.
// The method walks through the arguments, matches them against registered
// flags, and updates the associated variables via their pointers.
//
// Returns a ParseResult containing any errors encountered. Check result.Errors
// to determine if parsing succeeded:
//
//	result := parser.Parse(os.Args[1:])
//	if len(result.Errors) > 0 {
//		fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
//		os.Exit(1)
//	}
//
// Common errors include unknown flags, missing values, and type conversion failures.
func (p *Parser) Parse(args []string) *ParseResult {
	result := &ParseResult{
		Errors: []error{},
	}

	// Use the existing parser
	internalParser := parser.NewParser(p.definitions)
	parseResult, err := internalParser.Parse(args)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	// Apply setters to user variables
	if err := p.applySetters(parseResult); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result
}

// applySetters walks the parse result and applies setters
func (p *Parser) applySetters(parseResult *parser.ParseResult) error {
	// Build a map of flag names to setters for quick lookup
	setterMap := make(map[string]func(string) error)
	for _, s := range p.setters {
		for _, name := range s.names {
			setterMap[name] = s.fn
		}
	}

	// Walk all flags and apply setters
	var walkErr error
	parseResult.Walk(func(fv *parser.FlagValue) {
		if walkErr != nil {
			return
		}

		// Check if this is a slice context flag
		for _, name := range fv.Definition.Names {
			if sliceCtx, ok := p.sliceContexts[name]; ok {
				// This is a slice context - append a new element
				if fv.Present {
					sliceVal := reflect.ValueOf(sliceCtx.slicePtr).Elem()
					newElem := reflect.New(sliceCtx.elemType)
					sliceVal.Set(reflect.Append(sliceVal, newElem.Elem()))
					
					currentIndex := sliceVal.Len() - 1
					
					// Create setters for this specific element
					builderVal := reflect.ValueOf(sliceCtx.builder)
					elemPtr := sliceVal.Index(currentIndex).Addr()
					
					builderResults := builderVal.Call([]reflect.Value{
						reflect.ValueOf(New()),
						elemPtr,
					})
					
					elemParser := builderResults[0].Interface().(*Parser)
					
					// Update setter map with child setters that point to current element
					for _, s := range elemParser.setters {
						for _, childName := range s.names {
							setterMap[childName] = s.fn
						}
					}
				}
				return
			}
		}

		// Find setter for this flag
		var setterFn func(string) error
		for _, name := range fv.Definition.Names {
			if fn, ok := setterMap[name]; ok {
				setterFn = fn
				break
			}
		}

		if setterFn == nil {
			return
		}

		// Apply the setter
		if fv.Definition.Switch {
			// For switch flags, pass "true" if present
			if fv.Present {
				if err := setterFn("true"); err != nil {
					walkErr = fmt.Errorf("error setting flag %v: %w", fv.Definition.Names[0], err)
				}
			}
		} else {
			// For value flags, pass the value
			if fv.Present {
				if err := setterFn(fv.Value); err != nil {
					walkErr = fmt.Errorf("error setting flag %v: %w", fv.Definition.Names[0], err)
				}
			}
		}
	})

	return walkErr
}

// createSetter creates a setter function for the given pointer
func createSetter(ptr interface{}) (func(string) error, bool, error) {
	v := reflect.ValueOf(ptr)
	if v.Kind() != reflect.Ptr {
		return nil, false, fmt.Errorf("expected pointer, got %T", ptr)
	}

	elem := v.Elem()
	if !elem.CanSet() {
		return nil, false, fmt.Errorf("cannot set value")
	}

	// Check if it's a slice
	if elem.Kind() == reflect.Slice {
		return createSliceSetter(elem), false, nil
	}

	// Check for TextUnmarshaler interface on the pointer type
	unmarshalerType := reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	if v.Type().Implements(unmarshalerType) {
		return createTextUnmarshalerSetter(ptr.(encoding.TextUnmarshaler)), false, nil
	}

	// Special handling for time.Duration
	if elem.Type() == reflect.TypeOf(time.Duration(0)) {
		return createDurationSetter(elem), false, nil
	}

	// Handle basic types
	switch elem.Kind() {
	case reflect.Bool:
		return createBoolSetter(elem), true, nil
	case reflect.String:
		return createStringSetter(elem), false, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return createIntSetter(elem), false, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return createUintSetter(elem), false, nil
	case reflect.Float32, reflect.Float64:
		return createFloatSetter(elem), false, nil
	default:
		return nil, false, fmt.Errorf("unsupported type: %v", elem.Type())
	}
}

// createBoolSetter creates a setter for bool types
func createBoolSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		elem.SetBool(b)
		return nil
	}
}

// createStringSetter creates a setter for string types
func createStringSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		elem.SetString(value)
		return nil
	}
}

// createIntSetter creates a setter for int types
func createIntSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		elem.SetInt(i)
		return nil
	}
}

// createUintSetter creates a setter for uint types
func createUintSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		elem.SetUint(u)
		return nil
	}
}

// createFloatSetter creates a setter for float types
func createFloatSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		elem.SetFloat(f)
		return nil
	}
}

// createDurationSetter creates a setter for time.Duration
func createDurationSetter(elem reflect.Value) func(string) error {
	return func(value string) error {
		d, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		elem.Set(reflect.ValueOf(d))
		return nil
	}
}

// createSliceSetter creates a setter for slice types
func createSliceSetter(elem reflect.Value) func(string) error {
	elemType := elem.Type().Elem()

	return func(value string) error {
		// Parse the value according to element type
		var newElem reflect.Value

		switch elemType.Kind() {
		case reflect.String:
			// Check if value contains commas (comma-separated list)
			if strings.Contains(value, ",") {
				parts := strings.Split(value, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					elem.Set(reflect.Append(elem, reflect.ValueOf(part)))
				}
				return nil
			}
			newElem = reflect.ValueOf(value)

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			i, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return err
			}
			newElem = reflect.ValueOf(i).Convert(elemType)

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			u, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			newElem = reflect.ValueOf(u).Convert(elemType)

		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			newElem = reflect.ValueOf(f).Convert(elemType)

		default:
			return fmt.Errorf("unsupported slice element type: %v", elemType)
		}

		elem.Set(reflect.Append(elem, newElem))
		return nil
	}
}

// createTextUnmarshalerSetter creates a setter for types implementing encoding.TextUnmarshaler
func createTextUnmarshalerSetter(ptr encoding.TextUnmarshaler) func(string) error {
	return func(value string) error {
		return ptr.UnmarshalText([]byte(value))
	}
}

// normalizeFlagName adds appropriate dashes to a flag name
// Single character names get a single dash (-v)
// Multi-character names get double dashes (--verbose)
func normalizeFlagName(name string) string {
	// If already has dashes, return as-is
	if strings.HasPrefix(name, "-") {
		return name
	}
	
	// Single character gets single dash
	if len(name) == 1 {
		return "-" + name
	}
	
	// Multi-character gets double dash
	return "--" + name
}
