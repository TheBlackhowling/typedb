# Contributing to [PROJECT_NAME]

Thank you for your interest in contributing! This document provides guidelines for contributing to this project.

## Getting Started

1. **Fork and Clone:**
   ```bash
   git clone https://github.com/YOUR_ORG/YOUR_REPO.git
   cd YOUR_REPO
   ```

2. **Read Documentation:**
   - `README.md` - Project overview
   - `CONTEXT.md` - Context guide for AI assistants and contributors
   - This file - Contribution guidelines

3. **Set Up Development Environment:**
   - [Add setup instructions here]
   - Install dependencies: `[command]`
   - Run tests: `[command]`

## Development Workflow

### Branching Strategy

- **Main Branch:** `main` (protected, requires PR)
- **Feature Branches:** `feature/feature-name`
- **Bug Fixes:** `fix/bug-description`
- **Documentation:** `docs/topic`
- **Chores/Tooling:** `chore/task-name`

### Making Changes

1. **Create a Branch:**
   ```bash
   git checkout main
   git pull
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes:**
   - Write code/documentation
   - Follow existing patterns and conventions
   - Add tests if applicable

3. **Update Changelog:**
   - **CRITICAL:** Update `versions/unreleased.md` with your changes
   - Use proper format (Added/Changed/Fixed/Removed/Deprecated)
   - Include specific details

4. **Commit Changes:**
   ```bash
   git add .
   git commit -m "Add feature: feature-name"
   ```
   - Use clear, descriptive commit messages
   - Reference issues if applicable: `git commit -m "Fix: bug-name (closes #123)"`

5. **Push and Create PR:**
   ```bash
   git push -u origin feature/your-feature-name
   ```
   - Create a Pull Request targeting `main`
   - Use the PR template provided
   - Ensure changelog entry is visible in PR diff

## Changelog Guidelines

### Format

Follow [Keep a Changelog](https://keepachangelog.com/en/1.0.0/) format:

```markdown
## Added
- New feature: Feature name
  - Specific detail about what was added
  - Another detail

## Changed
- Updated feature: Feature name
  - What changed and why

## Fixed
- Fixed bug: Bug description
  - What was fixed

## Removed
- Removed feature: Feature name
  - Why it was removed
```

### Categories

- **Added** - New features, files, or content
- **Changed** - Modifications to existing features
- **Fixed** - Bug fixes and corrections
- **Removed** - Removed features or content
- **Deprecated** - Features that will be removed in future versions

### Versioning

Versions are automatically determined by the GitHub Actions workflow:
- **MAJOR** (X.0.0): Breaking changes, removals, deprecations
- **MINOR** (0.X.0): New features, significant additions
- **PATCH** (0.0.X): Bug fixes, minor changes, workflow improvements

## Pull Request Process

### Before Creating PR

- [ ] All changes are complete
- [ ] Changelog updated in `versions/unreleased.md`
- [ ] Code/documentation follows project conventions
- [ ] Tests pass (if applicable)
- [ ] Documentation updated (if needed)

### PR Description

Use the PR template provided (`.github/pull_request_template.md`). Include:
- Summary of changes
- Detailed breakdown by category
- Impact assessment
- Testing notes
- Related issues/PRs

### Review Process

1. PR is created targeting `main`
2. Automated checks run (if configured)
3. Maintainers review
4. Changes requested (if needed)
5. PR approved and merged
6. Version automatically released by workflow

## Code Style

[Add your project's code style guidelines here]

- Follow existing patterns
- Use consistent formatting
- Add comments for complex logic
- Keep functions focused and small

## Testing

[Add your project's testing guidelines here]

- Write tests for new features
- Ensure existing tests pass
- Add integration tests for significant changes

## Documentation

- Update relevant documentation with changes
- Add examples for new features
- Keep README.md up to date
- Document breaking changes clearly

## Questions?

- Open an issue for questions
- Check existing documentation
- Review `CONTEXT.md` for project structure

## Code of Conduct

[Add your code of conduct here or link to it]

Be respectful, inclusive, and constructive in all interactions.

---

Thank you for contributing! 🎉

