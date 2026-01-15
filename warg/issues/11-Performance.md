# Issue #11: Performance Considerations

## Problem
For complex flag trees and large argument lists, performance matters.

## Implications
- Affects algorithm choices
- Affects data structure design
- Affects caching strategy
- CLI mode must be fast (used in scripts)

## Areas to Consider
1. **Flag lookup:** Hash map vs tree traversal?
2. **Context resolution:** Cache parent chain?
3. **Struct tag parsing:** Parse once or every time?
4. **Memory allocation:** Pool objects or not?
5. **Reflection overhead:** Library mode uses reflection extensively

## Requirements
1. Parse typical CLI args in < 1ms
2. Support 100+ flags without degradation
3. Support 10+ nesting levels
4. Minimize allocations in hot paths

## Decision Needed
Are there performance constraints that affect design choices?
