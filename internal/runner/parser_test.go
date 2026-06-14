package runner

import (
	"testing"
)

func TestParseTestMD_FullExample(t *testing.T) {
	content := `## Description
Verify that the CLI binary compiles cleanly using the Makefile.

## Priority
blocking

## Commands
` + "```bash" + `
make clean && make build
test -x ./bin/revv || (echo "FAIL: binary not found" && exit 1)
echo "PASS: compilation successful"
` + "```" + `

## Expected Output
Exit code 0. Output ends with "PASS: compilation successful".
`

	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pt.Description != "Verify that the CLI binary compiles cleanly using the Makefile." {
		t.Errorf("unexpected description: %q", pt.Description)
	}

	if pt.Priority != "blocking" {
		t.Errorf("expected priority 'blocking', got %q", pt.Priority)
	}

	if pt.NoCommands {
		t.Error("expected NoCommands=false for test with commands")
	}

	if pt.Commands == "" {
		t.Error("expected non-empty commands")
	}

	if !contains(pt.Commands, "make clean") {
		t.Errorf("expected commands to contain 'make clean', got: %q", pt.Commands)
	}

	if pt.Expected == "" {
		t.Error("expected non-empty expected output")
	}
}

func TestParseTestMD_NoCommandsTest(t *testing.T) {
	content := `## Description
Manually verify the UI looks correct.

## Priority
warning

## Expected Output
Reviewer confirms visual correctness.
`

	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !pt.NoCommands {
		t.Error("expected NoCommands=true for test without commands")
	}

	if pt.Priority != "warning" {
		t.Errorf("expected priority 'warning', got %q", pt.Priority)
	}
}

func TestParseTestMD_DefaultPriority(t *testing.T) {
	content := `## Description
A test without explicit priority.

## Commands
echo "hello"
`

	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pt.Priority != "warning" {
		t.Errorf("expected default priority 'warning', got %q", pt.Priority)
	}
}

func TestParseTestMD_UnknownPriorityNormalized(t *testing.T) {
	content := `## Description
Test with unknown priority.

## Priority
critical
`

	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pt.Priority != "blocking" {
		t.Errorf("expected 'critical' normalized to 'blocking', got %q", pt.Priority)
	}
}

func TestParseTestMD_CommandsWithoutCodeBlock(t *testing.T) {
	content := `## Description
Simple test.

## Priority
blocking

## Commands
go test ./...
echo "PASS"

## Expected Output
PASS
`

	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if pt.Commands == "" {
		t.Error("expected non-empty commands even without code block")
	}

	if pt.NoCommands {
		t.Error("expected NoCommands=false")
	}
}

func TestExtractCodeBlock(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bash block",
			input:    "```bash\necho hello\n```",
			expected: "echo hello",
		},
		{
			name:     "plain block",
			input:    "```\nfoo\nbar\n```",
			expected: "foo\nbar",
		},
		{
			name:     "no block",
			input:    "just text",
			expected: "",
		},
		{
			name:     "empty block",
			input:    "```\n```",
			expected: "",
		},
		{
			name:     "multiple blocks",
			input:    "```bash\necho 1\n```\nSome intermediate text\n```bash\necho 2\n```",
			expected: "echo 1\necho 2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCodeBlock(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestNormalizeType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"automated", "automated"},
		{"auto", "automated"},
		{"docker", "automated"},
		{"command", "automated"},
		{"commands", "automated"},
		{"browser", "browser"},
		{"ui", "browser"},
		{"e2e", "browser"},
		{"visual", "browser"},
		{"manual", "browser"},
		{"human", "browser"},
		{"steps", "browser"},
		{"anything_else", "automated"},
	}

	for _, tt := range tests {
		got := normalizeType(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeType(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseTestMD_TypeNormalization(t *testing.T) {
	content := `## Description
Test type parsing.

## Type
ui

## Commands
echo "run ui"
`
	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pt.Type != "browser" {
		t.Errorf("expected Type 'browser' from 'ui', got %q", pt.Type)
	}
	// Since type is browser (not automated), NoCommands should be true (skipped by default execution)
	if !pt.NoCommands {
		t.Errorf("expected NoCommands=true for non-automated test, got false")
	}

	// Test default inference: no type, but has commands -> automated
	contentDefaultAuto := `## Description
Test inference.
## Commands
echo "hello"
`
	pt2, err := ParseTestMD(contentDefaultAuto)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pt2.Type != "automated" {
		t.Errorf("expected Type 'automated', got %q", pt2.Type)
	}

	// Test default inference: no type, no commands -> browser
	contentDefaultManual := `## Description
Test inference.
`
	pt3, err := ParseTestMD(contentDefaultManual)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pt3.Type != "browser" {
		t.Errorf("expected Type 'browser', got %q", pt3.Type)
	}
}

func TestParseTestMD_HeaderHijacking(t *testing.T) {
	content := `## Description
This is a test.

## Commands
` + "```bash" + `
## Fake Header
echo "inside code block"
` + "```" + `

## Expected Output
PASS
`
	pt, err := ParseTestMD(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !contains(pt.Commands, "## Fake Header") {
		t.Errorf("expected commands to contain hijacked header, got: %q", pt.Commands)
	}
	if _, ok := parseSections(content)["fake header"]; ok {
		t.Error("expected 'fake header' section to NOT be parsed")
	}
}
