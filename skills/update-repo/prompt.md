# Skill Prompt: Update Repo

This prompt guides the updating of existing `revv` QA tests within a repository. It focuses on auditing existing test files, checking them against recent development history, and ensuring that they remain correct and aligned with the codebase's current state.

## Context

When this skill is triggered, you must gather and examine the following information:

1. **Existing `.revv/` Directory Structure:**
   - Locate and list all test files under `.revv/` (e.g. `.revv/<category>/<test_name>/test.md`).
   - Read and parse each test file's fields (`## Description`, `## Priority`, `## Type`, `## Commands`/`## Steps`, `## Expected Output`).
   - Inspect `.revv/Dockerfile` and any scripts in `.revv/helpers/`.

2. **Commit History:**
   - Execute `git log --oneline -N` where N defaults to 10 (or is overridden by user specification).
   - Analyze commit messages to understand recent features, bug fixes, or refactorings.

3. **Codebase Tree and Source Code:**
   - Map the current layout of the source code (files and directories).
   - If specific components or API endpoints have changed, locate their implementations.

4. **Additional Markdown Files:**
   - Examine any system or documentation updates that describe new architectural decisions or user guides.

## Output

Your output must be a clear list of actions for every existing test file, along with proposals for new test definitions. For each test, you must decide to:

1. **KEEP**: If the test is still valid and fully covers the corresponding codebase functionality.
2. **UPDATE**: If the test is still relevant but requires modifications (e.g. command arguments changed, output format changed, priority needs adjustments, new setup steps required). You must provide the updated file content.
3. **DELETE**: If the test is no longer applicable (e.g. the associated feature or code path was completely removed).

Additionally, suggest any new tests (with full test definitions in `test.md` format) that should be added to cover recent changes identified in the commit log or codebase structure.

## Rules

### 1. Minimal Disruption Principle
- Preserve relevant tests whenever possible. Do not rewrite test logic or descriptions unless they are outdated or incorrect.
- Only modify what is directly affected by codebase changes. Unrelated tests must remain untouched to prevent regression or loss of history.
- Ensure that the logic remains correct under the updated state of the codebase.

### 2. Decision Justification
For every test that is updated or deleted, you must provide a brief, technical explanation referencing the codebase change or commit that prompted the decision. For example:
- "Updated `.revv/build/compile_check/test.md` because the build tool was switched from `go` to `make`."
- "Deleted `.revv/unit/json_parsing/test.md` because the deprecated JSON parsing library was replaced by standard library encoders."

### 3. Test MD Format Compliance
Any updated or proposed new test must strictly follow the `revv` test format:
- `## Description`: Clear intent and validation reason.
- `## Priority`: Either `blocking` or `warning`.
- `## Type`: Either `automated`, `browser`, or `manual`.
- `## Commands` (if Type is `automated`): Bash commands to run within the Docker container sandbox.
- `## Steps` (if Type is `browser` or `manual`): Step-by-step instructions.
- `## Expected Output`: Successful validation criteria.

### 4. Dockerfile and Helper Alignment
- If build configurations or project dependencies have changed in the codebase (e.g., node version bumped, package manager changed to pnpm), make sure to update `.revv/Dockerfile` accordingly.
- Keep helper scripts in `.revv/helpers/` up to date.

### 5. Concrete Examples of Actions

#### Example 1: UPDATE action
*Old test:*
```markdown
## Description
Verify output of CLI version flag.

## Priority
warning

## Type
automated

## Commands
```bash
./bin/myapp -version
```

## Expected Output
Prints version information and exits with 0.
```

*Code Change:* The app was updated to use `--version` instead of `-version`.
*Action:* UPDATE the commands block.
*New Commands block:*
```bash
./bin/myapp --version
```

#### Example 2: DELETE action
*Old test:* `.revv/unit/xml_parser/test.md`
*Code Change:* The XML import features were completely removed from the codebase.
*Action:* DELETE the test since there is no XML parsing feature to verify anymore.

### 6. Review Interface
- Present a summary table of the planned updates:
  | Test Path | Action (Keep/Update/Delete) | Reason |
  | --- | --- | --- |
- Under the summary table, output the exact file modifications or new files to be written.
