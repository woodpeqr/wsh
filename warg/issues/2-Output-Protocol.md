# Issue #2: Output Protocol for CLI Mode

## Problem
When warg parses arguments for an external program, how does it communicate the results back?

## Implications
- Affects how every external program integrates with warg
- Must be parseable by bash, Python, Ruby, etc.
- Must handle complex nested structures
- Must distinguish between "flag not set" vs "flag set with empty value"
- Must handle errors (unknown flag, missing value, etc.)

## Requirements
1. Machine-readable output format
2. Easy to parse in bash (most important use case)
3. Preserve hierarchy/context information
4. Clear error reporting
5. Support both switches and values

## Options to Evaluate

### Option A: Shell Variable Assignments
```bash
eval "$(warg ...)"
# Sets variables like:
WARG_NAME="value"
WARG_VERBOSE=1
WARG_GIT=1
WARG_GIT_COMMIT=1
WARG_GIT_MESSAGE="fix bug"
```
**Pros:** Native bash integration, easy to use
**Cons:** Security risk with `eval`, namespace pollution, limited nesting representation

### Option B: JSON Output
```json
{
  "name": "value",
  "verbose": true,
  "git": {
    "commit": true,
    "message": "fix bug"
  }
}
```
**Pros:** Standard format, preserves structure
**Cons:** Requires JSON parser in bash (jq), more complex

### Option C: Key-Value Pairs (Flat)
```
name=value
verbose=true
git=true
git.commit=true
git.message=fix bug
```
**Pros:** Simple to parse, no eval needed
**Cons:** String values with spaces/newlines need escaping, lost hierarchy

### Option D: Null-Delimited Records
```
name\0value\0verbose\0true\0git.commit\0true\0git.message\0fix bug\0
```
**Pros:** Safe for all values, easy to parse
**Cons:** Less human-readable, requires careful parsing

## Decision Needed
What format should warg output? Support multiple formats via flag?
