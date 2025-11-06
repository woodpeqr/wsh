package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWithLogging(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	var logMessages []string
	logger := func(format string, args ...interface{}) {
		logMessages = append(logMessages, fmt.Sprintf(format, args...))
	}

	// Create test script
	tmpfile, err := os.CreateTemp("", "log_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export TEST="value"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithLogging(logger)(defaultScriptExecutor)
	_, err = executor(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	// Check that logging occurred
	if len(logMessages) == 0 {
		t.Error("Expected log messages")
	}

	hasExecuting := false
	hasCompleted := false
	for _, msg := range logMessages {
		if strings.Contains(msg, "Executing") {
			hasExecuting = true
		}
		if strings.Contains(msg, "completed") {
			hasCompleted = true
		}
	}

	if !hasExecuting || !hasCompleted {
		t.Errorf("Expected both executing and completed messages, got: %v", logMessages)
	}
}

func TestWithTimeout_Success(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a quick script
	tmpfile, err := os.CreateTemp("", "timeout_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export QUICK="fast"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithTimeout(5 * time.Second)(defaultScriptExecutor)
	env, err := executor(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if env["QUICK"] != "fast" {
		t.Errorf("Expected QUICK=fast, got %v", env)
	}
}

func TestWithTimeout_Timeout(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a slow script
	tmpfile, err := os.CreateTemp("", "timeout_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`sleep 10 && export SLOW="value"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithTimeout(100 * time.Millisecond)(defaultScriptExecutor)
	_, err = executor(zshPath, tmpfile.Name())

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error message, got: %v", err)
	}
}

func TestWithErrorRecovery(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a script that should work
	tmpfile, err := os.CreateTemp("", "recovery_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export SAFE="value"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithErrorRecovery()(defaultScriptExecutor)
	env, err := executor(zshPath, tmpfile.Name())

	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if env["SAFE"] != "value" {
		t.Errorf("Expected SAFE=value, got %v", env)
	}
}

func TestWithRetry_Success(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a successful script
	tmpfile, err := os.CreateTemp("", "retry_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export RETRY="success"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithRetry(2, 10*time.Millisecond)(defaultScriptExecutor)
	env, err := executor(zshPath, tmpfile.Name())

	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if env["RETRY"] != "success" {
		t.Errorf("Expected RETRY=success, got %v", env)
	}
}

func TestWithRetry_Failure(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a failing script
	tmpfile, err := os.CreateTemp("", "retry_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`exit 1`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithRetry(2, 10*time.Millisecond)(defaultScriptExecutor)
	_, err = executor(zshPath, tmpfile.Name())

	if err == nil {
		t.Error("Expected error after retries")
	}

	if !strings.Contains(err.Error(), "failed after") {
		t.Errorf("Expected retry error message, got: %v", err)
	}
}

func TestWithEnvFilter(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a script that sets multiple variables
	tmpfile, err := os.CreateTemp("", "filter_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	script := `
export KEEP_ME="yes"
export FILTER_ME="no"
export KEEP_ALSO="yes"
`
	if _, err := tmpfile.Write([]byte(script)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Filter to only keep variables starting with KEEP
	predicate := func(key, value string) bool {
		return strings.HasPrefix(key, "KEEP")
	}

	executor := WithEnvFilter(predicate)(defaultScriptExecutor)
	env, err := executor(zshPath, tmpfile.Name())

	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if _, exists := env["KEEP_ME"]; !exists {
		t.Error("Expected KEEP_ME to be present")
	}

	if _, exists := env["KEEP_ALSO"]; !exists {
		t.Error("Expected KEEP_ALSO to be present")
	}

	if _, exists := env["FILTER_ME"]; exists {
		t.Error("Expected FILTER_ME to be filtered out")
	}
}

func TestWithCaching(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create a script
	tmpfile, err := os.CreateTemp("", "cache_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export CACHED="value"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	executor := WithCaching()(defaultScriptExecutor)

	// First execution
	env1, err := executor(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	// Second execution (should be cached)
	env2, err := executor(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if env1["CACHED"] != env2["CACHED"] {
		t.Error("Cached values should be the same")
	}
}

func TestMiddleware_Chaining(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	var logMessages []string
	logger := func(format string, args ...interface{}) {
		logMessages = append(logMessages, fmt.Sprintf(format, args...))
	}

	// Create a script
	tmpfile, err := os.CreateTemp("", "chain_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`export CHAIN="test"`)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	// Chain multiple middleware
	executor := defaultScriptExecutor
	executor = WithLogging(logger)(executor)
	executor = WithErrorRecovery()(executor)
	executor = WithTimeout(5 * time.Second)(executor)

	env, err := executor(zshPath, tmpfile.Name())
	if err != nil {
		t.Fatalf("executor() error = %v", err)
	}

	if env["CHAIN"] != "test" {
		t.Errorf("Expected CHAIN=test, got %v", env)
	}

	// Check that logging middleware was called
	if len(logMessages) == 0 {
		t.Error("Expected log messages from chained middleware")
	}
}

func TestWithMiddleware_Integration(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	var logMessages []string
	logger := func(format string, args ...interface{}) {
		logMessages = append(logMessages, fmt.Sprintf(format, args...))
	}

	// Create loader with multiple middleware
	loader := NewWshrcLoader(zshPath,
		WithMiddleware(
			WithLogging(logger),
			WithErrorRecovery(),
			WithTimeout(5*time.Second),
		),
	)

	// Create temporary directory with a script
	tmpdir, err := os.MkdirTemp("", "middleware_integration_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	script := `export MIDDLEWARE_TEST="integrated"`
	if err := os.WriteFile(filepath.Join(tmpdir, "01.sh"), []byte(script), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !strings.Contains(result, "MIDDLEWARE_TEST") {
		t.Error("Expected result to contain MIDDLEWARE_TEST")
	}

	// Verify logging middleware was called
	if len(logMessages) == 0 {
		t.Error("Expected middleware logging to occur")
	}
}
