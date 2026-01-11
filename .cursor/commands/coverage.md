# Get Code Coverage

Generates a code coverage report for the typedb package.

## What it does

1. Runs all tests with coverage profiling enabled
2. Generates a coverage profile file (`coverage.out`)
3. Displays coverage summary by function
4. Shows overall coverage percentage

## Command

```bash
go test -coverprofile='coverage.out' -covermode=atomic -v
go tool cover -func coverage.out
```

## Notes

- The coverage profile file (`coverage.out`) is ignored by `.gitignore` and should not be committed
- Coverage is calculated using atomic mode for accurate concurrent test coverage
- Use quotes around `'coverage.out'` in PowerShell to prevent path parsing issues
- To generate an HTML coverage report: `go tool cover -html=coverage.out -o coverage.html`

## Example Output

```
github.com/TheBlackHowling/typedb/query.go:14:		QueryAll			100.0%
github.com/TheBlackHowling/typedb/query.go:46:		QueryFirst			100.0%
github.com/TheBlackHowling/typedb/query.go:77:		QueryOne			100.0%
...
total:							(statements)			77.2%
```
