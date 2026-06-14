package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vipinsingh/revv/internal/runner"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w

	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		outChan <- buf.String()
	}()

	fn()

	w.Close()
	os.Stdout = oldStdout
	return <-outChan
}

func TestOutputJSON(t *testing.T) {
	results := []runner.TestResult{
		{
			Category: "unit",
			Name:     "check_1",
			Priority: "blocking",
			Passed:   true,
			Duration: 500 * time.Millisecond,
		},
		{
			Category: "unit",
			Name:     "check_2",
			Priority: "warning",
			Passed:   false,
			Duration: 1 * time.Second,
			Error:    "exit code 1",
			Output:   "error log line 1\nerror log line 2",
		},
		{
			Category: "visual",
			Name:     "check_3",
			Priority: "warning",
			Skipped:  true,
			Passed:   true,
			Duration: 0,
		},
	}

	outStr := captureStdout(t, func() {
		err := outputJSON(results)
		if err != nil {
			t.Errorf("outputJSON failed: %v", err)
		}
	})

	var parsed jsonOutput
	if err := json.Unmarshal([]byte(outStr), &parsed); err != nil {
		t.Fatalf("failed to parse JSON output: %v\nOutput: %q", err, outStr)
	}

	if len(parsed.Results) != 3 {
		t.Errorf("expected 3 results, got %d", len(parsed.Results))
	}

	if parsed.Summary.Passed != 1 {
		t.Errorf("expected 1 passed in summary, got %d", parsed.Summary.Passed)
	}
	if parsed.Summary.Failed != 1 {
		t.Errorf("expected 1 failed in summary, got %d", parsed.Summary.Failed)
	}
	if parsed.Summary.Skipped != 1 {
		t.Errorf("expected 1 skipped in summary, got %d", parsed.Summary.Skipped)
	}
	if parsed.Summary.BlockingTotal != 1 {
		t.Errorf("expected 1 blocking total, got %d", parsed.Summary.BlockingTotal)
	}
	if parsed.Summary.BlockingPass != 1 {
		t.Errorf("expected 1 blocking passed, got %d", parsed.Summary.BlockingPass)
	}

	// Verify details
	res2 := parsed.Results[1]
	if res2.Passed {
		t.Error("expected check_2 to show passed=false")
	}
	if res2.Error != "exit code 1" {
		t.Errorf("expected error 'exit code 1', got %q", res2.Error)
	}
	if !strings.Contains(res2.Output, "error log line 1") {
		t.Errorf("expected output to contain error log, got %q", res2.Output)
	}
}

func TestOutputTable(t *testing.T) {
	results := []runner.TestResult{
		{
			Category: "unit",
			Name:     "check_1",
			Priority: "blocking",
			Passed:   true,
			Duration: 500 * time.Millisecond,
		},
		{
			Category: "unit",
			Name:     "check_2",
			Priority: "warning",
			Passed:   false,
			Duration: 1 * time.Second,
			Error:    "exit code 1",
			Output:   "error log line 1\nerror log line 2",
		},
		{
			Category: "visual",
			Name:     "check_3",
			Priority: "warning",
			Skipped:  true,
			Passed:   true,
			Duration: 0,
		},
	}

	// Test non-verbose
	outStr := captureStdout(t, func() {
		outputTable(results, false)
	})

	if !strings.Contains(outStr, "unit/check_1") || !strings.Contains(outStr, "PASS") {
		t.Errorf("expected check_1 table line, got %q", outStr)
	}
	if !strings.Contains(outStr, "unit/check_2") || !strings.Contains(outStr, "FAIL") {
		t.Errorf("expected check_2 table line, got %q", outStr)
	}
	if strings.Contains(outStr, "error log line 1") {
		t.Error("expected error logs to be hidden in non-verbose mode")
	}

	// Test verbose mode
	outStrVerbose := captureStdout(t, func() {
		outputTable(results, true)
	})

	if !strings.Contains(outStrVerbose, "error log line 1") {
		t.Error("expected error logs to be visible in verbose mode")
	}
}

func TestRunExec_MissingConfigs(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Change cwd to tempDir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(oldCwd)

	cmd := newExecCmd()

	// 1. Run without .revv/
	err = runExec(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "no .revv/ directory found") {
		t.Errorf("expected error about missing .revv/, got: %v", err)
	}

	// Create .revv/ directory
	revvDir := filepath.Join(tempDir, ".revv")
	if err := os.Mkdir(revvDir, 0755); err != nil {
		t.Fatalf("failed to mkdir .revv: %v", err)
	}

	// 2. Run without Dockerfile
	err = runExec(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "no .revv/Dockerfile found") {
		t.Errorf("expected error about missing Dockerfile, got: %v", err)
	}
}

