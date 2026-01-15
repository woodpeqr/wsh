# Issue #12: Versioning & Compatibility

## Problem
As warg evolves, flag definitions and output formats may change.

## Implications
- Shell scripts using warg should not break
- Go code using library should have stable API
- Need to support multiple versions simultaneously?

## Requirements
1. Semantic versioning
2. Stable output format (or version flag)
3. Deprecation warnings for old formats
4. Migration guides

## Questions
1. Should warg support reading old definition formats?
2. Should CLI mode have a version flag to lock behavior?
3. How to handle breaking changes in library API?

## Decision Needed
Versioning and compatibility strategy.
