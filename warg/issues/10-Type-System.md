# Issue #10: Type System for Values

## Problem
Flags can have values beyond just strings and bools:
- Integers
- Floats
- Enums/choices
- File paths
- URLs
- Lists/arrays
- etc.

How should warg handle type validation and conversion?

## Implications
- CLI mode needs string validation
- Library mode needs type conversion
- Error messages must indicate type mismatches
- Some types need special handling (files must exist?)

## Requirements (CLI Mode)
1. Validate string values against expected type
2. Provide clear error on type mismatch
3. Support common types out of box
4. Allow custom validation?

## Requirements (Library Mode)
1. Convert to Go types automatically
2. Support all common Go types
3. Support custom types via interfaces?
4. Handle pointer types (optional values)

## Types to Support
- **Primitives:** bool, string, int, float
- **Enums:** restricted set of string values
- **Lists:** comma-separated or repeated flags?
- **Paths:** file/directory existence checks?
- **Complex:** JSON objects, key=value maps?

## Repeated Flags
```bash
-v -v -v              # Verbosity level via counting?
--file a --file b     # Multiple values?
```

Should `-v` count increments, or does each need to be defined?

## Decision Needed
Type system design and validation strategy.
