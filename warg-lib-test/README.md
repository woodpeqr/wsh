# warg Library Integration Test

This directory contains an integration test that will validate warg works correctly when imported as a library by external Go projects.

## Current Status: TODO

⚠️ **This test is currently a placeholder.** The library API for warg does not exist yet.

### What's Missing

The warg project currently only has CLI-specific functions exposed (for parsing flag definitions). These are meant for the CLI tool, not for Go programs that want to use warg as a library.

The library API should allow Go programs to:
- Register flags via struct tags
- Parse command-line arguments  
- Extract typed values
- Use context/hierarchical flags

### Example of Desired Library API

```go
package main

import "V-Woodpecker-V/wsh/warg"

type Config struct {
    Verbose bool   `warg:"-v,--verbose; Verbose output"`
    Name    string `warg:"-n,--name; User name"`
    Git     struct {
        Commit  bool   `warg:"-c,--commit; Commit changes"`
        Message string `warg:"-m,--message; Commit message"`
    } `warg:"-G,--git; Git operations"`
}

func main() {
    var cfg Config
    warg.Parse(&cfg)
    
    if cfg.Verbose {
        println("Verbose mode enabled")
    }
    // Use cfg fields...
}
```

## Running This Test

```bash
cd ../warg-lib-test
go run main.go
```

Currently outputs a TODO message explaining the status.

## What Needs to Happen

See `warg/TODO.md` for the full task list:

1. **Issue #1**: Separate CLI and library APIs
   - Move CLI parsing functions to `internal/cli/`
   - Create public library API in `flags/` package

2. **Issue #2**: Fix this integration test
   - Once library API exists, write real integration tests
   - Test struct tag parsing
   - Test argument parsing
   - Test value extraction

## When to Re-enable

This test should be re-enabled and filled out after:
- [ ] Library API is implemented
- [ ] Struct tag parsing works
- [ ] Argument parsing works
- [ ] Value extraction works

Until then, `just test-lib` will show the TODO message.
