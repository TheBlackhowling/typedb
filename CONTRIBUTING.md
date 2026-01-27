# Contributing to typedb

**Note:** We are not currently accepting code contributions. However, we welcome feature requests, suggestions, and bug reports!

## Feature Requests & Suggestions

We'd love to hear your ideas for improving typedb:

- **Feature Requests**: Open an issue describing the feature you'd like to see
- **Suggestions**: Share your ideas for improvements, optimizations, or new capabilities  
- **Bug Reports**: Report any issues you encounter

Your feedback helps shape the future of typedb!

---

## Development Guidelines (For Maintainers)

The following sections are for maintainers and internal development reference.

## Getting Started

1. **Fork and Clone:**
   ```bash
   git clone https://github.com/TheBlackHowling/typedb.git
   cd typedb
   ```

2. **Read Documentation:**
   - `README.md` - Project overview and API documentation
   - `CONTEXT.md` - Context guide for AI assistants and contributors
   - `IMPLEMENTATION_DEPENDENCIES.md` - Implementation dependency chart
   - This file - Contribution guidelines

3. **Set Up Development Environment:**
   - **Go 1.23.12+** required (minimum for security fixes)
   - **Go 1.24+ recommended** for best compatibility
   - Install Go: https://golang.org/doc/install
   - No external dependencies required (only `database/sql` from standard library)
   - Run tests: `go test ./...`
   - Run tests with coverage: `go test -cover ./...`

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

### Go Conventions
- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` for formatting (or `goimports` for imports)
- Run `golangci-lint` if configured (optional)

### Project-Specific Guidelines
- **Keep types and implementations separate** - `types.go` should only have type definitions
- **Document exported functions** - Add godoc comments for all public APIs
- **Use generics appropriately** - Leverage Go 1.18+ generics for type safety
- **Handle errors explicitly** - Don't ignore errors, return them appropriately
- **Follow dependency order** - Implement files in order specified in `IMPLEMENTATION_DEPENDENCIES.md`
- **Keep functions focused** - Single responsibility principle
- **Use interfaces** - Define interfaces for testability and flexibility

### Naming Conventions
- Use descriptive names for exported types/functions
- Prefix internal helpers with lowercase (unexported)
- Use `T` for generic type parameters
- Use `ctx` for context.Context parameters
- Use `exec` for Executor parameters

## Testing

### Testing Strategy
- **Unit tests** - Test isolated components (reflection helpers, deserialization, validation)
- **Integration tests** - Test database operations (use `sqlmock` or real test database)
- **Test each layer** - Write tests as you implement each layer from `IMPLEMENTATION_DEPENDENCIES.md`

### Test Structure
- Place tests in `*_test.go` files alongside source files
- Use table-driven tests where appropriate
- Mock dependencies using interfaces
- Use `testify` if needed (but prefer standard library)

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for specific package
go test ./typedb

# Run tests with verbose output
go test -v ./...
```

### Test Database Setup
- Use `sqlmock` for unit testing database interactions
- Use Docker/test containers for integration tests (optional)
- Test against multiple database drivers when possible

## Documentation

### Code Documentation
- Add godoc comments for all exported types, functions, and methods
- Include usage examples in godoc where helpful
- Document parameters and return values
- Explain complex algorithms or design decisions

### Project Documentation
- Update `README.md` with new features and API changes
- Update `IMPLEMENTATION_DEPENDENCIES.md` if dependency order changes
- Add examples to `examples/` directory for new features
- Document breaking changes clearly in changelog

### Design Documentation
- Design decisions are documented in `TechnicalDocumentation` repository
- Update design docs if making significant architectural changes
- Link to design docs from code comments when referencing design decisions

## Questions?

- Open an issue for questions
- Check existing documentation
- Review `CONTEXT.md` for project structure

## Implementation Workflow

### Before Starting Implementation
1. Review `IMPLEMENTATION_DEPENDENCIES.md` to understand dependency order
2. Read relevant design documents in `TechnicalDocumentation` repository
3. Check existing code patterns (if any)
4. Create a feature branch: `git checkout -b feature/layer-name`

### During Implementation
1. Implement files in dependency order
2. Write tests as you go
3. Keep types and implementations separate
4. Document exported functions
5. Update `versions/unreleased.md` with changes

### After Implementation
1. Run all tests: `go test ./...`
2. Check code formatting: `gofmt -l .`
3. Update changelog
4. Create PR following PR process below

## Code of Conduct

Be respectful, inclusive, and constructive in all interactions.

- Be welcoming to newcomers
- Provide constructive feedback
- Focus on what's best for the project
- Respect different viewpoints and experiences

---

Thank you for contributing! 🎉

