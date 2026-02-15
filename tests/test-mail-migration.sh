#!/bin/bash
# Phase 10: Mail Migration - Integration Test Suite
#
# Validates all Phase 10 success criteria:
# SC1: IMAP migration with folder/flag/date preservation
# SC2: Contact and calendar import
# SC3: MailCow API migration
# SC4: CLI wizard with dry-run and progress bars

set -e

TESTS_PASSED=0
TESTS_FAILED=0
START_TIME=$(date +%s)

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_section() {
    echo ""
    echo -e "${YELLOW}=== $1 ===${NC}"
}

# Change to setup directory
cd "$(dirname "$0")/../deploy/setup" || exit 1

log_section "Prerequisites Check"

# Check for Go compiler
if ! command -v go &> /dev/null; then
    log_fail "Go compiler not found"
    exit 1
else
    log_success "Go compiler found: $(go version)"
fi

# Check for darkpipe-setup source
if [ ! -f "cmd/darkpipe-setup/main.go" ]; then
    log_fail "darkpipe-setup source not found"
    exit 1
else
    log_success "darkpipe-setup source found"
fi

log_section "SC1: IMAP Migration with Folder/Flag/Date Preservation"

log_info "Building darkpipe-setup binary..."
if go build -o /tmp/darkpipe-setup ./cmd/darkpipe-setup/; then
    log_success "darkpipe-setup binary compiles"
else
    log_fail "darkpipe-setup binary compilation failed"
fi

log_info "Testing migrate command registration..."
if /tmp/darkpipe-setup migrate --help &> /dev/null; then
    log_success "migrate --help shows command documentation"
else
    log_fail "migrate --help failed"
fi

log_info "Testing provider listing..."
OUTPUT=$(/tmp/darkpipe-setup migrate 2>&1)
SUPPORTED_PROVIDERS=("gmail" "outlook" "icloud" "mailcow" "mailu" "docker-mailserver" "generic")
ALL_FOUND=true
for provider in "${SUPPORTED_PROVIDERS[@]}"; do
    if echo "$OUTPUT" | grep -q "$provider"; then
        log_success "Provider '$provider' listed"
    else
        log_fail "Provider '$provider' not found in listing"
        ALL_FOUND=false
    fi
done

log_info "Testing nonexistent provider error handling..."
if /tmp/darkpipe-setup migrate --from nonexistent 2>&1 | grep -q "not found"; then
    log_success "nonexistent provider returns error"
else
    log_fail "nonexistent provider should return error"
fi

log_info "Testing dry-run behavior..."
# Note: Cannot test actual migration without credentials, but we can verify --apply flag exists
if /tmp/darkpipe-setup migrate --help | grep -q "\-\-apply"; then
    log_success "--apply flag recognized for migration execution"
else
    log_fail "--apply flag not found"
fi

log_info "Running IMAP sync unit tests..."
if go test ./pkg/mailmigrate/... -run TestIMAP -count=1 &> /dev/null; then
    log_success "IMAP sync tests pass"
else
    log_fail "IMAP sync tests failed"
fi

log_info "Running folder mapping tests..."
if go test ./pkg/mailmigrate/... -run TestMapping -count=1 &> /dev/null; then
    log_success "Folder mapping tests pass"
else
    log_fail "Folder mapping tests failed"
fi

log_info "Running state resume tests..."
if go test ./pkg/mailmigrate/... -run TestState -count=1 &> /dev/null; then
    log_success "State resume tests pass"
else
    log_fail "State resume tests failed"
fi

log_section "SC2: Contact and Calendar Import"

log_info "Running CalDAV sync tests..."
if go test ./pkg/mailmigrate/... -run TestCalDAV -count=1 &> /dev/null; then
    log_success "CalDAV sync tests pass"
else
    log_fail "CalDAV sync tests failed"
fi

log_info "Running CardDAV sync tests..."
if go test ./pkg/mailmigrate/... -run TestCardDAV -count=1 &> /dev/null; then
    log_success "CardDAV sync tests pass"
else
    log_fail "CardDAV sync tests failed"
fi

log_info "Running VCF import tests..."
if go test ./pkg/mailmigrate/... -run TestImportVCF -count=1 &> /dev/null; then
    log_success "VCF import tests pass"
else
    log_fail "VCF import tests failed"
fi

log_info "Running ICS import tests..."
if go test ./pkg/mailmigrate/... -run TestImportICS -count=1 &> /dev/null; then
    log_success "ICS import tests pass"
else
    log_fail "ICS import tests failed"
fi

log_info "Running contact merge tests..."
if go test ./pkg/mailmigrate/... -run TestMerge -count=1 &> /dev/null; then
    log_success "Contact merge tests pass"
else
    log_fail "Contact merge tests failed"
fi

log_section "SC3: MailCow API Migration"

log_info "Running MailCow provider tests..."
if go test ./pkg/providers/... -run TestMailCow -count=1 &> /dev/null; then
    log_success "MailCow provider tests pass"
else
    log_fail "MailCow provider tests failed"
fi

log_info "Running provider registry tests..."
if go test ./pkg/providers/... -run TestRegistry -count=1 &> /dev/null; then
    log_success "Provider registry tests pass (mailcow registered)"
else
    log_fail "Provider registry tests failed"
fi

log_section "SC4: CLI Wizard with Dry-Run and Progress Bars"

log_info "Testing migrate command flags..."
FLAGS=(
    "--from"
    "--apply"
    "--folder-map"
    "--labels-as-folders"
    "--contacts-mode"
    "--vcf-file"
    "--ics-file"
    "--batch-size"
    "--state-file"
)

for flag in "${FLAGS[@]}"; do
    if /tmp/darkpipe-setup migrate --help | grep -q -- "$flag"; then
        log_success "Flag $flag recognized"
    else
        log_fail "Flag $flag not found"
    fi
done

log_info "Running all provider tests..."
if go test ./pkg/providers/... -count=1 &> /dev/null; then
    log_success "All provider tests pass"
else
    log_fail "Provider tests failed"
fi

log_info "Running full test suite..."
if go test ./... -count=1 &> /dev/null; then
    log_success "Full test suite passes"
else
    log_fail "Full test suite failed"
fi

log_section "Summary"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

echo ""
echo -e "${BLUE}Tests Passed:${NC} ${GREEN}$TESTS_PASSED${NC}"
echo -e "${BLUE}Tests Failed:${NC} ${RED}$TESTS_FAILED${NC}"
echo -e "${BLUE}Total Time:${NC} ${DURATION}s"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}Phase 10 test suite: ALL TESTS PASSED${NC}"
    exit 0
else
    echo -e "${RED}Phase 10 test suite: SOME TESTS FAILED${NC}"
    exit 1
fi
