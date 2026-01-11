# Context Guide - typedb

## Purpose
This document provides essential context for new AI assistants or contributors joining this project. It focuses on **where to find information** and **how to contribute**. Use this as a navigation guide to understand the project structure and contribution workflow.

---

## Project Overview

**typedb** is a Go library project. A type-safe, generic database query library for Go that prioritizes SQL-first development with minimal abstraction.

**Current Phase:** Early Development (Pre-v1.0.0)  
**Focus:** Core implementation following dependency chart in `IMPLEMENTATION_DEPENDENCIES.md`

---

## Documentation Structure

### Root Level Files
- **`README.md`** - Main project overview, links to all documentation
- **`CHANGELOG.md`** - Summary of changes with links to detailed version files
- **`CONTEXT.md`** - This file (priming document for new contexts)
- **`CONTRIBUTING.md`** - Contribution guidelines and workflow

### Changelog Documentation (`/versions/`)
- **`versions/unreleased.md`** - **CRITICAL:** Ongoing changes - add all new changes here
- **`versions/[MAJOR.MINOR.PATCH].md`** - Detailed changes for each released version
- See Changelog section below for detailed process

### Project Documentation
- **`IMPLEMENTATION_DEPENDENCIES.md`** - **CRITICAL:** Implementation dependency chart and order
- **`README.md`** - Project overview, quick start, API documentation
- **`CONTRIBUTING.md`** - Contribution guidelines and workflow

### Design Documentation (External)
Design documents are maintained in the `TechnicalDocumentation` repository:
- `docs/backlog/typedb-design-draft.md` - Complete design documentation
- `docs/backlog/typedb-loader-pattern-discussion.md` - Model loading patterns
- `docs/backlog/typedb-complex-models-design.md` - Multi-table models and JOINs
- `docs/backlog/typedb-model-method-awareness.md` - Model method discovery

### Source Code Structure
- **`*.go`** - Go source files (to be implemented)
  - `errors.go` - Error type definitions
  - `types.go` - Core types and interfaces
  - `registry.go` - Model registration system
  - `reflect.go` - Reflection utilities
  - `deserialize.go` - Row → Model conversion
  - `executor.go` - Database connection and query execution
  - `validate.go` - Model validation
  - `query.go` - Type-safe query helpers
  - `load.go` - Load methods
  - `model.go` - Model struct methods
- **`*_test.go`** - Test files
- **`examples/`** - Example applications (outside main package)

---

## How to Find Information

### For Implementation Questions
1. Start with **`IMPLEMENTATION_DEPENDENCIES.md`** for dependency order and structure
2. Check **`README.md`** for API overview and usage examples
3. Review design docs in **`TechnicalDocumentation`** repository for detailed design decisions

### For Design Decisions
1. Check **`TechnicalDocumentation/docs/backlog/typedb-design-draft.md`** for core design
2. Review **`TechnicalDocumentation/docs/backlog/typedb-loader-pattern-discussion.md`** for Load method patterns
3. See **`TechnicalDocumentation/docs/backlog/typedb-complex-models-design.md`** for complex model handling

### For Contribution Workflow
1. Read **`CONTRIBUTING.md`** for detailed guidelines
2. Review **`CONTEXT.md`** (this file) for workflow overview
3. Check **`CHANGELOG.md`** for recent changes and patterns

---

## Contribution Guidelines

### Before Making Changes

1. **Get Latest Changes:**
   ```bash
   git checkout main && git pull
   ```
   Or use the Cursor command: `/gml`

2. **Create a New Branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```
   **CRITICAL:** Always create a new branch for your work. Never commit directly to `main`.

3. **Review Existing Documentation:**
   - Read `README.md` for project overview
   - Check `CONTRIBUTING.md` for detailed guidelines
   - Review recent changes in `CHANGELOG.md`

### CRITICAL: Always Update Changelog

**Before committing ANY changes, you MUST update `versions/unreleased.md` with detailed changelog entries.**

The changelog follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format with these categories:
- **Added** - New features, files, or content
- **Changed** - Modifications to existing features
- **Fixed** - Bug fixes and corrections
- **Removed** - Removed features or content
- **Deprecated** - Features that will be removed in future versions

**Example changelog entry:**
```markdown
## Added
- New feature: User authentication system
  - Created login page with email/password authentication
  - Added JWT token generation and validation
  - Implemented password reset functionality
  - Updated API documentation with new endpoints
