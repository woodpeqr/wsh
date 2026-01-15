# TODO: Self-Host warg CLI & Enhance lib API

## Overview

Refactor the warg CLI tool to use its own `lib` package for flag parsing, and enhance the lib API to better support hierarchical, repeatable contexts with type-safe struct-based parsing.

**Type**: Feature Enhancement + Refactoring Task  
**Priority**: High  
**Constraint**: All existing tests must pass without modification

---

## Current Problems

### 1. Manual Parsing in warg CLI

**Location**: `/Users/vojta.vojacek/repos/wsh/warg/cmd/warg/main.go`

The `parseAddFlags()` function (lines 184-259) manually parses the `-A` inline format:

```bash
warg -A -n v,verbose -s -d "Verbose" -A -n n,name -d "Name" -- -v --name Alice
```

**Current behavior:**
- `-A` / `--add` starts a flag definition
- `-n` / `--name` specifies flag names (only valid within `-A` context)
- `-s` / `--switch` marks flag as switch (only valid within `-A` context)
- `-d` / `--description` provides help text (only valid within `-A` context)
- `-A` can be repeated to define multiple flags

**Problem:** This uses ad-hoc manual parsing instead of the lib package. warg should dogfood its own library.

### 2. Contexts Are Not Repeatable

**Current lib API:**
```go
parser := lib.New().
    Context([]string{"G", "git"}, "Git operations", func(git *Parser) *Parser {
        return git.
            Flag(&commitFlag, []string{"c", "commit"}, "Commit").
            Flag(&message, []string{"m"}, "Message")
    })
```

**Problem:** 
- Context can only appear once: `-G -c -m "msg"` works
- But `-G -c -G -p` (two separate git contexts) doesn't work as expected
- The `-A` use case REQUIRES repeatability to define multiple flags

### 3. No Type-Safe Way to Group Related Flags

**Current approach - scattered variables:**
```go
var verbose bool
var name string
var output string
var gitCommit bool
var gitMessage string

parser := lib.New().
    Flag(&verbose, []string{"v", "verbose"}, "Verbose").
    Flag(&name, []string{"n", "name"}, "Name").
    Flag(&output, []string{"o", "output"}, "Output").
    Context([]string{"G", "git"}, "Git ops", func(git *Parser) *Parser {
        return git.
            Flag(&gitCommit, []string{"c", "commit"}, "Commit").
            Flag(&gitMessage, []string{"m"}, "Message")
    })
```

**Problems:**
- All variables scattered in scope
- No clear grouping showing relationship
- Can't pass around "all git flags" as a single value
- No type safety for hierarchical structure

**Desired approach - struct-based:**
```go
type Config struct {
    Verbose bool
    Name    string
    Output  string
    Git     GitConfig
}

type GitConfig struct {
    Commit  bool
    Message string
}

var config Config
parser := lib.New().
    Flag(&config.Verbose, []string{"v", "verbose"}, "Verbose").
    Flag(&config.Name, []string{"n", "name"}, "Name").
    Flag(&config.Output, []string{"o", "output"}, "Output").
    Context(&config.Git, []string{"G", "git"}, "Git ops", ???)
```

**But:** How do we define child flags of `Git` without struct tags or reflection magic?

---

## Design Constraints & Preferences

### User Preferences
1. ✅ **Like:** Context API with builder functions - clean and explicit
2. ✅ **Like:** Having all args grouped in a single struct - type-safe, organized
3. ❌ **Dislike:** Struct tags - too much magic, hard to refactor, IDE support lacking
4. ✅ **Want:** Type safety - no `map[string]interface{}`

### Technical Constraints
1. ✅ Go doesn't support dynamic struct creation
2. ✅ All existing tests must pass unchanged
3. ✅ API should feel natural and idiomatic
4. ✅ Should reduce complexity, not increase it

---

## Proposed Solutions

### Option A: Struct-Based Context with Builder

Allow Context to accept a struct pointer and use builder to define children:

```go
type GitConfig struct {
    Commit  bool
    Message string
}

type Config struct {
    Verbose bool
    Name    string
    Git     GitConfig
}

var config Config
parser := lib.New().
    Flag(&config.Verbose, []string{"v", "verbose"}, "Verbose").
    Flag(&config.Name, []string{"n", "name"}, "Name").
    Context(&config.Git, []string{"G", "git"}, "Git ops", func(git *Parser) *Parser {
        return git.
            Flag(&config.Git.Commit, []string{"c", "commit"}, "Commit").
            Flag(&config.Git.Message, []string{"m"}, "Message")
    })
```

