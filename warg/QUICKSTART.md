# Quick Start: Using warg as a Library

## Installation

In your Go project, import the warg library:

```go
import "V-Woodpecker-V/wsh/warg/lib"
```

If you're developing locally, add a replace directive to your `go.mod`:

```bash
go mod edit -replace V-Woodpecker-V/wsh/warg=../warg
```

## 5-Minute Tutorial

### 1. Basic Flags

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
    
    parser := lib.New().
        Flag(&verbose, []string{"v", "verbose"}, "Enable verbose output").
        Flag(&name, []string{"n", "name"}, "User name")
    
    result := parser.Parse(os.Args[1:])
    if len(result.Errors) > 0 {
        fmt.Fprintf(os.Stderr, "Error: %v\n", result.Errors[0])
        os.Exit(1)
    }
    
    if verbose {
        fmt.Println("Verbose mode enabled")
    }
    fmt.Printf("Hello, %s!\n", name)
}
```

Run it:
```bash
$ go run main.go -v --name Alice
Verbose mode enabled
Hello, Alice!
```

### 2. Multiple Values (Slices)

```go
var tags []string

parser := lib.New().
    Flag(&tags, []string{"t", "tag"}, "Add tag (repeatable)")

parser.Parse(os.Args[1:])
fmt.Printf("Tags: %v\n", tags)
```

Run it:
```bash
$ go run main.go --tag go --tag cli
Tags: [go cli]

$ go run main.go --tag go,cli,parser
Tags: [go cli parser]
```

### 3. Hierarchical Contexts

```go
var verbose bool
var gitCommit bool
var message string

parser := lib.New().
    Flag(&verbose, []string{"v"}, "Verbose").
    Context([]string{"G", "git"}, "Git operations", func(git *lib.Parser) *lib.Parser {
        return git.
            Flag(&gitCommit, []string{"c", "commit"}, "Commit").
            Flag(&message, []string{"m"}, "Message")
    })

parser.Parse(os.Args[1:])
```

Run it:
```bash
$ go run main.go -v -G -c -m "Initial commit"
# or combined:
$ go run main.go -vGcm "Initial commit"
```

### 4. Time Duration

```go
import "time"

var timeout time.Duration

parser := lib.New().
    Flag(&timeout, []string{"t", "timeout"}, "Operation timeout")

parser.Parse(os.Args[1:])
fmt.Printf("Timeout: %v\n", timeout)
```

Run it:
```bash
$ go run main.go --timeout 5m30s
Timeout: 5m30s

$ go run main.go --timeout 2h
Timeout: 2h0m0s
```

## Common Patterns

### Pattern 1: Reusable Parser Fragments

```go
func commonFlags(p *lib.Parser) *lib.Parser {
    var verbose bool
    var debug bool
    return p.
        Flag(&verbose, []string{"v", "verbose"}, "Verbose").
        Flag(&debug, []string{"d", "debug"}, "Debug")
}

func main() {
    var port int
    
    parser := commonFlags(lib.New()).
        Flag(&port, []string{"p", "port"}, "Port")
    
    parser.Parse(os.Args[1:])
}
```

### Pattern 2: Configuration Struct

```go
type Config struct {
    Verbose bool
    Name    string
    Port    int
}

func parseArgs() Config {
    var cfg Config
    
    parser := lib.New().
        Flag(&cfg.Verbose, []string{"v", "verbose"}, "Verbose").
        Flag(&cfg.Name, []string{"n", "name"}, "Name").
        Flag(&cfg.Port, []string{"p", "port"}, "Port")
    
    result := parser.Parse(os.Args[1:])
    if len(result.Errors) > 0 {
        log.Fatal(result.Errors[0])
    }
    
    return cfg
}

func main() {
    cfg := parseArgs()
    fmt.Printf("Config: %+v\n", cfg)
}
```

### Pattern 3: Conditional Parsing

```go
func main() {
    var mode string
    
    // First, parse just the mode
    modeParser := lib.New().
        Flag(&mode, []string{"m", "mode"}, "Operation mode")
    
    modeParser.Parse(os.Args[1:])
    
    // Then parse mode-specific flags
    if mode == "server" {
        var port int
        serverParser := lib.New().
            Flag(&port, []string{"p", "port"}, "Server port")
        serverParser.Parse(os.Args[1:])
        // Start server on port...
    } else if mode == "client" {
        var host string
        clientParser := lib.New().
            Flag(&host, []string{"h", "host"}, "Server host")
        clientParser.Parse(os.Args[1:])
        // Connect to host...
    }
}
```

## Complete Examples

See the `examples/` directory:
- `examples/simple_demo/` - All basic features in one file
- `examples/full_demo/` - Separate examples for each feature

Run them:
```bash
cd warg/examples/simple_demo
go run main.go -v --name Alice --count 42 -t go -t rust

cd warg/examples/full_demo
go run main.go
```

## Supported Types Reference

| Type | Example | Usage |
|------|---------|-------|
| `bool` | `var v bool` | `-v` or `--verbose` |
| `string` | `var s string` | `--name Alice` |
| `int` | `var i int` | `--count 42` |
| `float64` | `var f float64` | `--rate 3.14` |
| `[]string` | `var ss []string` | `--tag go --tag cli` |
| `[]int` | `var is []int` | `--port 80 --port 443` |
| `time.Duration` | `var d time.Duration` | `--timeout 5m30s` |

## Next Steps

- Read the full documentation: `lib/README.md`
- Check the solution design: `solutions/Go-Library-Flag-Parsing-Functional.md`
- Run the integration tests: `just test-lib` (30+ comprehensive test cases)
- View integration test examples: `../warg-integration/integration_test.go`

## Getting Help

If something doesn't work:
1. Check error messages in `result.Errors`
2. Verify you're passing pointers to `Flag()`
3. Make sure flag names are strings without spaces
4. See if your type is supported (check the table above)

For questions about the design or architecture, see:
- `DESIGN.md` - Overall architecture
- `INSTRUCTIONS.md` - Development guidelines
- `solutions/Go-Library-Flag-Parsing-Functional.md` - Detailed solution spec
