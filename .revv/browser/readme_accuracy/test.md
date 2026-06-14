## Description
Verify that the README accurately describes revv's current architecture — 4 skills, Go binary, Docker execution. Outdated docs mislead contributors.

## Priority
warning

## Type
browser

## Steps
1. Open the README.md file in the repository root
2. Verify the architecture diagram mentions all 4 skills: init-repo, update-repo, add-tests, run-tests
3. Verify the "For Maintainers" section references the init-repo skill URL
4. Verify the "For Contributors" section explains the "revv run" flow
5. Verify the component table lists all 4 skills plus the Go binary (`revv exec`)
6. Take a screenshot of the architecture section for the report

## Expected Output
README accurately reflects the current 4-skill architecture with no references to old skill names (revv-update, revv-run).
