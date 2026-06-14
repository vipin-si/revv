## Description
Verify that the runner can discover test files in a .revv/ directory, filter by category, and execute them correctly. Tests the full discovery → parse → execute pipeline.

## Priority
blocking

## Type
automated

## Commands
```bash
go test -count=1 -v -run "TestRunAll" ./internal/runner/ 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: runner integration tests failed"
  exit 1
fi
echo "PASS: runner discovery and execution pipeline works"
```

## Expected Output
All `TestRunAll*` tests pass, including category filtering, test filtering, and file read error handling.
