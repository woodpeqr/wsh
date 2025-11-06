package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWshrcLoader_LoadFile(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "wshrc_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	tmpfile.Close()

	script := loader.loadFile(tmpfile.Name())
	expected := "source " + tmpfile.Name() + " 2>/dev/null"

	if script != expected {
		t.Errorf("loadFile() = %q, want %q", script, expected)
	}
}

func TestWshrcLoader_FindScripts(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create a temporary directory with test files
	tmpdir, err := os.MkdirTemp("", "wshrc_test_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create test files
	testFiles := []string{"script1.sh", "script2.sh", "README.md"}
	hiddenFile := ".hidden"
	subdir := "subdir"

	for _, f := range testFiles {
		path := filepath.Join(tmpdir, f)
		if err := os.WriteFile(path, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create hidden file
	if err := os.WriteFile(filepath.Join(tmpdir, hiddenFile), []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Create subdirectory
	if err := os.Mkdir(filepath.Join(tmpdir, subdir), 0755); err != nil {
		t.Fatal(err)
	}

	scripts, err := loader.findScripts(tmpdir)
	if err != nil {
		t.Fatalf("findScripts() error = %v", err)
	}

	// Should find 3 regular files, not hidden file or subdirectory
	if len(scripts) != 3 {
		t.Errorf("Expected 3 scripts, got %d", len(scripts))
	}

	// Check that hidden file is not included
	for _, script := range scripts {
		if strings.Contains(script, hiddenFile) {
			t.Errorf("Hidden file should not be included: %s", script)
		}
		if strings.Contains(script, subdir) {
			t.Errorf("Subdirectory should not be included: %s", script)
		}
	}
}

func TestWshrcLoader_LoadNonExistent(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Try to load a non-existent file
	script, err := loader.Load("/nonexistent/path/to/.wshrc")
	if err != nil {
		t.Errorf("Load() should not error on non-existent file, got: %v", err)
	}

	if script != "" {
		t.Errorf("Load() should return empty script for non-existent file, got: %q", script)
	}
}

func TestWshrcLoader_LoadSingleFile(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "wshrc_test_*.sh")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	content := `export TEST_VAR="value"`
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	tmpfile.Close()

	script, err := loader.Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !strings.Contains(script, "source "+tmpfile.Name()) {
		t.Errorf("Load() should contain source command, got: %q", script)
	}
}

func TestWshrcLoader_LoadDirectory(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create a temporary directory with test scripts
	tmpdir, err := os.MkdirTemp("", "wshrc_test_dir_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create test scripts
	script1 := `export VAR1="value1"`
	script2 := `export VAR2="value2"`

	if err := os.WriteFile(filepath.Join(tmpdir, "01.sh"), []byte(script1), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpdir, "02.sh"), []byte(script2), 0644); err != nil {
		t.Fatal(err)
	}

	script, err := loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Script should contain exports for both variables
	if !strings.Contains(script, "VAR1") {
		t.Error("Expected script to contain VAR1")
	}
	if !strings.Contains(script, "VAR2") {
		t.Error("Expected script to contain VAR2")
	}
	if !strings.Contains(script, "export") {
		t.Error("Expected script to contain export statements")
	}
}

func TestWshrcLoader_LoadEmptyDirectory(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create an empty temporary directory
	tmpdir, err := os.MkdirTemp("", "wshrc_test_empty_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	script, err := loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if script != "" {
		t.Errorf("Load() should return empty script for empty directory, got: %q", script)
	}
}

func TestWshrcLoader_ExecuteScriptWithError(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath)

	// Create a temporary directory
	tmpdir, err := os.MkdirTemp("", "wshrc_test_error_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	// Create a script that will fail
	badScript := `exit 1`
	if err := os.WriteFile(filepath.Join(tmpdir, "bad.sh"), []byte(badScript), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = loader.Load(tmpdir)
	if err == nil {
		t.Error("Load() should return error for failing script")
	}
}

// Test execution strategies

func TestSequentialExecutionStrategy(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create temporary scripts
	tmpdir, err := os.MkdirTemp("", "strategy_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	script1 := filepath.Join(tmpdir, "01.sh")
	script2 := filepath.Join(tmpdir, "02.sh")

	if err := os.WriteFile(script1, []byte(`export SEQ1="first"`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script2, []byte(`export SEQ2="second"`), 0644); err != nil {
		t.Fatal(err)
	}

	scripts := []string{script1, script2}
	executor := defaultScriptExecutor

	env, err := SequentialExecutionStrategy(zshPath, scripts, executor)
	if err != nil {
		t.Fatalf("SequentialExecutionStrategy() error = %v", err)
	}

	if env["SEQ1"] != "first" || env["SEQ2"] != "second" {
		t.Errorf("Expected SEQ1=first and SEQ2=second, got %v", env)
	}
}

func TestParallelExecutionStrategy(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Create temporary scripts
	tmpdir, err := os.MkdirTemp("", "strategy_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	script1 := filepath.Join(tmpdir, "01.sh")
	script2 := filepath.Join(tmpdir, "02.sh")

	if err := os.WriteFile(script1, []byte(`export PAR1="first"`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(script2, []byte(`export PAR2="second"`), 0644); err != nil {
		t.Fatal(err)
	}

	scripts := []string{script1, script2}
	executor := defaultScriptExecutor

	env, err := ParallelExecutionStrategy(zshPath, scripts, executor)
	if err != nil {
		t.Fatalf("ParallelExecutionStrategy() error = %v", err)
	}

	if env["PAR1"] != "first" || env["PAR2"] != "second" {
		t.Errorf("Expected PAR1=first and PAR2=second, got %v", env)
	}
}

func TestWshrcLoader_WithExecutionStrategy(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	loader := NewWshrcLoader(zshPath,
		WithExecutionStrategy(SequentialExecutionStrategy),
	)

	// Create temporary directory with scripts
	tmpdir, err := os.MkdirTemp("", "strategy_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	if err := os.WriteFile(filepath.Join(tmpdir, "01.sh"), []byte(`export STRAT="sequential"`), 0644); err != nil {
		t.Fatal(err)
	}

	script, err := loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !strings.Contains(script, "STRAT") {
		t.Error("Expected script to contain STRAT variable")
	}
}

func TestWshrcLoader_WithCustomExecutor(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	// Custom executor that returns a fixed environment
	customExecutor := func(zshPath, scriptPath string) (map[string]string, error) {
		return map[string]string{"CUSTOM": "executor"}, nil
	}

	loader := NewWshrcLoader(zshPath,
		WithScriptExecutor(customExecutor),
	)

	// Create temporary directory with a script
	tmpdir, err := os.MkdirTemp("", "custom_exec_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	if err := os.WriteFile(filepath.Join(tmpdir, "01.sh"), []byte(`export IGNORED="value"`), 0644); err != nil {
		t.Fatal(err)
	}

	script, err := loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Should contain CUSTOM from our custom executor
	if !strings.Contains(script, "CUSTOM") {
		t.Error("Expected script to contain CUSTOM from custom executor")
	}

	// Should not contain IGNORED since our custom executor doesn't run the script
	if strings.Contains(script, "IGNORED") {
		t.Error("Should not contain IGNORED variable")
	}
}

func TestWshrcLoader_WithMiddleware(t *testing.T) {
	zshPath, err := findZsh()
	if err != nil {
		t.Skip("zsh not available")
	}

	var executionLog []string

	loggingMiddleware := func(next ScriptExecutor) ScriptExecutor {
		return func(zshPath, scriptPath string) (map[string]string, error) {
			executionLog = append(executionLog, fmt.Sprintf("executing: %s", filepath.Base(scriptPath)))
			return next(zshPath, scriptPath)
		}
	}

	loader := NewWshrcLoader(zshPath,
		WithMiddleware(loggingMiddleware),
	)

	// Create temporary directory with scripts
	tmpdir, err := os.MkdirTemp("", "middleware_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpdir)

	if err := os.WriteFile(filepath.Join(tmpdir, "01.sh"), []byte(`export MW="test"`), 0644); err != nil {
		t.Fatal(err)
	}

	_, err = loader.Load(tmpdir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check that middleware was called
	if len(executionLog) == 0 {
		t.Error("Expected middleware to be called")
	}
}