**Pros:**
- ✅ No struct tags
- ✅ Type-safe struct grouping
- ✅ Explicit builder function (preferred style)
- ✅ Natural Go code

**Cons:**
- ❌ Repetition: `&config.Git` and `&config.Git.Commit`
- ❌ Doesn't solve repeatability yet

### Option B: Slice of Structs for Repeatable Contexts

Extend Option A to support slices:

```go
type AddFlagDef struct {
    Names       string
    IsSwitch    bool
    Description string
}

var addFlags []AddFlagDef
parser := lib.New().
    Context(&addFlags, []string{"A", "add"}, "Add flag", func(a *Parser) *Parser {
        // How do we reference the "current" AddFlagDef being populated?
        // We need a way to get the slice element being filled...
        return a.
            Flag(&???.Names, []string{"n", "name"}, "Names").
            Flag(&???.IsSwitch, []string{"s", "switch"}, "Switch").
            Flag(&???.Description, []string{"d", "description"}, "Description")
    })
```

**Problem:** 
- The builder function doesn't know WHICH slice element to populate
- We'd need to pass an index or reference to "current element"
- Gets messy quickly

### Option C: Factory Function Pattern

```go
type AddFlagDef struct {
    Names       string
    IsSwitch    bool
    Description string
}

var addFlags []AddFlagDef
parser := lib.New().
    RepeatableContext(
        &addFlags, 
        []string{"A", "add"}, 
        "Add flag",
        func() interface{} { return &AddFlagDef{} },  // Factory
        func(item interface{}, p *Parser) *Parser {    // Builder
            def := item.(*AddFlagDef)
            return p.
                Flag(&def.Names, []string{"n", "name"}, "Names").
                Flag(&def.IsSwitch, []string{"s", "switch"}, "Switch").
                Flag(&def.Description, []string{"d", "description"}, "Description")
        },
    )
```

**Pros:**
- ✅ No struct tags
- ✅ Type-safe
- ✅ Explicitly handles repeatability
- ✅ Builder gets reference to current item

**Cons:**
- ❌ Complex API
- ❌ Type assertion needed (`item.(*AddFlagDef)`)
- ❌ Verbose

### Option D: Struct Tags (Not Preferred)

```go
type AddFlagDef struct {
    Names       string `warg:"n,name;Comma-separated flag names"`
    IsSwitch    bool   `warg:"s,switch;Flag is a switch"`
    Description string `warg:"d,description;Help message"`
}

var addFlags []AddFlagDef
parser := lib.New().
    Flag(&addFlags, []string{"A", "add"}, "Add a flag definition")
```

**Pros:**
- ✅ Concise
- ✅ Naturally handles slices/repeatability
- ✅ Type-safe

**Cons:**
- ❌ **User dislikes struct tags**
- ❌ Magic/implicit behavior
- ❌ Hard to refactor

### Option E: Hybrid - Context Builder with Slice Support

Modify Context to detect slices and handle them specially:

```go
type AddFlagDef struct {
    Names       string
    IsSwitch    bool
    Description string
}

var addFlags []AddFlagDef
parser := lib.New().
    Context(&addFlags, []string{"A", "add"}, "Add flag", func(p *Parser, current *AddFlagDef) *Parser {
        return p.
            Flag(&current.Names, []string{"n", "name"}, "Names").
            Flag(&current.IsSwitch, []string{"s", "switch"}, "Switch").
            Flag(&current.Description, []string{"d", "description"}, "Description")
    })
```

**How it works:**
- Detect if first param is a slice pointer
- Builder function signature changes to include `current` item reference
- Each context invocation appends to slice and populates that element

**Pros:**
- ✅ No struct tags
- ✅ Builder function (preferred style)
- ✅ Type-safe with generics or reflection
- ✅ Natural for repeatable contexts

**Cons:**
- ❌ Builder function signature differs between single/repeatable contexts
- ❌ Needs runtime type detection or generics

---

## Recommended Approach

### Phase 1: Enhance Context API for Repeatability

**Modify Context() signature to support both single and repeatable contexts:**

