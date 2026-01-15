# Issue #5: Flag Name Inference (Short vs Long)

## Problem
Flags are defined with names like `["-n", "--name"]`. The parser must infer:
- Which is short form (single `-`)
- Which is long form (double `--`)
- What if user provides `["n", "name"]` without dashes?
- What if both are short or both are long?

## Implications
- Affects validation during flag definition
- Affects error messages
- Affects both library mode (struct tags) and CLI mode (definition format)

## Requirements
1. Must support flags with only short form, only long form, or both
2. Must validate that short forms are single character (or allow multi-char?)
3. Must handle edge cases gracefully

## Edge Cases
1. `["--name"]` - only long form, no short form (valid?)
2. `["-n"]` - only short form, no long form (valid?)
3. `["-nn", "--name"]` - two-character short form (valid? or error?)
4. `["n", "name"]` - no dashes (should warg add them? or error?)
5. `["-n", "-N"]` - two short forms, different case (valid?)

## Options to Evaluate

### Option A: Strict Inference
- Names starting with `--` are long
- Names starting with `-` (single) are short
- Short forms must be exactly 1 character after `-`
- Must have at least one name

### Option B: Flexible Inference
- Allow multi-character short forms (e.g., `-nn`)
- Auto-add dashes if missing based on length
- Allow flags with no short form or no long form

### Option C: Explicit Type Annotation
```
-D "n:short,name:long;string;Description"
```
User explicitly marks each name as short or long.

## Decision Needed
How strict should name validation be?
