## Description
Verify that `make build` produces a working binary at `bin/revv` using the project's Makefile. This tests the standard build path that contributors use.

## Priority
blocking

## Type
automated

## Commands
```bash
make build 2>&1
test -x ./bin/revv || (echo "FAIL: binary not found at ./bin/revv after make build" && exit 1)
echo "PASS: make build produced binary at ./bin/revv"
```

## Expected Output
`make build` exits 0 and `bin/revv` is an executable file.
