## Description
Verify that `revv exec` without a valid .revv/ directory returns a clear error message instead of crashing or hanging. Contributors should get an actionable error.

## Priority
warning

## Type
automated

## Commands
```bash
go build -o /tmp/revv_err ./cmd/revv 2>&1
OUTPUT=$(/tmp/revv_err exec --dir /tmp/nonexistent_revv_dir 2>&1) || true
echo "$OUTPUT"
echo "$OUTPUT" | grep -qi "error\|not found\|no such\|does not exist" || (echo "FAIL: missing directory did not produce clear error" && exit 1)
echo "PASS: exec with missing dir gives clear error"
```

## Expected Output
`revv exec` returns a non-zero exit code with a message indicating the directory doesn't exist.
