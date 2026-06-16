# Skill Prompt: Review Init

You are reviewing the quality of a `.revv/` test suite that was just generated (or updated) for a repository. Your job is to find problems that would cause tests to fail, miss bugs, or waste a contributor's time.

## When This Runs

This skill is called:
- Automatically at the end of `revv init` (to review the generated test suite)
- Automatically at the end of `revv update` (to review changes to the test suite)
- Manually when someone says "revv review" or "review tests"

## Step 1: Gather Context

Before launching reviewers, collect:

1. **The project's build config** — `go.mod`, `package.json`, `Cargo.toml`, `Makefile`, etc.
2. **The Dockerfile** — `.revv/Dockerfile`
3. **All test.md files** — `find .revv -name "test.md" | sort`
4. **The CLI help** — if it's a CLI project, note the available commands
5. **The existing unit tests** — `find . -name "*_test.go" -o -name "*.test.js" -o -name "test_*.py" | head -30`

You'll pass this context to each reviewer.

## Step 2: Launch 3 Parallel Reviewers

Launch all three as **parallel subagents** using the `research` subagent type. All three run simultaneously. Each reviewer gets the same context but evaluates a different dimension.

```
┌────────────────────────────────────────────────────────┐
│                  .revv/ Test Suite                      │
├──────────────────┬──────────────────┬──────────────────┤
│  Command         │  Coverage &      │  QA Realism      │
│  Correctness     │  Redundancy      │                  │
├──────────────────┼──────────────────┼──────────────────┤
│ • Bash syntax    │ • CLI coverage   │ • Would QA run   │
│ • Paths exist    │ • Redundant w/   │   this test?     │
│ • Alpine compat  │   unit tests     │ • Pass/fail      │
│ • Dockerfile env │ • Missing cats   │   specificity    │
│ • Exit codes     │ • Test count     │ • Priority right │
└────────┬─────────┴────────┬─────────┴────────┬─────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            ▼
                  Consolidated Review
               (prioritized by severity)
```

---

## Reviewer Prompts

### Reviewer 1: Command Correctness

```
You are a DevOps engineer who builds and debugs Docker containers. Your job is to review a set of test.md files and verify that every shell command would actually work inside the project's Docker container.

Context:
- Dockerfile: {DOCKERFILE_CONTENT}
- Base image: {BASE_IMAGE} (this tells you what OS, package manager, and tools are available)
- Project type: {PROJECT_TYPE} (Go/Node/Python/Rust)
- Build config: {BUILD_CONFIG_CONTENT}

Read every test.md file listed below and check:

{LIST_OF_TEST_FILES_WITH_CONTENT}

For each test, verify:

1. **Bash syntax**: Would `bash -n` pass on the ## Commands block? Look for:
   - Unclosed quotes or brackets
   - Missing `then`/`fi` or `do`/`done`
   - Incorrect variable expansion
   - Pipes and redirections that won't work in `sh` (Alpine uses ash, not bash)

2. **Binary paths**: Does the test reference binaries that exist in the Docker image?
   - Is the pre-built binary at the right path? (e.g., `/workspace/bin/myapp` vs `./bin/myapp`)
   - Are tools like `curl`, `jq`, `grep`, `diff` available in the base image?
   - For Alpine: `apk` not `apt-get`. No `bash` by default (only `sh`/`ash`).

3. **Environment assumptions**:
   - Does the test assume a HOME directory exists? (Docker root HOME = /root)
   - Does it assume network access? (Docker containers may not have internet)
   - Does it assume other services are running? (databases, servers)
   - Does it create temp directories that might not exist?

4. **Exit code handling**:
   - Does the test correctly exit non-zero on failure?
   - Are there commands that could silently fail? (e.g., `grep` returns 1 on no match)
   - Is `set -e` needed but missing?
   - Could the test pass even when it should fail?

5. **Alpine compatibility**:
   - Does it use GNU-specific flags? (`grep -P`, `sed -i ''`, `readlink -f`)
   - Does it use `bash` features in a `sh` environment?
   - Does it reference paths that exist on Debian but not Alpine?

For each issue found, report:
- **Test file**: path to the test.md
- **Line/Command**: the specific command that's wrong
- **Issue**: what would fail and why
- **Severity**: CRITICAL (test will fail) / HIGH (test may fail) / MEDIUM (test is fragile) / LOW (style issue)
- **Fix**: the corrected command

End with:
- Count of tests with no issues vs. tests with issues
- List of tests that would DEFINITELY fail in Docker
- Overall command correctness grade: A (all commands verified) to F (most would fail)

Send your complete review back as a message.
```

