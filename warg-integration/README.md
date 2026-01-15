# warg Integration Tests

This package contains integration tests for the warg library, demonstrating its usage as an external Go module.

## Purpose

These tests validate that:
- The warg library can be imported and used from external projects
- All public APIs work correctly
- Type inference works as expected
- Slices, contexts, and custom types function properly
- Error handling behaves correctly

## Running the Tests

```bash
# Run all integration tests
go test -v

# Run specific test
go test -v -run TestBasicTypes

# Run with coverage
go test -v -cover

# Run from the warg project root
cd ../warg
just test-lib
```

## Test Coverage

The integration tests cover:

- ✅ **Basic Types** - bool, string, int, float
- ✅ **Slice Types** - []string, []int, []float64 with repeated flags
- ✅ **Hierarchical Contexts** - Nested flag structures
- ✅ **Custom Types** - time.Duration with various formats
- ✅ **Immutability** - Parser instances are independent
- ✅ **Error Handling** - Unknown flags, missing values, invalid types
- ✅ **Edge Cases** - Empty args, unsigned ints, long names
- ✅ **Combined Short Flags** - -vdi style combinations
- ✅ **Real World Scenarios** - Complete application flag sets

## Dependencies

This project uses:
- `github.com/stretchr/testify` - For assertions and test utilities
- `V-Woodpecker-V/wsh/warg/lib` - The warg library being tested

## Structure

```
warg-integration/
├── README.md              # This file
├── go.mod                 # Module definition with local replace
├── integration_test.go    # All integration tests
└── go.sum                 # Dependency checksums
```

## Example Test

```go
func TestBasicTypes(t *testing.T) {
    var verbose bool
    var name string
    
    parser := lib.New().
        Flag(&verbose, []string{"v", "verbose"}, "Verbose").
        Flag(&name, []string{"n", "name"}, "Name")
    
    result := parser.Parse([]string{"-v", "--name", "Alice"})
    
    require.Empty(t, result.Errors, "should parse without errors")
    assert.True(t, verbose)
    assert.Equal(t, "Alice", name)
}
```

## Integration with CI/CD

These tests are automatically run as part of the warg project's test suite:

```bash
cd ../warg
just test-all  # Runs unit, CLI, and library integration tests
```

The `test-lib` justfile recipe runs these integration tests.