```go
// For single context
Context(contextPtr, names, desc, builder)

// For repeatable context (slice)
Context(slicePtr, names, desc, builderWithCurrentItem)
```

Use reflection to detect if first parameter is a slice, and adjust behavior accordingly.

### Phase 2: Refactor warg CLI to Use Enhanced API

```go
type AddFlagDef struct {
    Names       string
    IsSwitch    bool
    Description string
}

func main() {
    args := os.Args[1:]
    
    var helpFlag bool
    var addFlags []AddFlagDef
    
    parser := lib.New().
        Flag(&helpFlag, []string{"h", "help"}, "Show help").
        Context(&addFlags, []string{"A", "add"}, "Add flag definition", 
            func(p *Parser, def *AddFlagDef) *Parser {
                return p.
                    Flag(&def.Names, []string{"n", "name"}, "Flag names").
                    Flag(&def.IsSwitch, []string{"s", "switch"}, "Switch flag").
                    Flag(&def.Description, []string{"d", "description"}, "Description")
            })
    
    // ... rest of implementation
}
```

### Phase 3: Support Non-Repeatable Struct Contexts (Nice to Have)

Allow grouping flags even for single contexts:

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
    Context(&config.Git, []string{"G", "git"}, "Git ops", 
        func(p *Parser) *Parser {
            return p.
                Flag(&config.Git.Commit, []string{"c"}, "Commit").
                Flag(&config.Git.Message, []string{"m"}, "Message")
        })
```

---

## Implementation Tasks

### Task 1: Design Final API
- [ ] Decide between Option A, C, or E (or hybrid)
- [ ] Define Context signature for repeatable contexts
- [ ] Consider backwards compatibility
- [ ] Write API examples for common use cases

### Task 2: Implement Repeatable Context Support
- [ ] Modify `Context()` method in lib/parser.go
- [ ] Add slice detection logic
- [ ] Handle builder function with "current item" parameter
- [ ] Create appropriate flag definitions for child flags
- [ ] Implement setter logic to populate slice elements

### Task 3: Add Tests for Repeatable Contexts
- [ ] Test single context (existing behavior)
- [ ] Test repeatable context with slice
- [ ] Test multiple invocations of same context
- [ ] Test child flag validation (only valid within context)
- [ ] Test error cases

### Task 4: Refactor warg CLI
- [ ] Define AddFlagDef struct
- [ ] Replace parseAddFlags() with lib Context API
- [ ] Remove manual parsing code
- [ ] Validate all tests still pass

### Task 5: Documentation
- [ ] Update lib package documentation
- [ ] Add examples for repeatable contexts
- [ ] Document struct-based context usage
- [ ] Update DESIGN.md if needed

---

## Open Questions

1. **Builder function signature**: How to handle different signatures for single vs repeatable contexts?
   - Option A: Use reflection to detect and call appropriately
   - Option B: Use interface{} and type assertion
   - Option C: Use Go generics (requires Go 1.18+)

2. **Backwards compatibility**: Should existing Context() calls continue to work?
   - Yes: Need to support both old and new signatures
   - No: Breaking change, but cleaner API

3. **Child flag scope**: Should child flags be accessible outside their context?
   - Current: Child flags only valid when context is present
   - Desired: Same behavior, but enforced through struct scoping

4. **Help text generation**: How should repeatable contexts appear in help?
   ```
   -A, --add       Add a flag definition (repeatable)
     -n, --name      Flag names
     -s, --switch    Switch flag
     -d, --description  Description
   ```

5. **Empty contexts**: What if `-A` appears but no child flags are set?
   - Error immediately?
   - Allow and validate later?
   - Return partial struct with zero values?

---

## Success Criteria

- ✅ warg CLI uses lib to parse its own `-A` flags
- ✅ All existing tests pass without modification
- ✅ API supports repeatable contexts
- ✅ Type-safe struct-based flag grouping
- ✅ No struct tags required (or optional)
- ✅ Code is cleaner and more maintainable
- ✅ Help text generation works correctly
- ✅ Proves lib API is complete and production-ready

---

## Notes

- The `-A` use case is perfect for testing repeatable contexts
- Struct-based parsing enables better IDE support and type safety
- This enhancement makes lib more powerful for complex CLI tools
- Consider how this affects the JSON/heredoc input formats in warg
