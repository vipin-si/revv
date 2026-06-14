## Description
Verify that the test.md parser correctly handles all required sections (Description, Priority, Type, Commands, Expected Output) and properly extracts code blocks. The parser is the core of revv — if it breaks, all test execution breaks.

## Priority
blocking

## Type
automated

## Commands
```bash
go test -count=1 -v -run "TestParseTestMD" ./internal/runner/ 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: parser tests failed"
  exit 1
fi
echo "PASS: test.md parser works correctly"
```

## Expected Output
All `TestParseTestMD*` tests pass, including type normalization, code block extraction, and header hijacking protection.
