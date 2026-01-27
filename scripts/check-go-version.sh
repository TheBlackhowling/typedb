#!/bin/bash
# Check that Go version meets minimum requirement (1.23.12+)

set -e

MIN_VERSION="1.23.12"
CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')

echo "Checking Go version..."
echo "Current: $CURRENT_VERSION"
echo "Required: $MIN_VERSION or later"

# Compare versions
compare_versions() {
    local version1=$1
    local version2=$2
    
    # Extract major.minor.patch
    IFS='.' read -ra V1 <<< "$version1"
    IFS='.' read -ra V2 <<< "$version2"
    
    # Compare major
    if [ "${V1[0]}" -lt "${V2[0]}" ]; then
        return 1
    elif [ "${V1[0]}" -gt "${V2[0]}" ]; then
        return 0
    fi
    
    # Compare minor
    if [ "${V1[1]}" -lt "${V2[1]}" ]; then
        return 1
    elif [ "${V1[1]}" -gt "${V2[1]}" ]; then
        return 0
    fi
    
    # Compare patch
    if [ "${V1[2]}" -lt "${V2[2]}" ]; then
        return 1
    fi
    
    return 0
}

if compare_versions "$CURRENT_VERSION" "$MIN_VERSION"; then
    echo "✅ Go version meets requirement"
    exit 0
else
    echo "❌ Go version $CURRENT_VERSION is below minimum requirement $MIN_VERSION"
    echo "Please upgrade to Go $MIN_VERSION or later to address security vulnerabilities:"
    echo "  - GO-2025-3849 (requires Go 1.23.12+)"
    echo "  - GO-2025-3750 (requires Go 1.23.10+)"
    exit 1
fi
