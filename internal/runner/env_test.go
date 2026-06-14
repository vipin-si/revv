package runner

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDetectEnvVars(t *testing.T) {
	// Setup test inputs
	testContents := []string{
		`## Commands
echo $DATABASE_URL
echo ${API_KEY}
echo $home
echo $123_INVALID
echo ${UNMATCHED_BRACKET
echo $EXTRA_BRACKET}
echo $PORT-suffix
`,
	}

	// We'll set some environment variables on the host
	t.Setenv("DATABASE_URL", "postgres://localhost:5432")
	t.Setenv("API_KEY", "secret123")
	t.Setenv("EXTRA_BRACKET", "bracket_val")
	t.Setenv("PORT", "8080")

	// Create a temp .env file
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	envContent := `
# A comment line
DATABASE_URL=ignored_because_host_has_priority
ENV_FILE_VAR=env_file_val
SPACED_VAR = spaced_val
QUOTED_VAR="quoted_val"
SINGLE_QUOTED_VAR='single_quoted_val'
INLINE_COMMENT_VAR=val_with_comment # inline comment is stripped
`
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		t.Fatalf("failed to write temp .env: %v", err)
	}

	envFiles := []string{envPath}

	envPairs, statuses := DetectEnvVars(testContents, envFiles)

	// Expected detected variables:
	// - DATABASE_URL (host)
	// - API_KEY (host)
	// - EXTRA_BRACKET (host)
	// - PORT (host) - matches $PORT from $PORT-suffix
	//
	// Lowercase $home should be detected
	// $123_INVALID should be ignored
	// ENV_FILE_VAR should be in statuses but since it wasn't referenced in testContents,
	// wait, let's check: DetectEnvVars only checks status for variables referenced in testContents!
	// Let's verify this in the source code.
	// Yes, `seen` is built from testContents, then `names` are sorted keys of `seen`.
	// Only these names are checked for host/env file values!
	// So ENV_FILE_VAR, SPACED_VAR, etc., are NOT in `statuses` unless referenced.
	//
	// Let's add them to testContents to test their retrieval from .env file.
	testContentsWithEnv := append(testContents, `
echo $ENV_FILE_VAR
echo $SPACED_VAR
echo $QUOTED_VAR
echo $SINGLE_QUOTED_VAR
echo $INLINE_COMMENT_VAR
`)

	envPairs, statuses = DetectEnvVars(testContentsWithEnv, envFiles)

	// Let's check status mapping
	statusMap := make(map[string]EnvVarStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	expectedVars := []struct {
		name   string
		set    bool
		source string
	}{
		{"DATABASE_URL", true, "host"},
		{"API_KEY", true, "host"},
		{"EXTRA_BRACKET", true, "host"},
		{"PORT", true, "host"},
		{"ENV_FILE_VAR", true, ".env"},
		{"SPACED_VAR", true, ".env"},
		{"QUOTED_VAR", true, ".env"},
		{"SINGLE_QUOTED_VAR", true, ".env"},
		{"INLINE_COMMENT_VAR", true, ".env"},
		{"home", false, ""}, // Lowercase now detected
	}

	for _, ev := range expectedVars {
		s, ok := statusMap[ev.name]
		if !ok {
			t.Errorf("expected var %s to be detected", ev.name)
			continue
		}
		if s.Set != ev.set {
			t.Errorf("var %s: expected Set=%v, got %v", ev.name, ev.set, s.Set)
		}
		if s.Source != ev.source {
			t.Errorf("var %s: expected Source=%q, got %q", ev.name, ev.source, s.Source)
		}
	}

	// Verify values in envPairs
	valMap := make(map[string]string)
	for _, pair := range envPairs {
		// split by first '='
		idx := 1
		for i, c := range pair {
			if c == '=' {
				idx = i
				break
			}
		}
		valMap[pair[:idx]] = pair[idx+1:]
	}

	if valMap["DATABASE_URL"] != "postgres://localhost:5432" {
		t.Errorf("expected DATABASE_URL from host, got %q", valMap["DATABASE_URL"])
	}
	if valMap["ENV_FILE_VAR"] != "env_file_val" {
		t.Errorf("expected ENV_FILE_VAR=env_file_val, got %q", valMap["ENV_FILE_VAR"])
	}
	if valMap["SPACED_VAR"] != "spaced_val" {
		t.Errorf("expected SPACED_VAR=spaced_val, got %q", valMap["SPACED_VAR"])
	}
	if valMap["QUOTED_VAR"] != "quoted_val" {
		t.Errorf("expected QUOTED_VAR=quoted_val, got %q", valMap["QUOTED_VAR"])
	}
	if valMap["SINGLE_QUOTED_VAR"] != "single_quoted_val" {
		t.Errorf("expected SINGLE_QUOTED_VAR=single_quoted_val, got %q", valMap["SINGLE_QUOTED_VAR"])
	}
	// Inline comment should be stripped
	if valMap["INLINE_COMMENT_VAR"] != "val_with_comment" {
		t.Errorf("expected INLINE_COMMENT_VAR with inline comment stripped, got %q", valMap["INLINE_COMMENT_VAR"])
	}
}

func TestLoadEnvFile_MissingFile(t *testing.T) {
	vars := make(map[string]string)
	loadEnvFile("nonexistent_file_path", vars)
	if len(vars) != 0 {
		t.Errorf("expected empty map for missing env file, got: %v", vars)
	}
}

func TestLoadEnvFile_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := `
NO_EQUALS_SIGN
=LEFTHAND_EMPTY
`
	if err := os.WriteFile(envPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	vars := make(map[string]string)
	loadEnvFile(envPath, vars)

	expected := map[string]string{
		"": "LEFTHAND_EMPTY",
	}

	if !reflect.DeepEqual(vars, expected) {
		t.Errorf("expected %v, got %v", expected, vars)
	}
}

func TestStripInlineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no comment",
			input:    "KEY=value",
			expected: "KEY=value",
		},
		{
			name:     "simple comment",
			input:    "KEY=value # comment here",
			expected: "KEY=value ",
		},
		{
			name:     "comment in double quotes",
			input:    `KEY="value # no comment" # actual comment`,
			expected: `KEY="value # no comment" `,
		},
		{
			name:     "comment in single quotes",
			input:    `KEY='value # no comment' # actual comment`,
			expected: `KEY='value # no comment' `,
		},
		{
			name:     "escaped double quote inside double quotes",
			input:    `KEY="val\"#no comment" # comment`,
			expected: `KEY="val\"#no comment" `,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripInlineComment(tt.input)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
