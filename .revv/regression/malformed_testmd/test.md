## Description
Verify that the parser handles malformed test.md files gracefully — missing sections, empty files, code blocks inside headers, and unknown type values. The parser should never crash on bad input.

## Priority
warning

## Type
automated

## Commands
```bash
go test -count=1 -v -run "TestParseTestMD_HeaderHijacking\|TestNormalizeType\|TestParseTestMD_TypeNormalization" ./internal/runner/ 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: parser edge case tests failed"
  exit 1
fi
echo "PASS: parser handles malformed input gracefully"
```

## Expected Output
All edge case tests pass — unknown types default to "automated", code blocks inside headers don't create fake sections, empty files don't crash.
