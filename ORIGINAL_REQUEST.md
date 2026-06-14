# Original User Request

## Initial Request — 2026-06-14T12:58:59-07:00

Redesign revv from a 2-skill architecture (revv-update, revv-run) into a 4-skill architecture (init-repo, update-repo, add-tests, run-tests). Each skill has a lightweight SKILL.md trigger file and a detailed prompt.md with the full instructions.

Working directory: /Users/vipinsingh/Documents/Antigravity/open source/revv
Integrity mode: development

## Requirements

### R1. Create 4 new skills with SKILL.md + prompt.md

Create the following skill directories, each containing a `SKILL.md` (YAML frontmatter with name + description, then a short body that says "Read `prompt.md` in this directory and follow the instructions") and a `prompt.md` (the detailed, lengthy prompt):

**Skill 1: `skills/init-repo/`**
- Trigger: "revv init", "set up revv", "initialize revv"
- Purpose: First-time setup for a repo with no `.revv/` directory
- Context the prompt must tell the LLM to gather:
  - All markdown files (README.md, AGENTS.md, CONTRIBUTING.md, etc.)
  - Code tree (`find . -type f` or `ls -R`, excluding node_modules, .git, etc.)
  - Key code files (build configs like Makefile, package.json, Cargo.toml, go.mod)
- Output: Generate `.revv/` directory with Dockerfile and test.md files organized by category
- The prompt must include the full test.md format specification with `## Type` (automated | browser | manual), `## Description`, `## Priority` (blocking | warning), `## Commands` or `## Steps`, `## Expected Output`
- The prompt must include Dockerfile generation rules (base image, install deps, COPY source, pre-build)
- The prompt must include AGENTS.md generation (append if exists, create if not, skip if revv section already present — search for "vssinghh/revv")
- The AGENTS.md pointer content must reference all 4 skills by their raw GitHub URLs: `https://raw.githubusercontent.com/vssinghh/revv/main/skills/{skill-name}/SKILL.md`

**Skill 2: `skills/update-repo/`**
- Trigger: "revv update", "refresh tests", "update test suite"
- Purpose: Periodic refresh of the test suite based on recent changes
- Context the prompt must tell the LLM to gather:
  - Existing `.revv/` directory (all test.md files, Dockerfile, helpers)
  - Last N commits: `git log --oneline -N` (default N=10)
  - All markdown files
  - Code tree
- Output: For each existing test — keep, update, or delete. Plus suggested new tests.
- The prompt must tell the LLM to preserve tests that are still relevant and only modify what's changed

**Skill 3: `skills/add-tests/`**
- Trigger: "revv add", "add tests", "test this PR", "test this commit"
- Purpose: Add tests specific to the current PR or commit
- Context the prompt must tell the LLM to gather:
  - Existing `.revv/` directory
  - Current PR/commit diff: `git diff HEAD~1` or `git diff main...HEAD`
  - All markdown files
  - Code tree
- Output: Decision on whether new tests are needed. If yes, create the test.md files and add them. If no, explain why existing coverage is sufficient.

**Skill 4: `skills/run-tests/`**
- Trigger: "revv run", "run tests", "test my changes"
- Purpose: Execute the full test suite
- Context the prompt must tell the LLM to gather:
  - Existing `.revv/` directory
  - Current PR/commit context
  - All markdown files, code tree
- Output: First call the add-tests skill to check for new tests. Then:
  - For `## Type: automated` tests: build the Go binary (`which revv || go build -o /tmp/revv ./cmd/revv`) and run `revv exec --verbose`. The binary runs these in parallel Docker containers.
  - For `## Type: browser` tests: execute via Chrome DevTools MCP tools (navigate_page, click, fill, screenshot, etc.)
  - For `## Type: manual` tests: print steps for human verification
  - Analyze any failures (explain why, suggest fix)
  - Report summary: passed/failed, blocking vs warning, coverage gaps

### R2. Delete old skills

Delete the following directories that are being replaced:
- `skills/revv-update/`
- `skills/revv-run/`

### R3. Update README.md

Update the README.md to reflect the new 4-skill architecture:
- Update the architecture diagram and table to show 4 skills instead of 2
- Update the "How It Works" flow diagram
- Update the "For Maintainers" section — the paste prompt should reference `init-repo` skill
- Keep the Test Format section, FAQ, and License as-is
- Keep all references to the Go binary (`revv exec`) for parallel Docker execution

### R4. Prompts must be comprehensive and high-quality

Each `prompt.md` must be a detailed, production-quality prompt. They should be long enough to fully specify the behavior (expect 100-200 lines each). They are the core product — the instructions that turn any IDE's LLM into a QA engineer. Do not write skeleton prompts with TODOs.

## Acceptance Criteria

### Skill Structure
- [ ] `skills/init-repo/SKILL.md` exists with valid YAML frontmatter (name, description)
- [ ] `skills/init-repo/prompt.md` exists and is at least 80 lines long
- [ ] `skills/update-repo/SKILL.md` exists with valid YAML frontmatter
- [ ] `skills/update-repo/prompt.md` exists and is at least 80 lines long
- [ ] `skills/add-tests/SKILL.md` exists with valid YAML frontmatter
- [ ] `skills/add-tests/prompt.md` exists and is at least 80 lines long
- [ ] `skills/run-tests/SKILL.md` exists with valid YAML frontmatter
- [ ] `skills/run-tests/prompt.md` exists and is at least 80 lines long
- [ ] `skills/revv-update/` directory does NOT exist (deleted)
- [ ] `skills/revv-run/` directory does NOT exist (deleted)

### Prompt Quality
- [ ] Each prompt.md includes a "Context" section listing exactly what the LLM should read
- [ ] Each prompt.md includes an "Output" section specifying exactly what to generate
- [ ] Each prompt.md includes a "Rules" section with constraints and gotchas
- [ ] init-repo prompt includes the full test.md format specification (## Type, ## Description, ## Priority, ## Commands/Steps, ## Expected Output)
- [ ] init-repo prompt includes AGENTS.md generation rules (append if exists, create if not, skip duplicates)
- [ ] run-tests prompt references the Go binary for automated tests and Chrome DevTools for browser tests
- [ ] add-tests prompt includes the decision logic for whether new tests are needed

### README
- [ ] README.md architecture section shows 4 skills (init-repo, update-repo, add-tests, run-tests)
- [ ] README.md maintainer section references init-repo skill
- [ ] README.md still references the Go binary for parallel Docker execution

### Build Verification
- [ ] `go build -o /dev/null ./cmd/revv` succeeds (binary unchanged)
- [ ] `go test ./...` passes (no Go code changes expected)
