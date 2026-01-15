# Issue #8: Error Handling & Reporting

## Problem
Many things can go wrong during parsing:
- Unknown flag
- Missing required value
- Invalid value for type
- Context mismatch
- Malformed flag definition
- etc.

How should warg report errors in both modes?

## Implications
- Good errors are critical for UX
- CLI mode vs library mode may need different approaches
- Error messages should suggest fixes
- Must work well with bash error handling

## Requirements (CLI Mode)
1. Exit with non-zero status on error
2. Write error to stderr
3. Provide context (which flag, which argument)
4. Suggest valid alternatives when possible
5. Support different verbosity levels?

## Requirements (Library Mode)
1. Return typed errors
2. Support error wrapping
3. Provide programmatic access to error details
4. Support validation errors

## Error Categories
1. **Definition errors** - invalid flag structure (caught early)
2. **Parse errors** - unknown flag, missing value (during parsing)
3. **Validation errors** - invalid value format (post-parse)
4. **Usage errors** - conflicting flags, required flag missing

## Decision Needed
Error handling strategy and message formats.
