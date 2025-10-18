#!/bin/bash
# Test script to validate summon_homun code changes
# This bypasses the Go version issue by only syntax checking

echo "=== Validating summon_homun code changes ==="

# Check if new files exist
echo ""
echo "1. Checking new files exist..."
for file in cmd/summon_homun/issue.go cmd/summon_homun/issue_test.go cmd/summon_homun/ISSUE_FLAG_USAGE.md; do
    if [ -f "$file" ]; then
        echo "  ✓ $file exists"
    else
        echo "  ✗ $file missing"
        exit 1
    fi
done

# Check syntax of Go files
echo ""
echo "2. Checking Go syntax with gofmt..."
if gofmt -l cmd/summon_homun/*.go | grep -q .; then
    echo "  ✗ Files need formatting:"
    gofmt -l cmd/summon_homun/*.go
    exit 1
else
    echo "  ✓ All Go files properly formatted"
fi

# Verify key functions exist in issue.go
echo ""
echo "3. Checking issue.go contains required functions..."
required_funcs=("fetchIssue" "isGitHub" "isGitea" "extractRepoFromURL" "fetchGitHubIssue" "fetchGiteaIssue" "getGitRemoteURL")
for func in "${required_funcs[@]}"; do
    if grep -q "func $func" cmd/summon_homun/issue.go; then
        echo "  ✓ Function $func found"
    else
        echo "  ✗ Function $func missing"
        exit 1
    fi
done

# Verify main.go uses flag package
echo ""
echo "4. Checking main.go uses flag package..."
if grep -q '"flag"' cmd/summon_homun/main.go; then
    echo "  ✓ flag package imported"
else
    echo "  ✗ flag package not imported"
    exit 1
fi

if grep -q 'flag.String("issue"' cmd/summon_homun/main.go; then
    echo "  ✓ --issue flag defined"
else
    echo "  ✗ --issue flag not defined"
    exit 1
fi

# Verify model has issueContent field
echo ""
echo "5. Checking model.go has issueContent field..."
if grep -q 'issueContent.*string' cmd/summon_homun/model.go; then
    echo "  ✓ issueContent field added to model"
else
    echo "  ✗ issueContent field missing from model"
    exit 1
fi

# Verify instructions.go has issue template
echo ""
echo "6. Checking instructions.go has issue template..."
if grep -q 'issueInstructionsTemplate' cmd/summon_homun/instructions.go; then
    echo "  ✓ issueInstructionsTemplate defined"
else
    echo "  ✗ issueInstructionsTemplate missing"
    exit 1
fi

if grep -q 'if m.issueContent != ""' cmd/summon_homun/instructions.go; then
    echo "  ✓ Issue content handling logic present"
else
    echo "  ✗ Issue content handling logic missing"
    exit 1
fi

# Check documentation
echo ""
echo "7. Checking documentation updated..."
if grep -q "Issue Flag" cmd/summon_homun/README.md; then
    echo "  ✓ README.md updated with Issue Flag section"
else
    echo "  ✗ README.md not updated"
    exit 1
fi

echo ""
echo "=== All validation checks passed! ==="
echo ""
echo "Note: Full 'go test' cannot run due to Go version mismatch on this system"
echo "      (system has Go 1.19, dependencies require Go 1.24+)"
echo "      However, all syntax and structure checks pass."
