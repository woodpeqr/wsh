# lib - Type-Safe Flag Parsing Library for Go

The `lib` package provides a functional, type-inferred flag parsing library with support for hierarchical contexts and repeatable slice contexts.

## Features

- **Type Inference**: Automatically detects variable types
- **Immutable API**: Functional composition with builder pattern
- **Hierarchical Contexts**: Group related flags together
- **Repeatable Contexts**: Support for slice-based repeatable flag groups
- **Slice Support**: Native support for repeatable flags (`[]string`, `[]int`, etc.)
- **Custom Types**: Support for `time.Duration` and `encoding.TextUnmarshaler`
- **Combined Short Flags**: Parse `-vdi` as `-v -d -i`

## Basic Usage

```go
package main

import (
    "fmt"
    "os"
    "V-Woodpecker-V/wsh/warg/lib"
)

func main() {
    var verbose bool
    var name string
    var count int
    
    parser := lib.New().
        Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
        Flag(&name, []string{"n", "name"}, "User name").
        Flag(&count, []string{"c", "count"}, "Count")
    
    result := parser.Parse(os.Args[1:])
    if len(result.Errors) > 0 {
        fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
        os.Exit(1)
    }
    
    fmt.Printf("verbose: %v, name: %s, count: %d\n", verbose, name, count)
}
```

## Single Struct Context

Group related flags in a struct:

```go
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
    Context(&config.Git, []string{"G", "git"}, "Git operations", 
        func(p *lib.Parser, git *GitConfig) *lib.Parser {
            return p.
                Flag(&git.Commit, []string{"c", "commit"}, "Commit").
                Flag(&git.Message, []string{"m"}, "Message")
        })

// Usage: -v -G -c -m "Fix bug"
```

## Repeatable Slice Context

Define repeatable flag groups using slices:

```go
type AddFlagDef struct {
    Names       string
    Description string
    IsSwitch    bool
}

var addFlags []AddFlagDef

parser := lib.New().
    Context(&addFlags, []string{"A", "add"}, "Add flag definition",
        func(p *lib.Parser, def *AddFlagDef) *lib.Parser {
            return p.
                Flag(&def.Names, []string{"n", "name"}, "Names").
                Flag(&def.Description, []string{"d", "description"}, "Description").
                Flag(&def.IsSwitch, []string{"s", "switch"}, "Switch flag")
        })

// Usage: -A -n "v,verbose" -d "Verbose" -s -A -n "n,name" -d "Name"
// Result: addFlags contains 2 elements
```

## Supported Types

### Basic Types
- `bool` - Switch flags (no value)
- `string`, `int`, `uint`, `float64` - Value flags
- `int8`, `int16`, `int32`, `int64`
- `uint8`, `uint16`, `uint32`, `uint64`
- `float32`

### Slice Types
Repeatable flags append values to slices:
```go
var tags []string
parser := lib.New().Flag(&tags, []string{"t", "tag"}, "Tags")
// Usage: -t go -t cli -t parser
// Result: tags = ["go", "cli", "parser"]
```

### Custom Types
- `time.Duration` - Parses durations like "5m", "30s", "1h30m"
- Any type implementing `encoding.TextUnmarshaler`

## Flag Name Conventions

- Single character names become short flags: `"v"` → `-v`
- Multi-character names become long flags: `"verbose"` → `--verbose`
- Names with dashes are preserved: `"user-name"` → `--user-name`

## Error Handling

```go
result := parser.Parse(os.Args[1:])
if len(result.Errors) > 0 {
    for _, err := range result.Errors {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
    }
    os.Exit(1)
}
```

Common errors:
- Unknown flags
- Missing values for value flags
- Type conversion errors (e.g., "abc" for an int flag)

## Examples

See the `examples/` directory for complete working examples:
- `simple_demo/` - Basic usage
- `full_demo/` - Advanced features including contexts and custom types

## API Reference

### `lib.New() *Parser`
Creates a new Parser instance.

### `Parser.Flag(ptr interface{}, names []string, description string) *Parser`
Registers a flag with automatic type inference. Returns a new Parser (immutable).

### `Parser.Context(ptr interface{}, names []string, description string, builder interface{}) *Parser`
Creates a hierarchical context. The `ptr` can be:
- `*Struct` for single contexts
- `*[]Struct` for repeatable contexts

The builder function signature:
```go
func(p *Parser, parent *YourType) *Parser
```

### `Parser.Parse(args []string) *ParseResult`
Parses the arguments and returns a result with any errors.

## Design Philosophy

1. **Immutable**: Each method returns a new Parser, enabling safe composition
2. **Type-Safe**: Uses Go's type system instead of reflection magic
3. **Explicit**: Builder functions make relationships clear
4. **Functional**: Compose parsers using standard Go functions
5. **No Magic**: No struct tags, no code generation

## Comparison with Standard `flag` Package

| Feature | `lib` | `flag` |
|---------|-------|--------|
| Combined short flags | ✅ `-vdi` | ❌ |
| Hierarchical contexts | ✅ | ❌ |
| Repeatable contexts | ✅ | ❌ |
| Immutable API | ✅ | ❌ |
| Type inference | ✅ | ❌ (manual) |
| Slice support | ✅ Native | ⚠️ Manual |
| Custom types | ✅ Interface | ✅ Interface |