### Reviewer 2: Coverage & Redundancy

```
You are a test architect reviewing a test suite for completeness and efficiency. Your job is to find gaps in coverage and tests that duplicate work the project's own unit tests already do.

Context:
- Project type: {PROJECT_TYPE}
- CLI commands/subcommands: {CLI_COMMANDS}
- Source tree: {SOURCE_TREE}
- Existing unit test files: {EXISTING_TEST_FILES}
- Build system: {BUILD_SYSTEM} (e.g., mage, make, npm)

Read every test.md file listed below:

{LIST_OF_TEST_FILES_WITH_CONTENT}

Check these dimensions:

1. **CLI coverage**: For CLI projects, list every subcommand from `--help` output. For each:
   - Is there at least one test that exercises it?
   - Flag untested commands as coverage gaps

2. **Redundancy with unit tests**: For each test.md:
   - Does it just run `go test -v -run TestFoo ./...`? If so, it's REDUNDANT — it's testing the exact same thing the project's own unit tests already test
   - A good revv test should test BEHAVIOR from the outside (run the binary, check output) not delegate to internal tests
   - Exception: a test called `unit_tests` that runs the full suite is fine as a build gate — but it shouldn't be in `integration/`

3. **Missing categories**: Based on the project type, are expected categories present?
   - `build/` — does every project have at least one? (required)
   - `sanity/` — does it have basic smoke tests? (required)
   - `integration/` — if the project has multiple components, are they tested together?
   - `regression/` — are edge cases covered?
   - `security/` — basic secret/vuln scanning?
   - `browser/` — if the project has a web UI, are there browser tests?

4. **Test count**: Is the count appropriate for the project size?
   - < 10 source files: 5-10 tests
   - 10-50 files: 10-20 tests
   - 50+ files: 15-30 tests
   - Flag if under or over

5. **Category balance**: Are tests distributed well or clustered?
   - All tests in `build/` = bad (no behavioral testing)
   - No `regression/` tests = suspicious (every project has edge cases)

For each issue found, report:
- **Type**: GAP (missing test) / REDUNDANT (duplicates unit tests) / IMBALANCE (wrong category distribution)
- **Detail**: what's missing or wrong
- **Severity**: HIGH (significant gap) / MEDIUM (notable omission) / LOW (nice-to-have)
- **Suggestion**: specific test to add or change

End with:
- CLI command coverage: X/Y commands tested
- Redundancy count: N tests that just delegate to unit tests
- Category coverage: which categories are present/missing
- Overall coverage grade: A (comprehensive) to F (major gaps)

Send your complete review back as a message.
```

### Reviewer 3: QA Realism

```
You are a senior QA engineer with 10 years of experience. You've seen hundreds of test suites. Your job is to evaluate whether these tests would actually catch bugs in production, or whether they're checkbox tests that look good but catch nothing.

Read every test.md file listed below:

{LIST_OF_TEST_FILES_WITH_CONTENT}

For each test, evaluate:

1. **Would a QA engineer actually run this?**
   - Is this test checking something that could realistically break?
   - Or is it testing something trivially obvious that would never fail?
   - A test that checks `--help` output exists is useful. A test that checks `--help` contains a specific word that's hardcoded in source is fragile.

2. **Can this test actually fail?**
   - Are the assertions specific enough?
   - Could the test pass even when the feature is broken?
   - Example of a bad test: `go build ./... && echo PASS` — this passes even if the binary is broken, as long as it compiles
   - Example of a good test: `go build -o /tmp/app . && /tmp/app --version | grep -q "v" || exit 1`

3. **Is the priority correct?**
   - `blocking` = if this fails, the PR should NOT merge. Is that true for this test?
   - `warning` = advisory only. Should any of these actually be blocking?
   - Common mistake: making regression tests `blocking` when they should be `warning`
   - Common mistake: making integration tests `warning` when they should be `blocking`

4. **Is the description helpful?**
   - Could a contributor who sees "FAIL: integration/serve_startup" understand what broke from the description alone?
   - Does the Expected Output tell them what success looks like?

5. **Test isolation**: Does each test check ONE thing?
   - Flag tests that do too much (compile + run + check output + verify config)
   - Each test should have a single failure mode

6. **Failure diagnostics**: When a test fails, will the contributor know WHY?
   - Does the test print the actual output on failure? (not just "FAIL")
   - Does it print what was expected vs. what was received?
   - Are error messages specific? ("FAIL: binary not found at /workspace/bin/app" vs "FAIL")

For each issue found, report:
- **Test file**: path to the test.md
- **Issue type**: UNTESTABLE (can't fail) / WRONG_PRIORITY / VAGUE_ASSERTION / POOR_DIAGNOSTICS / OVER_SCOPED
- **Detail**: what's wrong
- **Severity**: HIGH (test is misleading) / MEDIUM (test is weak) / LOW (minor improvement)
- **Fix**: specific improvement

End with:
- Count of strong tests vs. weak tests
- List of tests that could never fail (worst offenders)
- Overall QA realism grade: A (would trust this suite) to F (checkbox theater)

Send your complete review back as a message.
```

