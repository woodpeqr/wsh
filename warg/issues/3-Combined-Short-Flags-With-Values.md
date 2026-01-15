# Issue #3: Combined Short Flags with Values

## Problem
When short flags are combined (e.g., `-abc`), what happens if one requires a value?

**Example:**
```bash
# If -a is switch, -b is switch, -c requires value
-abc value    # Does this work?
-abcvalue     # Or this?
-ab -c value  # Or must it be separate?
```

## Implications
- Affects parsing algorithm complexity
- Affects user experience and intuition
- Must be consistent with Unix conventions
- Edge cases multiply with context flags

## Requirements
1. Behavior must be predictable
2. Should align with common Unix tools where possible
3. Must work with context flags (capital letters)
4. Error messages must be clear when misused

## Options to Evaluate

### Option A: Value Must Follow Combined Flags
```bash
-abc value    # -c takes "value"
-abcvalue     # ERROR: unexpected 'v'
```
**Pros:** Simple rule, clear separation
**Cons:** Can't have value immediately after

### Option B: Value Can Be Attached or Separate
```bash
-abc value    # -c takes "value"
-abcvalue     # -c takes "value"
```
**Pros:** Flexible, matches some tools (tar -xvf vs -xvffile)
**Cons:** Ambiguous when combined with context flags

### Option C: Value Flag Must Be Last in Combination
```bash
-abc value    # OK: -c takes "value"
-acb value    # ERROR: -c must be last in combination
```
**Pros:** Clear rule, prevents ambiguity
**Cons:** Restrictive, user must remember position

## Edge Cases
1. Context flag in middle: `-Gcm "msg"` - does `-G` enter context, then `-c` and `-m`?
2. Multiple value flags: `-am "msg"` - if both need values, how to handle?
3. Attached values: `-Gcm"msg"` - should this work?

## Decision Needed
What rules govern value flags in combinations?
