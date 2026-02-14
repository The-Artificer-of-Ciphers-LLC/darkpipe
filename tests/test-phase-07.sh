#!/bin/bash
# Phase 7: Build System & Deployment - Integration Tests
# Tests that all Phase 7 artifacts are valid and functional
#
# Usage:
#   ./tests/test-phase-07.sh
#
# Requirements:
#   - Docker 27+ (for build tests)
#   - python3 (for YAML/XML validation)
#   - Go 1.24+ (for setup tool compilation)
#   - xmllint (for XML validation)

set -euo pipefail

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Helper functions
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

print_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

print_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((TESTS_SKIPPED++))
}

# Change to repo root
cd "$(dirname "$0")/.."
REPO_ROOT=$(pwd)

echo "======================================"
echo "Phase 7 Integration Test Suite"
echo "======================================"
echo ""

# Test 1: Dockerfiles exist
print_test "Checking Dockerfiles exist"
DOCKERFILES=(
    "cloud-relay/Dockerfile"
    "home-device/stalwart/Dockerfile"
    "home-device/maddy/Dockerfile"
    "home-device/postfix-dovecot/Dockerfile"
)

for dockerfile in "${DOCKERFILES[@]}"; do
    if [ -f "$dockerfile" ]; then
        print_pass "Found $dockerfile"
    else
        print_fail "Missing $dockerfile"
    fi
done

# Test 2: Dockerfiles have required labels
print_test "Checking Dockerfile OCI labels"
for dockerfile in "${DOCKERFILES[@]}"; do
    if grep -q "org.opencontainers.image.source" "$dockerfile" && \
       grep -q "org.opencontainers.image.version" "$dockerfile" && \
       grep -q "org.opencontainers.image.licenses" "$dockerfile"; then
        print_pass "$dockerfile has OCI labels"
    else
        print_fail "$dockerfile missing OCI labels"
    fi
done

# Test 3: Dockerfiles use TARGETARCH
print_test "Checking Dockerfiles use TARGETARCH for multi-arch builds"
for dockerfile in "${DOCKERFILES[@]}"; do
    if grep -q "ARG TARGETARCH" "$dockerfile"; then
        print_pass "$dockerfile uses TARGETARCH"
    else
        print_skip "$dockerfile doesn't use TARGETARCH (may not build Go code)"
    fi
done

# Test 4: Build Dockerfiles (if Docker available)
print_test "Building Dockerfiles"
if command -v docker &> /dev/null; then
    # Build cloud-relay
    if docker build --quiet --build-arg TARGETARCH=amd64 -f cloud-relay/Dockerfile . -t test-cloud-relay:latest > /dev/null 2>&1; then
        print_pass "cloud-relay Dockerfile builds successfully"
        # Check image size
        IMAGE_SIZE=$(docker image inspect test-cloud-relay:latest --format '{{.Size}}' | awk '{print int($1/1024/1024)}')
        if [ "$IMAGE_SIZE" -lt 100 ]; then
            print_pass "cloud-relay image size: ${IMAGE_SIZE}MB (under 100MB target)"
        else
            print_fail "cloud-relay image size: ${IMAGE_SIZE}MB (exceeds 100MB target)"
        fi
        docker rmi test-cloud-relay:latest > /dev/null 2>&1
    else
        print_fail "cloud-relay Dockerfile build failed"
    fi

    # Build home-device Dockerfiles (context is home-device/[server]/)
    for server in stalwart maddy postfix-dovecot; do
        if docker build --quiet -f "home-device/$server/Dockerfile" "home-device/$server/" -t "test-home-$server:latest" > /dev/null 2>&1; then
            print_pass "home-device/$server Dockerfile builds successfully"
            docker rmi "test-home-$server:latest" > /dev/null 2>&1
        else
            print_fail "home-device/$server Dockerfile build failed"
        fi
    done
else
    print_skip "Docker not available, skipping build tests"
fi

# Test 5: Entrypoints have setup detection
print_test "Checking entrypoints have setup detection"
ENTRYPOINTS=(
    "cloud-relay/entrypoint.sh"
    "home-device/stalwart/entrypoint-wrapper.sh"
    "home-device/maddy/entrypoint-wrapper.sh"
    "home-device/postfix-dovecot/entrypoint.sh"
)

for entrypoint in "${ENTRYPOINTS[@]}"; do
    if [ -f "$entrypoint" ]; then
        if grep -q ".darkpipe-configured" "$entrypoint" || grep -q "darkpipe-setup" "$entrypoint"; then
            print_pass "$entrypoint has setup detection"
        else
            print_fail "$entrypoint missing setup detection"
        fi
    else
        print_skip "$entrypoint not found"
    fi
done

