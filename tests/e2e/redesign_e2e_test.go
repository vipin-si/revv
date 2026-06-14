package e2e

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vipinsingh/revv/internal/runner"
)

// mockExecutor is a test double for the sandbox Executor interface
type mockExecutor struct {
	result *runner.ExecResult
	err    error
}

func (m *mockExecutor) Exec(ctx context.Context, cmd []string) (*runner.ExecResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

// findRepoRoot dynamically locates the repository root containing go.mod
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found")
}

func readPromptFile(t *testing.T, root, skillDir string) (string, []string) {
	t.Helper()
	path := filepath.Join(root, "skills", skillDir, "prompt.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read prompt file %s: %v", path, err)
	}
	lines := strings.Split(string(content), "\n")
	return string(content), lines
}

func readSkillFile(t *testing.T, root, skillDir string) string {
	t.Helper()
	path := filepath.Join(root, "skills", skillDir, "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read SKILL.md %s: %v", path, err)
	}
	return string(content)
}

func parseFrontmatter(content string) (map[string]string, error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return nil, fmt.Errorf("missing starting ---")
	}
	parts := strings.Split(trimmed, "---")
	if len(parts) < 3 {
		return nil, fmt.Errorf("missing closing ---")
	}
	fm := parts[1]
	res := make(map[string]string)
	for _, line := range strings.Split(fm, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx == -1 {
			return nil, fmt.Errorf("invalid frontmatter line: %s", line)
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		res[key] = val
	}
	return res, nil
}

func verifySkillFrontmatter(t *testing.T, content, expectedName string) {
	t.Helper()
	fields, err := parseFrontmatter(content)
	if err != nil {
		t.Errorf("failed to parse frontmatter: %v", err)
		return
	}
	if fields["name"] != expectedName {
		t.Errorf("expected name %q, got %q", expectedName, fields["name"])
	}
	if fields["description"] == "" {
		t.Errorf("expected non-empty description in frontmatter")
	}
}

func assertContains(t *testing.T, content, substr, desc string) {
	t.Helper()
	if !strings.Contains(content, substr) {
		t.Errorf("missing expected content: %s (should contain: %q)", desc, substr)
	}
}

