# Branch Protection Setup for Benchmark Workflow

To ensure benchmarks run only when a merge is attempted and block merges if they fail, configure branch protection rules.

## Setup Instructions

1. Go to your repository on GitHub
2. Navigate to **Settings** > **Branches** > **Branch protection rules**
3. Click **Add rule** or edit the existing rule for the `main` branch
4. Enable the following settings:
   - ✅ **Require status checks to pass before merging**
   - ✅ **Require branches to be up to date before merging** (optional but recommended)
5. Under **Status checks that are required**, add:
   - `benchmark` (the job name from the workflow)

## How It Works

- The benchmark workflow runs when a PR is opened, updated, or marked as ready for review
- The workflow will **block the merge** if:
  - Benchmarks fail to run (compilation errors, test failures, etc.)
  - Performance regressions >10% are detected compared to baseline
- The merge button will be disabled until the benchmark check passes
- This ensures benchmarks run before merge without running on every push

## Alternative: GitHub Merge Queue

If you have GitHub Enterprise and want benchmarks to run only during the merge queue process:

1. Enable **Merge Queue** in branch protection rules
2. The workflow will run as part of the merge queue
3. Merges will be blocked if benchmarks fail

## Manual Override

If you need to merge without running benchmarks (e.g., documentation-only changes), you can:
- Use `workflow_dispatch` to manually trigger benchmarks
- Temporarily disable the required status check (not recommended)
