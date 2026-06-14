# Skill Prompt: Run Tests

This prompt guides the execution and evaluation of the repository's QA suite. It orchestrates test preparation, calls test-creation checks, and runs different types of tests using Go-based automation or browser automation.

## Context

When this skill is triggered, you must perform the following discovery steps:

1. **Locate Sandbox Configurations:**
   - Inspect the `.revv/` directory, verifying the presence of `.revv/Dockerfile`, `.revv/helpers/`, and tests in subfolders.
   - Verify the location of the codebase and key build scripts.

2. **Retrieve Commit Context & Pull Request Details:**
   - Determine current changes (e.g. uncommitted files, recent commit history) to understand the delta.

3. **Check for New Tests:**
   - Always call the `add-tests` skill first to determine if new tests need to be created or if any existing ones need updates before execution.

## Output

Perform the following execution flow:

1. **Pre-Execution Check:**
   - Run the `add-tests` skill. Follow its instructions to add or modify test definitions if local codebase changes warrant it.

2. **Categorize and Execute Tests:**
   - **Automated Tests** (`## Type: automated`): Run them using the `revv` command-line tool.
   - **Browser Tests** (`## Type: browser`): Execute them interactively using Chrome DevTools MCP tools. If Chrome DevTools is unavailable, print the steps for the user but still report the test as type `browser`.

3. **Summary & Diagnostics:**
   - Produce a detailed report summarizing passes, failures, duration, and blockers.
   - Analyze any failures by investigating execution logs or screenshots, and propose precise code or test fixes.

## Rules

### 1. Automated Test Execution Rules

- **Go Binary Compilation**:
  Before running automated tests, check if the `revv` binary is available in the path. If not, build it from the repository source and store it in `/tmp`:
  ```bash
  which revv || go build -o /tmp/revv ./cmd/revv
  ```
- **Execution Command**:
  Execute tests in parallel using the `revv` execution engine. Ensure `--verbose` is provided to capture logs:
  ```bash
  /tmp/revv exec --verbose
  ```
  *(Note: If the `revv` command is globally installed, you can run `revv exec --verbose`).*
- **Sandbox Environment**: All automated tests run within isolated Docker containers created based on the `.revv/Dockerfile` build definition. Do not attempt to run automated commands directly on the host system.

### 2. Browser Test Execution Rules

If any test is configured with `## Type: browser`, execute it directly via the agent's browser control tools (such as Chrome DevTools MCP plugins):

- **Setup Phase**: Look for a `## Setup` section. If present, execute the commands (such as starting a local development server or spinning up database containers) before launching the browser.
- **Interactions**:
  - Open a tab and navigate to the application using `new_page` or `navigate_page`.
  - Interact with elements using `click`, `fill`, or `type` selectors.
  - Verify page state using `get_text` or `evaluate_javascript`.
  - Take visual state captures using `screenshot` and embed or link them in the final test report.
- **Degraded/Fallback Mode**: If Chrome DevTools MCP tools are unavailable in the current environment, print the browser test steps for the user. The test type remains `browser` and should be marked as "browser - needs human verification".

### 3. Failure Analysis and Reporting

- **Summary Structure**:
  Present the results table at the top of your response:
  | Category | Test Name | Type | Status (Pass/Fail/Pending) | Priority (Blocking/Warning) |
  | --- | --- | --- | --- | --- |

- **Failure Diagnostics**:
  For each failed test:
  1. Retrieve and display the error log or stdout.
  2. Inspect the test commands and the modified code files.
  3. Determine the root cause: is it a code bug, a test configuration issue, or a broken dependency?
  4. Write a concrete recommendation/fix.

- **Severity Classification**:
  - If ANY `blocking` test fails → report overall status as **FAIL** and clearly state: "Blocking tests failed — do not merge."
  - If only `warning` tests fail → report overall status as **WARN** and list the issues.
  - If all tests pass → report overall status as **PASS**.

- **Browser Test Failures**:
  - Include a screenshot of the failure state when possible.
  - Note whether the failure is deterministic (same result on retry) or flaky.
  - If a browser test has a `## Script` section and it fails, fall back to `## Steps` and re-run via LLM interpretation.

- **Retry Policy**:
  - Do NOT retry automated tests — they are deterministic.
  - Browser tests may be retried once if the first attempt fails, to account for timing or rendering issues.

- **Final Output**:
  End your report with a clear verdict:
  ```
  ## Verdict
  [PASS | WARN | FAIL]
  [One-line summary of results]
  ```
