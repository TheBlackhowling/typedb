# Go Package Release Strategy

This document outlines the release strategy for typedb as a Go package published on GitHub and available via the Go module proxy.

## Overview

typedb follows semantic versioning and integrates with Go's module system, GitHub releases, and the Go module proxy (pkg.go.dev).

**Important:** Tagging and GitHub releases are **manual**. The automated workflow only prepares version files and updates the changelog. Without creating a git tag, there is no actual release - the version won't be available via Go modules or pkg.go.dev.

## Versioning Strategy

### Semantic Versioning (SemVer)

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (X.0.0): Breaking changes that require code changes
- **MINOR** (0.X.0): New features, backward compatible
- **PATCH** (0.0.X): Bug fixes, backward compatible

### Current Status

- **Current Version:** Pre-v1.0.0 (0.x.x)
- **Stability:** API may change before v1.0.0
- **v1.0.0 Target:** When API is stable and production-ready

### Version Determination

Our automated release workflow determines version bumps based on changelog entries:

- **MAJOR:** `Removed` or `Deprecated` sections in changelog
- **MINOR:** `Added` sections (new features)
- **PATCH:** `Fixed` sections (bug fixes, minor changes)

## Release Process

### 1. Pre-Release Checklist

Before creating a release:

- [ ] All changes documented in `versions/unreleased.md`
- [ ] Tests passing (unit + integration)
- [ ] Code coverage maintained or improved
- [ ] Documentation updated (README, examples)
- [ ] Breaking changes clearly documented
- [ ] Migration guide provided (if breaking changes)

### 2. Automated Version Preparation Workflow

Our GitHub Actions workflow (`version-release.yml`) prepares version files and changelog:

1. **Trigger:** Manual via `workflow_dispatch` or automatic on merge to main
2. **Version Calculation:** Analyzes `versions/unreleased.md` to determine bump type
3. **Version File:** Creates `versions/[VERSION].md` with commit SHA and PR link
4. **Changelog Update:** Updates `CHANGELOG.md` with summary
5. **Cleanup:** Clears `versions/unreleased.md` (if source was unreleased.md)
6. **Commit & Push:** Commits changes to main branch

**Important:** The workflow does NOT create git tags or GitHub releases. These must be done manually.

### 3. Manual Tagging and Release

**Tagging is manual and required for an actual release.** Without a tag:
- ❌ No GitHub release is created
- ❌ Version is not available via `go get github.com/TheBlackHowling/typedb@vX.Y.Z`
- ❌ Version is not indexed by pkg.go.dev

**To create a release after the workflow has prepared version files:**

```bash
# 1. Ensure you're on main with latest changes (including version files)
git checkout main
git pull origin main

# 2. Verify version file exists
ls versions/v*.md

# 3. Create and push git tag
VERSION="0.1.10"  # Use the version created by the workflow
git tag v$VERSION
git push origin v$VERSION

# 4. (Optional) Create GitHub release with notes
gh release create v$VERSION --notes-file versions/$VERSION.md
```

**Or using GitHub CLI in one step:**
```bash
VERSION="0.1.10"
gh release create v$VERSION \
  --notes-file versions/$VERSION.md \
  --title "Release v$VERSION"
```

This will:
- Create the git tag
- Create the GitHub release
- Make the version available via Go modules
- Trigger pkg.go.dev indexing

### 4. Complete Release Workflow

**Step 1: Prepare Version Files (Automated)**
```bash
# Option A: Let workflow run automatically on PR merge
# Option B: Trigger manually
gh workflow run version-release.yml
```

**Step 2: Create Tag and Release (Manual)**
```bash
# Wait for workflow to complete, then:
VERSION="0.1.10"  # Check versions/ directory for latest
gh release create v$VERSION --notes-file versions/$VERSION.md
```

**Step 3: Verify Release**
```bash
# Check tag exists
git fetch --tags
git tag -l "v*"

# Check GitHub release
gh release list

# Verify Go module availability (may take a few hours)
go list -m -versions github.com/TheBlackHowling/typedb
```

## Go Module Integration

### Module Path

```
github.com/TheBlackHowling/typedb
```

### Version Tags

- **Format:** `v[MAJOR].[MINOR].[PATCH]`
- **Examples:** `v0.1.9`, `v1.0.0`, `v1.2.3`
- **Required:** Must start with `v` for Go modules

### Using typedb

Users import typedb with version:

```go
import "github.com/TheBlackHowling/typedb"
```

Then in `go.mod`:
```go
require github.com/TheBlackHowling/typedb v0.1.9
```

Go will automatically:
1. Fetch from GitHub
2. Cache in Go module proxy
3. Make available via pkg.go.dev

## Go Module Proxy (pkg.go.dev)

### Automatic Publishing

Once a version tag is pushed to GitHub:

1. **Go Module Proxy** automatically discovers the tag
2. **pkg.go.dev** indexes the package (usually within hours)
3. **Documentation** is generated from Go doc comments
4. **Version History** is tracked automatically

### Verification

Check if a version is available:

```bash
# Check module versions
go list -m -versions github.com/TheBlackHowling/typedb

# View on pkg.go.dev
# https://pkg.go.dev/github.com/TheBlackHowling/typedb@v0.1.9
```

### Documentation Requirements

For pkg.go.dev to display properly:

- ✅ Package-level documentation: `// Package typedb ...`
- ✅ Exported function documentation: `// FunctionName ...`
- ✅ Exported type documentation: `// TypeName ...`
- ✅ Example functions: `func ExampleFunctionName()`

