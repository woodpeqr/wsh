# wsh - Woodpecker Shell

A zsh wrapper with support for modular configuration through `.wshrc` files or directories.

## Project Structure

```
wsh/
├── main.go         # Entry point and CLI argument parsing
├── shell.go        # Shell struct and execution logic
├── wshrc.go        # .wshrc loading and processing
├── env.go          # Environment variable manipulation
├── *_test.go       # Unit tests for each module
└── README.md       # This file
```

## Architecture

### Core Components

#### 1. Shell (`shell.go`)
The main shell orchestrator that:
- Manages zsh path and wshrc configuration
- Coordinates initialization and execution
- Handles command vs interactive mode
- Manages exit codes

**Key struct:**
```go
type Shell struct {
    ZshPath     string        // Path to zsh executable
    WshrcPath   string        // Path to .wshrc (file or directory)
    Env         *Environment  // Environment handler
    WshrcLoader *WshrcLoader  // Configuration loader
}
```

#### 2. WshrcLoader (`wshrc.go`)
Handles loading and processing of `.wshrc` configurations:
- Detects if `.wshrc` is a file or directory
- For files: generates a simple source command
- For directories: executes all scripts in parallel and merges environments
- Filters hidden files and subdirectories

**Key methods:**
- `Load(path)` - Main entry point for loading configuration
- `loadFile(path)` - Handles single file configuration
- `loadDirectory(path)` - Handles directory-based configuration
- `executeScriptsInParallel(scripts)` - Runs multiple scripts concurrently

#### 3. Environment (`env.go`)
Manages environment variable operations:
- Captures current environment
- Executes scripts and captures their environment
- Parses null-delimited environment output
- Builds export scripts for environment changes
- Merges multiple environment maps

**Key methods:**
- `GetCurrent()` - Returns current environment as map
- `ExecuteAndCapture(zshPath, scriptPath)` - Runs script and captures env
- `BuildExportScript(current, new)` - Generates export statements
- `Merge(envChan)` - Combines multiple environment maps

## Usage

### Command Line

```bash
# Interactive mode
./wsh

# Execute command
./wsh -c "echo hello"

# Execute with arguments
./wsh -c "echo $1" arg1
```

### Configuration

#### Single File Mode
Create `~/.wshrc` as a file:
```bash
export MY_VAR="value"
alias ll="ls -la"
```

#### Directory Mode
Create `~/.wshrc/` as a directory with multiple scripts:
```bash
~/.wshrc/
├── 01-env.sh      # Environment variables
├── 02-paths.sh    # PATH configuration
└── 03-aliases.sh  # Aliases and functions
```

**Key features:**
- All scripts execute in parallel for faster startup
- Environment changes are merged automatically
- Hidden files (starting with `.`) are ignored
- Subdirectories are ignored
- Only regular files are processed

## Testing

Run all tests:
```bash
go test -v ./...
```

Check coverage:
```bash
go test -cover ./...
```

Run specific test:
```bash
go test -v -run TestEnvironment_ParseEnvLine
```

### Test Coverage

Current coverage: **84.2%**

Test files:
- `env_test.go` - Environment variable operations
- `wshrc_test.go` - Configuration loading
- `shell_test.go` - Shell execution and integration tests

## Building

```bash
go build -o wsh
```

## Functional Programming Patterns

wsh leverages several functional programming patterns for flexibility and composability:

### 1. Functional Options Pattern

Constructors accept variadic option functions for flexible configuration:

```go
shell, err := NewShell(
    WithZshPath("/custom/zsh"),
    WithWshrcPath("~/.customrc"),
)
```

**Benefits:**
- Backward compatible (default usage stays simple)
- Easy to test with custom configurations
- Self-documenting API
- Optional parameters without bloat

### 2. Strategy Pattern

Different execution strategies can be swapped at runtime:

```go
loader := NewWshrcLoader(zshPath,
    WithExecutionStrategy(SequentialExecutionStrategy), // or ParallelExecutionStrategy
)
```

**Use cases:**
- Parallel execution (default) for speed
- Sequential execution for debugging
- Custom strategies for dependency management

### 3. Middleware Pattern

Wrap script executors with composable middleware:

```go
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithTimeout(10 * time.Second),
        WithLogging(log.Printf),
        WithErrorRecovery(),
    ),
)
```

**Available middleware:**
- `WithLogging` - Log execution with timing
- `WithTimeout` - Prevent hanging scripts
- `WithErrorRecovery` - Recover from panics
- `WithRetry` - Retry failed executions
- `WithEnvFilter` - Filter environment variables
- `WithCaching` - Cache script results

**Benefits:**
- Composable - mix and match as needed
- Reusable - write once, use everywhere
- Testable - easy to verify behavior
- Non-invasive - doesn't change core logic

See [EXAMPLES.md](EXAMPLES.md) for detailed usage examples.

## Design Decisions

### Why Parallel Execution?
When `.wshrc` is a directory, scripts are executed in parallel by default to minimize startup time. Each script runs in isolation and their environments are merged at the end. This can be changed to sequential execution via the strategy pattern.

### Why Null-Delimited Environment Parsing?
Using `env -0` ensures environment variables with newlines or special characters are parsed correctly.

### Why Separate Structs?
The code is organized into three main components (Shell, WshrcLoader, Environment) for:
- **Separation of concerns**: Each struct has a single responsibility
- **Testability**: Individual components can be tested in isolation
- **Maintainability**: Changes to one component don't affect others
- **Extensibility**: New features can be added without major refactoring

### Why Functional Patterns?
Functional programming patterns provide:
- **Flexibility**: Easy to customize behavior without modifying code
- **Composability**: Combine small pieces to create complex behavior
- **Testability**: Easy to mock and test individual components
- **Maintainability**: Clear separation of concerns and responsibilities

## Future Enhancements

Areas for potential expansion:
- Subcommands for built-in utilities (arg parser, etc.)
- Pipeline transformation pattern for environment processing
- Plugin system using middleware
- Configuration validation
- Performance profiling via middleware
- Shell function support
