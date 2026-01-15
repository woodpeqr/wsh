# Issue #7: Help Text Generation

## Problem
Users will need `--help` functionality. How should warg generate and format help text?

## Implications
- Every flag definition must support help generation
- Must show hierarchy clearly
- Must be readable for complex flag trees
- Must integrate with both CLI and library modes

## Requirements
1. Auto-generate from flag definitions
2. Show short and long forms
3. Indicate required vs optional
4. Show default values
5. Display context hierarchy
6. Wrap long descriptions
7. Support custom help text templates?

## Format Questions

### How to show hierarchy?
```
Options:
  -n, --name <string>     User name
  -v, --verbose           Verbose output
  -G, --git               Git operations
      -c, --commit        Commit changes
      -m, --message <str> Commit message
      -s, --stash         Stash changes
```
Indentation shows context?

### How to show context in usage?
```
Usage: program [OPTIONS]
       program -G [GIT_OPTIONS]

Options:
  ...
```
Separate usage lines per context?

### Should help be context-aware?
```bash
program --help              # Show all flags
program --git --help        # Show only git context flags?
```

## Decision Needed
Help text format and generation strategy.
