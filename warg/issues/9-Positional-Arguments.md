# Issue #9: Positional Arguments

## Problem
Many CLIs have both flags and positional arguments:
```bash
git commit -m "message" file1.txt file2.txt
          ^--- flags ---^ ^-- positionals --^
```

How should warg handle positional arguments?

## Implications
- Affects parsing algorithm significantly
- Must distinguish flags from positionals
- Must preserve order (usually)
- Flags and positionals can be interleaved or separated by `--`

## Requirements
1. Support flags before, after, or mixed with positionals
2. Support `--` to end flag parsing (Unix convention)
3. Allow flag definitions to specify expected positional count
4. Handle greedy flags (last flag takes multiple values?)

## Options to Evaluate

### Option A: Everything After Flags
```bash
program -a -b positional1 positional2
# Once no more flags found, rest are positional
```

### Option B: Explicit Separator
```bash
program -a -b -- positional1 positional2
# `--` marks end of flags
```

### Option C: Interleaved Parsing
```bash
program -a positional1 -b positional2
# Flags can appear anywhere
```

### Option D: No Positional Support
```bash
# warg only handles flags, program handles positionals
program $(warg parse "$@")
```

## Edge Cases
1. What if a positional looks like a flag? (`--`)
2. Can context flags affect positional parsing?
3. Should positionals be returned in output?
4. Can positionals have validation rules?

## Decision Needed
How to handle positional arguments, if at all?
