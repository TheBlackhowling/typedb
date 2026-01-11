# Unreleased Changes

This file tracks changes that have been made but not yet released in a version.

When a version is released, the content from this file will be moved to the version file and this file will be cleared.

## How to Use

Add changelog entries here as you make changes. When ready to release:

1. The changelog action will automatically use this content when creating a version
2. Or include changelog content in your PR description
3. Or manually trigger a release via workflow_dispatch

---

## Current Unreleased Changes

## Added
- GitHub Actions workflow for testing and coverage
  - Runs tests on push and pull requests
  - Tests against multiple Go versions (1.18-1.22)
  - Generates coverage reports
  - Uploads coverage artifacts
  - Includes race detector for concurrent testing
