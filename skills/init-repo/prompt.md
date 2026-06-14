# Skill Prompt: Init Repo

This prompt guides the initialization of `revv` for a code repository. It analyzes the existing codebase, detects configurations, and builds a standard testing sandbox structure.

## Context

When this skill is triggered, you must perform the following discovery steps:

1. **Locate Build Configurations:** Search the repository root and subdirectories for key build and dependency configurations:
   - Go: `go.mod`
   - Rust: `Cargo.toml`
   - Node.js: `package.json`
   - Python: `requirements.txt`, `pyproject.toml`, `setup.py`
   - Makefile: `Makefile`
   - Docker: `Dockerfile`, `docker-compose.yml`

2. **Map the Code Tree:**
   - Execute file listing commands (e.g., `find . -maxdepth 3` or `ls -R`) to understand the layout.
   - List key directories: source code directories (`src/`, `cmd/`, `internal/`, `pkg/`, `lib/`), test directories (`tests/`, `test/`), and documentation.

3. **Read Documentation:**
   - Read `README.md`, `CONTRIBUTING.md`, or any setup guides to understand how to build and test the project.

4. **Verify Existing Configuration:**
   - Check if a `.revv/` directory or `AGENTS.md` already exists in the repository.

## Output

Generate the following structure in the repository root:

1. **`.revv/` Directory:**
   - **`Dockerfile`**: A Docker container definition customized for this project's build and test dependencies.
   - **`test.md` files**: Initial basic tests placed in structured categories (e.g., `.revv/build/compile_check/test.md`, `.revv/unit/basic_tests/test.md`, `.revv/manual/smoke_test/test.md`).

2. **`AGENTS.md` File:**
   - Provide integration instructions for AI agents, pointing to the 4 skills required for repository QA operations.

## Rules

### 1. Test MD Format Specification

Every `test.md` file created inside `.revv/` must follow this exact markdown structure:

- **`## Description`**: A clear explanation of what the test verifies, why it is important, and what failure signifies.
- **`## Priority`**: Must be either `blocking` (for critical paths, gating releases) or `warning` (advisory, non-gating).
- **`## Type`**: Must be one of `automated`, `browser`, or `manual`.
- **`## Commands`** (Required if Type is `automated`): A block of actual, executable shell commands (e.g., in a `bash` markdown block) that return exit code `0` on success, non-zero on failure.
- **`## Steps`** (Required if Type is `browser` or `manual`): A numbered list of human-readable steps to reproduce the test.
- **`## Expected Output`**: A description of what the successful result looks like.

Example for automated test:
```markdown
## Description
Verify that the project compiles successfully.

## Priority
blocking

## Type
automated

## Commands
```bash
make build
```

## Expected Output
Build finishes successfully with exit code 0.
```

### 2. Dockerfile Generation Rules

The generated `.revv/Dockerfile` must satisfy the following:

- **Base Image**: Select an appropriate, minimal official base image (e.g., `golang:1.22-alpine`, `node:20-alpine`, `python:3.11-slim`) based on the detected language and project build configuration.
- **System Dependencies**: Install necessary compilation or execution tools (e.g., `make`, `git`, `gcc`, `libc-dev`, `g++`).
- **Working Directory**: Must be `/workspace`.
- **Copy Source**: Copy the repository source code into the `/workspace` directory in the container.
- **Pre-Build / RUN Command**: Execute the project's build/compilation step (e.g., `RUN go build -v ./...` or `RUN npm install && npm run build`) during image creation. This caches compilation artifacts so that running tests in container sandboxes is instantaneous.

### 3. AGENTS.md Generation Rules

- **Append or Create**:
  - If `AGENTS.md` does not exist, create it.
  - If `AGENTS.md` exists, **append** the revv section to the end of the file.
- **Deduplication**:
  - Search `AGENTS.md` for "vssinghh/revv". If this string is already present, skip appending or generating the revv section to avoid duplication.
- **Pointer Content**:
  - The `AGENTS.md` section for revv must reference all 4 skills by their exact raw GitHub URLs:
    - **init-repo**: `https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md`
    - **update-repo**: `https://raw.githubusercontent.com/vssinghh/revv/main/skills/update-repo/SKILL.md`
    - **add-tests**: `https://raw.githubusercontent.com/vssinghh/revv/main/skills/add-tests/SKILL.md`
    - **run-tests**: `https://raw.githubusercontent.com/vssinghh/revv/main/skills/run-tests/SKILL.md`

Example portion to add to `AGENTS.md`:
```markdown
---

## revv — Automated QA

This repository uses [revv](https://github.com/vssinghh/revv) for automated QA testing.
Tests are defined in `.revv/` and run within Docker containers or through browser automation.

When executing QA workflows, refer to the following skills:
- **revv init**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/init-repo/SKILL.md
- **revv update**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/update-repo/SKILL.md
- **revv add-tests**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/add-tests/SKILL.md
- **revv run**: https://raw.githubusercontent.com/vssinghh/revv/main/skills/run-tests/SKILL.md
```
