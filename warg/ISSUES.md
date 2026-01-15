# warg - Open Issues & Design Problems

This document provides an index of open design issues. Each issue is tracked in a separate file in the `issues/` directory.

## Issues Index

1. [CLI Definition Format](issues/1-CLI-Definition-Format.md) - How external programs define flag structures
2. [Output Protocol](issues/2-Output-Protocol.md) - How warg communicates parsed results back
3. [Combined Short Flags with Values](issues/3-Combined-Short-Flags-With-Values.md) - Parsing `-abc value` scenarios
4. [Context Flag Behavior](issues/4-Context-Flag-Behavior.md) - Semantic meaning of context flags
5. [Flag Name Inference](issues/5-Flag-Name-Inference.md) - Validating short vs long flag names
6. [Struct Tag Format](issues/6-Struct-Tag-Format.md) - Library mode struct tag syntax
7. [Help Text Generation](issues/7-Help-Text-Generation.md) - Auto-generating help documentation
8. [Error Handling](issues/8-Error-Handling.md) - Error reporting strategy for both modes
9. [Positional Arguments](issues/9-Positional-Arguments.md) - Handling non-flag arguments
10. [Type System](issues/10-Type-System.md) - Type validation and conversion
11. [Performance](issues/11-Performance.md) - Performance constraints and optimization
12. [Versioning & Compatibility](issues/12-Versioning-Compatibility.md) - Backwards compatibility strategy

## Priority

**Blockers** (need decisions before implementation):
1. Issue #1 (CLI Definition Format)
2. Issue #2 (Output Protocol)
3. Issue #3 (Combined Flags with Values)
4. Issue #4 (Context Flag Behavior)
5. Issue #6 (Struct Tag Format)
6. Issue #9 (Positional Arguments)
7. Issue #10 (Type System)

**Can iterate on later**:
- Issue #7 (Help Text)
- Issue #5 (Name Inference)
- Issue #8 (Error Handling)
- Issue #11 (Performance)
- Issue #12 (Versioning)

---

This document previously contained all issues inline. They have been split into individual files for easier tracking and discussion.
