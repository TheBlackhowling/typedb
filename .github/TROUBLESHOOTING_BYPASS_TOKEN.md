# Troubleshooting Bypass Token Issues

## Issue: "Using bypass token for push" message doesn't appear

If the workflow doesn't show "Using bypass token for push", it means the `CHANGELOG_BYPASS_TOKEN` secret is not being detected.

### Check 1: Verify Secret Exists

1. Go to: https://github.com/TheBlackHowling/typedb/settings/secrets/actions
2. Verify `CHANGELOG_BYPASS_TOKEN` exists in the list
3. Check that the name is **exactly** `CHANGELOG_BYPASS_TOKEN` (case-sensitive, no spaces)

### Check 2: Verify Secret Scope

- Repository secrets are available to all workflows in the repo
- Organization secrets need to be explicitly enabled for the repository
- If using an org secret, make sure it's enabled for `typedb` repository

### Check 3: Check Workflow Logs

After running the workflow, check the "Push changes" step logs. You should see:
- `✅ CHANGELOG_BYPASS_TOKEN secret is set (length: XX)` - Token is detected
- `❌ CHANGELOG_BYPASS_TOKEN secret is NOT set or is empty` - Token is NOT detected

### Check 4: Secret Name Typo

Common mistakes:
- `CHANGELOG_BYPASS_TOKEN` ✅ (correct)
- `CHANGELOG_BYPASS_TOKENS` ❌ (wrong - extra S)
- `CHANGELOG_BYPASS` ❌ (wrong - missing _TOKEN)
- `changelog_bypass_token` ❌ (wrong - lowercase)

### Check 5: Recreate Secret

If the secret exists but isn't being detected:

1. Delete the existing `CHANGELOG_BYPASS_TOKEN` secret
2. Create a new one with the exact name `CHANGELOG_BYPASS_TOKEN`
3. Paste your PAT token
4. Save
5. Re-run the workflow

### Check 6: Verify Token Format

The PAT token should be:
- 40 characters long (for classic tokens)
- Starts with `ghp_` (for classic tokens)
- No spaces or line breaks

### Check 7: Workflow Permissions

Make sure the workflow has access to secrets:
- Go to workflow file → Check if `permissions` block exists
- Secrets are available by default, but verify workflow can access them

## If Token is Detected But Push Still Fails

If you see "Using bypass token for push" but still get 403:

1. **Verify username in bypass list**: The username (not org name) must be in the bypass list
2. **Check PAT scopes**: Must have `repo` scope (full control)
3. **Verify token belongs to bypassed user**: The PAT must belong to the user in the bypass list
4. **Check org-level restrictions**: Org settings might override repo settings
