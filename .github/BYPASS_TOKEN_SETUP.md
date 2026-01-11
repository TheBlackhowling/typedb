# Setting Up Bypass Token for Changelog Updates

## Important: Org Owner vs Org Admin

- **Org Owner**: Has full control, can bypass branch protection
- **Org Admin**: Has admin access but may not have bypass permissions by default
- **For bypass to work**: The user associated with the PAT must be explicitly added to the bypass list

## Create a Personal Access Token (PAT)

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Generate a new token with these **required** permissions:
   - ✅ `repo` (full control of private repositories) - **REQUIRED**
   - ✅ `workflow` (optional, for workflow permissions)
3. **Important**: Use a token from a user account (not a bot account)
4. Copy the token immediately (you won't see it again)

## Add Token to Repository Secrets

1. Go to `https://github.com/TheBlackHowling/typedb/settings/secrets/actions`
2. Click "New repository secret"
3. Name: `CHANGELOG_BYPASS_TOKEN`
4. Value: Paste your PAT token
5. Save

## Grant Bypass Permissions (CRITICAL STEP)

1. Go to `https://github.com/TheBlackHowling/typedb/settings/branches`
2. Find the rule protecting `main` branch
3. Scroll to **"Allow specified actors to bypass required pull requests"**
4. Click **"Add actor"** or **"Edit"**
5. **Add the GitHub username** that owns the PAT token (not the org name)
6. Save the branch protection rule

## Troubleshooting

### Issue: Push still fails with 403 error

**Check 1: Verify PAT has correct scopes**
- PAT must have `repo` scope (full control)
- Classic tokens need `repo` scope, fine-grained tokens need `Contents: Write`

**Check 2: Verify user is in bypass list**
- The **username** (not org name) must be in the bypass list
- Go to branch protection → "Allow specified actors to bypass" → verify username is listed

**Check 3: Verify token is correct**
- Check that `CHANGELOG_BYPASS_TOKEN` secret exists and is set correctly
- The token must belong to the user added to the bypass list

**Check 4: Check workflow logs**
- Look for "Using bypass token for push" message
- Check the remote URL is being set (should show username, not token)

**Check 5: Org-level restrictions**
- Some orgs have additional restrictions that override repo-level settings
- Check org settings → "Member privileges" → "Repository creation" and "Base permissions"

### Alternative: Use GitHub App

For better security and more reliable bypass, create a GitHub App with bypass permissions instead of a PAT.
