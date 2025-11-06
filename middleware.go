package main

import (
	"context"
	"fmt"
	"os"
	"time"
)

// Common middleware implementations for ScriptExecutor

// WithLogging adds logging to script execution
func WithLogging(logger func(string, ...interface{})) ScriptMiddleware {
	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			logger("Executing script: %s", scriptPath)
			start := time.Now()

			env, err := next(zshPath, scriptPath)

			duration := time.Since(start)
			if err != nil {
				logger("Script failed: %s (took %v) - error: %v", scriptPath, duration, err)
			} else {
				logger("Script completed: %s (took %v)", scriptPath, duration)
			}

			return env, err
		}
	}
}

// WithTimeout adds a timeout to script execution
func WithTimeout(duration time.Duration) ScriptMiddleware {
	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			ctx, cancel := context.WithTimeout(context.Background(), duration)
			defer cancel()

			resultChan := make(chan struct {
				env map[string]string
				err error
			}, 1)

			go func() {
				env, err := next(zshPath, scriptPath)
				resultChan <- struct {
					env map[string]string
					err error
				}{env, err}
			}()

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("script execution timed out after %v: %s", duration, scriptPath)
			case result := <-resultChan:
				return result.env, result.err
			}
		}
	}
}

// WithErrorRecovery adds error recovery to script execution
func WithErrorRecovery() ScriptMiddleware {
	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			defer func() {
				if r := recover(); r != nil {
					fmt.Fprintf(os.Stderr, "wsh: panic recovered while executing %s: %v\n", scriptPath, r)
				}
			}()

			return next(zshPath, scriptPath)
		}
	}
}

// WithRetry adds retry logic to script execution
func WithRetry(maxRetries int, delay time.Duration) ScriptMiddleware {
	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			var lastErr error

			for attempt := 0; attempt <= maxRetries; attempt++ {
				env, err := next(zshPath, scriptPath)
				if err == nil {
					return env, nil
				}

				lastErr = err

				if attempt < maxRetries {
					time.Sleep(delay)
				}
			}

			return nil, fmt.Errorf("script failed after %d attempts: %w", maxRetries+1, lastErr)
		}
	}
}

// WithEnvFilter filters environment variables based on a predicate
func WithEnvFilter(predicate func(key, value string) bool) ScriptMiddleware {
	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			env, err := next(zshPath, scriptPath)
			if err != nil {
				return nil, err
			}

			filtered := make(map[string]string)
			for k, v := range env {
				if predicate(k, v) {
					filtered[k] = v
				}
			}

			return filtered, nil
		}
	}
}

// WithCaching caches script execution results (useful for expensive operations)
func WithCaching() ScriptMiddleware {
	cache := make(map[string]map[string]string)

	return func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			// Check cache
			if cached, ok := cache[scriptPath]; ok {
				return cached, nil
			}

			// Execute and cache
			env, err := next(zshPath, scriptPath)
			if err != nil {
				return nil, err
			}

			cache[scriptPath] = env
			return env, nil
		}
	}
}