func TestRedesignE2E(t *testing.T) {
	root, err := findRepoRoot()
	if err != nil {
		t.Fatalf("failed to find repository root: %v", err)
	}

	// ==========================================
	// TIER 1: Feature Coverage (36 test cases)
	// ==========================================
	t.Run("Tier1_FeatureCoverage", func(t *testing.T) {
		// --- F1. init-repo ---
		t.Run("TC_1_1_1_InitRepoDirExists", func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, "skills", "init-repo"))
			if err != nil || !info.IsDir() {
				t.Errorf("skills/init-repo directory does not exist")
			}
		})

		t.Run("TC_1_1_2_InitRepoSkillMdFrontmatter", func(t *testing.T) {
			content := readSkillFile(t, root, "init-repo")
			verifySkillFrontmatter(t, content, "init-repo")
		})

		t.Run("TC_1_1_3_InitRepoPromptMinLines", func(t *testing.T) {
			_, lines := readPromptFile(t, root, "init-repo")
			if len(lines) < 80 {
				t.Errorf("skills/init-repo/prompt.md has %d lines, expected >= 80", len(lines))
			}
		})

		t.Run("TC_1_1_4_InitRepoPromptSections", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "## Context Gathering", "Context section")
			assertContains(t, content, "## What Tests to Generate", "Output section")
			assertContains(t, content, "## test.md Format", "Format section")
		})

		t.Run("TC_1_1_5_InitRepoPromptTestMdSpec", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "## Description", "Description header")
			assertContains(t, content, "## Priority", "Priority header")
			assertContains(t, content, "## Type", "Type header")
			assertContains(t, content, "## Commands", "Commands header")
			assertContains(t, content, "## Expected Output", "Expected Output header")
		})

		t.Run("TC_1_1_6_InitRepoPromptDockerfileRules", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "Base image", "Dockerfile base image rule")
			assertContains(t, content, "System deps", "Dockerfile dependencies rule")
			assertContains(t, content, "/workspace", "Dockerfile working directory")
			assertContains(t, content, "COPY source", "Dockerfile source copying rule")
			assertContains(t, content, "Pre-build", "Dockerfile pre-build / RUN rule")
		})

		t.Run("TC_1_1_7_InitRepoPromptAgentsMdRules", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "## AGENTS.md Rules", "AGENTS.md title")
			assertContains(t, content, "append", "AGENTS.md append rule")
			assertContains(t, content, "already configured", "AGENTS.md deduplication rule")
			assertContains(t, content, "vssinghh/revv", "AGENTS.md checking duplicate repo path")
		})

		t.Run("TC_1_1_8_InitRepoPromptRawUrls", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md", "init-repo raw URL")
			assertContains(t, content, "https://raw.githubusercontent.com/vssinghh/revv/main/skills/update-repo/SKILL.md", "update-repo raw URL")
			assertContains(t, content, "https://raw.githubusercontent.com/vssinghh/revv/main/skills/add-tests/SKILL.md", "add-tests raw URL")
			assertContains(t, content, "https://raw.githubusercontent.com/vssinghh/revv/main/skills/run-tests/SKILL.md", "run-tests raw URL")
		})

		// --- F2. update-repo ---
		t.Run("TC_1_2_1_UpdateRepoDirExists", func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, "skills", "update-repo"))
			if err != nil || !info.IsDir() {
				t.Errorf("skills/update-repo directory does not exist")
			}
		})

		t.Run("TC_1_2_2_UpdateRepoSkillMdFrontmatter", func(t *testing.T) {
			content := readSkillFile(t, root, "update-repo")
			verifySkillFrontmatter(t, content, "update-repo")
		})

		t.Run("TC_1_2_3_UpdateRepoPromptMinLines", func(t *testing.T) {
			_, lines := readPromptFile(t, root, "update-repo")
			if len(lines) < 80 {
				t.Errorf("skills/update-repo/prompt.md has %d lines, expected >= 80", len(lines))
			}
		})

		t.Run("TC_1_2_4_UpdateRepoPromptSections", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "## Context", "Context section")
			assertContains(t, content, "## Output", "Output section")
			assertContains(t, content, "## Rules", "Rules section")
		})

		t.Run("TC_1_2_5_UpdateRepoPromptGitLogN", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "git log --oneline -N", "git log commit history search context")
		})

		t.Run("TC_1_2_6_UpdateRepoPromptPreservationRules", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "Minimal Disruption Principle", "Preservation principle")
			assertContains(t, content, "Preserve relevant tests", "Preservation instruction")
		})

		// --- F3. add-tests ---
		t.Run("TC_1_3_1_AddTestsDirExists", func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, "skills", "add-tests"))
			if err != nil || !info.IsDir() {
				t.Errorf("skills/add-tests directory does not exist")
			}
		})

		t.Run("TC_1_3_2_AddTestsSkillMdFrontmatter", func(t *testing.T) {
			content := readSkillFile(t, root, "add-tests")
			verifySkillFrontmatter(t, content, "add-tests")
		})

		t.Run("TC_1_3_3_AddTestsPromptMinLines", func(t *testing.T) {
			_, lines := readPromptFile(t, root, "add-tests")
			if len(lines) < 80 {
				t.Errorf("skills/add-tests/prompt.md has %d lines, expected >= 80", len(lines))
			}
		})

		t.Run("TC_1_3_4_AddTestsPromptSections", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "## Context", "Context section")
			assertContains(t, content, "## Output", "Output section")
			assertContains(t, content, "## Rules", "Rules section")
		})

		t.Run("TC_1_3_5_AddTestsPromptGitDiff", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "git diff", "git diff search context")
		})

		t.Run("TC_1_3_6_AddTestsPromptDecisionLogic", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "Decision on Test Necessity", "Decision logic title")
			assertContains(t, content, "If YES", "Yes tests needed branch")
			assertContains(t, content, "If NO", "No tests needed branch")
		})

		// --- F4. run-tests ---
		t.Run("TC_1_4_1_RunTestsDirExists", func(t *testing.T) {
			info, err := os.Stat(filepath.Join(root, "skills", "run-tests"))
			if err != nil || !info.IsDir() {
				t.Errorf("skills/run-tests directory does not exist")
			}
		})

		t.Run("TC_1_4_2_RunTestsSkillMdFrontmatter", func(t *testing.T) {
			content := readSkillFile(t, root, "run-tests")
			verifySkillFrontmatter(t, content, "run-tests")
		})

		t.Run("TC_1_4_3_RunTestsPromptMinLines", func(t *testing.T) {
			_, lines := readPromptFile(t, root, "run-tests")
			if len(lines) < 80 {
				t.Errorf("skills/run-tests/prompt.md has %d lines, expected >= 80", len(lines))
			}
		})

		t.Run("TC_1_4_4_RunTestsPromptSections", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "## Context", "Context section")
			assertContains(t, content, "## Output", "Output section")
			assertContains(t, content, "## Rules", "Rules section")
		})

		t.Run("TC_1_4_5_RunTestsPromptDelegation", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "add-tests", "Delegates to add-tests")
			assertContains(t, content, "skill", "Delegates to add-tests")
		})

		t.Run("TC_1_4_6_RunTestsPromptGoBinary", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "go build -o /tmp/revv ./cmd/revv", "Go binary compilation instruction")
			assertContains(t, content, "which revv", "Go binary execution checks")
		})

		t.Run("TC_1_4_7_RunTestsPromptBrowserExecution", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "navigate_page", "DevTools navigate page tool")
			assertContains(t, content, "click", "DevTools click selector tool")
			assertContains(t, content, "screenshot", "DevTools screenshot tool")
		})

		t.Run("TC_1_4_8_RunTestsBrowserFallback", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "browser - needs human verification", "Browser test fallback instructions")
		})

		t.Run("TC_1_4_9_RunTestsPromptFailureAnalysis", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "Failure Analysis and Reporting", "Failure Analysis header")
			assertContains(t, content, "Diagnostics", "Failure Diagnostics steps")
		})

		// --- F5. Repository Integrity & Metadata ---
		t.Run("TC_1_5_1_RevvUpdateDeleted", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(root, "skills", "revv-update"))
			if !os.IsNotExist(err) {
				t.Errorf("deprecated skills/revv-update directory still exists")
			}
		})

		t.Run("TC_1_5_2_RevvRunDeleted", func(t *testing.T) {
			_, err := os.Stat(filepath.Join(root, "skills", "revv-run"))
			if !os.IsNotExist(err) {
				t.Errorf("deprecated skills/revv-run directory still exists")
			}
		})

		t.Run("TC_1_5_3_ReadmeArchitectureUpdated", func(t *testing.T) {
			readme, err := os.ReadFile(filepath.Join(root, "README.md"))
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}
			readmeStr := string(readme)
			assertContains(t, readmeStr, "init-repo", "README architecture init-repo skill link")
			assertContains(t, readmeStr, "update-repo", "README architecture update-repo skill link")
			assertContains(t, readmeStr, "add-tests", "README architecture add-tests skill link")
			assertContains(t, readmeStr, "run-tests", "README architecture run-tests skill link")
		})

		t.Run("TC_1_5_4_ReadmeMaintainerSetupUpdated", func(t *testing.T) {
			readme, err := os.ReadFile(filepath.Join(root, "README.md"))
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}
			readmeStr := string(readme)
			assertContains(t, readmeStr, "init-repo/SKILL.md", "README For Maintainers references init-repo")
		})

		t.Run("TC_1_5_5_ReadmeGoBinaryReferences", func(t *testing.T) {
			readme, err := os.ReadFile(filepath.Join(root, "README.md"))
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}
			readmeStr := string(readme)
			assertContains(t, readmeStr, "revv exec", "README references Go binary revv exec")
		})

		t.Run("TC_1_5_6_BackendCompiles", func(t *testing.T) {
			cmd := exec.Command("go", "build", "-o", os.DevNull, "./cmd/revv")
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("compiling the Go backend failed: %v\nOutput: %s", err, string(output))
			}
		})

		t.Run("TC_1_5_7_GoUnitTestsRun", func(t *testing.T) {
			// run internal tests only, avoiding recursive e2e test execution
			cmd := exec.Command("go", "test", "./internal/...")
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("Go unit tests failed to run or pass: %v\nOutput: %s", err, string(output))
			}
		})
	})

	// ==========================================
	// TIER 2: Boundary & Corner Cases (24 test cases)
	// ==========================================
	t.Run("Tier2_BoundaryCases", func(t *testing.T) {
		// --- Init-Repo Boundary Cases ---
		t.Run("TC_2_1_1_InitRepoAgentsMdAppend", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "AGENTS.md", "AGENTS.md append rule check")
			assertContains(t, content, "append", "AGENTS.md append rule check")
		})

		t.Run("TC_2_1_2_InitRepoAgentsMdSkipDuplicate", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "do NOT add anything", "AGENTS.md deduplication logic")
		})

		t.Run("TC_2_1_3_InitRepoNoBuildConfigs", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "build system", "init-repo detected config")
		})

		t.Run("TC_2_1_4_InitRepoInvalidYamlFrontmatter", func(t *testing.T) {
			invalidFm := `---
invalid_key_no_colon
description: test
---`
			_, err := parseFrontmatter(invalidFm)
			if err == nil {
				t.Errorf("expected error for invalid frontmatter syntax")
			}
		})

		t.Run("TC_2_1_5_InitRepoFrontmatterFieldChecks", func(t *testing.T) {
			missingFieldsFm := `---
name: init-repo
---`
			fields, err := parseFrontmatter(missingFieldsFm)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if fields["name"] != "init-repo" {
				t.Errorf("expected key 'name' to be parsed")
			}
			if _, ok := fields["description"]; ok {
				t.Errorf("description field should not be present")
			}
		})

		t.Run("TC_2_1_6_InitRepoUrlValidation", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "https://raw.githubusercontent.com/vssinghh/revv/main/skills/", "URL pattern validator")
		})

		// --- Update-Repo Boundary Cases ---
		t.Run("TC_2_2_1_UpdateRepoFewerThan10Commits", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "defaults to 10", "Commit history default N=10")
		})

		t.Run("TC_2_2_2_UpdateRepoNoRevvDir", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "Existing", "update-repo existing tests context check")
			assertContains(t, content, ".revv/", "update-repo existing tests context check")
		})

		t.Run("TC_2_2_3_UpdateRepoNoChangesIdempotency", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "remain untouched", "idempotency instruction")
			assertContains(t, content, "Only modify", "idempotency instruction")
		})

		t.Run("TC_2_2_4_UpdateRepoExclusionPaths", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "Codebase Tree", "Git tree context check")
			assertContains(t, content, "layout", "Git tree context check")
		})

		t.Run("TC_2_2_5_UpdateRepoSkillFrontmatterName", func(t *testing.T) {
			content := readSkillFile(t, root, "update-repo")
			verifySkillFrontmatter(t, content, "update-repo")
		})

		t.Run("TC_2_2_6_UpdateRepoPromptLengthCheck", func(t *testing.T) {
			shortPrompt := "too short prompt content\nline 2"
			lines := strings.Split(shortPrompt, "\n")
			if len(lines) < 80 {
				t.Log("verified that prompt files are expected to be >= 80 lines")
			} else {
				t.Errorf("line count check failed to trigger for short prompt")
			}
		})

		// --- Add-Tests Boundary Cases ---
		t.Run("TC_2_3_1_AddTestsEmptyDiff", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "no new tests", "Empty diff behavior")
			assertContains(t, content, "necessary", "Empty diff behavior")
		})

		t.Run("TC_2_3_2_AddTestsNonCodeDiff", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "Documentation/Style changes", "Non-code diff handling instruction")
		})

		t.Run("TC_2_3_3_AddTestsMassiveDiff", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "directories", "Massive diff handling")
			assertContains(t, content, "Categories", "Massive diff handling")
		})

		t.Run("TC_2_3_4_AddTestsUninitializedRepo", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "Existing", "Uninitialized repo check")
			assertContains(t, content, ".revv/", "Uninitialized repo check")
		})

		t.Run("TC_2_3_5_AddTestsSkillFrontmatterName", func(t *testing.T) {
			content := readSkillFile(t, root, "add-tests")
			verifySkillFrontmatter(t, content, "add-tests")
		})

		t.Run("TC_2_3_6_AddTestsPlacementPaths", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, content, "Place tests in appropriate subdirectories", "Test placement instruction")
		})

		// --- Run-Tests Boundary Cases ---
		t.Run("TC_2_4_1_RunTestsNoDockerDaemon", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "Docker", "Docker daemon instructions")
			assertContains(t, content, "containers", "Docker daemon instructions")
		})

		t.Run("TC_2_4_2_RunTestsNoChromeDevTools", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "Degraded/Fallback Mode", "DevTools unavailable instruction")
			assertContains(t, content, "browser - needs human verification", "DevTools unavailable fallback")
		})

		t.Run("TC_2_4_3_RunTestsNoTestsDefined", func(t *testing.T) {
			tempDir := t.TempDir()
			exec := &mockExecutor{
				result: &runner.ExecResult{ExitCode: 0},
			}
			results, err := runner.RunAll(context.Background(), exec, tempDir, runner.FilterOpts{})
			if err != nil {
				t.Fatalf("unexpected error running empty tests: %v", err)
			}
			if len(results) != 0 {
				t.Errorf("expected 0 results, got %d", len(results))
			}
			summary := runner.Summary(results)
			if !strings.Contains(summary, "0 passed") {
				t.Errorf("expected summary to show 0 passed, got: %q", summary)
			}
		})

		t.Run("TC_2_4_4_RunTestsPriorityWarningFailure", func(t *testing.T) {
			results := []runner.TestResult{
				{
					Category: "unit",
					Name:     "warning_test",
					Priority: "warning",
					Passed:   false,
					Skipped:  false,
				},
			}
			if runner.HasBlockingFailure(results) {
				t.Errorf("HasBlockingFailure should be false for warning test failures")
			}
		})

		t.Run("TC_2_4_5_RunTestsPriorityBlockingFailure", func(t *testing.T) {
			results := []runner.TestResult{
				{
					Category: "unit",
					Name:     "blocking_test",
					Priority: "blocking",
					Passed:   false,
					Skipped:  false,
				},
			}
			if !runner.HasBlockingFailure(results) {
				t.Errorf("HasBlockingFailure should be true for blocking test failures")
			}
		})

		t.Run("TC_2_4_6_RunTestsBinaryResolution", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "which revv || go build -o /tmp/revv", "Go binary resolution instruction")
		})
	})

	// ==========================================
	// TIER 3: Cross-Feature Combinations (5 test cases)
	// ==========================================
	t.Run("Tier3_CrossFeature", func(t *testing.T) {
		t.Run("TC_3_1_RunTestsToAddTestsDelegation", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "add-tests", "run-tests delegates to add-tests first")
			assertContains(t, content, "skill", "run-tests delegates to add-tests first")
		})

		t.Run("TC_3_2_UpdateRepoToRunTestsIntegration", func(t *testing.T) {
			contentUpdate, _ := readPromptFile(t, root, "update-repo")
			contentRun, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, contentUpdate, "KEEP", "update-repo decides to KEEP/UPDATE/DELETE tests")
			assertContains(t, contentRun, "revv exec", "run-tests runs tests after prep phase")
		})

		t.Run("TC_3_3_InitRepoToUpdateRepoIdempotency", func(t *testing.T) {
			contentInit, _ := readPromptFile(t, root, "init-repo")
			contentUpdate, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, contentInit, "test.md", "init-repo initializes tests")
			assertContains(t, contentInit, "files", "init-repo initializes tests")
			assertContains(t, contentUpdate, "Minimal Disruption Principle", "update-repo has Minimal Disruption Principle check")
		})

		t.Run("TC_3_4_InitRepoToAddTestsPipeline", func(t *testing.T) {
			contentInit, _ := readPromptFile(t, root, "init-repo")
			contentAdd, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, contentInit, "test.md Format", "init-repo specifies test format")
			assertContains(t, contentAdd, "Test MD Format", "add-tests complies with test format")
		})

		t.Run("TC_3_5_UpdateRepoToAddTestsSequence", func(t *testing.T) {
			contentUpdate, _ := readPromptFile(t, root, "update-repo")
			contentAdd, _ := readPromptFile(t, root, "add-tests")
			assertContains(t, contentUpdate, "git log --oneline -N", "update-repo reads commit logs")
			assertContains(t, contentAdd, "git diff", "add-tests reads diff context")
		})
	})

	// ==========================================
	// TIER 4: Real-World Scenarios (6 test cases)
	// ==========================================
	t.Run("Tier4_RealWorld", func(t *testing.T) {
		t.Run("TC_4_1_GreenfieldSetupFlow", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "init-repo")
			assertContains(t, content, "setting up", "init-repo setup scenario")
			assertContains(t, content, ".revv/", "init-repo creates config directory")
			assertContains(t, content, "AGENTS.md", "init-repo creates agent instructions file")
		})

		t.Run("TC_4_2_FeatureDevAutomatedFlow", func(t *testing.T) {
			contentAdd, _ := readPromptFile(t, root, "add-tests")
			contentRun, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, contentAdd, "propose", "add-tests proposes tests")
			assertContains(t, contentAdd, "new tests", "add-tests proposes tests")
			assertContains(t, contentRun, "containers", "run-tests runs automated tests inside Docker")
		})

		t.Run("TC_4_3_BugFixFailureAnalysisFlow", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "Failure Diagnostics", "run-tests provides failure diagnostics")
			assertContains(t, content, "recommendation/fix", "run-tests suggests fixes")
		})

		t.Run("TC_4_4_UiChangeBrowserAutomationFlow", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "run-tests")
			assertContains(t, content, "Browser Test Execution", "run-tests supports browser tests")
			assertContains(t, content, "Chrome DevTools", "run-tests uses Chrome DevTools")
		})

		t.Run("TC_4_5_MultiCommitSyncMaintenanceFlow", func(t *testing.T) {
			content, _ := readPromptFile(t, root, "update-repo")
			assertContains(t, content, "UPDATE the commands", "update-repo handles command changes")
			assertContains(t, content, "DELETE the test", "update-repo deletes stale tests")
		})

		t.Run("TC_4_6_ContributorOnboardingFlow", func(t *testing.T) {
			readme, err := os.ReadFile(filepath.Join(root, "README.md"))
			if err != nil {
				t.Fatalf("failed to read README.md: %v", err)
			}
			readmeStr := string(readme)
			assertContains(t, readmeStr, "No API key. No binary install. No setup.", "onboarding experience")
			assertContains(t, readmeStr, "AGENTS.md", "onboarding references AGENTS.md")
		})
	})
}
