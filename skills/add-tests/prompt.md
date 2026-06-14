# Skill Prompt: Add Tests

This prompt guides the process of determining if new QA tests are required to cover recent code modifications and, if so, adding them to the `.revv/` suite in the correct format.

## Context

When this skill is triggered, you must perform the following analysis:

1. **Detect Changes (Git Diff):**
   - Retrieve the change diff of the current working copy or recent commits.
   - Run `git diff HEAD~1` to see the changes in the last commit, or `git diff main...HEAD` (or target branch) for a pull request diff.
   - Analyze the files modified: are they source files (`.go`, `.js`, `.py`, etc.), build configurations, or documentation?

2. **Inspect Existing `.revv/` Tests:**
   - Scan the existing `.revv/` test structure to understand what coverage is already present.
   - Map existing test definitions to modified code files to check if existing tests already cover the logic.

3. **Codebase Tree and Markdown Files:**
   - Verify the location of the modifications inside the codebase structure.
   - Check if new features or APIs have been described in recently updated markdown documentation.

## Output

Your output must consist of a clear decision and corresponding actions:

1. **Decision on Test Necessity:**
   - State clearly whether new tests are needed to cover the changes.
   - **If YES**: Specify which files are being added and in what directories (e.g. `.revv/unit/new_feature/test.md`).
   - **If NO**: Provide a detailed explanation justifying why no new tests are necessary (e.g. "Only refactored internal helper comments without changing execution behavior", or "The existing test `.revv/build/compile_check/test.md` already covers this change").

2. **Test File Creation:**
   - If tests are needed, write the complete, valid `test.md` content for each proposed test, matching the standard `revv` format.

## Rules

### 1. Decision Criteria (When to Add Tests)
- **Feature Additions**: Any new command-line option, API endpoint, logic branch, or module **must** have at least one corresponding test.
- **Bug Fixes**: A bug fix should have a regression test that fails prior to the fix and passes after the fix.
- **Pure Refactoring**: If code structure changes but behavior does not, verify existing tests cover it and explain why new ones are not needed.
- **Documentation/Style changes**: Do not add tests for changes that do not affect code compilation or execution.

### 2. Test Placement and Categories
- Place tests in appropriate subdirectories under `.revv/` based on their category:
  - `build/`: Compilations and sanity checks.
  - `unit/`: Narrow testing of components or individual functions.
  - `integration/`: Testing interactions between systems.
  - `browser/`: Web UI workflows.
  - `manual/`: Non-automated UI or user flows.

### 3. Test MD Format Compliance
Every created test must follow the standard `revv` format:
- `## Description`: Summary of target behavior.
- `## Priority`: `blocking` or `warning`.
- `## Type`: `automated`, `browser`, or `manual`.
- `## Commands` (if automated) or `## Steps` (if browser/manual).
- `## Expected Output`: Verifiable outcome.

### 4. Concrete Examples of Test Decisions

#### Example 1: New CLI Parameter Added
*Diff:*
```diff
--- a/cmd/app/main.go
+++ b/cmd/app/main.go
+ var timeout = flag.Duration("timeout", 5*time.Second, "operation timeout")
```
*Decision:* YES, add test.
*New Test File:* `.revv/unit/cli_timeout/test.md`
```markdown
## Description
Verify CLI accepts custom timeout duration.

## Priority
blocking

## Type
automated

## Commands
```bash
./bin/myapp --timeout 10s
```

## Expected Output
Exits with code 0 without timeout errors.
```

#### Example 2: Documentation Typo Corrected
*Diff:*
```diff
--- a/README.md
+++ b/README.md
- This is a exmple of the app.
+ This is an example of the app.
```
*Decision:* NO, do not add test.
*Explanation:* The changes are restricted entirely to user documentation in `README.md` and do not alter the compilation or runtime behavior of the software.
