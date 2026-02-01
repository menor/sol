#!/bin/bash
#
# Smoke test for Sol CLI
#
# Runs basic tests against a real Upsun project to verify functionality.
# Tests are idempotent - they clean up after themselves.
#
# Usage:
#   SMOKE_TEST_PROJECT=your-project-id ./scripts/smoke-test.sh
#
# Or set the project ID in your environment and just run:
#   ./scripts/smoke-test.sh

# Don't use set -e - we want to continue on test failures

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration - require project ID
if [ -z "$SMOKE_TEST_PROJECT" ]; then
    echo "Error: SMOKE_TEST_PROJECT environment variable is required."
    echo ""
    echo "Usage:"
    echo "  SMOKE_TEST_PROJECT=your-project-id ./scripts/smoke-test.sh"
    exit 1
fi
PROJECT="$SMOKE_TEST_PROJECT"
SOL="./sol"
TEST_VAR_NAME="SOL_SMOKE_TEST_VAR"
PASSED=0
FAILED=0

# Helper functions
pass() {
    echo -e "${GREEN}PASS${NC}: $1"
    PASSED=$((PASSED + 1))
}

fail() {
    echo -e "${RED}FAIL${NC}: $1"
    FAILED=$((FAILED + 1))
}

skip() {
    echo -e "${YELLOW}SKIP${NC}: $1"
}

section() {
    echo ""
    echo "=== $1 ==="
}

# Check dependencies
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed."
    echo "Install with: brew install jq"
    exit 1
fi

# Check if sol binary exists
if [ ! -f "$SOL" ]; then
    echo "Building sol..."
    go build -o sol .
fi

echo "Sol Smoke Test"
echo "Project: $PROJECT"
echo ""

# =============================================================================
section "Authentication"
# =============================================================================

if $SOL auth:info | grep -q '"authenticated": true'; then
    pass "auth:info - authenticated"
else
    echo ""
    echo "Not authenticated. Please run:"
    echo "  $SOL auth:login"
    echo ""
    echo "Then re-run this script."
    exit 1
fi

# =============================================================================
section "Projects"
# =============================================================================

if $SOL project:list > /dev/null 2>&1; then
    pass "project:list"
else
    fail "project:list"
fi

if $SOL project:info "$PROJECT" > /dev/null 2>&1; then
    pass "project:info"
else
    fail "project:info"
fi

# =============================================================================
section "Environments"
# =============================================================================

if $SOL environment:list --project "$PROJECT" > /dev/null 2>&1; then
    pass "environment:list"
else
    fail "environment:list"
fi

# Get first environment for further tests
ENV=$($SOL environment:list --project "$PROJECT" 2>/dev/null | jq -r '.[0].id // empty')
if [ -n "$ENV" ]; then
    if $SOL environment:info "$ENV" --project "$PROJECT" > /dev/null 2>&1; then
        pass "environment:info ($ENV)"
    else
        fail "environment:info"
    fi
else
    skip "environment:info - no environments found"
fi

# =============================================================================
section "Activities"
# =============================================================================

if $SOL activity:list --project "$PROJECT" --limit 3 > /dev/null 2>&1; then
    pass "activity:list"
else
    fail "activity:list"
fi

if $SOL activity:list --project "$PROJECT" --state complete --limit 1 > /dev/null 2>&1; then
    pass "activity:list --state filter"
else
    fail "activity:list --state filter"
fi

# Get an activity ID for log test
ACTIVITY_ID=$($SOL activity:list --project "$PROJECT" --limit 1 2>/dev/null | jq -r '.[0].id // empty')
if [ -n "$ACTIVITY_ID" ]; then
    if $SOL activity:log "$ACTIVITY_ID" --project "$PROJECT" > /dev/null 2>&1; then
        pass "activity:log"
    else
        fail "activity:log"
    fi
else
    skip "activity:log - no activities found"
fi

# =============================================================================
section "Variables (Project Level)"
# =============================================================================

# Clean up any leftover test variable first (idempotency)
$SOL variable:delete "$TEST_VAR_NAME" --project "$PROJECT" --level project > /dev/null 2>&1 || true

if $SOL variable:list --project "$PROJECT" --level project > /dev/null 2>&1; then
    pass "variable:list (project)"
else
    fail "variable:list (project)"
fi

# Create test variable
if $SOL variable:set "$TEST_VAR_NAME" "smoke-test-value" --project "$PROJECT" --level project > /dev/null 2>&1; then
    pass "variable:set (create)"
else
    fail "variable:set (create)"
fi

# Read it back
if $SOL variable:get "$TEST_VAR_NAME" --project "$PROJECT" --level project 2>/dev/null | grep -q "smoke-test-value"; then
    pass "variable:get"
else
    fail "variable:get"
fi

# Update it
if $SOL variable:set "$TEST_VAR_NAME" "updated-value" --project "$PROJECT" --level project > /dev/null 2>&1; then
    pass "variable:set (update)"
else
    fail "variable:set (update)"
fi

# Delete it (cleanup)
if $SOL variable:delete "$TEST_VAR_NAME" --project "$PROJECT" --level project > /dev/null 2>&1; then
    pass "variable:delete"
else
    fail "variable:delete"
fi

# Verify deletion
if $SOL variable:get "$TEST_VAR_NAME" --project "$PROJECT" --level project > /dev/null 2>&1; then
    fail "variable:delete (verify) - variable still exists"
else
    pass "variable:delete (verify)"
fi

# =============================================================================
section "Results"
# =============================================================================

echo ""
TOTAL=$((PASSED + FAILED))
echo -e "Passed: ${GREEN}$PASSED${NC}/$TOTAL"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}/$TOTAL"
    exit 1
else
    echo -e "${GREEN}All smoke tests passed!${NC}"
    exit 0
fi
