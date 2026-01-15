# INSTRUCTIONS.md

This document provides guidelines for working on the warg project, particularly for AI assistants and future contributors.

## General Workflow

When given a task, follow these steps:

1. **Understand the Context**: Read relevant documentation (DESIGN.md, existing code)
2. **Implement the Solution**: Write production code
3. **Write Tests**: Always accompany implementation with comprehensive tests
4. **Validate**: Ensure all tests pass
5. **Document**: Create appropriate documentation

## Task Types

### Working on Issues

When told to work on an issue (by number or full name):

1. **Locate the Issue**
   - Look in the `issues/` directory
   - Example: Issue #1 → `issues/1-CLI-Definition-Format.md`
   - Read the entire issue thoroughly

2. **Understand the Problem**
   - Identify the core problem statement
   - Note all requirements and constraints
   - Consider implications and edge cases
   - Review related code in the codebase

3. **Think Through Solutions**
   - Consider multiple approaches if appropriate
   - Evaluate trade-offs (complexity, performance, maintainability)
   - Choose the best solution based on project goals
   - Don't rush to implementation - think first

4. **Implement the Solution**
   - Write clean, minimal, production-ready code
   - Follow existing code style and patterns
   - Add appropriate error handling
   - Keep changes surgical and focused

5. **Write Comprehensive Tests**
   - Test happy paths
   - Test edge cases and error conditions
   - Ensure test coverage is thorough
   - Tests should be clear and maintainable
   - Run tests and verify they all pass

6. **Update Examples**
   - Update example shell scripts in `examples/` directory when making changes to CLI interface
   - Ensure all examples still work with the new implementation
   - Keep examples up to date with current syntax and usage patterns

6. **Create Solution Documentation (When Appropriate)**
   - **Create a solution .md file ONLY if the task is architectural in nature**
   - Architectural tasks include:
     - Design decisions that affect system structure
     - Implementation approaches for complex features
     - Solutions that involve multiple components
     - Tasks requiring significant trade-off analysis
   - **Do NOT create solution .md files for:**
     - Simple bug fixes
     - Small feature additions
     - Code refactoring (unless major architecture change)
     - Documentation updates
     - Test additions
   
   **When to create solution documentation:**
   - If task output is primarily code/tests → No solution doc needed
   - If task involves significant design decisions → Create solution doc
   - If uncertain, err on the side of NOT creating documentation
   
   **If creating solution documentation:**
   - Create a markdown file in `solutions/` directory
   - Name it exactly the same as the issue file
   - Example: `issues/1-CLI-Definition-Format.md` → `solutions/1-CLI-Definition-Format.md`
   - Document:
     - Decision summary
     - Implementation approach
     - Trade-offs considered
     - Usage examples
     - Testing strategy
     - Future considerations

7. **Update Examples and Documentation**
   - Update example shell scripts in `examples/` directory when making changes to CLI interface
   - Ensure all examples still work with the new implementation
   - Keep examples up to date with current syntax and usage patterns
   - Update relevant documentation files (README, DESIGN.md, etc.) if needed

### Refactoring Tasks

When asked to refactor existing code:

⚠️ **CRITICAL RULE: DO NOT CHANGE TESTS** ⚠️

1. **Preserve Test Behavior**
   - Tests define the contract/behavior
   - Tests must remain unchanged
   - If tests fail after refactor, fix the production code, not the tests

2. **Refactoring Process**
   - Run existing tests first to establish baseline
   - Make incremental changes to production code
   - Run tests after each change
   - Only proceed if all tests still pass

3. **When Tests Can Change**
   - Only if explicitly asked to modify tests
   - Only if fixing incorrect test behavior (rare)
   - Must get explicit confirmation first

4. **Focus Areas**
   - Improve code structure
   - Reduce complexity
   - Enhance readability
   - Optimize performance (if needed)
   - Remove duplication

### Adding New Features

1. **Design First**
   - Consider how it fits with existing architecture
   - Check DESIGN.md for guidance
   - Ensure consistency with project goals

2. **Implement with Tests**
   - Write tests alongside or before implementation (TDD encouraged)
   - Test both success and failure cases
   - Consider integration with existing features

3. **Document**
   - Update relevant documentation
   - Add code comments for complex logic (sparingly)
   - Include usage examples

## Code Quality Standards

### General Principles

- **Minimal Changes**: Make the smallest change needed to solve the problem
- **Surgical Precision**: Only modify what's necessary
- **Consistency**: Follow existing patterns and conventions
- **Clarity**: Code should be self-documenting when possible
- **Testing**: All code must be tested

### Go-Specific Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Write clear error messages
- Prefer explicit over implicit
- Keep functions small and focused
- Use meaningful names

### Testing Guidelines

- Test file naming: `*_test.go`
- Use table-driven tests when appropriate
- Test names should describe what they test: `TestParseInlineDefinition`
- Use subtests for multiple cases: `t.Run("case name", ...)`
- Assert clearly with helpful error messages
- Avoid test interdependencies

### Documentation Guidelines

- Use clear, concise language
- Include code examples
- Document edge cases and limitations
- Explain "why" not just "what"
- Keep documentation up to date

## Project Structure

