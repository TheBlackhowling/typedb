# Get Latest Changes from Main (gml)

This command checks out the `main` branch and pulls the latest changes from the remote repository.

```bash
git checkout main && git pull
```

## Usage

Type `/gml` in Cursor's chat to execute this command.

## What it does

1. Switches to the `main` branch
2. Pulls the latest changes from the remote repository
3. Updates your local repository with any new commits, including version files and CHANGELOG.md updates created by the workflow after PR merges

## When to use

- Before starting new work to ensure you have the latest codebase
- After a PR merges to get the latest version files
- When you want to sync your local repository with the remote main branch

