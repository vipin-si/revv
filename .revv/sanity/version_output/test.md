## Description
Verify that `revv version` prints version information including the commit hash. This confirms ldflags injection works correctly during build.

## Priority
blocking

## Type
automated

## Commands
```bash
go build -o /tmp/revv_ver ./cmd/revv 2>&1
OUTPUT=$(/tmp/revv_ver version 2>&1)
echo "$OUTPUT"
echo "$OUTPUT" | grep -q "revv" || (echo "FAIL: version output doesn't contain 'revv'" && exit 1)
echo "PASS: version command works"
```

## Expected Output
Output like `revv dev (commit: abc1234, built at: 2026-01-01T00:00:00Z)`.