```
warg/
├── cmd/warg/          # CLI application entry point
├── flags/             # Core flag parsing library
├── internal/          # Internal packages
├── issues/            # Issue descriptions (read-only)
├── solutions/         # Solution documentation (write here)
├── DESIGN.md          # Architecture and design decisions
├── ISSUES.md          # Issue tracking overview
└── INSTRUCTIONS.md    # This file
```

## Issue Resolution Workflow Example

Given: "Solve issue #5"

```bash
# 1. Read the issue
cat issues/5-Flag-Name-Inference.md

# 2. Understand the context
cat DESIGN.md
ls -la flags/

# 3. Think through the solution
# - What are the requirements?
# - What are the edge cases?
# - How does this fit with existing code?
# - Are there multiple approaches?

# 4. Implement
# Create/modify files in flags/ or other relevant directories

# 5. Write tests
# Create/modify *_test.go files

# 6. Validate
just test-all

# 7. Document solution
cat > solutions/5-Flag-Name-Inference.md <<EOF
# Solution to Issue #5: Flag Name Inference
...
EOF
```

## Common Pitfalls to Avoid

❌ **Don't:**
- Change tests during refactoring
- Rush to implementation without understanding
- Make sweeping changes when small changes suffice
- Add features not requested
- Skip testing edge cases
- Leave commented-out code
- Ignore existing patterns
- Create files in /tmp or system directories for testing

✅ **Do:**
- Read and understand the issue completely
- Think through multiple solutions
- Make minimal, focused changes
- Write comprehensive tests
- Follow existing code style
- Document decisions and trade-offs
- Validate with tests before considering done
- Use existing project directories (like `examples/`) for test scripts
- Clean up temporary test files when done

## Testing Commands

### Unit Tests

```bash
# Run all tests
just test

# Run tests with verbose output (already verbose by default)
go test ./... -v

# Run tests for specific package
just test-package flags

# Run specific test
go test ./flags/ -run TestParseInlineDefinition

# Run tests with coverage
just test-coverage
```

### Integration Testing

The warg project has two primary interfaces that need different testing approaches:

#### Testing CLI Interface (Bash Scripts)

For the CLI side, use the justfile recipe that runs integration tests:

```bash
# Run CLI integration tests
just test-cli
```

The `test-cli` recipe in `justfile` performs the following tests:
- Builds the warg CLI tool
- Tests all demo commands (inline, JSON, heredoc formats)
- Verifies output structure correctness
- Validates hierarchical flag structures
- Tests error handling

You can also create custom bash scripts for specific CLI testing scenarios. The justfile recipe provides a good template.

#### Testing Library Interface (Go Module)

For the library side, use the justfile recipe:

```bash
# Run library integration tests
just test-lib
```

**Note**: A library integration test already exists at `../warg-lib-test/`

The `test-lib` recipe runs the integration test that imports warg as a module and validates all public APIs work correctly.

To manually run the integration test:

```bash
cd ../warg-lib-test
go run main.go
```

To create a new integration test from scratch:

```bash
# Create a test directory (e.g., at the same level as warg)
cd /Users/vojta.vojacek/repos/wsh
mkdir -p my-warg-test
cd my-warg-test

# Initialize a new Go module
go mod init my-warg-test

# Add warg as a dependency (use local path during development)
go mod edit -replace V-Woodpecker-V/wsh/warg=../warg
go get V-Woodpecker-V/wsh/warg/flags
To create a new integration test from scratch, see the example at `../warg-lib-test/` for reference.

### Testing Strategy Summary

| Interface | Testing Method | Purpose |
|-----------|---------------|---------|
| **Unit Tests** | `just test` | Test individual functions and packages |
| **CLI Interface** | `just test-cli` | Test real-world CLI usage and argument parsing |
| **Library Interface** | `just test-lib` | Test that warg can be imported and used as a library |

**Best Practice**: Always test both interfaces when making changes:
1. Run unit tests: `just test`
2. Run CLI integration: `just test-cli`
3. Run library integration: `just test-lib`

Or run all three levels at once:
```bash
just test-all
```

This ensures the library works both as a standalone CLI tool and as an importable Go library.

## Building

```bash
# Build the CLI tool
just build

# Build and install to ~/bin
just install

# Build from root workspace
cd .. && just build_warg
```

## Questions Before Starting

Before implementing a solution, consider:

1. **Scope**: What exactly needs to be changed?
2. **Impact**: What other parts of the system are affected?
3. **Testing**: How will this be tested?
4. **Backwards Compatibility**: Will this break existing code?
5. **Documentation**: What documentation needs updating?

## Final Checklist

Before considering a task complete:

- [ ] Issue/requirement fully understood
- [ ] Solution implemented and working
- [ ] All unit tests written and passing (`go test ./...`)
- [ ] Code follows project conventions
- [ ] No unnecessary changes made
- [ ] Documentation created/updated (if needed)
- [ ] Solution markdown file created (ONLY if architectural task)
- [ ] Validated with all three test levels:
  - [ ] Unit tests: `just test`
  - [ ] CLI integration: `just test-cli`
  - [ ] Library integration: `just test-lib`
  - [ ] Or run all at once: `just test-all`

## Getting Help

If uncertain about:
- **Architecture decisions**: Check DESIGN.md
- **Existing patterns**: Look at similar code in the codebase
- **Requirements**: Re-read the issue carefully
- **Trade-offs**: Document multiple options and ask for guidance

Remember: It's better to ask for clarification than to implement the wrong solution.
