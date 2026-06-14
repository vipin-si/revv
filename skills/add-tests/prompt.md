# Skill Prompt: Add Tests

You are reviewing a PR or commit to decide: does this change need new tests? Your job is to look at what changed, check if existing tests already cover it, and if not, write the test.

## Step 1: Understand the Change

Read the diff to understand what the developer actually changed:

```bash
# For the last commit:
git diff HEAD~1

# For a PR (changes since branching from main):
git diff main...HEAD

# If uncommitted changes:
git diff
```

Classify each changed file:

| File type | Test implication |
|---|---|
| Source code (`.go`, `.js`, `.py`, `.rs`) | Likely needs tests |
| Build config (`Makefile`, `package.json`, `go.mod`) | May need build test update |
| Documentation (`README.md`, `CONTRIBUTING.md`) | No tests needed |
| Test files (`*_test.go`, `*.test.js`) | No revv tests needed (they ARE tests) |
| CI/CD (`.github/workflows/`) | No tests needed |
| Config (`.env.example`, `.gitignore`) | No tests needed |

## Step 2: Check Existing Coverage

Read the current test suite:

```bash
find .revv -name "test.md" | sort
```

For each changed source file, ask:
1. Is there already a test that exercises this code path?
2. Would the existing test catch a regression if this change broke something?
3. Does the existing test need updating (e.g., new flag, changed output)?

Map the changes to existing tests. If coverage exists, no new test needed.

## Step 3: Make the Decision

### Decision on Test Necessity

State clearly: **YES, add tests** or **NO, existing coverage is sufficient.**

### If YES — write the tests

For each new test, provide:

1. **Justification**: What changed and why existing tests don't cover it
2. **Test path**: Where it goes (e.g., `.revv/regression/nil_pointer_fix/test.md`)
3. **Full test.md content**: Complete, valid, ready to write to disk

### If NO — explain why

Don't just say "no tests needed." Explain specifically:

- "The change in `parser.go` only renames an internal variable. The existing test `.revv/integration/parser_integration/test.md` already tests the parse output, which is unchanged."
- "This is a documentation-only change (`README.md`). No code behavior changed."
- "The refactor in `runner.go` moves code between functions but the public API is identical. Existing tests in `.revv/integration/runner_pipeline/test.md` cover the same entry points."

## Decision Criteria

### Always add tests for:

1. **New features**: Any new CLI command, flag, API endpoint, or user-facing behavior
   ```diff
   + rootCmd.AddCommand(newLintCmd())
   ```
   → Add `.revv/sanity/lint_command/test.md`

2. **Bug fixes**: The fix should have a regression test that would catch the bug if it reappeared
   ```diff
   - if len(items) > 0 {
   + if len(items) >= 0 {
   ```
   → Add `.revv/regression/empty_items_fix/test.md`

3. **New error handling**: If a new error path was added, test that it returns the right error
   ```diff
   + if cfg == nil {
   +   return fmt.Errorf("config file not found")
   + }
   ```
   → Add `.revv/regression/missing_config_error/test.md`

4. **Changed output format**: If the output shape changed, existing tests may pass but miss the regression
   ```diff
   - fmt.Printf("Version: %s\n", version)
   + fmt.Printf(`{"version": "%s"}\n`, version)
   ```
   → Update existing test OR add new one for JSON format

5. **New dependencies**: If a new package/tool is required, the Dockerfile may need updating and a build test should verify it

### Never add tests for:

1. **Documentation-only changes**: README, CONTRIBUTING, comments, docstrings
2. **Style/formatting**: Code reformatting, import reordering, whitespace
3. **CI/CD changes**: Workflow files, Dockerfiles not in `.revv/`
4. **Existing test changes**: If someone modifies `*_test.go` files, those are already tests

### Use judgment for:

1. **Refactoring**: If behavior is unchanged, existing tests should cover it. But if the refactor changes internal APIs that tests reference, update the test commands.
2. **Dependency bumps**: Usually no test needed unless the bump changes behavior or requires a Dockerfile update.
3. **Config file changes**: If a new config option is added and the app reads it, consider a test.

## Test Writing Rules

Follow the same format as [init-repo](https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md):

### Test MD Format Compliance

```markdown
## Description
[What this test verifies and why. Reference the commit/PR that motivated it.]

## Priority
[blocking | warning]

## Type
[automated | browser]

## Commands
```bash
[Real commands. Exit 0 = pass.]
```

## Expected Output
[What success looks like.]
```

### Category Placement

| Change type | Category |
|---|---|
| New build requirement | `build/` |
| New CLI command/flag | `sanity/` |
| New API/service interaction | `integration/` |
| Bug fix | `regression/` |
| Security-related | `security/` |
| UI change | `browser/` |

### Naming

The directory name IS the test name. Make it descriptive:
- ✅ `regression/nil_pointer_parser_fix`
- ✅ `sanity/json_output_flag`
- ❌ `test1`
- ❌ `new_test`

## Output Format

Present your decision clearly:

```markdown
## Decision: [YES | NO]

### Justification
[Why tests are or aren't needed, referencing specific files and changes]

### Changes to existing tests
[If any existing tests need updating, list them with the specific changes]

### New tests
[If adding tests, show the full test.md for each]

### Dockerfile changes
[If the Dockerfile needs updating, show the change]
```

## Edge Cases

- **Multiple unrelated changes in one PR**: Evaluate each change independently. Some may need tests, others may not.
- **Reverted commit**: If a commit was reverted, check if any tests were added for the original commit. They may now need to be deleted.
- **Merge commits**: Skip merge commits — they don't introduce new code.
- **Large refactors**: If 20+ files changed but behavior is the same, don't add 20 tests. Verify existing tests pass and explain why no new tests are needed.
