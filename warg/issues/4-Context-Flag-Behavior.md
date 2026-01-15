# Issue #4: Context Flag Behavior

## Problem
When a flag is defined as a "context flag," what exactly does it do? Is it:
- A switch that also opens a context?
- Purely a namespace with no value?
- Something else?

## Implications
- Affects whether context flags themselves can be "set" or "unset"
- Affects help text generation
- Affects struct tag design for library mode
- Affects how context flags are represented in output

## Requirements
1. Must support infinite nesting
2. Must work in both CLI and library modes
3. Must be queryable (did user specify --git?)
4. Context resolution must be efficient

## Scenarios to Consider

### Scenario A: Context as Pure Namespace
```bash
-Gc    # git.commit = true, git = <not a value>
```
Git context is entered but `-G` itself has no boolean value.

**Question:** Can you ask "was --git specified?" No, only "was something in git context specified?"

### Scenario B: Context as Switch + Namespace
```bash
-Gc    # git = true, git.commit = true
-G     # git = true (valid on its own)
```
Git context is entered AND `-G` acts as a switch flag.

**Question:** What does `-G` alone mean? Just entering context with no subflags?

### Scenario C: Context Requires Child Flags
```bash
-G     # ERROR: context flag requires at least one child
-Gc    # OK
```
Context flags are invalid without children.

**Question:** Is this too restrictive?

## Decision Needed
What is the semantic meaning of a context flag?