# Test 6: Entrypoints have Docker secrets support
print_test "Checking entrypoints support Docker secrets (_FILE pattern)"
for entrypoint in "${ENTRYPOINTS[@]}"; do
    if [ -f "$entrypoint" ]; then
        if grep -q "_FILE" "$entrypoint"; then
            print_pass "$entrypoint supports Docker secrets"
        else
            print_fail "$entrypoint missing Docker secrets support"
        fi
    else
        print_skip "$entrypoint not found"
    fi
done

# Test 7: .dockerignore files exist
print_test "Checking .dockerignore files exist"
DOCKERIGNORES=(
    ".dockerignore"
    "cloud-relay/.dockerignore"
    "home-device/.dockerignore"
)

for dockerignore in "${DOCKERIGNORES[@]}"; do
    if [ -f "$dockerignore" ]; then
        if grep -q ".git" "$dockerignore" && grep -q "secrets" "$dockerignore"; then
            print_pass "$dockerignore exists and excludes .git and secrets"
        else
            print_fail "$dockerignore missing critical exclusions"
        fi
    else
        print_fail "$dockerignore not found"
    fi
done

# Test 8: GitHub Actions workflows are valid YAML
print_test "Validating GitHub Actions workflow YAML syntax"
WORKFLOWS=(
    ".github/workflows/build-custom.yml"
    ".github/workflows/build-prebuilt.yml"
    ".github/workflows/release.yml"
)

if command -v python3 &> /dev/null; then
    for workflow in "${WORKFLOWS[@]}"; do
        if [ -f "$workflow" ]; then
            if python3 -c "import yaml; yaml.safe_load(open('$workflow'))" 2>/dev/null; then
                print_pass "$workflow is valid YAML"
            else
                print_fail "$workflow has YAML syntax errors"
            fi
        else
            print_fail "$workflow not found"
        fi
    done
else
    print_skip "python3 not available, skipping YAML validation"
fi

# Test 9: Workflows have multi-arch support
print_test "Checking workflows build for linux/amd64 and linux/arm64"
for workflow in "${WORKFLOWS[@]}"; do
    if [ -f "$workflow" ]; then
        if grep -q "linux/amd64" "$workflow" && grep -q "linux/arm64" "$workflow"; then
            print_pass "$workflow builds multi-arch images"
        else
            print_fail "$workflow missing multi-arch platform support"
        fi
    else
        print_skip "$workflow not found"
    fi
done

# Test 10: Workflows publish to GHCR (not Docker Hub)
print_test "Checking workflows publish to GHCR only"
for workflow in "${WORKFLOWS[@]}"; do
    if [ -f "$workflow" ]; then
        if grep -q "ghcr.io" "$workflow" && ! grep -q "docker.io" "$workflow"; then
            print_pass "$workflow publishes to GHCR only"
        else
            print_fail "$workflow may be publishing to Docker Hub (should be GHCR only)"
        fi
    else
        print_skip "$workflow not found"
    fi
done

# Test 11: Setup tool compiles
print_test "Checking setup tool compiles"
if command -v go &> /dev/null; then
    if [ -f "deploy/setup/cmd/darkpipe-setup/main.go" ]; then
        cd deploy/setup
        if go build -o /tmp/darkpipe-setup-test ./cmd/darkpipe-setup/ > /dev/null 2>&1; then
            print_pass "Setup tool compiles successfully"
            rm -f /tmp/darkpipe-setup-test
        else
            print_fail "Setup tool compilation failed"
        fi
        cd "$REPO_ROOT"
    else
        print_skip "Setup tool source not found (will be created in Plan 07-02)"
    fi
else
    print_skip "Go not available, skipping setup tool compilation"
fi

# Test 12: Setup tool has required packages
print_test "Checking setup tool dependencies"
if [ -f "deploy/setup/go.mod" ]; then
    if grep -q "github.com/spf13/cobra" deploy/setup/go.mod && \
       grep -q "github.com/AlecAivazis/survey" deploy/setup/go.mod && \
       grep -q "github.com/miekg/dns" deploy/setup/go.mod; then
        print_pass "Setup tool has required dependencies (cobra, survey, dns)"
    else
        print_fail "Setup tool missing required dependencies"
    fi
else
    print_skip "Setup tool go.mod not found"
fi

# Test 13: TrueNAS template is valid YAML
print_test "Validating TrueNAS Scale templates"
if [ -f "deploy/templates/truenas-scale/app.yaml" ]; then
    if command -v python3 &> /dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('deploy/templates/truenas-scale/app.yaml'))" 2>/dev/null; then
            print_pass "TrueNAS app.yaml is valid YAML"
        else
            print_fail "TrueNAS app.yaml has YAML syntax errors"
        fi
    else
        print_skip "python3 not available, skipping YAML validation"
    fi

    # Check for GHCR image references
    if grep -q "ghcr.io/trek-e/darkpipe" deploy/templates/truenas-scale/app.yaml; then
        print_pass "TrueNAS template references GHCR images"
    else
        print_fail "TrueNAS template doesn't reference GHCR images"
    fi
