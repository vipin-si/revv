# Skill Prompt: Update Repo

You are a QA engineer doing a periodic review of an existing test suite. The tests were set up a while ago — your job is to check if they're still accurate, fix what's drifted, delete what's dead, and add what's missing.

## Step 1: Understand What Changed

Before touching any tests, understand what happened since the last update:

1. **Read the commit history:**
   ```bash
   git log --oneline -N
   ```
   Where N defaults to 10. Read each commit message — look for:
   - New features or commands added
   - Bug fixes (may need regression tests)
   - Refactors (tests may need updated paths/commands)
   - Deleted features (tests should be removed)
   - Dependency changes (Dockerfile may need updating)

2. **Read the current test suite:**
   ```bash
   find .revv -name "test.md" | sort
   ```
   For each test, read its `## Commands` or `## Steps` and understand what it actually checks.

3. **Map the code tree:**
   ```bash
   find . -maxdepth 3 -not -path './.git/*' -not -path './node_modules/*'
   ```
   Look for new files, deleted files, or moved files that tests might reference.

4. **Read documentation:**
   - `README.md`, `CONTRIBUTING.md`, `AGENTS.md`
   - Any new docs that describe changed behavior

## Step 2: Audit Every Existing Test

For EACH test.md file in `.revv/`, make one of three decisions:

### ✅ KEEP — Test is still valid
The code it tests hasn't changed. The commands still work. The expected output is still correct.

No action needed. Don't touch it.

### ✏️ UPDATE — Test needs changes
The test is still relevant but something drifted. Common reasons:

| What changed | What to update |
|---|---|
| CLI flag renamed (`-v` → `--verbose`) | Update `## Commands` |
| Function moved to different package | Update test description |
| Output format changed (text → JSON) | Update `## Expected Output` |
| Build tool changed (`go build` → `make build`) | Update `## Commands` |
| Feature became critical path | Change priority `warning` → `blocking` |
| New setup step required (DB, env var) | Add `## Setup` section |

### 🗑️ DELETE — Test is dead
The feature it tests no longer exists. Common reasons:

- The code path was completely removed
- The CLI command was deleted
- The API endpoint was deprecated and removed
- The test was for a third-party dependency that was replaced

**Don't delete tests just because they seem redundant.** Only delete if the thing being tested is genuinely gone.

## Step 3: Propose New Tests

After auditing existing tests, check if the recent commits introduced anything that isn't covered:

- **New CLI command or flag** → needs a sanity test
- **New API endpoint** → needs an integration test
- **Bug fix** → needs a regression test (fails before fix, passes after)
- **New dependency** → Dockerfile may need updating
- **New config option** → needs a test that exercises it

Follow the same test format and categories from the [init-repo skill](https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md).

## Step 4: Update the Dockerfile

Check if `.revv/Dockerfile` needs changes:

| Code change | Dockerfile update needed |
|---|---|
| Go version bumped in `go.mod` | Update base image (`golang:1.22-alpine` → `golang:1.23-alpine`) |
| New system dependency (e.g., `libssl`) | Add `RUN apk add --no-cache libssl-dev` |
| Package manager changed (`npm` → `pnpm`) | Update install command |
| New build step added | Add `RUN` command |
| Source directory restructured | Update `COPY` paths |

If nothing changed that affects the build environment, leave the Dockerfile alone.

## Step 5: Present the Review

### Summary Table

Present this FIRST so the developer can see all decisions at a glance:

```markdown
| Test Path | Action | Reason |
|-----------|--------|--------|
| .revv/build/compile_check/test.md | ✅ KEEP | Build command unchanged |
| .revv/sanity/cli_help/test.md | ✏️ UPDATE | New `--json` flag added to help output |
| .revv/integration/xml_parser/test.md | 🗑️ DELETE | XML parsing removed in commit abc1234 |
| .revv/regression/json_edge_case/test.md | 🆕 NEW | Bug fix in commit def5678 needs regression test |
```

### Detailed Changes

For each UPDATE and NEW action, show the exact file content:

```markdown
### ✏️ UPDATE: .revv/sanity/cli_help/test.md

**Reason:** Commit `abc1234` added `--json` flag. Help output now includes it.

**Change:** Added `--json` to the grep check in Commands.

[Full updated test.md content here]
```

For each DELETE, explain why:

```markdown
### 🗑️ DELETE: .revv/integration/xml_parser/test.md

**Reason:** XML import feature completely removed in commit `def5678`.
The `internal/xml/` package no longer exists.
```

## Rules

### Minimal Disruption Principle
- Preserve relevant tests whenever possible. Don't rewrite something that's working.
- Only modify what is directly affected by codebase changes.
- If you're unsure whether a test is still valid, KEEP it and note your uncertainty.

### Decision Justification
Every UPDATE and DELETE must reference a specific commit or code change. Don't say "updated for consistency." Say "updated because commit `abc1234` renamed the `--verbose` flag to `-v`."

### Test MD Format Compliance
Any updated or new test must follow the standard format:
- `## Description`: What it tests and why it matters
- `## Priority`: `blocking` or `warning`
- `## Type`: `automated` or `browser`
- `## Commands` (if automated): Real shell commands, exit 0 = pass
- `## Steps` (if browser): Numbered steps for Chrome DevTools
- `## Expected Output`: What success looks like

### Don't Over-Update
If 10 commits happened and only 2 affect tests, you should have ~2 updates and ~8 KEEPs. A review that changes everything is suspicious. Question it.

### AGENTS.md
Check if the AGENTS.md revv section is still accurate. If new skills were added or URLs changed, update it. If it's fine, leave it alone.
