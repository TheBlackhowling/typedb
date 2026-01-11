# Repository Template

This is a template repository with best practices for changelog management, automated versioning, and PR workflows.

## Features

- ✅ **Automated Changelog System** - Structured changelog with version files
- ✅ **Semantic Versioning** - Automatic version determination (major/minor/patch)
- ✅ **GitHub Actions Workflow** - Automated version releases on PR merge
- ✅ **Cursor Commands** - AI-friendly commands for PR creation and workflow
- ✅ **PR Templates** - Standardized pull request templates
- ✅ **Context Documentation** - AI assistant priming documents

## Quick Start

### 1. Use This Template

Click "Use this template" on GitHub to create a new repository from this template, or:

```bash
# Clone this template
git clone https://github.com/YOUR_ORG/NewRepoTemplate.git your-new-repo
cd your-new-repo

# Remove existing git history and initialize new repo
rm -rf .git
git init
git add .
git commit -m "Initial commit from template"
```

### 2. Customize for Your Project

1. **Update README.md** - Replace this content with your project description
2. **Update CONTEXT.md** - Replace placeholders (`[PROJECT_NAME]`, `[PROJECT_TYPE]`, etc.)
3. **Update CONTRIBUTING.md** - Customize contribution guidelines
4. **Update Cursor Commands** - Edit `.cursor/commands/create-pr.md` to use your repository name
5. **Update Workflow** - Edit `.github/workflows/version-release.yml` if needed for your project type

### 3. Configure Repository

1. **Set Default Branch** - Ensure `main` is your default branch (or update workflow)
2. **Enable GitHub Actions** - Workflows will run automatically
3. **Configure Branch Protection** - Protect `main` branch, require PR reviews
4. **Set Up Secrets** - Configure `REPO_ACCESS_TOKEN` secret if using cross-repository workflows:
   - Go to Settings → Secrets and variables → Actions
   - Add a new repository secret named `REPO_ACCESS_TOKEN`
   - Use a Personal Access Token (PAT) with appropriate permissions
   - See `.github/workflows/CHANGELOG_ACTION_SETUP.md` for details

## How It Works

### Changelog System

- **`CHANGELOG.md`** - Summary changelog with links to version files
- **`versions/[VERSION].md`** - Detailed changes for each version (auto-generated from PR descriptions)

### Automated Versioning

When a PR is merged:
1. Workflow extracts changelog content from the PR description's "Changes Made" section
2. Determines version bump (major/minor/patch) based on change types
3. Creates version file with commit SHA and PR link
4. Updates `CHANGELOG.md` with summary

### Version Determination

- **MAJOR** (X.0.0): Removed or Deprecated sections, breaking changes
- **MINOR** (0.X.0): Added sections with new features/content
- **PATCH** (0.0.X): Fixed sections, minor changes, workflow improvements

## Usage

### Before Making Changes

1. Get latest changes:
   ```bash
   git checkout main && git pull
   ```
   Or use Cursor command: `/gml`

2. Create a branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Making Changes

1. Make your changes
2. Commit and push:
   ```bash
   git add .
   git commit -m "Add feature: feature-name"
   git push -u origin feature/your-feature-name
   ```

### Creating PRs

**Option 1: Using Cursor Command (Recommended)**
- Type `/create-pr` in Cursor chat
- AI assistant will analyze changes and generate PR summary
- PR will be created automatically

**Option 2: Manual**
```bash
gh pr create --web --base main --repo YOUR_ORG/YOUR_REPO
```

**Important:** Include a "Changes Made" section in your PR description with changelog entries. The workflow will extract this content automatically. See `.github/pull_request_template.md` for the format.

### After PR Merge

- Version is automatically released
- Version file created in `versions/`
- `CHANGELOG.md` updated

## Cursor Commands

- **`/gml`** - Get latest changes from main branch
- **`/create-pr`** - Create PR with AI-generated summary (only when task complete)
- **`/pre-commit-checklist`** - Show checklist to verify before committing

## File Structure

```
.
├── .github/
│   ├── workflows/
│   │   └── version-release.yml    # Automated versioning workflow
│   └── pull_request_template.md   # PR template
├── .cursor/
│   └── commands/                  # Cursor slash commands
│       ├── create-pr.md
│       ├── gml.md
│       └── pre-commit-checklist.md
├── versions/
│   └── [VERSION].md                # Version files (auto-generated)
├── CHANGELOG.md                    # Summary changelog
├── CONTEXT.md                      # AI assistant context guide
├── CONTRIBUTING.md                 # Contribution guidelines
├── README.md                       # This file
└── .gitignore                      # Git ignore rules
```

## Documentation

- **`CONTEXT.md`** - Context guide for AI assistants and contributors
- **`CONTRIBUTING.md`** - Detailed contribution guidelines
- **`CHANGELOG.md`** - Project changelog

## Best Practices

1. **Always document changes in PR** - Include "Changes Made" section in PR description
2. **Create branches** - Never commit directly to `main`
3. **Wait for PR approval** - Don't merge your own PRs
4. **Follow conventions** - Match existing patterns
5. **Test changes** - Verify before committing

## Customization

### For Different Project Types

**Game Development:**
- Update workflow version detection keywords for game-specific content
- Add game-specific documentation structure

**Documentation Projects:**
- Adjust version detection for documentation changes
- Add documentation-specific templates

**Software Projects:**
- Add code style guidelines
- Include testing requirements
- Add deployment documentation

## Support

For questions or issues:
1. Check `CONTEXT.md` for project structure
2. Review `CONTRIBUTING.md` for guidelines
3. Check `CHANGELOG.md` for recent changes

## License

[Add your license here]

---

**Note:** This template includes placeholders that should be customized for your specific project. Search for `[PROJECT_NAME]`, `YOUR_ORG`, `YOUR_REPO`, etc. and replace with your actual values.
