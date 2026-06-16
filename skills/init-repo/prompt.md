# Skill Prompt: Init Repo

You are setting up automated QA for a code repository using revv. Your job is to think like a senior QA engineer doing a first pass on a new project. You are not writing unit tests — the project already has those. You are writing the tests that a manual QA person would run before approving a release.

## Context Gathering

Before generating anything, read and understand the project:

1. **Read all documentation:**
   - `README.md` — what does this project do? How is it used?
   - `CONTRIBUTING.md` — how do contributors build and test?
   - `AGENTS.md` — does revv already exist here?
   - Any other `.md` files in the root

2. **Understand the build system:**
   - Go: `go.mod`, `Makefile`
   - Rust: `Cargo.toml`
   - Node.js: `package.json` (read `scripts` section carefully)
   - Python: `requirements.txt`, `pyproject.toml`, `setup.py`
   - Docker: `Dockerfile`, `docker-compose.yml`

3. **Map the code tree:**
   ```bash
   find . -maxdepth 3 -not -path './.git/*' -not -path './node_modules/*' -not -path './vendor/*'
   ```

4. **Read key source files** to understand what the project actually does:
   - Entry points (`main.go`, `index.js`, `app.py`, `src/main.rs`)
   - CLI commands or API routes
   - Configuration files

5. **Check if `.revv/` already exists.** If it does, this is the wrong skill — tell the user to use `revv update` instead.

## What Tests to Generate

Think like a QA engineer who just joined the team. You're asking: "If a contributor submits a PR, what should I check before approving it?"

### Test Categories

Generate tests in these categories, in this priority order. Skip categories that don't apply to the project.

#### 1. `build/` — Does it build? (blocking)
The most fundamental check. Every project must have at least one build test.

- **Does the project compile/build from a clean checkout?**
- **Does `make build` / `npm run build` / `cargo build` succeed?**
- **Is the output binary/bundle present and executable?**
- **Do all dependencies resolve correctly?**

Example: For a Go project:
```
.revv/build/compile_check/test.md   → go build ./...
.revv/build/binary_exists/test.md   → test -x ./bin/myapp
```

#### 2. `sanity/` — Does it work at all? (blocking)
Basic smoke tests that verify the project isn't completely broken.

- **Does the CLI show help output?** (`./myapp --help`)
- **Does the server start?** (`./myapp serve & sleep 2 && curl localhost:8080/health`)
- **Does the main command run without crashing?** (`./myapp --version`)
- **Do basic config files load without errors?**

Example: For a CLI tool:
```
.revv/sanity/cli_help/test.md       → ./myapp --help | grep "Usage"
.revv/sanity/version_output/test.md → ./myapp --version | grep -E "[0-9]+\.[0-9]+"
```

#### 3. `integration/` — Do the pieces work together? (blocking or warning)
Tests that verify different parts of the system interact correctly.

- **Does the API respond to requests?**
- **Does the database migration run?**
- **Does the authentication flow work end-to-end?**
- **Do background workers process jobs?**
- **Do file imports/exports produce correct output?**

Example: For a web API:
```
.revv/integration/api_health/test.md     → curl -f http://localhost:8080/health
.revv/integration/api_crud/test.md       → create, read, update, delete a resource
```

#### 4. `regression/` — Do known edge cases still work? (warning)
Tests based on past bugs, tricky inputs, or boundary conditions you spot in the code.

- **Does the parser handle malformed input?**
- **Does the app handle missing config gracefully?**
- **Does it work with empty input / zero values / very large input?**
- **Does error handling return proper messages (not stack traces)?**

Example: For a JSON parser:
```
.revv/regression/malformed_json/test.md    → feed broken JSON, expect graceful error
.revv/regression/empty_input/test.md       → feed empty string, expect no crash
```

#### 5. `security/` — Are there obvious security issues? (warning)
Basic security checks appropriate to the project.

- **Are secrets/tokens not hardcoded in source?** (`grep -r "password\|secret\|api_key" --include="*.go"`)
- **Does the server reject unauthenticated requests to protected routes?**
- **Are dependencies free of known vulnerabilities?** (`npm audit`, `go vuln check`)
- **Does the app sanitize user input?**

#### 6. `browser/` — Does the UI/app work? (blocking or warning)
For any project with a web UI, API with a frontend, or visual output. These are `## Type: browser` tests executed by the IDE via Chrome DevTools MCP.

- **Does the landing page load without errors?**
- **Does the login flow work?**
- **Do forms submit correctly?**
- **Does navigation between pages work?**
- **Does the app render correctly on mobile viewports?**
- **Are there console errors on page load?**
- **Does the UI look correct?** (the IDE takes screenshots and judges)
- **Are the docs accurate?** (the IDE can browse docs and compare to behavior)
- **Is the error messaging user-friendly?**

Example:
```
.revv/browser/landing_page/test.md   → open localhost:3000, verify title, no console errors
.revv/browser/login_flow/test.md     → enter credentials, verify redirect to dashboard
.revv/browser/docs_accuracy/test.md  → open README, verify instructions match actual behavior
```

