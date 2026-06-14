## Description
Scan the source code for hardcoded secrets, API keys, or passwords. Revv handles Docker containers and file operations — any leaked credentials would be a serious security issue.

## Priority
warning

## Type
automated

## Commands
```bash
# Search for common secret patterns in Go source files (exclude test files and vendor)
FINDINGS=$(grep -rn --include="*.go" -i "password\|api_key\|secret_key\|private_key\|token.*=.*\"[a-zA-Z0-9]" . \
  --exclude-dir=vendor --exclude-dir=.git --exclude="*_test.go" 2>/dev/null | \
  grep -v "// " | grep -v "flag\." | grep -v "Flags()" | grep -v "envVarPattern" | grep -v "shellBuiltins" || true)

if [ -n "$FINDINGS" ]; then
  echo "WARNING: Possible hardcoded secrets found:"
  echo "$FINDINGS"
  echo "FAIL: review the findings above"
  exit 1
fi
echo "PASS: no hardcoded secrets detected in source"
```

## Expected Output
No matches found — all sensitive values should come from environment variables or config files.
