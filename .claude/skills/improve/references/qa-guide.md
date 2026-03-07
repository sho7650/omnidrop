# QA Phase Detailed Guide

## Parsing Output by QA Tool






### golangci-lint

**Command**:
```bash
golangci-lint run 2>&1
```

**Parsing rules**:
- Format: `filepath:line:col: message (linter-name)`
- Severity: all → MEDIUM (golangci-lint does not distinguish severity natively)

**Example output**:
```
main.go:15:2: SA1019: package "io/ioutil" is deprecated (staticcheck)
```




### go vet

**Command**:
```bash
go vet ./... 2>&1
```

**Parsing rules**:
- Format: `filepath:line:col: message`
- Severity: always HIGH

**Example output**:
```
./main.go:15:2: printf: fmt.Sprintf format %d has arg s of wrong type string
```

### go test (Unit Tests)

**Command**:
```bash
go test ./... 2>&1
```

Look for `--- FAIL:` lines to identify individual test failures.
Look for `panic:` to identify runtime panics.
The summary line shows `FAIL` with the package name for failing packages, or `ok` for passing packages.
Example: `FAIL omnidrop/internal/auth 0.123s` or `ok omnidrop/internal/services 0.045s`

### Severity Classification

- CRITICAL: Security vulnerability (injection, SSRF, auth bypass, etc.)
- HIGH: Likely bug, lack of type safety, test failure, compilation error
- MEDIUM: Coding convention violation, readability issue
- LOW: Style improvement suggestion, performance hint

## Code Review Checklist

- [ ] No unjustified use of unsafe types (any, unknown casts, etc.)
- [ ] Error handling: async operations have proper error handling
- [ ] External input is validated
- [ ] No hardcoded magic numbers or strings
- [ ] Functions are within 50 lines
- [ ] Files are within 300 lines
- [ ] No circular imports/dependencies
- [ ] All errors checked (no `_` for error returns)
- [ ] Context passed to long-running operations
- [ ] Input validation on all endpoints
- [ ] Proper auth middleware

## Issue Aggregation Template

```markdown
# Issues - Round N

**Date**: YYYY-MM-DD HH:MM
**Found**: X issues | **Severity**: CRITICAL=0, HIGH=0, MEDIUM=0, LOW=0
**Sources**: lint=a, typecheck=b, unit-test=c, e2e=d, review=e

## Issues

### [HIGH] Example issue title
- **File**: `path/to/file:line`
- **Source**: lint | typecheck | unit-test | e2e | review
- **Detail**: Description of the problem
- **Suggestion**: Proposed fix (if any)
- **Status**: open
```
