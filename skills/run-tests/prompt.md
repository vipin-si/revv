# Skill Prompt: Run Tests

You are executing the QA test suite for a repository. Your job is to run every test, analyze every failure, and deliver a verdict that tells the developer exactly what's wrong and how to fix it.

## Step 1: Gather Context

Before running anything, understand what you're working with:

1. **Verify `.revv/` exists.** If it doesn't, tell the user to run `revv init` first and stop.

2. **Inventory the tests:**
   ```bash
   find .revv -name "test.md" | sort
   ```
   Read each test.md and categorize:
   - `automated` tests → will be run via the Go binary in Docker
   - `browser` tests → will be run via Chrome DevTools MCP

3. **Check current changes:**
   ```bash
   git status --short
   git log --oneline -5
   ```
   This tells you what the developer changed — useful for failure analysis later.

4. **Run `add-tests` first.**
   Before executing, read the [add-tests skill](https://raw.githubusercontent.com/vssinghh/revv/main/skills/add-tests/SKILL.md) and follow its instructions. This checks if the current changes need new tests. If it generates new tests, include them in this run.

## Execution Flow — Two Phases

Running tests happens in TWO mandatory phases. You MUST complete both:

1. **Phase 1 — Automated tests**: Run `revv exec` to execute all `automated` tests in Docker containers. The binary handles parallelism, isolation, and result collection.
2. **Phase 2 — Browser tests**: After `revv exec` finishes, YOU must run every `browser` test yourself via Chrome DevTools MCP. The binary CANNOT run these — it skips them. If you stop after Phase 1, browser tests are untested.

**Do not report results until both phases are complete.**

---

## Step 2: Phase 1 — Run Automated Tests (via Go binary)

Automated tests (`## Type: automated`) run inside Docker containers via the Go binary.

### Build the binary

```bash
# Check if revv is already available
which revv 2>/dev/null
if [ $? -ne 0 ]; then
  # Build from source
  go build -o /tmp/revv ./cmd/revv
  export PATH="/tmp:$PATH"
fi
```

### Execute

```bash
revv exec --verbose
```

Key flags:
- `--verbose` — see each test's output as it runs
- `--json` — get structured output for programmatic analysis
- `--category <name>` — run only tests in a specific category (e.g., `--category build`)
- `--test <category/name>` — run a single test (e.g., `--test build/compile_check`)
- `--timeout <duration>` — override the default 5-minute timeout (e.g., `--timeout 10m`)

### What happens under the hood

1. The binary reads `.revv/Dockerfile` and builds a Docker image
2. It discovers all `test.md` files where `## Type` is `automated`
3. It spins up parallel Docker containers — one per test
4. Each container runs the `## Commands` from the test.md
5. Exit code 0 = pass, non-zero = fail
6. Results are collected and returned
7. **Browser tests are SKIPPED** — they show as `SKIP (no commands)` in the output

### Important

- **Do NOT run automated test commands directly on the host.** They run inside Docker.
- If Docker is not running, tell the user: "Docker is required for automated tests. Please start Docker Desktop and try again."
- If the Dockerfile build fails, report it as a blocking failure with the build log.

---

## Step 3: Phase 2 — Run Browser Tests (via Chrome DevTools MCP)

**This step is MANDATORY.** The Go binary skipped all browser tests. Now you must run them.

Check the `revv exec` output for lines like:
```
─ browser/readme_accuracy    warning    SKIP   (no commands)
```
Each of those is a browser test you must execute now.

Browser tests (`## Type: browser`) run via Chrome DevTools MCP tools directly in the IDE.

### For each browser test:

1. **Check for `## Setup` section.** If present, run the setup commands first:
   ```bash
   # Example: start the dev server
   npm start &
   sleep 3
   ```
   Wait for the server to be ready before proceeding.

2. **Check for `## Script` section.** If present, run the script directly — it's a deterministic Playwright/automation script generated from a previous successful run. This is faster and more reliable than AI interpretation.

3. **If no `## Script`, use `## Steps`.** Read each step and execute via Chrome DevTools MCP tools:

   | Step instruction | Chrome DevTools tool |
   |---|---|
   | "Open/navigate to URL" | `navigate_page` or `new_page` |
   | "Click element" | `click` with CSS/text selector |
   | "Type/enter text" | `fill` with selector and value |
   | "Verify text exists" | `get_text` + check content |
   | "Check page title" | `evaluate_javascript('document.title')` |
   | "Verify no console errors" | `evaluate_javascript('window.__consoleErrors')` |
   | "Take screenshot" | `screenshot` — embed in report |
   | "Check element visible" | `evaluate_javascript('!!document.querySelector("...")') ` |

4. **After successful `## Steps` execution**, generate a `## Script` section and write it back to the test.md file. This makes future runs deterministic. The script should use the exact selectors and values discovered during this run.

5. **If `## Script` fails**, delete the `## Script` section and fall back to `## Steps`. Re-run via AI interpretation. If the steps succeed, generate a new `## Script`.

### Fallback

If Chrome DevTools MCP tools are not available in this environment:
- Print the test steps for the user to follow manually
- Mark the test as "browser - needs human verification"
- Do NOT count it as a failure

## Step 4: Analyze Failures

For EVERY failed test, provide:

### 1. What failed
```
Test: build/compile_check
Type: automated
Priority: blocking
```

### 2. Error output
Show the actual error — the raw stdout/stderr from the test execution. Don't summarize, show the real output.

### 3. Root cause analysis
Determine WHY it failed:

| Cause | How to identify | Example |
|---|---|---|
| **Code bug** | Test was passing before, fails after recent changes | New function has a nil pointer |
| **Test misconfiguration** | Test commands are wrong for this project | test.md references `make build` but project uses `go build` |
| **Missing dependency** | Import or tool not available | `gcc` not installed in Dockerfile |
| **Flaky test** | Browser test fails intermittently | Timing issue with page load |
| **Environment issue** | Docker not running, port in use | "Cannot connect to Docker daemon" |

### 4. Recommended fix
Be specific. Don't say "fix the test." Say exactly what to change:

```diff
- RUN go build ./...
+ RUN CGO_ENABLED=0 go build ./...
```

Or if it's a code bug:
```
The function ParseTestMD on line 42 of parser.go doesn't handle
the case where ## Commands is empty. Add a nil check before
calling extractCodeBlock.
```

## Step 5: Report Results

### Results table (always show this first)

```markdown
| Category | Test Name | Type | Status | Priority |
|----------|-----------|------|--------|----------|
| build | compile_check | automated | ✅ Pass | blocking |
| build | unit_tests | automated | ❌ Fail | blocking |
| sanity | cli_help | automated | ✅ Pass | blocking |
| browser | readme_accuracy | browser | ⏳ Pending | warning |
```

Status icons:
- ✅ Pass
- ❌ Fail
- ⏳ Pending (browser test needs human verification)
- 🔄 Retried (browser test passed on retry)

### Summary line

```
Results: 12 passed, 1 failed, 1 pending (14 total)
Blocking: 7/7 passed | Warning: 5/6 passed, 1 pending
```

### Verdict

The final line of your report must be one of:

```
## Verdict: PASS ✅
All tests passed. Safe to merge.
```

```
## Verdict: WARN ⚠️
All blocking tests passed. 2 warning tests failed — review recommended but not required.
```

```
## Verdict: FAIL ❌
1 blocking test failed. Do not merge until fixed.
[List the failing blocking tests]
```

### Retry policy

- **Automated tests**: NEVER retry. They are deterministic — if they fail, it's a real failure.
- **Browser tests**: Retry ONCE if the first attempt fails (timing/rendering issues). If it fails again, it's a real failure.

## Edge Cases

- **No tests found**: If `.revv/` exists but has no test.md files, tell the user: "No tests found. Run `revv init` to generate tests."
- **All tests skipped**: If all tests are `browser` type and Chrome DevTools is unavailable, report all as pending and suggest the user run in an IDE with browser tools.
- **Docker build fails**: Report as a blocking failure. Include the Dockerfile and build log. This usually means the Dockerfile needs updating for new dependencies.
- **Test timeout**: If a test runs longer than the timeout (default 5 minutes), kill it and report as failed with "timed out after 5m".
- **Partial failure**: If some tests pass and some fail, still report all results. Don't stop at the first failure.