func TestRunExec_MissingConfigs_JSONMode(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Change cwd to tempDir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(oldCwd)

	cmd := newExecCmd()
	cmd.Flags().Set("json", "true")

	// 1. Run without .revv/
	var outStr string
	outStr = captureStdout(t, func() {
		err = runExec(cmd, nil)
	})
	if err != nil {
		t.Errorf("expected nil error in JSON mode, got: %v", err)
	}
	var payload map[string]string
	if err := json.Unmarshal([]byte(outStr), &payload); err != nil {
		t.Fatalf("failed to parse JSON error output: %v\nOutput: %q", err, outStr)
	}
	if !strings.Contains(payload["error"], "no .revv/ directory found") {
		t.Errorf("expected error to contain missing directory message, got: %q", payload["error"])
	}

	// Create .revv/ directory
	revvDir := filepath.Join(tempDir, ".revv")
	if err := os.Mkdir(revvDir, 0755); err != nil {
		t.Fatalf("failed to mkdir .revv: %v", err)
	}

	// 2. Run without Dockerfile
	outStr = captureStdout(t, func() {
		err = runExec(cmd, nil)
	})
	if err != nil {
		t.Errorf("expected nil error in JSON mode, got: %v", err)
	}
	if err := json.Unmarshal([]byte(outStr), &payload); err != nil {
		t.Fatalf("failed to parse JSON error output: %v\nOutput: %q", err, outStr)
	}
	if !strings.Contains(payload["error"], "no .revv/Dockerfile found") {
		t.Errorf("expected error to contain missing Dockerfile message, got: %q", payload["error"])
	}
}

func TestExecute(t *testing.T) {
	// Execute just calls root command execute. Since we didn't set args, it runs help by default.
	// We can set os.Args to a version call and check it executes.
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"revv", "version"}
	out := captureStdout(t, func() {
		err := Execute()
		if err != nil {
			t.Errorf("Execute failed: %v", err)
		}
	})

	if !strings.Contains(out, "revv") {
		t.Errorf("expected version output, got %q", out)
	}
}

func TestNewVersionCmd(t *testing.T) {
	cmd := newVersionCmd()
	if cmd.Use != "version" {
		t.Errorf("expected use version, got %q", cmd.Use)
	}

	out := captureStdout(t, func() {
		cmd.Run(cmd, nil)
	})

	if !strings.Contains(out, "revv") {
		t.Errorf("expected version output, got %q", out)
	}
}

func TestNewExecCmd(t *testing.T) {
	cmd := newExecCmd()
	if cmd.Use != "exec" {
		t.Errorf("expected use exec, got %q", cmd.Use)
	}
}

type mockContextKey string

func TestRunExec_DockerAvailabilityCheckFails(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get cwd: %v", err)
	}

	// Change cwd to tempDir
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(oldCwd)

	// Create .revv/ and Dockerfile
	revvDir := filepath.Join(tempDir, ".revv")
	if err := os.Mkdir(revvDir, 0755); err != nil {
		t.Fatalf("failed to mkdir .revv: %v", err)
	}
	dockerfile := filepath.Join(revvDir, "Dockerfile")
	if err := os.WriteFile(dockerfile, []byte("FROM alpine"), 0644); err != nil {
		t.Fatalf("failed to write Dockerfile: %v", err)
	}

	// Set invalid Docker host to force connection failure
	oldDockerHost := os.Getenv("DOCKER_HOST")
	os.Setenv("DOCKER_HOST", "tcp://localhost:9999")
	defer func() {
		if oldDockerHost == "" {
			os.Unsetenv("DOCKER_HOST")
		} else {
			os.Setenv("DOCKER_HOST", oldDockerHost)
		}
	}()

	// Redirect stdin to send "no" so it declines installation
	oldStdin := os.Stdin
	rStdin, wStdin, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe for stdin: %v", err)
	}
	os.Stdin = rStdin
	defer func() {
		os.Stdin = oldStdin
		wStdin.Close()
		rStdin.Close()
	}()

	go func() {
		_, _ = wStdin.Write([]byte("no\n"))
	}()

	cmd := newExecCmd()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cmd.SetContext(ctx)

	err = runExec(cmd, nil)
	if err == nil {
		t.Errorf("expected connection error or context timeout error, got nil")
	} else if !strings.Contains(err.Error(), "installation declined") {
		t.Errorf("expected installation declined error, got: %v", err)
	}
}
