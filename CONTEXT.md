# Context Guide - [PROJECT_NAME]

## Purpose
This document provides essential context for new AI assistants or contributors joining this project. It focuses on **where to find information** and **how to contribute**. Use this as a navigation guide to understand the project structure and contribution workflow.

---

## Project Overview

**[PROJECT_NAME]** is a [PROJECT_TYPE] project. [Brief description of what this project does and its current state.]

**Current Phase:** [Development Phase]  
**Focus:** [Current focus areas]

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
- **[Add your documentation structure here]**
- Example: `/docs/` - Project documentation
- Example: `/src/` - Source code
- Example: `/tests/` - Test files

---

## How to Find Information

### For [Topic] Questions
1. Start with **[Location]** for overview
2. Check **[Specific Location]** for detailed information
3. Review **[Additional Location]** for related content

**[Customize this section based on your project structure]**

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

*This document should be customized for your specific project. Update placeholders like [PROJECT_NAME], [PROJECT_TYPE], etc. with your actual project information.*