---

## Step 3: Consolidate Findings

After all three reviewers complete, merge their findings into a single report. Use this format:

```markdown
# revv Test Suite Review — Pass {N}

**Repository**: {REPO_NAME}
**Tests reviewed**: {COUNT}
**Review date**: {DATE}

## Scoring Summary

| Dimension | Grade | Source |
|:----------|:------|:-------|
| **Command Correctness** | X/5 | Reviewer 1 |
| **Coverage & Completeness** | X/5 | Reviewer 2 |
| **QA Realism** | X/5 | Reviewer 3 |
| **Overall** | **X/5** | |

**Verdict**: [PASS — ready to use / NEEDS FIXES — {N} issues to address]

---

## 🔴 Must Fix (will fail or miss bugs)

### 1. [Issue title]
**Source**: [Command Correctness / Coverage / QA Realism]
**Severity**: 🔴 CRITICAL / HIGH

[Description]

**Fix**: [Specific, actionable fix — include the corrected command or new test.md content]

---

## 🟡 Should Fix (weakens the suite)

### N. [Issue title]
...

---

## 🟢 Nice to Have

- [Brief description] — [fix]

---

## Cross-Reviewer Consensus

Issues flagged by multiple reviewers (highest confidence):

| Issue | Flagged By | Consensus Severity |
|:------|:-----------|:-------------------|
| [Issue] | Reviewer 1 + 3 | 🔴 CRITICAL |

---

## Top 3 Actions to Improve This Test Suite

1. [Most impactful action]
2. [Second most impactful]
3. [Third most impactful]
```

Save the report as an artifact.

## Step 4: Apply Fixes

After presenting the report:

1. **Fix all 🔴 Must Fix issues** — rewrite the test.md files with corrected commands, add missing tests, remove redundant ones
2. **Fix all 🟡 Should Fix issues** — unless there's a good reason to skip
3. **Skip 🟢 Nice to Have** — unless trivial

## Step 5: Re-Review (Loop)

After applying fixes, run the review again:

- **Pass 2**: Add to each reviewer's prompt: "This is a SECOND review pass. The test suite has been revised. Verify fixes from Pass 1 are applied. Be stricter — focus on issues you may have missed."
- **Pass 3** (only if Pass 2 still has 🔴 issues): Final check. Only flag issues that would cause real failures.

### When to Stop

- Stop after **Pass 2** if no remaining 🔴 issues
- Run **Pass 3** only if Pass 2 still has 🔴 items
- Never run more than 3 passes

---

## Integration with Other Skills

| Skill | How it connects |
|:------|:---------------|
| `init-repo` | Init generates `.revv/`, then calls this skill to review. Loop up to 3 passes. |
| `update-repo` | After updating tests, call this skill to verify changes. Single pass. |
| `add-tests` | After adding new tests, call this skill to review just the new ones. Single pass. |

## Tips

- **Cross-reviewer consensus is gold** — if Command Correctness AND QA Realism both flag the same test, it's definitely broken
- **Redundancy is the #1 issue** — most generated test suites delegate too much to existing unit tests instead of testing behavior
- **Alpine is the #1 failure cause** — most command errors come from assuming Debian/Ubuntu in an Alpine container
- **The review should be adversarial** — the reviewers should try to break the tests, not validate them
