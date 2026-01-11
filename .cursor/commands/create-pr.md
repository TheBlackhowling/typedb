# Create Pull Request (create-pr)

This command analyzes your changes using Cursor AI to generate an Intention and Summary, then creates a PR with those values embedded for the GitHub Action to use.

## Usage

Type `/create-pr` in Cursor's chat. The AI assistant will:

1. **Analyze Commits**: Review all commits on your branch compared to main
2. **Generate Intention**: Create a concise intention statement describing the purpose
3. **Generate Summary**: Create a verbose summary of changes (for PR description)
4. **Create PR**: Create the PR with Intention and Summary embedded in the body

## What it does

1. **Analyzes changes**: 
   - Reviews commit messages and file diffs
   - Examines file structure and changes
   - Understands the context of modifications

2. **Generates AI-powered content**:
   - **Intention**: Brief statement of what this PR aims to accomplish
   - **Summary**: Detailed description formatted as bullet points (for Overview section)
     - Each bullet point should be a complete sentence
     - Semantically group related changes together
     - Use multiple bullets to cover different aspects

3. **Creates PR**: Automatically creates the PR with embedded metadata:
   ```markdown
   <!-- PR_METADATA_START -->
   INTENTION: [AI-generated intention]
   SUMMARY: [AI-generated verbose summary]
   <!-- PR_METADATA_END -->
   ```

4. **GitHub Action Enhancement**: The workflow will use these values to generate a better semantic changelog

## PR Body Format

The PR will be created with this structure:

```markdown
<!-- PR_METADATA_START -->
INTENTION: [Concise intention statement]
SUMMARY: [Verbose summary of changes]
<!-- PR_METADATA_END -->

[GitHub Action will generate the rest]
```

## When to use

**⚠️ IMPORTANT: Only create PR when task is complete.**

**Two scenarios:**

1. **User-Prompted Work:**
   - User explicitly tells you when task is complete
   - Wait for user confirmation before creating PR
   - Example: User says "let's create a PR" or "this task is done"

2. **Feature/Task-Based Work:**
   - Create PR only after completing pre-commit checklist
   - Verify all changes are committed and pushed
   - Ensure changelog is updated

**Do NOT use this command:**
- Mid-task or while work is in progress
- Without user confirmation (for user-prompted work)
- Without completing checklist (for feature work)
- If changelog is missing

## How Intention and Summary are Generated

The AI analyzes:
- **Commit messages**: Understanding what was changed and why
- **File changes**: Patterns in modifications (new features, fixes, refactors)
- **File structure**: Context from directory organization
- **Change patterns**: Detecting breaking changes, new features, bug fixes

**Intention Example:**
```
Add automated PR summary generation using Cursor AI analysis
```

**Summary Example (must be bullet points):**
```
- Enhanced PR workflow with Cursor AI integration for better context
- Added comprehensive file categorization system (Documentation, Source, Assets, Migrations, Configuration, Other)
- Implemented automatic preservation of Cursor-generated metadata in PR descriptions
- Added entry-level engineer learning program with structured modules covering JavaScript fundamentals through real-world projects
```

**IMPORTANT**: The Summary MUST be formatted as bullet points (using `- ` prefix). Each bullet should be a complete, meaningful sentence that semantically groups related changes together.

## Manual PR Creation

If you prefer to create PRs manually with Intention and Summary:

```bash
# Generate Intention and Summary first (ask Cursor to analyze commits)
# Then create PR with metadata directly (no temporary file needed):
gh pr create --title "Your PR Title" --body "<!-- PR_METADATA_START -->
INTENTION: Your intention here
SUMMARY: Your verbose summary here
<!-- PR_METADATA_END -->" --base main --repo Blackhowling-Dev/NewRepoTemplate
```

Or using a here-string (PowerShell):
```powershell
gh pr create --title "Your PR Title" --body @"
<!-- PR_METADATA_START -->
INTENTION: Your intention here
SUMMARY: Your verbose summary here
<!-- PR_METADATA_END -->
"@ --base main --repo Blackhowling-Dev/NewRepoTemplate
```

## GitHub Action Integration

The GitHub Action (`pr-summary-and-description.yml`) will:
1. Extract Intention and Summary from PR metadata
2. Use Summary as the Overview section (preserving bullet point format from Cursor)
3. Generate file categorization with additions/deletions by category
4. Add PR Summary section with Intention wrapped in `<!-- INTENTION_START -->` and `<!-- INTENTION_END -->` tags for reliable changelog extraction

## Repository

- **Repository**: `Blackhowling-Dev/NewRepoTemplate`
- **Base Branch**: `main`
- **Workflow**: `.github/workflows/pr-summary-and-description.yml`
- **Action Source**: Uses `Blackhowling-Dev/TechnicalDocumentation/.github/actions/pr-summary-and-description` action
