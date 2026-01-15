# Issue #6: Struct Tag Parsing for Library Mode

## Problem
Library mode uses struct tags like:
```go
Add struct {
    Names []string `warg:"-n,--names; names of the flag"`
} `warg:"-A,--add; add a new flag"`
```

How should the tag format work? How to represent:
- Type (inferred from field type?)
- Context flags (nested structs?)
- Required vs optional flags?
- Default values?
- Validation rules?

## Implications
- This is the **primary interface** for Go users
- Tag format affects ergonomics of library mode
- Must be expressive enough for complex flag trees
- Must integrate with Go's reflection system

## Requirements
1. Concise syntax for common cases
2. Support all features (context, types, validation)
3. Type inference from Go types where possible
4. Clear error messages for invalid tags
5. Support nested structs for context flags

## Tag Format Questions

### How to specify names?
```go
`warg:"-n,--name"`           // Comma-separated
`warg:"-n --name"`           // Space-separated
`warg:"names:-n,--name"`     // Explicit key
```

### How to specify description?
```go
`warg:"-n,--name; Description here"`        // Semicolon separator
`warg:"-n,--name desc:'Description'"`       // Key-value
`warg:"-n,--name" desc:"Description"`       // Multiple tags
```

### How to specify type?
```go
Value string `warg:"-v,--value; Description"`     // Infer from field type
Value string `warg:"-v,--value type:string; Description"`  // Explicit
```

### How to represent context?
```go
Git struct {
    Commit bool `warg:"-c,--commit; Commit changes"`
} `warg:"-G,--git; Git operations"`  // Nested struct = context
```

### How to handle optional fields?
```go
Value *string `warg:"-v,--value; Optional value"`   // Pointer = optional?
Value string  `warg:"-v,--value; Required value"`   // Non-pointer = required?
```

## Decision Needed
Define the struct tag format specification.
