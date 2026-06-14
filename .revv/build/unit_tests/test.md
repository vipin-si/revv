## Description
Verify that all Go unit and integration tests pass. This catches regressions in the core logic — parser, runner, sandbox, CLI, and git utilities.

## Priority
blocking

## Type
automated

## Commands
```bash
go test -count=1 -v ./internal/... 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: one or more Go tests failed"
  exit 1
fi
echo "PASS: all Go tests passed"
```

## Expected Output
All test packages report `ok` with zero failures.
