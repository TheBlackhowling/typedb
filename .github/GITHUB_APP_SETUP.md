# Setting Up GitHub App for Changelog Bypass

This guide walks you through creating a GitHub App that can bypass branch protection rules for automated changelog updates.

## Why Use a GitHub App?

- ✅ Can be added to bypass lists even when individual users cannot
- ✅ More secure than PATs (scoped permissions, easier to revoke)
- ✅ Better audit trail (actions show as the app, not a user)
- ✅ Fine-grained permissions control

## Step 1: Create the GitHub App

1. Go to your organization settings:
   - `https://github.com/organizations/TheBlackHowling/settings/apps`
   - Or: Organization → Settings → Developer settings → GitHub Apps

2. Click **"New GitHub App"**

3. Fill in the app details:

   **Basic Information:**
   - **GitHub App name**: `typedb-changelog-automation` (or your preferred name)
   - **Homepage URL**: `https://github.com/TheBlackHowling/typedb`
   - **User authorization callback URL**: Leave empty (not needed for this use case)
   - **Webhook**: 
     - ✅ **Active**: Unchecked (we don't need webhooks)
     - **Webhook URL**: Leave empty
   - **Webhook secret**: Leave empty

   **Permissions:**

   **Repository permissions:**
   - **Contents**: ✅ **Read and write** (required to push changes)
   - **Pull requests**: ✅ **Read** (optional, for PR info)
   - **Metadata**: ✅ **Read-only** (default, required)

   **Organization permissions:**
   - Leave all as **No access** (not needed)

   **Where can this GitHub App be installed?**
   - ✅ **Only on this account** (TheBlackHowling organization)

4. Click **"Create GitHub App"**

## Step 2: Generate and Save Private Key

1. After creating the app, you'll see a page with app details
2. Scroll down to **"Private keys"** section
3. Click **"Generate a private key"**
4. **IMPORTANT**: Download the `.pem` file immediately - you won't be able to download it again!
5. Save it securely (you'll need it for the workflow)

## Step 3: Note Your App ID

1. On the app settings page, find **"App ID"** (it's a number like `123456`)
2. Copy this number - you'll need it for the workflow

## Step 4: Install the App on Your Repository

1. On the app settings page, click **"Install App"** (in the left sidebar)
2. Select **"Only select repositories"**
3. Choose `typedb` repository
4. Click **"Install"**
5. On the installation page, you can grant additional permissions if needed (usually not required)

## Step 5: Add App to Branch Protection Bypass List

1. Go to repository settings: `https://github.com/TheBlackHowling/typedb/settings/branches`
2. Find the rule protecting `main` branch
3. Scroll to **"Allow specified actors to bypass required pull requests"**
4. Click **"Add actor"** or **"Edit"**
5. Search for your GitHub App name: `typedb-changelog-automation`
6. Add the app to the bypass list
7. Save the branch protection rule

## Step 6: Add Secrets to Repository

You need to add two secrets to your repository:

1. Go to: `https://github.com/TheBlackHowling/typedb/settings/secrets/actions`

2. Add **`APP_ID`** secret:
   - Click **"New repository secret"**
   - Name: `APP_ID`
   - Value: The App ID number from Step 3
   - Click **"Add secret"**

3. Add **`APP_PRIVATE_KEY`** secret:
   - Click **"New repository secret"**
   - Name: `APP_PRIVATE_KEY`
   - Value: Open the `.pem` file you downloaded in Step 2, copy the entire contents (including `-----BEGIN RSA PRIVATE KEY-----` and `-----END RSA PRIVATE KEY-----`)
   - Click **"Add secret"**

## Step 7: Update Workflow to Use GitHub App

Update your `version-release.yml` workflow to use the GitHub App instead of a PAT:

```yaml
- name: Generate GitHub App token
  id: generate-token
  uses: actions/create-github-app-token@v1
  with:
    app-id: ${{ secrets.APP_ID }}
    private-key: ${{ secrets.APP_PRIVATE_KEY }}
    owner: TheBlackHowling

- name: Push changes
  if: steps.run_changelog.outputs.has_content == 'true' && steps.run_changelog.outputs.version != '' && env.TARGET_BRANCH != ''
  run: |
    if [ "${{ github.event_name }}" == "push" ] && [[ "${{ github.ref }}" == refs/tags/* ]]; then
      echo "Cannot push to tag, changes should be on a branch"
      exit 1
    else
      echo "=== Using GitHub App Token for Push ==="
      git remote set-url origin https://x-access-token:${{ steps.generate-token.outputs.token }}@github.com/${{ github.repository }}.git
      git push origin HEAD:${{ env.TARGET_BRANCH }}
    fi
```

## Step 8: Remove Old Bypass Token (Optional)

If you were using `CHANGELOG_BYPASS_TOKEN`, you can now:
1. Remove it from repository secrets (or org secrets)
2. Remove the bypass token debug code from the workflow

## Troubleshooting

### Issue: App token generation fails

**Check 1: Verify secrets are correct**
- `APP_ID` should be a number (like `123456`)
- `APP_PRIVATE_KEY` should include the full PEM file contents with headers

**Check 2: Verify app is installed**
- Go to: `https://github.com/organizations/TheBlackHowling/settings/installations`
- Verify `typedb-changelog-automation` is installed on `typedb` repository

**Check 3: Verify app permissions**
- Go to app settings → Permissions
- Ensure "Contents" has "Read and write" permission

### Issue: Push still fails with 403

**Check 1: Verify app is in bypass list**
- Go to branch protection settings
- Verify the GitHub App is listed in "Allow specified actors to bypass"

**Check 2: Check workflow logs**
- Look for "Using GitHub App Token for Push" message
- Check if token was generated successfully

**Check 3: Verify installation permissions**
- Go to app installation settings
- Ensure repository has "Read and write" access to contents

## Security Best Practices

1. **Rotate keys periodically**: Generate new private keys and update secrets
2. **Limit app installation**: Only install on repositories that need it
3. **Minimal permissions**: Only grant permissions that are absolutely necessary
4. **Monitor app activity**: Check app activity logs regularly
5. **Store private key securely**: Never commit the `.pem` file to the repository

## Additional Resources

- [GitHub Apps Documentation](https://docs.github.com/en/apps)
- [Creating GitHub Apps](https://docs.github.com/en/apps/creating-github-apps)
- [GitHub App Permissions](https://docs.github.com/en/apps/creating-github-apps/setting-permissions-for-github-apps)
