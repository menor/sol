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
# Or create a .env file with SMOKE_TEST_PROJECT and just run:
#   ./scripts/smoke-test.sh

# Don't use set -e - we want to continue on test failures

# Load .env file if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

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
    echo ""
    echo "Or create a .env file with:"
    echo "  SMOKE_TEST_PROJECT=your-project-id"
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
section "Deployment Commands"
# =============================================================================

# Test that deployment commands exist and parse correctly
if $SOL redeploy --help > /dev/null 2>&1; then
    pass "redeploy --help"
else
    fail "redeploy --help"
fi

if $SOL environment:activate --help > /dev/null 2>&1; then
    pass "environment:activate --help"
else
    fail "environment:activate --help"
fi

if $SOL environment:deactivate --help > /dev/null 2>&1; then
    pass "environment:deactivate --help"
else
    fail "environment:deactivate --help"
fi

if $SOL environment:delete --help > /dev/null 2>&1; then
    pass "environment:delete --help"
else
    fail "environment:delete --help"
fi

if $SOL push --help > /dev/null 2>&1; then
    pass "push --help"
else
    fail "push --help"
fi

# Test redeploy on main environment (safe operation)
if [ -n "$ENV" ]; then
    echo "Testing redeploy on $ENV (this may take a moment)..."
    REDEPLOY_OUTPUT=$($SOL redeploy --project "$PROJECT" --environment "$ENV" 2>&1)
    if echo "$REDEPLOY_OUTPUT" | grep -q '"type": "environment.redeploy"'; then
        ACTIVITY_ID=$(echo "$REDEPLOY_OUTPUT" | jq -r '.id // empty')
        pass "redeploy (activity: $ACTIVITY_ID)"
    else
        fail "redeploy"
        echo "  Output: $REDEPLOY_OUTPUT"
    fi
else
    skip "redeploy - no environment found"
fi

# Test activate/deactivate cycle on a non-main environment
# Look for smoke-test environment, create if it doesn't exist
TEST_ENV="smoke-test"
TEST_ENV_EXISTS=$($SOL environment:list --project "$PROJECT" 2>/dev/null | jq -r --arg env "$TEST_ENV" '[.[] | select(.id == $env)][0].id // empty')

if [ -z "$TEST_ENV_EXISTS" ]; then
    echo "Creating smoke-test environment (this may take a few minutes)..."
    if $SOL environment:branch "$TEST_ENV" --project "$PROJECT" --parent main --title "Smoke Test" --wait > /dev/null 2>&1; then
        pass "environment:branch (created $TEST_ENV)"
    else
        fail "environment:branch"
        skip "environment:activate - branch creation failed"
        skip "environment:deactivate - branch creation failed"
        TEST_ENV=""
    fi
else
    pass "environment:branch (smoke-test already exists)"
fi

if [ -n "$TEST_ENV" ]; then
    # Get current status
    ENV_STATUS=$($SOL environment:info "$TEST_ENV" --project "$PROJECT" 2>/dev/null | jq -r '.status // empty')

    if [ "$ENV_STATUS" = "active" ]; then
        # Deactivate first, then activate
        echo "Testing deactivate on $TEST_ENV..."
        if $SOL environment:deactivate "$TEST_ENV" --project "$PROJECT" --wait > /dev/null 2>&1; then
            pass "environment:deactivate ($TEST_ENV)"

            echo "Testing activate on $TEST_ENV..."
            if $SOL environment:activate "$TEST_ENV" --project "$PROJECT" --wait > /dev/null 2>&1; then
                pass "environment:activate ($TEST_ENV)"
            else
                fail "environment:activate ($TEST_ENV)"
            fi
        else
            fail "environment:deactivate ($TEST_ENV)"
            skip "environment:activate - deactivate failed"
        fi
    elif [ "$ENV_STATUS" = "inactive" ]; then
        # Activate first, then deactivate
        echo "Testing activate on $TEST_ENV..."
        if $SOL environment:activate "$TEST_ENV" --project "$PROJECT" --wait > /dev/null 2>&1; then
            pass "environment:activate ($TEST_ENV)"

            echo "Testing deactivate on $TEST_ENV..."
            if $SOL environment:deactivate "$TEST_ENV" --project "$PROJECT" --wait > /dev/null 2>&1; then
                pass "environment:deactivate ($TEST_ENV)"
            else
                fail "environment:deactivate ($TEST_ENV)"
            fi
        else
            fail "environment:activate ($TEST_ENV)"
            skip "environment:deactivate - activate failed"
        fi
    else
        skip "environment:activate/deactivate - environment status: $ENV_STATUS"
    fi
else
    skip "environment:activate - no non-production environment found"
    skip "environment:deactivate - no non-production environment found"
fi

# Skip destructive/complex tests
skip "environment:delete (not tested - destructive)"
skip "push (not tested - requires git repo setup)"

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
