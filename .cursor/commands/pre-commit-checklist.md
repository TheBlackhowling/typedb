# Pre-Commit Checklist

Use this checklist before committing changes to ensure nothing is forgotten.

## Before Committing

- [ ] All changes are complete
- [ ] **CRITICAL: Updated `versions/unreleased.md`** with detailed changes
- [ ] Changelog entry follows proper format (Added/Changed/Fixed/Removed)
- [ ] Changelog includes specific details about what changed
- [ ] All related files are updated (indexes, cross-references, etc.)
- [ ] No temporary files or debug code included

## Before Creating PR

- [ ] All commits are pushed to branch
- [ ] **CRITICAL: Verify `versions/unreleased.md` is included in PR**
- [ ] PR description explains the changes
- [ ] Check that changelog entry is visible in PR diff

## Quick Check Command

Before committing, run:
```bash
git status
git diff versions/unreleased.md
```

If `versions/unreleased.md` shows no changes, you may have forgotten to update it!

