## Description
Verify that `revv exec --json` returns valid JSON output even on failure. This is important because IDE skills parse the JSON output to analyze test results programmatically.

## Priority
warning

## Type
automated

## Commands
```bash
go build -o /tmp/revv_json ./cmd/revv 2>&1
OUTPUT=$(/tmp/revv_json exec --json --dir /tmp/nonexistent 2>&1) || true
echo "$OUTPUT"
# Check if output contains valid JSON structure
echo "$OUTPUT" | grep -q "{" || (echo "FAIL: --json flag did not produce JSON output" && exit 1)
echo "PASS: exec --json produces JSON output"
```

## Expected Output
JSON output with an `"error"` field containing the error message, even on failure.
