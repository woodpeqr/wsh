# Solution: Functional Flag Parsing for Go Libraries (Type-Inferred)

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Quick Start for New Readers](#quick-start-for-new-readers)
3. [Problem Statement](#problem-statement)
4. [Solution Design](#solution-design)
   - [Core Concept](#core-concept-type-inferred-functional-builder)
   - [Implementation Details](#1-immutable-flag-definitions-reuses-existing-structure)
5. [Usage Examples](#5-complete-usage-examples)
6. [Supported Types](#supported-types)
7. [Key Architectural Insight](#key-architectural-insight-flagvaluevalue-as-the-bridge)
8. [Architecture Convergence](#architecture-convergence-cli-and-library-paths)
9. [Implementation Phases](#implementation-phases)
10. [Comparison with Standard Library](#differences-from-standard-library)
11. [Trade-offs](#trade-offs)
12. [Common Questions (FAQ)](#common-questions-faq)
13. [Conclusion](#conclusion)

## Executive Summary

This document proposes a **type-inferred functional flag parsing API** for Go libraries that:
- Uses a **single `Flag()` method** that automatically detects types from pointers
- Maintains **familiar UX** similar to the standard `flag` package
- Follows **functional programming** paradigm (immutability, composition)
- **Reuses 100% of existing parser code** - no refactoring needed
- Supports **advanced features**: slices, contexts, custom types
- **Estimated effort**: ~700 lines of new code, 5-6 days

**Key Innovation**: Parser produces `FlagValue.Value` as strings (CLI JSON), library consumes those same strings via type-aware setters (Go variables).

## Quick Start for New Readers

If you're new to this architecture, here's what you need to know:

### What We're Building

```go
// User writes this simple code:
var verbose bool
var name string
var ports []int

parser := flags.New().
    Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
    Flag(&name, []string{"n", "name"}, "User name").
    Flag(&ports, []string{"p", "port"}, "Ports (repeatable)")

parser.Parse(os.Args[1:])
// Now verbose, name, and ports are populated!
```

### How It Works (3 Steps)

1. **Type Inference** - Library detects `bool`, `string`, `[]int` from pointers
2. **Same Parser** - Uses existing `parser.Parse()` (CLI and Library share code)
3. **Setter Application** - Converts strings → types, updates variables

### Why This Is Simple

- ✅ **No parser changes** - Existing parser already does everything we need
- ✅ **Strings are key** - All CLI args are strings, setters convert to types
- ✅ **Slices just work** - Parser creates multiple `FlagValue`s, setter accumulates

**Read on for detailed design...**

### Visual Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                        USER CODE                                │
│  var ports []int                                                │
│  parser := flags.New().Flag(&ports, []string{"p"}, "Ports")    │
│  parser.Parse([]string{"--port", "80", "--port", "8080"})      │
└────────────────────────────┬────────────────────────────────────┘
                             │
                ┌────────────┴────────────┐
                │  Type Inference         │
                │  Detects: *[]int        │
                │  Creates: setter func   │
                └────────────┬────────────┘
                             │
                ┌────────────┴────────────────────────────────┐
                │  Parser (Existing Code - No Changes!)       │
                │  Creates: FlagValue{Value: "80"}            │
                │           FlagValue{Value: "8080"}          │
                └────────────┬────────────────────────────────┘
                             │
                ┌────────────┴────────────────────────────────┐
                │  Setter Application (New Code)              │
                │  Call 1: setter("80")   → parse → append    │
                │  Call 2: setter("8080") → parse → append    │
                └────────────┬────────────────────────────────┘
                             │
                             ▼
                   ports = []int{80, 8080} ✅
```

---

## Problem Statement

Design a flag parsing API for Go libraries that:
1. Provides similar UX to the standard `flag` package (variable + pointer passing)
2. **Uses type inference** - no need to specify Bool(), String(), etc.
3. Follows functional programming paradigm
4. Supports warg's unique features (hierarchical contexts, short flag combining, context resolution)
5. **Reuses existing CLI parsing infrastructure** - same `FlagDefinition` structure
6. Supports common use cases including slice types
7. Remains composable and testable

## Key Constraints

- User experience should feel familiar to Go developers using the standard library
- Must follow functional paradigm (immutability, pure functions, composition)
- Must support warg's advanced features (contexts, hierarchical flags, resolution)
- Must infer type from pointer (using reflection)
- **Must produce same `[]FlagDefinition` as CLI parser** for consistency
- CLI and Library paths should converge to same parsing logic

## Solution Design

### Core Concept: Type-Inferred Functional Builder

Use reflection to infer types from pointers, eliminating the need for type-specific methods:

```go
// Standard flag package (imperative, mutable)
var verbose bool
flag.BoolVar(&verbose, "verbose", false, "Enable verbose")
flag.Parse()

// Warg functional approach with TYPE INFERENCE
var verbose bool
var name string
var count int

parser := flags.New().
    Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
    Flag(&name, []string{"n", "name"}, "User name").
    Flag(&count, []string{"c", "count"}, "Count value").
    Parse(os.Args[1:])

// Type is automatically inferred from the pointer!
```

### 1. Immutable Flag Definitions (Reuses Existing Structure)

**Critical Design Decision**: Library and CLI parsers must produce identical `FlagDefinition` structures.

```go
package flags

import (
    "encoding"
    "fmt"
    "reflect"
    "strconv"
    "strings"
)

// Parser is immutable - each method returns a new Parser
// It builds the SAME []FlagDefinition that CLI parser produces
type Parser struct {
    defs []FlagDefinition  // Reuses existing type!
    setters []setter       // Internal: captures how to set values
}

// setter encapsulates the logic to set a value on a user pointer
type setter struct {
    names   []string
    fn      func(string) error
    isSlice bool  // NEW: track if this is a slice type
}

// New creates a new empty parser
func New() *Parser {
    return &Parser{
        defs:    []FlagDefinition{},
        setters: []setter{},
    }
}
```

### 2. Type-Inferred Flag Registration

Single `Flag()` method that uses reflection to determine type:

```go
// Flag registers a flag with automatic type inference
func (p *Parser) Flag(ptr interface{}, names []string, desc string) *Parser {
    t := reflect.TypeOf(ptr)
    
    if t.Kind() != reflect.Ptr {
        panic("Flag() requires a pointer")
    }
    
    elemType := t.Elem()
    
    // Create FlagDefinition based on type
    def := FlagDefinition{
        Names:       names,
        Description: desc,
        Children:    []FlagDefinition{},
    }
    
    var setterFn func(string) error
    var isSlice bool  // Track if this is a slice type
    
    // Type inference based on pointer element type
    switch elemType.Kind() {
    case reflect.Bool:
        def.Switch = true
        setterFn = makeBoolSetter(ptr.(*bool))
        
    case reflect.String:
        def.Switch = false
        setterFn = makeStringSetter(ptr.(*string))
        
    case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
        def.Switch = false
        setterFn = makeIntSetter(ptr, elemType)
        
    case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
        def.Switch = false
        setterFn = makeUintSetter(ptr, elemType)
        
    case reflect.Float64, reflect.Float32:
        def.Switch = false
        setterFn = makeFloatSetter(ptr, elemType)
        
    case reflect.Slice:
        // Slice support! Handle []string, []int, etc.
        def.Switch = false
        setterFn = makeSliceSetter(ptr, elemType)
        isSlice = true  // Mark as slice type
        
    default:
        // Check if it implements TextUnmarshaler (like time.Duration)
        if elemType.Implements(reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()) {
            def.Switch = false
            setterFn = makeTextUnmarshalerSetter(ptr)
        } else {
            panic(fmt.Sprintf("unsupported type: %v", elemType))
        }
    }
    
    return p.withFlag(def, setter{names: names, fn: setterFn, isSlice: isSlice})
}

// Helper functions for creating setters

func makeBoolSetter(ptr *bool) func(string) error {
    return func(value string) error {
        if value == "" {
            *ptr = true
            return nil
        }
        val, err := strconv.ParseBool(value)
        if err != nil {
            return err
        }
        *ptr = val
        return nil
    }
}

func makeStringSetter(ptr *string) func(string) error {
    return func(value string) error {
        *ptr = value
        return nil
    }
}

func makeIntSetter(ptr interface{}, elemType reflect.Type) func(string) error {
    return func(value string) error {
        val, err := strconv.ParseInt(value, 10, elemType.Bits())
        if err != nil {
            return err
        }
        reflect.ValueOf(ptr).Elem().SetInt(val)
        return nil
    }
}

func makeUintSetter(ptr interface{}, elemType reflect.Type) func(string) error {
    return func(value string) error {
        val, err := strconv.ParseUint(value, 10, elemType.Bits())
        if err != nil {
            return err
        }
        reflect.ValueOf(ptr).Elem().SetUint(val)
        return nil
    }
}

func makeFloatSetter(ptr interface{}, elemType reflect.Type) func(string) error {
    return func(value string) error {
        val, err := strconv.ParseFloat(value, elemType.Bits())
        if err != nil {
            return err
        }
        reflect.ValueOf(ptr).Elem().SetFloat(val)
        return nil
    }
}

func makeSliceSetter(ptr interface{}, elemType reflect.Type) func(string) error {
    sliceElemType := elemType.Elem()
    
    return func(value string) error {
        // Get current slice value
        sliceVal := reflect.ValueOf(ptr).Elem()
        
        // Parse the new element based on slice element type
        var newElem reflect.Value
        
        switch sliceElemType.Kind() {
        case reflect.String:
            // Support comma-separated values
            values := strings.Split(value, ",")
            for _, v := range values {
                sliceVal = reflect.Append(sliceVal, reflect.ValueOf(strings.TrimSpace(v)))
            }
            reflect.ValueOf(ptr).Elem().Set(sliceVal)
            return nil
            
        case reflect.Int:
            val, err := strconv.Atoi(value)
            if err != nil {
                return err
            }
            newElem = reflect.ValueOf(val)
            
        case reflect.Float64:
            val, err := strconv.ParseFloat(value, 64)
            if err != nil {
                return err
            }
            newElem = reflect.ValueOf(val)
            
        default:
            return fmt.Errorf("unsupported slice element type: %v", sliceElemType)
        }
        
        // Append to slice
        sliceVal = reflect.Append(sliceVal, newElem)
        reflect.ValueOf(ptr).Elem().Set(sliceVal)
        return nil
    }
}

func makeTextUnmarshalerSetter(ptr interface{}) func(string) error {
    return func(value string) error {
        unmarshaler := reflect.ValueOf(ptr).Interface().(encoding.TextUnmarshaler)
        return unmarshaler.UnmarshalText([]byte(value))
    }
}

// withFlag is the core immutability function - returns new Parser
func (p *Parser) withFlag(def FlagDefinition, s setter) *Parser {
    newDefs := make([]FlagDefinition, len(p.defs)+1)
    copy(newDefs, p.defs)
    newDefs[len(p.defs)] = def
    
    newSetters := make([]setter, len(p.setters)+1)
    copy(newSetters, p.setters)
    newSetters[len(p.setters)] = s
    
    return &Parser{
        defs:    newDefs,
        setters: newSetters,
    }
}
```

### 3. Hierarchical Context Support (Functional)

Contexts work the same way but use the single `Flag()` method:

```go
// Context creates a context flag with its own sub-parser
// The context itself is a bool flag that activates child flags
func (p *Parser) Context(names []string, desc string, builder func(*Parser) *Parser) *Parser {
    // Create a sub-parser for the context
    subParser := builder(New())
    
    // Create the context flag definition (reuses existing structure!)
    def := FlagDefinition{
        Names:       names,
        Description: desc,
        Switch:      true,  // Contexts are always switches
        Children:    subParser.defs,  // Nested FlagDefinitions!
    }
    
    // Context setter is a no-op (context itself has no value)
    s := setter{
        names: names,
        fn:    func(string) error { return nil },
    }
    
    return p.withFlag(def, s)
}
```

Usage example:

```go
var verbose bool
var gitCommit bool
var commitMsg string

parser := flags.New().
    Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
    Context([]string{"G", "git"}, "Git operations", func(ctx *flags.Parser) *flags.Parser {
        return ctx.
            Flag(&gitCommit, []string{"c", "commit"}, "Commit changes").
            Flag(&commitMsg, []string{"m", "message"}, "Commit message")
    })
```

### 4. Pure Parsing Function (Reuses Existing Parser!)

**Critical**: The library uses the existing parser directly, then applies setters:

```go
// ParseResult contains parsing outcome
type ParseResult struct {
    Remaining []string      // Unparsed arguments
    Errors    []error       // Any errors encountered
}

// Parse parses arguments using the EXISTING parser infrastructure
func (p *Parser) Parse(args []string) ParseResult {
    // Use the existing parser.Parser with our definitions!
    coreParser := parser.NewParser(p.defs)
    parserResult, err := coreParser.Parse(args)
    
    if err != nil {
        return ParseResult{Errors: []error{err}}
    }
    
    // Apply setters to user variables (walks all FlagValues)
    if err := p.applySetters(parserResult); err != nil {
        return ParseResult{Errors: []error{err}}
    }
    
    return ParseResult{Remaining: []string{}}  // TODO: handle remaining args
}

// applySetters walks the parse result and applies setters
// This is where FlagValue.Value is consumed!
func (p *Parser) applySetters(result *parser.ParseResult) error {
    // Group FlagValues by flag name (for accumulation)
    valuesByFlag := make(map[string][]*parser.FlagValue)
    
    result.Walk(func(fv *parser.FlagValue) {
        if !fv.Present {
            return
        }
        // Use first name as key
        key := fv.Definition.Names[0]
        valuesByFlag[key] = append(valuesByFlag[key], fv)
    })
    
    // Apply setters with accumulated values
    for _, s := range p.setters {
        // Find all FlagValues for this setter
        var values []*parser.FlagValue
        for _, name := range s.names {
            if fvs, ok := valuesByFlag[name]; ok {
                values = fvs
                break
            }
        }
        
        if len(values) == 0 {
            continue  // Flag not present in args
        }
        
        // Handle based on whether it's a slice type
        if s.isSlice {
            // Call setter for each occurrence (accumulates)
            for _, fv := range values {
                if err := s.fn(fv.Value); err != nil {  // <-- Using Value here!
                    return err
                }
            }
        } else {
            // Use last value (standard flag behavior)
            lastValue := values[len(values)-1]
            if err := s.fn(lastValue.Value); err != nil {  // <-- Using Value here!
                return err
            }
        }
    }
    
    return nil
}

// Must is a convenience wrapper that panics on error
func (p *Parser) Must(args []string) []string {
    result := p.Parse(args)
    if len(result.Errors) > 0 {
        panic(result.Errors[0])
    }
    return result.Remaining
}
```

**How FlagValue.Value flows**:
1. Parser populates `FlagValue.Value` from args as **strings** (e.g., "Alice", "go", "80", "30s")
2. Library walks all FlagValues
3. For each FlagValue, library calls `setter(fv.Value)` where Value is a **string**
4. Setter **parses the string** into the target type, then updates user's variable

**Examples of type conversion in setters**:

```go
// String: No conversion
setter("Alice") → *name = "Alice"

// Int: Parse string → int
setter("80") → strconv.Atoi("80") → *port = 80

// []int: Parse string → int, append
setter("80")   → strconv.Atoi("80")   → *ports = append(*ports, 80)
setter("8080") → strconv.Atoi("8080") → *ports = append(*ports, 8080)
// Result: ports = []int{80, 8080}

// time.Duration: Parse string → Duration
setter("30s") → time.ParseDuration("30s") → *timeout = 30 * time.Second
```

**Why this works for all types**:
- Command-line arguments are **always strings** (by definition)
- Parser captures them as strings (correct!)
- Setters convert strings to target types (just like standard `flag` package)
- This is **exactly** how `flag.IntVar`, `flag.Float64Var`, etc. work

**No changes to parser needed** - it already does the right thing!

### 5. Complete Usage Examples

#### Basic Usage (Type Inference!)

```go
package main

import (
    "fmt"
    "os"
    "V-Woodpecker-V/wsh/warg/flags"
)

func main() {
    var verbose bool
    var name string
    var count int
    
    // No need to specify Bool(), String(), Int() - types are inferred!
    parser := flags.New().
        Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
        Flag(&name, []string{"n", "name"}, "User name").
        Flag(&count, []string{"c", "count"}, "Count value")
    
    result := parser.Parse(os.Args[1:])
    
    if len(result.Errors) > 0 {
        fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
        os.Exit(1)
    }
    
    fmt.Printf("verbose: %v\n", verbose)
    fmt.Printf("name: %s\n", name)
    fmt.Printf("count: %d\n", count)
    fmt.Printf("remaining: %v\n", result.Remaining)
}
```

#### Slice Support (Repeatable Flags)

```go
func main() {
    var tags []string
    var ports []int
    
    parser := flags.New().
        Flag(&tags, []string{"t", "tag"}, "Add tag (repeatable)").
        Flag(&ports, []string{"p", "port"}, "Add port (repeatable)")
    
    // Supports: --tag go --tag cli --tag parser
    // Or: --tags go,cli,parser (comma-separated for strings)
    result := parser.Parse(os.Args[1:])
    
    fmt.Printf("tags: %v\n", tags)   // ["go", "cli", "parser"]
    fmt.Printf("ports: %v\n", ports) // [8080, 3000]
}
```

#### Hierarchical Context Usage

```go
func main() {
    var verbose bool
    var gitCommit bool
    var commitMsg string
    var gitPush bool
    
    parser := flags.New().
        Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
        Context([]string{"G", "git"}, "Git operations", func(git *flags.Parser) *flags.Parser {
            return git.
                Flag(&gitCommit, []string{"c", "commit"}, "Commit changes").
                Flag(&commitMsg, []string{"m", "message"}, "Commit message").
                Flag(&gitPush, []string{"p", "push"}, "Push changes")
        })
    
    // Supports: -v -Gcm "Fix bug" -p
    // Or: --verbose --git --commit --message "Fix bug" --push
    result := parser.Parse(os.Args[1:])
    
    if len(result.Errors) > 0 {
        fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
        os.Exit(1)
    }
}
```

#### Custom Types (TextUnmarshaler)

```go
import "time"

func main() {
    var timeout time.Duration
    
    // time.Duration implements encoding.TextUnmarshaler
    parser := flags.New().
        Flag(&timeout, []string{"t", "timeout"}, "Operation timeout")
    
    // Supports: --timeout 30s, --timeout 5m, etc.
    parser.Parse(os.Args[1:])
    
    fmt.Printf("timeout: %v\n", timeout)
}
```

## Functional Paradigm Adherence

### Immutability
✅ Each `Parser` method returns a new `Parser` instance
✅ Flag definitions are immutable once created
✅ No global state (unlike standard `flag` package)

### Pure Functions
✅ `Parse()` is a pure function that returns a result
✅ No side effects except setting user-provided pointers (which is the explicit contract)
✅ Same input always produces same output

### Composition
✅ Parsers can be composed via `Context()` 
✅ Can build reusable parser fragments:
```go
func commonFlags(p *flags.Parser) *flags.Parser {
    var v bool
    return p.Flag(&v, []string{"v", "verbose"}, "Verbose")
}

parser := commonFlags(flags.New()).
    Flag(&name, []string{"n", "name"}, "Name")
```

### Higher-Order Functions
✅ `Context()` takes a function that builds a sub-parser
✅ Enables flexible composition patterns

## Supported Types

### Primitive Types (Auto-Inferred)
| Go Type | Switch/Value | Notes |
|---------|-------------|-------|
| `bool` | Switch | No value needed, or accepts true/false/1/0 |
| `string` | Value | Any string value |
| `int`, `int8`, `int16`, `int32`, `int64` | Value | Parsed with base 10 |
| `uint`, `uint8`, `uint16`, `uint32`, `uint64` | Value | Parsed with base 10 |
| `float32`, `float64` | Value | Floating point numbers |

### Slice Types (Repeatable Flags)
| Go Type | Usage | Notes |
|---------|-------|-------|
| `[]string` | Repeatable | Supports: `--tag foo --tag bar` OR `--tags foo,bar` |
| `[]int` | Repeatable | Each occurrence appends to slice |
| `[]float64` | Repeatable | Each occurrence appends to slice |

**Slice Behavior**:
- Each flag occurrence appends a new value
- For `[]string`, comma-separated values are split automatically
- Example: `--tag go,cli --tag parser` → `["go", "cli", "parser"]`

### Custom Types (TextUnmarshaler)
Any type implementing `encoding.TextUnmarshaler` is automatically supported:
- `time.Duration` - Examples: `30s`, `5m`, `1h30m`
- `net.IP` - IP addresses
- `url.URL` - URLs
- Custom types you define

### Use Cases and Comparison

**Standard Library**:
- ❌ No slice support (must use custom Value implementation)
- ❌ No repeatable flags
- ❌ One name per flag registration

**Warg Library**:
- ✅ Native slice support for common types
- ✅ Repeatable flags work naturally
- ✅ Multiple names per flag (aliases)
- ✅ Hierarchical contexts

**Common Use Cases Covered**:
1. ✅ Multiple tags: `--tag go --tag cli --tag parser`
2. ✅ Multiple environment vars: `--env A=1 --env B=2`
3. ✅ Multiple ports: `--port 8080 --port 3000`
4. ✅ Multiple authors: `--author Alice --author Bob`
5. ✅ Comma-separated values: `--tags go,cli,parser`

**Edge Cases**:
- Empty slices remain empty if flag not provided
- Repeating non-slice flags: last value wins (standard behavior)
- Mixed comma-separated and individual: both work, values are merged

## Key Architectural Insight: FlagValue.Value as the Bridge

The existing `FlagValue` struct is **perfectly designed** for both CLI and Library use:

```go
type FlagValue struct {
    Definition *flags.FlagDefinition
    Present    bool
    Value      string  // <-- This is the key!
    Children   []*FlagValue
}
```

### How It Works for Both Modes

**CLI Mode (Current)**:
```
Args: --tag go --tag cli
  ↓
Parser creates:
  FlagValue { Value: "go" }
  FlagValue { Value: "cli" }
  ↓
JSON serialization:
  { "value": "go" }
  { "value": "cli" }
  ↓
Bash script reads JSON
```

**Library Mode (New)**:
```
Args: --tag go --tag cli
  ↓
Parser creates:
  FlagValue { Value: "go" }   ← Same as CLI!
  FlagValue { Value: "cli" }  ← Same as CLI!
  ↓
Library walks FlagValues:
  setter("go")   → appends to slice
  setter("cli")  → appends to slice
  ↓
User's variable: []string{"go", "cli"}
```

### Why No Refactoring Is Needed

1. **Parser already creates separate FlagValues** for repeated flags ✅
2. **FlagValue.Value already contains the string value** ✅
3. **ParseResult.Walk() already exists** for traversal ✅
4. **Library just consumes what parser produces** ✅

The parser doesn't need to know about slices - it just reports each occurrence.
The library recognizes slice types and accumulates the values.

**Perfect separation of concerns!**

### Critical Detail: All Values Are Strings

**Q: What about `[]int` or `[]float64`? FlagValue.Value is a string!**

**A: That's exactly correct, and it works perfectly!**

Command-line arguments are **always strings** by definition. The conversion happens in the setter:

**Example: `--port 80 --port 8080 --port 3000`**

```
Parser produces:
  FlagValue { Value: "80" }    ← String!
  FlagValue { Value: "8080" }  ← String!
  FlagValue { Value: "3000" }  ← String!

Library applies setters:
  Call 1: setter("80")
    → strconv.Atoi("80") = 80
    → *ports = append(*ports, 80)
    → ports = []int{80}
  
  Call 2: setter("8080")
    → strconv.Atoi("8080") = 8080
    → *ports = append(*ports, 8080)
    → ports = []int{80, 8080}
  
  Call 3: setter("3000")
    → strconv.Atoi("3000") = 3000
    → *ports = append(*ports, 3000)
    → ports = []int{80, 8080, 3000}

Result: ports = []int{80, 8080, 3000} ✅
```

**This is exactly how the standard `flag` package works:**
- `flag.IntVar` receives string "80", parses to int
- `flag.Float64Var` receives string "3.14", parses to float64
- `flag.DurationVar` receives string "30s", parses to Duration

**The setter function encapsulates type conversion:**
- `[]string`: No conversion, append directly
- `[]int`: `strconv.Atoi()`, then append
- `[]float64`: `strconv.ParseFloat()`, then append
- Custom types: Use `TextUnmarshaler` interface

**Parser stays type-agnostic (good design!) → Library handles type conversion**

### Current Architecture

```
CLI Path:                              Library Path (NEW):
┌─────────────────┐                   ┌──────────────────┐
│ Bash script     │                   │ Go code          │
│ ./warg <<EOF    │                   │ var v bool       │
│ -v, --verbose   │                   │ flags.New()...   │
│ EOF             │                   └────────┬─────────┘
└────────┬────────┘                            │
         │                                     │
         v                                     v
┌─────────────────┐                   ┌──────────────────┐
│ cli.Parse*()    │                   │ Flag() with      │
│ - ParseHeredoc  │                   │ reflection       │
│ - ParseJSON     │                   └────────┬─────────┘
│ - ParseInline   │                            │
└────────┬────────┘                            │
         │                                     │
         v                                     │
┌──────────────────────────────────────────────┴──────┐
│           []flags.FlagDefinition                    │
│  ┌────────────────────────────────────────────┐     │
│  │ Names: ["-v", "--verbose"]                 │     │
│  │ Switch: true                               │     │
│  │ Description: "Enable verbose"              │     │
│  │ Children: []FlagDefinition{}               │     │
│  └────────────────────────────────────────────┘     │
└────────────────────────┬────────────────────────────┘
                         │
                         v
              ┌──────────────────────┐
              │ parser.NewParser()   │
              │ parser.Parse(args)   │
              └──────────┬───────────┘
                         │
                         v
              ┌──────────────────────┐
              │ parser.ParseResult   │
              │ - FlagValues tree    │
              │ - Present flags      │
              │ - Values             │
              └──────────────────────┘
```

### Key Insight: Single Source of Truth

**Both paths produce identical `[]FlagDefinition`**, then use the **same parser logic**:

1. **CLI Path**: User writes heredoc/JSON → Parser converts to `[]FlagDefinition`
2. **Library Path**: User writes Go code → Reflection converts to `[]FlagDefinition`
3. **Both paths**: Use `parser.Parser` with `[]FlagDefinition` → Same behavior

### Refactoring Required

#### Critical Understanding: How FlagValue.Value Works

The existing parser **already handles everything correctly**! Here's how:

**Current behavior** (verified):
```bash
# Input: --tag go --tag cli --tag parser
# Parser output:
{
  "flags": [
    { "value": "go", "present": true },
    { "value": "cli", "present": true },
    { "value": "parser", "present": true }
  ]
}
```

The parser creates **separate FlagValue entries** for each occurrence. This is perfect because:

1. **CLI Mode**: Each FlagValue is serialized to JSON → bash script handles them
2. **Library Mode**: Library walks all FlagValues → accumulates for slices

**Key Insight**: `FlagValue.Value` is the bridge between parsing and consumption!
- CLI: Reads `Value` from JSON
- Library: Passes `Value` to setter function

#### 1. Zero Changes to Existing Code ✅

**No changes needed**:
- `flags/definition.go` - Already perfect
- `internal/parser/parser.go` - Already handles repeated flags correctly!
- `internal/parser/context.go` - Reusable as-is
- `internal/cli/definition.go` - CLI parsing stays unchanged
- `internal/parser/output.go` - JSON output stays same

**Existing `Walk()` method** - Already exists and works perfectly:
```go
// internal/parser/parser.go
func (r *ParseResult) Walk(fn func(*FlagValue)) {
    for _, flag := range r.Flags {
        flag.Walk(fn)
    }
}
```

**New files to add**:
- `flags/parser.go` - Library API (Parser type, Flag(), Context(), Parse())
- `flags/setters.go` - Type inference and setter functions
- `flags/parser_test.go` - Tests for library API

#### 2. How Library Handles Slices (No Parser Changes!)

The library needs to **accumulate** values when it sees multiple FlagValues with the same flag name:

```go
func (p *Parser) applySetters(result *parser.ParseResult) error {
    // Group FlagValues by flag name for accumulation
    valuesByFlag := make(map[string][]*parser.FlagValue)
    
    result.Walk(func(fv *parser.FlagValue) {
        if !fv.Present {
            return
        }
        // Group by first name
        key := fv.Definition.Names[0]
        valuesByFlag[key] = append(valuesByFlag[key], fv)
    })
    
    // Apply setters
    for _, s := range p.setters {
        flagValues := valuesByFlag[s.names[0]]
        
        if len(flagValues) == 0 {
            continue  // Flag not present
        }
        
        if s.isSlice {
            // For slice types, call setter for each occurrence
            for _, fv := range flagValues {
                if err := s.fn(fv.Value); err != nil {
                    return err
                }
            }
        } else {
            // For non-slice, use last occurrence (standard behavior)
            lastValue := flagValues[len(flagValues)-1]
            if err := s.fn(lastValue.Value); err != nil {
                return err
            }
        }
    }
    
    return nil
}
```

**Why this works**:
- Parser treats each `--tag foo` as a separate event ✅
- Library accumulates them into a slice ✅
- CLI sees each occurrence separately in JSON ✅
- No breaking changes ✅

### Implementation Strategy

#### Phase 1: Core Library Parser (1-2 days)
- Create `flags/parser.go` with `Parser`, `Flag()`, `Parse()`
- Implement type inference for basic types (bool, string, int, float)
- Reuse existing `parser.Parser` - no modifications needed
- Write tests comparing CLI and Library output

#### Phase 2: Slice Support (1 day)
- Add slice type detection in `Flag()`
- Implement accumulation logic in `Parse()`
- Handle comma-separated values for `[]string`
- Add tests for repeatable flags

#### Phase 3: Context Support (1 day)
- Implement `Context()` method
- Ensure nested `FlagDefinition` structures match CLI
- Test hierarchical flag resolution

#### Phase 4: Advanced Types (1 day)
- Support `TextUnmarshaler` interface
- Handle `time.Duration`, `net.IP`, etc.
- Add custom type tests

#### Phase 5: Polish & Documentation (1 day)
- Error messages
- Help text generation (reuse from CLI?)
- API documentation
- Migration guide from standard `flag`
- Examples

**Total Estimate**: 5-6 days for full implementation

### Testing Strategy

#### Convergence Tests (Critical!)

Ensure CLI and Library produce identical behavior:

```go
func TestCliLibraryConvergence(t *testing.T) {
    // Define flags via CLI heredoc
    cliDefs := cli.ParseHeredocDefinition(`
        -v, --verbose Enable verbose
        -n, --name [value] User name
    `)
    
    // Define same flags via Library
    var verbose bool
    var name string
    libParser := flags.New().
        Flag(&verbose, []string{"v", "verbose"}, "Enable verbose").
        Flag(&name, []string{"n", "name"}, "User name")
    
    // Compare FlagDefinition structures
    assertEqual(t, cliDefs, libParser.defs)
    
    // Parse same args with both
    args := []string{"-v", "--name", "Alice"}
    
    cliResult := parser.NewParser(cliDefs).Parse(args)
    libResult := libParser.Parse(args)
    
    // CLI result should match library variables
    assertEqual(t, cliResult.Find("verbose").Present, verbose)
    assertEqual(t, cliResult.Find("name").Value, name)
}
```

## Differences from Standard Library

| Aspect | Standard `flag` | Warg `flags` |
|--------|----------------|--------------|
| **State** | Global mutable `FlagSet` | Immutable `Parser` |
| **API Style** | Imperative (mutate flagset) | Functional (return new parser) |
| **Type Registration** | `Bool()`, `String()`, `Int()` | Single `Flag()` with type inference |
| **Names** | Single name per registration | Multiple names (aliases) per flag |
| **Chaining** | Not supported | Fluent chaining |
| **Contexts** | Not supported | Full hierarchical support |
| **Short flags** | No combining | `-abc` works |
| **Slices** | No (need custom Value) | Native support for `[]string`, `[]int`, etc. |
| **Repeatable flags** | No | Yes, appends to slices |

## Implementation Phases

### Phase 1: Core Parser with Type Inference (Minimal Viable)
Files to create:
- `flags/parser.go` - `Parser` type, `New()`, `Flag()`, `Parse()`
- `flags/setters.go` - Type inference and value setters
- `flags/parser_test.go` - Basic tests

Types supported: `bool`, `string`, `int`, `float64`

**No changes to existing code** - reuses `parser.Parser` directly

### Phase 2: Context Support
- Add `Context()` method to `Parser`
- Ensure nested `FlagDefinition` structures work
- Test hierarchical resolution

### Phase 3: Slice Support (Repeatable Flags)
- Detect slice types in `Flag()`
- Accumulate values on repeated flags
- Support comma-separated for `[]string`

### Phase 4: Advanced Types
- Support `TextUnmarshaler` interface
- Add `time.Duration`, `net.IP`, etc.
- Custom type parsers

### Phase 5: Polish
- Error messages
- Help text generation
- Documentation
- Examples

## Trade-offs

### ✅ Benefits
1. **Testable**: No global state, easy to test in parallel
2. **Composable**: Build reusable parser fragments
3. **Type-safe**: Type inference catches errors at runtime
4. **Immutable**: Thread-safe by default
5. **Functional**: Aligns with modern Go patterns
6. **Simple API**: Single `Flag()` method for all types
7. **Code Reuse**: Uses existing parser infrastructure (90%+ reuse)
8. **Consistent**: CLI and Library produce identical structures
9. **Slice Support**: Native support for repeatable flags

### ⚠️ Considerations
1. **Runtime Type Checking**: Type errors detected at runtime, not compile-time
2. **Reflection Overhead**: Minimal, only during parser construction
3. **Panic on Invalid Types**: Unsupported types cause panic at registration (could return error instead)
4. **Learning**: Developers must understand functional patterns (but API is simple)
5. **Pointers**: Still requires passing pointers (unavoidable for value setting)

### Comparison: Type Inference vs Explicit Types

**Explicit (original proposal)**:
```go
parser := flags.New().
    Bool(&verbose, []string{"v"}, "Verbose").
    String(&name, []string{"n"}, "Name").
    Int(&count, []string{"c"}, "Count")
```

**Type Inferred (this proposal)**:
```go
parser := flags.New().
    Flag(&verbose, []string{"v"}, "Verbose").
    Flag(&name, []string{"n"}, "Name").
    Flag(&count, []string{"c"}, "Count")
```

**Why Type Inference Wins**:
- ✅ Less API surface (1 method vs 10+ methods)
- ✅ Easier to extend (new types don't need new methods)
- ✅ More consistent (same pattern for everything)
- ✅ More like standard library (doesn't repeat type info)
- ⚠️ Type errors at runtime instead of compile-time (acceptable trade-off)

## Alternative Approaches Considered

### Option A: Pure Functional (No Pointers)
```go
result := parser.Parse(os.Args[1:])
verbose := result.Bool("verbose")  // Retrieve by name
```
❌ Loses compile-time type safety
❌ Loses compile-time flag existence checking
❌ Runtime lookups by string are error-prone
❌ Different API than standard library

### Option B: Fully Imperative (Like standard library)
```go
parser := flags.NewParser()
parser.BoolVar(&verbose, "v", false, "desc")
parser.Parse(os.Args[1:])
```
❌ Not functional
❌ Harder to compose
❌ Mutable state issues
❌ Can't build reusable fragments

### Option C: Explicit Type Methods (Original Proposal)
```go
parser := flags.New().
    Bool(&verbose, []string{"v"}, "desc").
    String(&name, []string{"n"}, "desc")
```
⚠️ Many methods (Bool, String, Int, Float, Duration, StringSlice, IntSlice...)
⚠️ Have to remember which method to use
⚠️ Repetitive (type already known from variable)

### Option D: Type-Inferred Single Method (Chosen)
```go
parser := flags.New().
    Flag(&verbose, []string{"v"}, "desc").
    Flag(&name, []string{"n"}, "desc")
```
✅ Single method for all types
✅ Type inferred from pointer
✅ Easy to extend (no new methods needed)
✅ Closest to standard library feel
✅ Best functional properties

## Future Extensions

### 1. Parser Combinators
```go
func Combine(parsers ...*Parser) *Parser {
    // Merge multiple parsers
}
```

### 2. Middleware Pattern
```go
type Middleware func(*Parser) *Parser

func WithLogging(p *Parser) *Parser {
    // Wrap parser with logging
}
```

### 3. Validation
```go
func (p *Parser) Validate(fn func() error) *Parser {
    // Add post-parse validation
}
```

## Testing Strategy

### Unit Tests
- Test each flag type registration
- Test immutability (original parser unchanged)
- Test context nesting
- Test flag resolution

### Integration Tests
- Test with real CLI arguments
- Test combined short flags
- Test context chain walking
- Test error cases

### Property-Based Tests
- Parse then serialize should be idempotent
- Order of flag registration shouldn't matter
- Parser immutability properties

## Documentation Requirements

1. **API Documentation**: Godoc for all public functions
2. **Examples**: Show common patterns (basic, slices, contexts, custom types)
3. **Migration Guide**: For users coming from `flag` package
4. **Best Practices**: Guide on composing parsers
5. **Type Support Matrix**: Document all supported types

## Common Questions (FAQ)

### Q1: Why is FlagValue.Value a string if I have []int?

**A:** Command-line arguments are **always strings** by definition (shell passes strings). The setter function handles type conversion:

```go
// Parser captures: FlagValue{Value: "80"}  ← String!
// Setter converts:  strconv.Atoi("80") → 80 → append to []int
```

This is **exactly** how the standard `flag` package works (flag.IntVar receives strings, parses to int).

### Q2: Do I need to change the existing parser code?

**A:** **No changes needed!** The parser already:
- Creates separate FlagValues for repeated flags ✅
- Stores values as strings ✅  
- Provides Walk() for traversal ✅

The library just consumes what the parser produces.

### Q3: How do slices work with repeated flags?

**A:** The parser creates one FlagValue per occurrence:
```
--port 80 --port 8080
↓
FlagValue{Value: "80"}
FlagValue{Value: "8080"}
```

Library walks both, calls setter twice:
```go
setter("80")   → *ports = append(*ports, 80)
setter("8080") → *ports = append(*ports, 8080)
// Result: ports = []int{80, 8080}
```

### Q4: Why use reflection instead of explicit type methods?

**A:** Single `Flag()` method is simpler:
- ✅ Less API surface (1 method vs 10+)
- ✅ No need to remember Bool/String/Int/etc
- ✅ Variable type already declares intent
- ✅ Easy to add new types without new methods

Trade-off: Type errors at runtime vs compile-time (acceptable for this use case).

### Q5: Can I use custom types?

**A:** Yes! Any type implementing `encoding.TextUnmarshaler` works automatically:
```go
var timeout time.Duration  // Implements TextUnmarshaler
parser.Flag(&timeout, []string{"t"}, "Timeout")
// Supports: --timeout 30s, --timeout 5m, etc.
```

### Q6: Is this thread-safe?

**A:** Yes! The Parser is **immutable** - each method returns a new instance. Multiple goroutines can safely use different parsers or even share read-only parser definitions.

### Q7: How does this compare to pflag or cobra?

| Feature | std flag | pflag/cobra | warg |
|---------|----------|-------------|------|
| Slice support | ❌ | ✅ | ✅ |
| Type inference | ❌ | ❌ | ✅ |
| Hierarchical contexts | ❌ | ❌ | ✅ |
| Short flag combining | ❌ | Partial | ✅ |
| Functional/Immutable | ❌ | ❌ | ✅ |
| Global state | ✅ Mutable | ✅ Mutable | ❌ |

## Conclusion

This type-inferred functional approach to flag parsing:
- **Maintains familiar pointer-based value setting** (like standard library)
- **Uses single `Flag()` method** for all types (simpler API)
- **Automatically infers types** from pointers (less repetition)
- **Embraces functional paradigm** (immutability, composition)
- **Supports warg's advanced features** (contexts, hierarchical flags, short flag combining)
- **Provides type safety** via runtime reflection (acceptable trade-off)
- **Enables testable, composable code** (no global state)
- **Reuses 90%+ of existing parser code** (minimal refactoring)
- **Ensures CLI and Library convergence** (same `FlagDefinition` structure)
- **Native slice support** for repeatable flags (common use case)

### Key Design Decisions

1. **Type Inference**: Single `Flag()` method uses reflection → simpler API, less repetition
2. **Immutable Builders**: Each method returns new Parser → functional, composable, thread-safe
3. **Reuse Existing Parser**: Library produces `[]FlagDefinition` → CLI and Library converge
4. **Native Slice Support**: Detect `[]T` types → handle repeatable flags naturally
5. **TextUnmarshaler Support**: Automatic custom type handling → extensible without new methods

### Implementation Effort

**Minimal Refactoring Required**:
- ✅ Existing parser code: **No changes**
- ✅ Existing CLI code: **No changes**
- ✅ Existing definition structs: **No changes**
- 🆕 New library API: **~500 lines** (parser.go + setters.go + tests)
- 🆕 Convergence tests: **~200 lines**

**Total**: ~700 lines of new code, **0 lines changed** in existing code.

The key insight is that **setting user pointers is the only necessary "side effect"**, and we can make everything else functional and immutable while reusing all existing infrastructure.
