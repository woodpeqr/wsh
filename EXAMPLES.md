# wsh - Usage Examples

This document demonstrates the functional programming patterns available in wsh.

## Basic Usage

### Default Configuration

```go
// Uses default settings: parallel execution, no middleware
shell, err := NewShell()
if err != nil {
    log.Fatal(err)
}

exitCode := shell.Run("echo hello", nil)
```

## Functional Options Pattern

### Custom Paths

```go
// Override zsh and wshrc paths
shell, err := NewShell(
    WithZshPath("/usr/local/bin/zsh"),
    WithWshrcPath("/etc/wshrc"),
)
```

### Testing Setup

```go
// Perfect for unit tests - use temporary directories
tmpDir := t.TempDir()
shell, err := NewShell(
    WithWshrcPath(filepath.Join(tmpDir, ".wshrc")),
)
```

### Custom Components

```go
// Inject custom environment handler or loader
customEnv := NewEnvironment()
shell, err := NewShell(
    WithEnvironment(customEnv),
)
```

## Execution Strategies

### Parallel Execution (Default)

```go
// Scripts execute concurrently for faster startup
loader := NewWshrcLoader(zshPath,
    WithExecutionStrategy(ParallelExecutionStrategy),
)
```

### Sequential Execution

```go
// Scripts execute one at a time - useful for debugging
loader := NewWshrcLoader(zshPath,
    WithExecutionStrategy(SequentialExecutionStrategy),
)
```

### Custom Strategy

```go
// Implement your own execution strategy
func CustomStrategy(zshPath string, scripts []string, executor ScriptExecutor) (map[string]string, error) {
    // Your custom logic here
    // Maybe execute in specific order, with dependencies, etc.
    return merged, nil
}

loader := NewWshrcLoader(zshPath,
    WithExecutionStrategy(CustomStrategy),
)
```

## Middleware

### Logging

```go
// Log all script executions with timing
logger := func(format string, args ...interface{}) {
    log.Printf(format, args...)
}

loader := NewWshrcLoader(zshPath,
    WithMiddleware(WithLogging(logger)),
)
```

### Timeout Protection

```go
// Prevent scripts from hanging forever
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithTimeout(5 * time.Second),
    ),
)
```

### Error Recovery

```go
// Recover from panics in scripts
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithErrorRecovery(),
    ),
)
```

### Retry Logic

```go
// Retry failed scripts
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithRetry(3, 100*time.Millisecond), // 3 retries, 100ms delay
    ),
)
```

### Environment Filtering

```go
// Only keep certain environment variables
onlyMyVars := func(key, value string) bool {
    return strings.HasPrefix(key, "MY_")
}

loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithEnvFilter(onlyMyVars),
    ),
)
```

### Caching

```go
// Cache script results for repeated executions
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithCaching(),
    ),
)
```

## Composing Middleware

Middleware can be chained together:

```go
// Production configuration: timeout + logging + recovery
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithTimeout(10 * time.Second),
        WithLogging(log.Printf),
        WithErrorRecovery(),
    ),
)
```

```go
// Development configuration: sequential + logging + retry
loader := NewWshrcLoader(zshPath,
    WithExecutionStrategy(SequentialExecutionStrategy),
    WithMiddleware(
        WithRetry(2, 50*time.Millisecond),
        WithLogging(log.Printf),
    ),
)
```

```go
// Testing configuration: custom executor + caching
mockExecutor := func(zshPath, scriptPath string) (map[string]string, error) {
    return map[string]string{"TEST": "value"}, nil
}

loader := NewWshrcLoader(zshPath,
    WithScriptExecutor(mockExecutor),
    WithMiddleware(
        WithCaching(),
        WithLogging(t.Logf),
    ),
)
```

## Custom Middleware

Create your own middleware for specific needs:

```go
// Metrics collection middleware
func WithMetrics(collector MetricsCollector) ScriptMiddleware {
    return func(next ScriptExecutor) ScriptExecutor {
        return func(zshPath, scriptPath string) (map[string]string, error) {
            start := time.Now()
            env, err := next(zshPath, scriptPath)

            collector.RecordDuration(scriptPath, time.Since(start))
            if err != nil {
                collector.RecordError(scriptPath)
            }

            return env, err
        }
    }
}

// Usage
loader := NewWshrcLoader(zshPath,
    WithMiddleware(
        WithMetrics(myMetricsCollector),
    ),
)
```

```go
// Rate limiting middleware
func WithRateLimit(maxPerSecond int) ScriptMiddleware {
    limiter := rate.NewLimiter(rate.Limit(maxPerSecond), 1)

    return func(next ScriptExecutor) ScriptExecutor {
        return func(zshPath, scriptPath string) (map[string]string, error) {
            if err := limiter.Wait(context.Background()); err != nil {
                return nil, err
            }
            return next(zshPath, scriptPath)
        }
    }
}
```

## Complete Example: Production Shell

```go
package main

import (
    "log"
    "os"
    "time"
)

func main() {
    // Production configuration with all bells and whistles
    shell, err := NewShell(
        WithWshrcPath(os.Getenv("WSH_RC_PATH")), // Allow override via env
        WithWshrcLoader(
            NewWshrcLoader(
                findZshOrDefault(),
                WithExecutionStrategy(ParallelExecutionStrategy),
                WithMiddleware(
                    WithTimeout(30 * time.Second),      // Prevent hangs
                    WithRetry(2, 100*time.Millisecond), // Retry transient failures
                    WithLogging(log.Printf),            // Log everything
                    WithErrorRecovery(),                // Recover from panics
                ),
            ),
        ),
    )

    if err != nil {
        log.Fatalf("Failed to initialize shell: %v", err)
    }

    // Run the shell
    os.Exit(shell.Run(os.Args[1], os.Args[2:]))
}
```

## Pipeline Processing Example

While not implemented yet, here's how the pipeline pattern could work:

```go
// Transform environment variables through a pipeline
type EnvTransform func(map[string]string) map[string]string

func FilterPrefix(prefix string) EnvTransform {
    return func(env map[string]string) map[string]string {
        filtered := make(map[string]string)
        for k, v := range env {
            if strings.HasPrefix(k, prefix) {
                filtered[k] = v
            }
        }
        return filtered
    }
}

func AddPrefix(prefix string) EnvTransform {
    return func(env map[string]string) map[string]string {
        prefixed := make(map[string]string)
        for k, v := range env {
            prefixed[prefix+k] = v
        }
        return prefixed
    }
}

// Usage (if implemented)
env := GetEnvironment()
env = FilterPrefix("MY_")(env)
env = AddPrefix("APP_")(env)
// Result: MY_VAR becomes APP_MY_VAR
```
