package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// mockExecutor is a test double for the sandbox Executor interface.
type mockExecutor struct {
	result *ExecResult
	err    error
}

func (m *mockExecutor) Exec(ctx context.Context, cmd []string) (*ExecResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func TestDiscoverTests(t *testing.T) {
	dir := t.TempDir()

	// Create test structure: .revv/unit/build_check/test.md
	testDir := filepath.Join(dir, "unit", "build_check")
	os.MkdirAll(testDir, 0755)
	os.WriteFile(filepath.Join(testDir, "test.md"), []byte("## Description\ntest"), 0644)

	// Create another: .revv/visual/ui_check/test.md
	visualDir := filepath.Join(dir, "visual", "ui_check")
	os.MkdirAll(visualDir, 0755)
	os.WriteFile(filepath.Join(visualDir, "test.md"), []byte("## Description\nvisual test"), 0644)

	// Non-test files should be ignored
	os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM golang"), 0644)

	tests, err := DiscoverTests(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tests) != 2 {
		t.Fatalf("expected 2 tests, got %d: %v", len(tests), tests)
	}
}

func TestTestInfo(t *testing.T) {
	tests := []struct {
		revvDir  string
		testPath string
		wantCat  string
		wantName string
	}{
		{"/repo/.revv", "/repo/.revv/unit/build_check/test.md", "unit", "build_check"},
		{"/repo/.revv", "/repo/.revv/visual/ui_check/test.md", "visual", "ui_check"},
		{"/repo/.revv", "/repo/.revv/integration/api_health/test.md", "integration", "api_health"},
	}

	for _, tt := range tests {
		cat, name := TestInfo(tt.revvDir, tt.testPath)
		if cat != tt.wantCat || name != tt.wantName {
			t.Errorf("TestInfo(%q, %q) = (%q, %q), want (%q, %q)",
				tt.revvDir, tt.testPath, cat, name, tt.wantCat, tt.wantName)
		}
	}
}

func TestRunTest_Passing(t *testing.T) {
	executor := &mockExecutor{
		result: &ExecResult{
			ExitCode: 0,
			Stdout:   "PASS: test passed",
			Stderr:   "",
			Duration: 100 * time.Millisecond,
		},
	}

	content := `## Description
Build test.

## Priority
blocking

## Commands
make build
`

	result := RunTest(context.Background(), executor, "build", "compile", content)

	if !result.Passed {
		t.Errorf("expected test to pass, got error: %s", result.Error)
	}
	if result.Priority != "blocking" {
		t.Errorf("expected priority 'blocking', got %q", result.Priority)
	}
	if result.Skipped {
		t.Error("expected test to not be skipped")
	}
}

func TestRunTest_Failing(t *testing.T) {
	executor := &mockExecutor{
		result: &ExecResult{
			ExitCode: 1,
			Stdout:   "FAIL: build failed",
			Stderr:   "error: missing dependency",
			Duration: 200 * time.Millisecond,
		},
	}

	content := `## Description
Build test.

## Priority
blocking

## Commands
make build
`

	result := RunTest(context.Background(), executor, "build", "compile", content)

	if result.Passed {
		t.Error("expected test to fail")
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestRunTest_NoCommandsSkipped(t *testing.T) {
	executor := &mockExecutor{} // should never be called

	content := `## Description
Visual UI check with no commands.

## Priority
warning
`

	result := RunTest(context.Background(), executor, "visual", "ui_check", content)

	if !result.Skipped {
		t.Error("expected no-commands test to be skipped")
	}
	if !result.Passed {
		t.Error("expected skipped test to be marked as passed")
	}
}

func TestRunTest_Timeout(t *testing.T) {
	executor := &mockExecutor{
		result: &ExecResult{
			ExitCode: -1,
			Stdout:   "partial output",
			Duration: 5 * time.Second,
			TimedOut: true,
		},
	}

	content := `## Description
Slow test.

## Priority
blocking

## Commands
sleep 100
`

	result := RunTest(context.Background(), executor, "build", "slow_test", content)

	if result.Passed {
		t.Error("expected timed out test to fail")
	}
	if result.Error != "test timed out" {
		t.Errorf("expected timeout error, got %q", result.Error)
	}
}

func TestSummary(t *testing.T) {
	results := []TestResult{
		{Category: "build", Name: "compile", Priority: "blocking", Passed: true},
		{Category: "lint", Name: "vet", Priority: "warning", Passed: false},
		{Category: "visual", Name: "ui", Priority: "warning", Skipped: true, Passed: true},
	}

	summary := Summary(results)

	if summary == "" {
		t.Error("expected non-empty summary")
	}
}

func TestHasBlockingFailure(t *testing.T) {
	noFailure := []TestResult{
		{Priority: "blocking", Passed: true},
		{Priority: "warning", Passed: false},
	}
	if HasBlockingFailure(noFailure) {
		t.Error("expected no blocking failure")
	}

	withFailure := []TestResult{
		{Priority: "blocking", Passed: false},
	}
	if !HasBlockingFailure(withFailure) {
		t.Error("expected blocking failure")
	}
}

func TestRunAll_WithFilter(t *testing.T) {
	dir := t.TempDir()

	// Create two tests
	os.MkdirAll(filepath.Join(dir, "unit", "test1"), 0755)
	os.WriteFile(filepath.Join(dir, "unit", "test1", "test.md"), []byte(`## Description
Test 1.

## Priority
blocking

## Commands
echo pass
`), 0644)

	os.MkdirAll(filepath.Join(dir, "lint", "test2"), 0755)
	os.WriteFile(filepath.Join(dir, "lint", "test2", "test.md"), []byte(`## Description
Test 2.

## Priority
warning

## Commands
echo pass
`), 0644)

	executor := &mockExecutor{
		result: &ExecResult{ExitCode: 0, Stdout: "pass"},
	}

	// Filter by category
	results, err := RunAll(context.Background(), executor, dir, FilterOpts{Category: "unit"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with category filter, got %d", len(results))
	}
	if len(results) > 0 && results[0].Category != "unit" {
		t.Errorf("expected category 'unit', got %q", results[0].Category)
	}

	// Filter by specific test
	results, err = RunAll(context.Background(), executor, dir, FilterOpts{Test: "lint/test2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result with test filter, got %d", len(results))
	}
	if len(results) > 0 && results[0].Name != "test2" {
		t.Errorf("expected name 'test2', got %q", results[0].Name)
	}
}

func TestTestInfo_ErrorPath(t *testing.T) {
	//filepath.Rel will fail with one absolute path and one relative path
	cat, name := TestInfo("relative/path", "/absolute/path")
	if cat != "unknown" || name != "unknown" {
		t.Errorf("expected unknown/unknown, got %q/%q", cat, name)
	}

	// Test case where parts has length 1
	// E.g. filepath.Rel("/repo/.revv", "/repo/.revv/test.md") -> rel="test.md"
	// parts = ["test.md"]
	cat, name = TestInfo("/repo/.revv", "/repo/.revv/test.md")
	if cat != "default" || name != "test" {
		t.Errorf("expected default/test, got %q/%q", cat, name)
	}
}

func TestRunTest_ExecError(t *testing.T) {
	executor := &mockExecutor{
		err: os.ErrPermission,
	}
	content := `## Description
Test exec error.
## Commands
echo hello
`
	res := RunTest(context.Background(), executor, "unit", "test", content)
	if res.Passed {
		t.Errorf("expected failure when exec returns error")
	}
	if !contains(res.Error, "execution error") {
		t.Errorf("expected 'execution error' in error message, got %q", res.Error)
	}
}

func TestRunAll_FileReadError(t *testing.T) {
	dir := t.TempDir()
	testDir := filepath.Join(dir, "unit", "build_check")
	testFile := filepath.Join(testDir, "test.md")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	// Write a minimal file that parses but has no commands.
	// The runner should skip it (Passed=true, Skipped=true) since
	// tests without commands are browser/non-automated tests.
	if err := os.WriteFile(testFile, []byte("## Description\nTest with no commands.\n"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	executor := &mockExecutor{
		result: &ExecResult{ExitCode: 0},
	}

	results, err := RunAll(context.Background(), executor, dir, FilterOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if !results[0].Skipped {
		t.Errorf("expected test to be skipped when file has no commands")
	}
	if results[0].Error != "no commands to execute" {
		t.Errorf("expected 'no commands to execute' error, got %q", results[0].Error)
	}
}
