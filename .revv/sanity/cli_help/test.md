## Description
Verify that the compiled revv binary prints help output and has the expected subcommands (exec, version). A CLI that can't show help is broken.

## Priority
blocking

## Type
automated

## Commands
```bash
go build -o /tmp/revv_sanity ./cmd/revv 2>&1
/tmp/revv_sanity --help 2>&1 | grep -q "revv" || (echo "FAIL: --help does not contain 'revv'" && exit 1)
/tmp/revv_sanity --help 2>&1 | grep -q "exec" || (echo "FAIL: --help does not mention 'exec' subcommand" && exit 1)
/tmp/revv_sanity --help 2>&1 | grep -q "version" || (echo "FAIL: --help does not mention 'version' subcommand" && exit 1)
echo "PASS: CLI help output contains expected commands"
```

## Expected Output
Help output includes "revv", "exec", and "version" subcommands.
