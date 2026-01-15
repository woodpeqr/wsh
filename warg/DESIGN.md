# warg - Architecture & Design

## Overview
A dual-purpose argument parsing library that works both as a Go library and as a CLI tool for any language.

## Core Features

### 1. Short Flag Combining
Unlike Go's standard `flag` package, supports combining short flags:
```bash
-a -b -c  →  -abc
```

### 2. Hierarchical Context System
Flags can have subflags with context inheritance. Context flags are capital letters by convention in short form.

**Example:**
```bash
--git --commit --message "fix"  →  -Gcm "fix"
```

Where:
- `-G` = `--git` (context flag by convention)
- `-c` = `--commit` (flag within git context)
- `-m` = `--message` (flag within git context, requires value)

### 3. Context Resolution
When a flag is not found in the current context, parser walks up to parent context and retries until:
- Flag is found, OR
- No parent context exists (error)

**Rule:** If flag exists in current context, use it. Otherwise check parent.

### 4. Infinite Nesting Depth
Context nesting is theoretically unlimited:
```bash
-Gcs  # git → commit → sign (if defined that way)
-Gcs  # git → commit, also git → sign (if both c and s are in git context)
```
The structure depends entirely on flag definitions.

## Data Model

### Flag Definition (Language-Agnostic)
Each flag has:
- **Names**: `[]string` - one or more names (e.g., `["-n", "--name"]`)
  - Short form: starts with single `-`
  - Long form: starts with `--`
- **Type**: 
  - CLI mode: `bool` (switch) or `string` (value)
  - Library mode: Generic type `T`
- **Description**: `string` - help text
- **Children**: `[]FlagDef` - optional subflags (recursive)

### Interface Layers

#### WFlag (CLI Mode)
For use by external programs via CLI interface.
- Works with string values
- Returns parsed results via stdout/JSON/protocol TBD

#### LibFlag[T] (Library Mode)  
For use within Go programs.
- Uses Go generics for type safety
- Directly sets typed pointers in user structs
- Uses struct tags for definition: `warg:"-n,--name; description"`

## Design Decisions (In Progress)

### CLI Definition Format
**Requirement:** Must be concise and bash-friendly.

**Options to explore:**
- Inline format: `warg -D "n,name;description" -D "g,git;Git context" ...`
- Config file: JSON/YAML schema
- Environment variables
- Heredoc or piped definition

**TBD:** How external programs pass flag definitions to warg CLI.

### Parsing Algorithm
```
For each arg token:
  1. If it's a combined short flag (e.g., -abc):
     - Split into individual flags
     - Process each in order
  2. For each flag:
     - Look up in current context
     - If not found, walk up context chain
     - If found, handle based on type:
       - Switch: Set to true
       - Value: Consume next arg token
       - Context: Push onto context stack
  3. If flag requires value but no more tokens: error
  4. If flag not found in any context: error
```

### Ambiguity Handling
- Current context takes precedence
- Parent contexts are fallback only
- Capital letters are context flags **by convention** only (not enforced)

## Implementation Status
- [ ] Core flag definition structures
- [ ] Context tree data structure
- [ ] Parsing algorithm
- [ ] CLI input format design
- [ ] Library mode with struct tags
- [ ] CLI mode with external definition
- [ ] Flag resolution with context chain
- [ ] Short flag combining logic
- [ ] Value vs switch handling
- [ ] Help text generation
