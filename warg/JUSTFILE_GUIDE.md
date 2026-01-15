# Justfile Guide for warg

This project uses [just](https://github.com/casey/just) as a command runner, similar to Make but simpler and cross-platform.

## Installation

```bash
# macOS
brew install just

# Linux
cargo install just

# or download from https://github.com/casey/just/releases
```

## Available Commands

Run `just` or `just --list` to see all available recipes:

```bash
$ just --list
Available recipes:
    build                # Build the CLI tool
    clean                # Clean build artifacts
    default              # Show all available recipes
    fmt                  # Format code
    install              # Install the CLI tool
    lint                 # Run linter
    test                 # Run all unit tests
    test-all             # Run all three test levels
    test-cli             # Run CLI integration tests
    test-coverage        # Run unit tests with coverage
    test-lib             # Run library integration tests
    test-package package # Run specific package tests
    tree                 # Show project structure
```

## Common Workflows

### Development

```bash
# Run all tests
just test-all

# Run only unit tests
just test

# Format code
just fmt

# Run linter
just lint

# Build the CLI
just build
```

### Testing

```bash
# Run all three test levels (unit, CLI, library)
just test-all

# Run only unit tests
just test

# Run only CLI integration tests
just test-cli

# Run only library integration tests
just test-lib

# Run tests with coverage
just test-coverage

# Run tests for specific package
just test-package flags
```

### Building and Installing

```bash
# Build the CLI tool (creates ./warg binary)
just build

# Build and install to ~/bin
just install

# Clean build artifacts
just clean
```

## Root Workspace Commands

From the root `/wsh` directory:

```bash
# Build warg and install to bin/warg/
just build_warg

# Run all warg tests
just test_warg

# Build all projects
just build_all

# Clean all artifacts
just clean
```

## Integration with INSTRUCTIONS.md

When following INSTRUCTIONS.md for issue resolution:

1. **After making changes**, run:
   ```bash
   just test-all
   ```

2. **Before considering task complete**, ensure:
   - [ ] `just test` passes (unit tests)
   - [ ] `just test-cli` passes (CLI integration)
   - [ ] `just test-lib` passes (library integration)

3. **Quick iteration**:
   ```bash
   just fmt && just test
   ```

## Why Justfile?

- **Simple**: Easy to read and write
- **Fast**: No unnecessary rebuilds
- **Cross-platform**: Works on macOS, Linux, Windows
- **Better than bash scripts**: Type checking, error handling, built-in features
- **Task-oriented**: Clear separation of concerns
- **Self-documenting**: `just --list` shows all available commands

## Learn More

- [Official Documentation](https://just.systems/)
- [GitHub Repository](https://github.com/casey/just)
