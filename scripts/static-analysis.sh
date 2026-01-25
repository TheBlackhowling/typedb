#!/bin/bash
# Static analysis script for typedb
# Run all static analysis tools before 1.0.0 release

set -e

echo "🔍 Running static analysis for typedb..."
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0

# Check if tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo -e "${YELLOW}⚠️  $1 not found. Install with: go install $2${NC}"
        return 1
    fi
    return 0
}

# 1. go vet (built-in)
echo "1️⃣  Running go vet..."
if go vet ./...; then
    echo -e "${GREEN}✅ go vet passed${NC}"
else
    echo -e "${RED}❌ go vet failed${NC}"
    ERRORS=$((ERRORS + 1))
fi
echo ""

# 2. golangci-lint
echo "2️⃣  Running golangci-lint..."
if check_tool golangci-lint "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; then
    if golangci-lint run; then
        echo -e "${GREEN}✅ golangci-lint passed${NC}"
    else
        echo -e "${RED}❌ golangci-lint failed${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Skipping golangci-lint (not installed)${NC}"
fi
echo ""

# 3. staticcheck
echo "3️⃣  Running staticcheck..."
if check_tool staticcheck "honnef.co/go/tools/cmd/staticcheck@latest"; then
    if staticcheck ./...; then
        echo -e "${GREEN}✅ staticcheck passed${NC}"
    else
        echo -e "${RED}❌ staticcheck failed${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Skipping staticcheck (not installed)${NC}"
fi
echo ""

# 4. gosec (security)
echo "4️⃣  Running gosec (security analysis)..."
if check_tool gosec "github.com/securego/gosec/v2/cmd/gosec@latest"; then
    if gosec -quiet ./...; then
        echo -e "${GREEN}✅ gosec passed${NC}"
    else
        echo -e "${RED}❌ gosec found security issues${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Skipping gosec (not installed)${NC}"
fi
echo ""

# 5. govulncheck (vulnerability scanning)
echo "5️⃣  Running govulncheck..."
if check_tool govulncheck "golang.org/x/vuln/cmd/govulncheck@latest"; then
    if govulncheck ./...; then
        echo -e "${GREEN}✅ govulncheck passed${NC}"
    else
        echo -e "${RED}❌ govulncheck found vulnerabilities${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Skipping govulncheck (not installed)${NC}"
fi
echo ""

# 6. errcheck
echo "6️⃣  Running errcheck..."
if check_tool errcheck "github.com/kisielk/errcheck@latest"; then
    if errcheck ./...; then
        echo -e "${GREEN}✅ errcheck passed${NC}"
    else
        echo -e "${RED}❌ errcheck found unchecked errors${NC}"
        ERRORS=$((ERRORS + 1))
    fi
else
    echo -e "${YELLOW}⚠️  Skipping errcheck (not installed)${NC}"
fi
echo ""

# Summary
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}✅ All static analysis checks passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Static analysis found $ERRORS issue(s)${NC}"
    echo ""
    echo "Install missing tools with:"
    echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    echo "  go install honnef.co/go/tools/cmd/staticcheck@latest"
    echo "  go install github.com/securego/gosec/v2/cmd/gosec@latest"
    echo "  go install golang.org/x/vuln/cmd/govulncheck@latest"
    echo "  go install github.com/kisielk/errcheck@latest"
    exit 1
fi