else
    print_skip "TrueNAS template not found (will be created in Plan 07-03)"
fi

if [ -f "deploy/templates/truenas-scale/questions.yaml" ]; then
    if command -v python3 &> /dev/null; then
        if python3 -c "import yaml; yaml.safe_load(open('deploy/templates/truenas-scale/questions.yaml'))" 2>/dev/null; then
            print_pass "TrueNAS questions.yaml is valid YAML"
        else
            print_fail "TrueNAS questions.yaml has YAML syntax errors"
        fi
    else
        print_skip "python3 not available, skipping YAML validation"
    fi
else
    print_skip "TrueNAS questions.yaml not found"
fi

# Test 14: Unraid template is valid XML
print_test "Validating Unraid template"
if [ -f "deploy/templates/unraid/darkpipe.xml" ]; then
    if command -v xmllint &> /dev/null; then
        if xmllint --noout deploy/templates/unraid/darkpipe.xml 2>/dev/null; then
            print_pass "Unraid template is valid XML"
        else
            print_fail "Unraid template has XML syntax errors"
        fi
    else
        print_skip "xmllint not available, skipping XML validation"
    fi

    # Check for GHCR image references
    if grep -q "ghcr.io/trek-e/darkpipe" deploy/templates/unraid/darkpipe.xml; then
        print_pass "Unraid template references GHCR images"
    else
        print_fail "Unraid template doesn't reference GHCR images"
    fi
else
    print_skip "Unraid template not found (will be created in Plan 07-03)"
fi

# Test 15: Platform guides exist
print_test "Checking platform deployment guides exist"
GUIDES=(
    "deploy/platform-guides/raspberry-pi.md"
    "deploy/platform-guides/truenas-scale.md"
    "deploy/platform-guides/unraid.md"
    "deploy/platform-guides/proxmox-lxc.md"
    "deploy/platform-guides/synology-nas.md"
    "deploy/platform-guides/mac-silicon.md"
)

for guide in "${GUIDES[@]}"; do
    if [ -f "$guide" ]; then
        # Check for required sections
        if grep -q "Prerequisites" "$guide" && grep -q "Quick Start" "$guide"; then
            print_pass "$guide exists with Prerequisites and Quick Start sections"
        else
            print_fail "$guide missing required sections"
        fi
    else
        print_skip "$guide not found (will be created in Plan 07-03)"
    fi
done

# Test 16: Raspberry Pi guide mentions memory optimization
print_test "Checking RPi4 guide has memory optimization recommendations"
if [ -f "deploy/platform-guides/raspberry-pi.md" ]; then
    if grep -qi "4GB" deploy/platform-guides/raspberry-pi.md && \
       grep -qi "memory" deploy/platform-guides/raspberry-pi.md && \
       grep -qi "swap" deploy/platform-guides/raspberry-pi.md; then
        print_pass "RPi4 guide includes memory optimization recommendations"
    else
        print_fail "RPi4 guide missing memory optimization details"
    fi
else
    print_skip "RPi4 guide not found"
fi

# Test 17: TrueNAS guide mentions 24.10+ requirement
print_test "Checking TrueNAS guide mentions 24.10+ (Electric Eel) requirement"
if [ -f "deploy/platform-guides/truenas-scale.md" ]; then
    if grep -qi "24.10" deploy/platform-guides/truenas-scale.md && \
       grep -qi "Electric Eel" deploy/platform-guides/truenas-scale.md; then
        print_pass "TrueNAS guide mentions 24.10+ requirement"
    else
        print_fail "TrueNAS guide missing 24.10+ requirement"
    fi
else
    print_skip "TrueNAS guide not found"
fi

# Test 18: Unraid guide mentions Docker Compose for full stack
print_test "Checking Unraid guide mentions Docker Compose alternative"
if [ -f "deploy/platform-guides/unraid.md" ]; then
    if grep -qi "Docker Compose" deploy/platform-guides/unraid.md || \
       grep -qi "Compose Manager" deploy/platform-guides/unraid.md; then
        print_pass "Unraid guide mentions Docker Compose for full stack"
    else
        print_fail "Unraid guide missing Docker Compose information"
    fi
else
    print_skip "Unraid guide not found"
fi

echo ""
echo "======================================"
echo "Test Summary"
echo "======================================"
echo -e "${GREEN}Passed:${NC}  $TESTS_PASSED"
echo -e "${RED}Failed:${NC}  $TESTS_FAILED"
echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"
echo "======================================"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