## Release Types

### Standard Release

Regular releases with changelog entries:

- Triggered by: Merge to main with changelog updates
- Version bump: Based on changelog content
- Tag format: `v0.1.10`, `v0.2.0`, etc.

### Pre-Release Versions

For testing before stable release:

- Format: `v0.1.10-alpha.1`, `v1.0.0-rc.1`
- Usage: `go get github.com/TheBlackHowling/typedb@v0.1.10-alpha.1`
- Note: Not recommended for production use

### Major Version Releases

For breaking changes:

- Requires: Migration guide in `docs/MIGRATION.md`
- Communication: Clear release notes
- Deprecation: Announce breaking changes in advance (if possible)

## Changelog Management

### Structure

Changelog follows [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [Version] - YYYY-MM-DD

**Commit:** [SHA]
**Pull Request:** [#N]

**Summary:** Brief description

**Key Changes:**
- ✅ **Source Code**: Code changes included
- ✅ **Documentation**: Documentation updated

**Detailed Changes:** See [versions/VERSION.md](versions/VERSION.md)
```

### Categories

- **Added** - New features
- **Changed** - Changes to existing features
- **Fixed** - Bug fixes
- **Removed** - Removed features
- **Deprecated** - Soon-to-be removed features

### Best Practices

1. **Update `versions/unreleased.md`** before committing
2. **Be specific** - Include what changed and why
3. **Link to PRs** - Reference pull requests for context
4. **Group related changes** - Keep similar changes together
5. **User-focused** - Write from user perspective

## Version Lifecycle

### Pre-v1.0.0 (Current)

- **Status:** Development/Alpha
- **API Stability:** May change
- **Versioning:** 0.x.x
- **Breaking Changes:** Allowed without major version bump

### v1.0.0+

- **Status:** Stable
- **API Stability:** Guaranteed within major version
- **Versioning:** x.y.z
- **Breaking Changes:** Require major version bump

## Release Checklist

### Before Release

- [ ] All PRs merged and tested
- [ ] Changelog updated in `versions/unreleased.md`
- [ ] Tests passing (unit + integration)
- [ ] Code coverage acceptable
- [ ] Documentation updated
- [ ] Examples working
- [ ] No known critical bugs

### During Version Preparation

- [ ] Workflow triggered (manual or automatic)
- [ ] Version calculated correctly
- [ ] Version file created (`versions/X.Y.Z.md`)
- [ ] Changelog updated
- [ ] Changes committed and pushed to main

### During Tagging and Release

- [ ] Version file exists in `versions/` directory
- [ ] Git tag created: `git tag vX.Y.Z`
- [ ] Tag pushed: `git push origin vX.Y.Z`
- [ ] GitHub release created (optional, can use `gh release create`)
- [ ] Release verified: `go list -m -versions github.com/TheBlackHowling/typedb`

### After Release

- [ ] Verify tag exists: `git tag -l "v*"`
- [ ] Verify GitHub release: `gh release list`
- [ ] Test module import: `go get github.com/TheBlackHowling/typedb@v[X.Y.Z]`
- [ ] Check pkg.go.dev (may take hours): `https://pkg.go.dev/github.com/TheBlackHowling/typedb@v[X.Y.Z]`
- [ ] Announce release (if significant)

## Automation

### Current Automation

- ✅ **Version Calculation** - Based on changelog
- ✅ **Version File Creation** - Creates `versions/X.Y.Z.md`
- ✅ **Changelog Update** - Automatic CHANGELOG.md update
- ✅ **Unreleased Cleanup** - Clears `versions/unreleased.md` when used

### Manual Steps Required

- ⚠️ **Tag Creation** - Must be done manually: `git tag vX.Y.Z && git push origin vX.Y.Z`
- ⚠️ **GitHub Release** - Must be done manually: `gh release create vX.Y.Z --notes-file versions/X.Y.Z.md`

### Future Enhancements

Potential improvements:

- [ ] Automated release notes generation
- [ ] Dependency update automation (Dependabot)
- [ ] Security advisory integration
- [ ] Release announcement automation
- [ ] Version compatibility checking

## Best Practices

### For Maintainers

1. **Regular Releases** - Don't let unreleased changes accumulate
2. **Clear Changelog** - Help users understand changes
3. **Test Before Release** - Verify everything works
4. **Document Breaking Changes** - Provide migration guides
5. **Communicate** - Announce significant releases

### For Contributors

1. **Update Changelog** - Always update `versions/unreleased.md`
2. **Follow SemVer** - Understand version impact of changes
3. **Test Thoroughly** - Ensure changes work
4. **Document Changes** - Help users understand updates

## Troubleshooting

### Tag Not Created

- Check workflow logs in GitHub Actions
- Verify `versions/unreleased.md` has content
- Ensure workflow has permission to create tags

### Version Not on pkg.go.dev

- Wait a few hours (indexing takes time)
- Verify tag exists: `git tag -l "v*"`
- Check tag format (must start with `v`)
- Verify module path in `go.mod`

### Import Errors

- Verify version exists: `go list -m -versions github.com/TheBlackHowling/typedb`
- Check `go.mod` version matches tag
- Run `go mod tidy` to update dependencies

## References

- [Semantic Versioning](https://semver.org/)
- [Go Modules](https://go.dev/ref/mod)
- [pkg.go.dev](https://pkg.go.dev/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Go Release Process](https://go.dev/doc/modules/release-workflow)