```

### Common Tasks

#### Adding New Features
1. Create branch: `git checkout -b feature/feature-name`
2. Make your changes
3. **Update `versions/unreleased.md`** with detailed changelog entry
4. Commit changes: `git commit -m "Add feature: feature-name"`
5. Push branch: `git push -u origin feature/feature-name`
6. Create PR (see PR Creation section below)

#### Updating Documentation
1. Create branch: `git checkout -b docs/update-topic`
2. Update documentation files
3. **Update `versions/unreleased.md`** with changelog entry
4. Commit: `git commit -m "Update documentation: topic"`
5. Push and create PR

#### Fixing Bugs
1. Create branch: `git checkout -b fix/bug-description`
2. Fix the bug
3. **Update `versions/unreleased.md`** with changelog entry under "Fixed"
4. Commit: `git commit -m "Fix: bug-description"`
5. Push and create PR

---

## Changelog Generation

### Process

1. **Before Committing:**
   - Update `versions/unreleased.md` with your changes
   - Use proper format: Added/Changed/Fixed/Removed/Deprecated
   - Include specific details about what changed

2. **When PR is Merged:**
   - GitHub Actions workflow automatically:
     - Determines version bump (major/minor/patch) based on change types
     - Creates `versions/[VERSION].md` file with commit SHA and PR link
     - Updates `CHANGELOG.md` with summary
     - Clears `versions/unreleased.md`

3. **Version Determination:**
   - **MAJOR** (X.0.0): Removed or Deprecated sections, breaking changes
   - **MINOR** (0.X.0): Added sections with new features/content
   - **PATCH** (0.0.X): Fixed sections, minor changes, workflow improvements

See `docs/VERSIONING.md` (if it exists) for detailed versioning strategy.

---

## PR Creation Workflow

### ⚠️ IMPORTANT: Only Create PR When Task is Complete

**Two scenarios:**

1. **User-Prompted Work:**
   - User explicitly tells you when task is complete
   - Wait for user confirmation before creating PR
   - Example: User says "let's create a PR" or "this task is done"

2. **Feature/Task-Based Work:**
   - Create PR only after completing pre-commit checklist
   - Verify all changes are committed and pushed
   - Ensure changelog is updated

**Do NOT create PRs:**
- Mid-task or while work is in progress
- Without user confirmation (for user-prompted work)
- Without completing checklist (for feature work)
- If changelog is missing

### Using Cursor Commands

**`/create-pr`** - Automatically generates a Copilot-style PR summary and creates the PR
- Analyzes changes and generates comprehensive summary
- Creates PR with detailed description
- **Only use when task is complete**

**`/gml`** - Get latest changes from main branch
- Switches to main and pulls latest changes
- Use before starting new work

**`/pre-commit-checklist`** - Shows checklist to verify before committing
- Ensures changelog is updated
- Verifies all changes are complete

### Manual PR Creation

If you prefer to create PRs manually:
```bash
gh pr create --web --base main --repo YOUR_ORG/YOUR_REPO
```

Or with auto-fill from commits:
```bash
gh pr create --fill --base main --repo YOUR_ORG/YOUR_REPO
```

---

## Quick Commands

### Git Commands
```bash
# Get latest changes
git checkout main && git pull

# Create new branch
git checkout -b feature/your-feature-name

# Check status
git status

# View changelog diff
git diff versions/unreleased.md
```

### Cursor Slash Commands
- `/gml` - Get latest changes from main
- `/create-pr` - Create PR with AI-generated summary (only when task complete)
- `/pre-commit-checklist` - Show pre-commit checklist

### Testing Commands
```bash
# Run all tests
go test -v ./...

# Run tests with race detection
CGO_ENABLED=1 go test -v -race ./...

# Generate coverage report
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

---

## Important Notes

1. **Always update changelog** - Every change must be documented in `versions/unreleased.md`
2. **Create branches** - Never commit directly to `main`
3. **Wait for PR approval** - Don't merge your own PRs unless explicitly allowed
4. **Follow conventions** - Match existing code/documentation style
5. **Test changes** - Verify your changes work before committing

---

## Getting Help

- Check `README.md` for project overview
- Review `CONTRIBUTING.md` for detailed guidelines
- Check `CHANGELOG.md` to see recent changes
- Review existing code/documentation for patterns

---

## Implementation Guidelines

### Follow Dependency Order
Always implement files in the order specified in `IMPLEMENTATION_DEPENDENCIES.md`:
1. Layer 0: `errors.go`, `types.go` (foundation)
2. Layer 1: `registry.go` (registration)
3. Layer 2: `reflect.go` (reflection utilities)
4. Layer 3: `deserialize.go` (deserialization)
5. Layer 4: `executor.go` (executor implementation)
6. Layer 5: `validate.go` (validation)
7. Layer 6: `query.go` (query helpers)
8. Layer 7: `load.go`, `model.go` (load methods)

### Key Design Principles
- **SQL-First**: Users write SQL, library handles type safety
- **No SQL Generation**: Library does NOT generate SQL queries
- **Database-Agnostic**: Works with any `database/sql` driver
- **No Global State**: All operations require explicit instances
- **Type-Safe Generics**: Use Go 1.18+ generics for type safety

### Testing Strategy
- Test each layer as you build it
- Use unit tests for isolated components
- Use integration tests for database operations
- Mock dependencies where appropriate

### Running Tests and Coverage

**Run all tests:**
```bash
go test -v ./...
```

**Run tests with race detection (requires CGO):**
```bash
CGO_ENABLED=1 go test -v -race ./...
```

**Generate coverage report:**
```bash
# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

**Note:** Coverage files (`coverage.out`, `coverage.html`) are ignored by `.gitignore` and should not be committed.

The CI workflow (`.github/workflows/test.yml`) automatically runs tests with race detection and generates coverage reports for Go 1.22.