### How Many Tests?

- **Small project** (< 10 source files): 5-10 tests
- **Medium project** (10-50 files): 10-20 tests
- **Large project** (50+ files): 15-30 tests

Quality over quantity. Each test should catch a real problem, not pad a number. Ask yourself: "Would a QA engineer actually check this?" If not, don't generate it.

### Test Writing Rules

1. **Commands must be REAL.** Every command in `## Commands` must work inside the Docker container. No pseudocode, no `# TODO`, no placeholders. If you're not sure a command works, be conservative.

2. **Be specific about failure.** Don't write `make build`. Write:
   ```bash
   make build 2>&1
   test -x ./bin/myapp || (echo "FAIL: binary not found at ./bin/myapp" && exit 1)
   echo "PASS: binary built successfully"
   ```

3. **Test one thing per test.** A test called `compile_check` should only check compilation, not also run the test suite.

4. **Use meaningful names.** The directory name IS the test name. `api_health` is good. `test1` is not.

5. **Set priority correctly:**
   - `blocking` = if this fails, the PR should NOT be merged
   - `warning` = something is wrong but not critical

6. **Set type correctly:**
   - `automated` = shell commands in Docker (the binary runs these in parallel)
   - `browser` = steps the IDE executes via Chrome DevTools MCP (UI, visual, subjective checks)

## test.md Format

Every `test.md` MUST have these sections in this order:

```markdown
## Description
[What this test verifies and why it matters. Be specific.]

## Priority
[blocking | warning]

## Type
[automated | browser]

## Commands
```bash
[Real shell commands. Exit 0 = pass, non-zero = fail.]
[Print PASS/FAIL messages for clarity.]
```

## Expected Output
[What success looks like. Be specific enough that someone can verify.]
```

For `browser` type tests, replace `## Commands` with `## Steps`:
```markdown
## Steps
1. [Action the IDE should take via Chrome DevTools]
2. [Next action]
3. [Verification step]
```

For `browser` tests that need a running server, add `## Setup`:
```markdown
## Setup
```bash
npm start &
sleep 3
```
```

## Dockerfile Rules

Generate `.revv/Dockerfile` with:

- **Base image**: Minimal official image matching the project language and version. **Read the version from the project's config** (e.g., `go.mod` for Go, `package.json` for Node, `Cargo.toml` for Rust). Examples: `golang:1.26-alpine`, `node:20-alpine`, `python:3.11-slim`, `rust:1.78-alpine`
- **System deps**: Install `make`, `git`, `gcc`, and any other tools the build needs
- **Working directory**: `/workspace`
- **COPY source**: `COPY . /workspace`
- **Pre-build**: Run the project's build step (`RUN make build`, `RUN npm install && npm run build`, etc.) so tests start instantly — no waiting for compilation
- **Do NOT** include secrets, API keys, or `.env` files in the image

## AGENTS.md Rules

After generating `.revv/`, update the repo's `AGENTS.md`:

1. **Check if `AGENTS.md` exists:**
   - If YES → read its contents
   - If NO → create it with an empty file first

2. **Check for existing revv section:**
   - Search for `vssinghh/revv` in the file
   - If found → do NOT add anything (already configured)
   - If not found → append the revv section below

3. **Section to append:**

```markdown
---

## revv — Automated QA

This repository uses [revv](https://github.com/vssinghh/revv) for automated QA testing.
Tests are defined in `.revv/` and run within Docker containers or through browser automation.

When executing QA workflows, read the skill instructions from these URLs and follow them:
- **revv init**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md
- **revv update**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/update-repo/SKILL.md
- **revv add-tests**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/add-tests/SKILL.md
- **revv run**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/run-tests/SKILL.md
```

## Final Checklist

Before presenting the generated files to the developer:

- [ ] Every test.md has all required sections (Description, Priority, Type, Commands/Steps, Expected Output)
- [ ] Every automated test's commands would actually work in the Dockerfile's environment
- [ ] At least one `build/` test exists (blocking)
- [ ] At least one `sanity/` test exists (blocking)
- [ ] At least one `browser/` test exists if the project has a UI
- [ ] Dockerfile pre-builds the project
- [ ] AGENTS.md is updated (not duplicated)
- [ ] Test names are descriptive (no `test1`, `check2`)
- [ ] No more than 30 tests total

## Review Loop — Quality Gate

**After generating all files, do NOT present them to the user yet.** First, run the review-init skill to validate quality.

1. Read the [review-init skill](https://raw.githubusercontent.com/vssinghh/revv/main/skills/review-init/SKILL.md) and follow its instructions.
2. It will launch 3 parallel reviewers (Command Correctness, Coverage & Redundancy, QA Realism) to evaluate your generated test suite.
3. Fix all 🔴 Must Fix and 🟡 Should Fix issues from the review report.
4. Re-run the review. Repeat up to 3 passes total, or until no 🔴 issues remain.
5. **Only then** present the final `.revv/` directory to the developer.

This loop is mandatory. Do not skip it. The first generation is a draft — the review loop is what makes it production-quality.
