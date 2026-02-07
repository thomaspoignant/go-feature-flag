# Missing Test Coverage Analysis

Based on codecov analysis, here are the missing tests that should be added:

## Critical Missing Tests (0% Coverage)

### 1. `addFlagset()` - 0.0% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:417`

**Missing test cases:**
- ✅ Success case: Add a new flagset successfully
- ❌ Error case: `NewGoFeatureFlagClient()` fails (line 420-424)
- ❌ Error case: `config.AddFlagSet()` fails (line 428-433)

**Test file:** `cmd/relayproxy/service/flagset_manager_test.go`

**Suggested test:**
```go
func TestFlagsetManager_addFlagset(t *testing.T) {
    // Test success case
    // Test NewGoFeatureFlagClient error
    // Test config.AddFlagSet error
}
```

### 2. `removeFlagset()` - 0.0% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:449`

**Missing test cases:**
- ✅ Success case: Remove a flagset successfully  
- ❌ Error case: Flagset not found in map (line 462-466)
- ❌ Error case: `GetFlagSetAPIKeys()` fails (line 452-456)
- ❌ Error case: `config.RemoveFlagSet()` fails (line 481-486)

**Test file:** `cmd/relayproxy/service/flagset_manager_test.go`

**Suggested test:**
```go
func TestFlagsetManager_removeFlagset(t *testing.T) {
    // Test success case
    // Test flagset not found
    // Test GetFlagSetAPIKeys error
    // Test config.RemoveFlagSet error
}
```

## Partial Coverage Tests

### 3. `processFlagsetChange()` - 50.0% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:346`

**Missing test cases:**
- ❌ Path when flagset doesn't exist (calls `addFlagset`) - line 348-350

**Note:** This is partially covered by integration tests, but unit test would be better.

### 4. `removeDeletedFlagsets()` - 66.7% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:364`

**Missing test cases:**
- ❌ Path when flagset needs to be removed (line 366-368)

**Note:** This is covered by integration tests, but unit test would improve coverage.

### 5. `processFlagsetAPIKeyChange()` - 69.2% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:373`

**Missing test cases:**
- ❌ Error case: `GetFlagSetAPIKeys()` returns error (line 378-380)
- ❌ Error case: `SetFlagSetAPIKeys()` returns error (line 389-391)

**Test file:** `cmd/relayproxy/service/flagset_manager_test.go`

### 6. `FlagSet()` - 87.5% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:185`

**Missing test cases:**
- ❌ Error case: Flagset name exists in APIKeysToFlagSetName but not in FlagSets map (line 198-200)

**Test file:** `cmd/relayproxy/service/flagset_manager_test.go`

### 7. `AllFlagSets()` - 83.3% coverage
**Location:** `cmd/relayproxy/service/flagset_manager.go:229`

**Missing test cases:**
- ❌ Error case: No flagsets configured (line 231-232)

**Test file:** `cmd/relayproxy/service/flagset_manager_test.go`

## Priority Recommendations

1. **High Priority:** Add unit tests for `addFlagset()` and `removeFlagset()` - these are critical new methods with 0% coverage
2. **Medium Priority:** Add error path tests for `processFlagsetAPIKeyChange()`, `FlagSet()`, and `AllFlagSets()`
3. **Low Priority:** Add unit tests for `processFlagsetChange()` and `removeDeletedFlagsets()` (already covered by integration tests)

## Test Strategy

The integration tests in `config_change_test.go` test the end-to-end functionality, but they don't test:
- Individual method error paths
- Edge cases that are hard to trigger through config file changes
- Direct method invocations

Unit tests should mock:
- `config.Config` methods (`AddFlagSet`, `RemoveFlagSet`, `GetFlagSetAPIKeys`, `SetFlagSetAPIKeys`)
- `NewGoFeatureFlagClient` (or use test fixtures)
- Logger (use `zaptest/observer` for log verification)
