## Description
Verify that environment variable detection correctly finds $VAR and ${VAR} patterns in test commands, ignores shell builtins ($HOME, $PATH), and loads .env files with inline comment stripping.

## Priority
warning

## Type
automated

## Commands
```bash
go test -count=1 -v -run "TestDetectEnvVars\|TestLoadEnvFile\|TestStripInlineComment" ./internal/runner/ 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: env var detection tests failed"
  exit 1
fi
echo "PASS: environment variable detection works correctly"
```

## Expected Output
All env-related tests pass, correctly detecting custom vars while skipping builtins.
