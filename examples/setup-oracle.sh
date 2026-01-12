#!/bin/bash
# Automated setup script for Oracle Database in Docker
# This script handles Oracle Container Registry authentication and license acceptance
# Oracle Database Free is free for development and testing purposes

set -e

ORACLE_REGISTRY="container-registry.oracle.com"
ORACLE_IMAGE="database/free:latest"

echo "🔐 Oracle Database Setup for Testing"
echo "======================================"
echo ""
echo "Oracle Database Free is free for development and testing purposes."
echo "This script will:"
echo "  1. Check if you're logged into Oracle Container Registry"
echo "  2. Pull the Oracle Database Free image"
echo "  3. Verify license acceptance"
echo ""

# Check if already logged in
if docker info | grep -q "$ORACLE_REGISTRY"; then
    echo "✅ Already authenticated to Oracle Container Registry"
else
    echo "📝 Oracle Container Registry Authentication Required"
    echo ""
    echo "To use Oracle Database Free (free for testing):"
    echo "  1. Create a free account at: https://container-registry.oracle.com/"
    echo "  2. Accept the license terms for 'database/free'"
    echo "  3. Run: docker login $ORACLE_REGISTRY"
    echo ""
    read -p "Have you created an account and accepted the license? (y/n) " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        echo "Please visit https://container-registry.oracle.com/ to:"
        echo "  1. Create a free account"
        echo "  2. Navigate to 'database/free'"
        echo "  3. Accept the license terms"
        echo ""
        echo "Then run this script again."
        exit 1
    fi
    
    echo ""
    echo "🔑 Logging into Oracle Container Registry..."
    docker login "$ORACLE_REGISTRY"
    
    if [ $? -ne 0 ]; then
        echo "❌ Login failed. Please check your credentials."
        exit 1
    fi
fi

echo ""
echo "📥 Pulling Oracle Database Free image..."
docker pull "$ORACLE_REGISTRY/$ORACLE_IMAGE"

if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Oracle Database Free image pulled successfully!"
    echo ""
    echo "You can now use Oracle in docker-compose:"
    echo "  docker-compose up oracle migrate-oracle test-oracle"
    echo ""
    echo "Or run all tests including Oracle:"
    echo "  ./run-tests.sh all"
    echo ""
else
    echo ""
    echo "❌ Failed to pull Oracle image."
    echo "Make sure you've accepted the license terms at:"
    echo "  https://container-registry.oracle.com/"
    exit 1
fi
