# HCP Casdoor Setup Guide

This guide walks through the full setup for the forked repository:

- Crowdin translation sync
- AWS ECR bootstrap in integration and production
- Manual build and release publishing to ECR

It assumes the repository is `hashicorp-forge/casdoor`, the Crowdin project is `hcp-casdoor`, and the AWS region is `us-east-2`.

## Quick Reference

| Item                 | Value                                                     |
| -------------------- | --------------------------------------------------------- |
| Crowdin project name | `hcp-casdoor`                                             |
| Crowdin project ID   | `892410`                                                  |
| Crowdin project URL  | `https://crowdin.com/project/hcp-casdoor`                 |
| Crowdin project type | File-based                                                |
| Crowdin visibility   | Private                                                   |
| Source language      | English                                                   |
| AWS region           | `us-east-2`                                               |
| Integration account  | `995429640676`                                            |
| Integration role ARN | `arn:aws:iam::995429640676:role/DoormatGithubActionsRole` |
| Production account   | `877995958936`                                            |
| Production role ARN  | `arn:aws:iam::877995958936:role/DoormatAdminRole`         |
| ECR repositories     | `casdoor`, `casdoor-all-in-one`                           |

## What Is Already Configured

The repository already contains these workflows and files:

- `.github/workflows/sync.yml` for Crowdin sync
- `.github/workflows/bootstrap-ecr.yml` for ECR repository bootstrap
- `.github/workflows/build.yml` for manual build and release publishing
- `crowdin.yml` for backend translation files
- `web/crowdin.yml` for frontend translation files

The Crowdin configs read the project ID from `CROWDIN_PROJECT_ID`, so the GitHub workflow supplies the value instead of hardcoding it in two places.

## Prerequisites

Before you run anything, make sure you have:

1. Admin or maintainer access to the GitHub repository.
2. A Crowdin account and a private file-based project named `hcp-casdoor`.
3. Access to create and manage GitHub repository secrets.
4. Access to run GitHub Actions workflows in the repository.

You do not need AWS access keys in GitHub secrets. The workflows use Doormat to obtain AWS credentials at runtime.

## 1. Create or Verify the Crowdin Project

1. Log in to Crowdin and open the project `hcp-casdoor`.
2. Make the project private.
3. Choose a file-based project.
4. Set the source language to English.
5. Select the initial target languages you want to support.
6. Skip the onboarding app unless you want the extra guided setup experience.
7. Confirm the project ID is `892410`.
8. Confirm the project URL is `https://crowdin.com/project/hcp-casdoor`.

Notes:

- Selecting the top 30 languages is fine to start.
- You can add more target languages later.
- One shared Crowdin project is the correct choice here because both the frontend and backend translation files point to the same project ID.

## 2. Add the Required GitHub Secret

Add this repository secret in GitHub:

- `CROWDIN_PERSONAL_TOKEN`

This token lets the Crowdin GitHub Action talk to Crowdin. Keep it in GitHub Secrets only. Do not store it in the repository.

You do not need to add these as repository secrets:

- `CROWDIN_PROJECT_ID` because the workflow sets it to `892410`
- `GITHUB_TOKEN` because GitHub provides it automatically to workflows
- AWS access keys because Doormat handles AWS authentication

## 3. Set the GitHub Actions Permissions

The Crowdin workflow creates pull requests automatically. To allow that, check the repository settings:

1. Open the repository settings.
2. Go to the Actions settings page.
3. Make sure workflow permissions allow write access.
4. Allow GitHub Actions to create and approve pull requests if your repository policy requires it.

The Crowdin workflow uses the built-in `GITHUB_TOKEN`, so no personal access token is needed for GitHub itself in the current setup.

## 4. Run the Crowdin Sync Workflow

The Crowdin workflow is manual now.

1. Open the repository Actions tab.
2. Select `Crowdin Action`.
3. Run the workflow with `workflow_dispatch`.
4. Confirm that the workflow uses the repository secret `CROWDIN_PERSONAL_TOKEN`.
5. Confirm that it creates or updates a pull request into `master`.

What the workflow does:

- Uploads the backend source files from `crowdin.yml`.
- Uploads the frontend source files from `web/crowdin.yml`.
- Downloads translations.
- Pushes the translation changes to the localization branch.
- Opens a pull request automatically.

If the workflow fails, check these items first:

- The Crowdin project ID is still `892410`.
- The `CROWDIN_PERSONAL_TOKEN` secret exists.
- Repository Actions permissions allow pull request creation.
- The project is private and file-based.

## 5. Bootstrap ECR in Both AWS Accounts

The ECR bootstrap workflow creates the repositories in the two AWS accounts if they do not already exist.

1. Open the repository Actions tab.
2. Select `Bootstrap ECR`.
3. Run the workflow with `workflow_dispatch`.
4. Keep the default role ARNs unless you intentionally changed them.
5. Confirm the workflow runs in `us-east-2`.
6. Review the job summary for the repository URIs and ARNs.

The workflow creates these repositories in both accounts:

- `casdoor`
- `casdoor-all-in-one`

The integration account and default role are:

- Account: `995429640676`
- Role ARN: `arn:aws:iam::995429640676:role/DoormatGithubActionsRole`

The production account and default role are:

- Account: `877995958936`
- Role ARN: `arn:aws:iam::877995958936:role/DoormatAdminRole`

No AWS secret is needed for this workflow. The workflow uses the pinned Doormat GitHub Action to assume the AWS role and export temporary credentials for the job.

If you need to verify the result, open AWS and confirm that each account has:

- `casdoor`
- `casdoor-all-in-one`

## 6. Run the Manual Build and Release Workflow

The `Build` workflow is also manual.

1. Open the repository Actions tab.
2. Select `Build`.
3. Run the workflow with `workflow_dispatch`.
4. Leave the role ARN inputs at their defaults unless the AWS roles changed.
5. Review the job output after the run finishes.

The workflow does the following:

- Runs tests and the backend/frontend build steps.
- Runs semantic release.
- Assumes the integration AWS role and production AWS role through Doormat.
- Logs in to both ECR registries.
- Pushes the standard image and the all-in-one image to both accounts.

Important note:

- The Docker publish steps only run when semantic release reports a new release.
- If no new release is detected, the publish jobs skip.

## 7. Understand the Crowdin Authentication Choice

This repository uses the built-in `GITHUB_TOKEN` for Crowdin pull request creation.

### Why this is the simplest option

- It is created automatically by GitHub.
- It does not require another long-lived secret.
- It is scoped to the repository.
- It is the lowest-friction setup for same-repo pull requests.

### Tradeoffs

- It only works inside the repository workflow context.
- It depends on repository Actions permissions.
- It is not ideal if you need a stable human identity on commits.

### When to use other options later

- Use a personal access token if you need a human identity or cross-repo access.
- Use a GitHub App token if you want tighter least-privilege controls and are willing to spend more time on setup.

### Current setup steps for `GITHUB_TOKEN`

1. Keep `permissions: contents: write` and `pull-requests: write` in the Crowdin workflow.
2. Leave `env: GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}` in the workflow.
3. Make sure repository settings allow GitHub Actions to create pull requests.

## 8. Keep the Public Links Updated

The repository now points at the new Crowdin project URL:

- `https://crowdin.com/project/hcp-casdoor`

The README no longer shows the Crowdin badge. The web app still links to the Crowdin project from the translation notice.

If the Crowdin project ever changes again, update these locations:

- README project link
- Web app Crowdin link
- Crowdin workflow project ID

## 9. Final Setup Checklist

Use this checklist to confirm everything is ready:

- [ ] Crowdin project `hcp-casdoor` exists and is private
- [ ] Crowdin project ID is `892410`
- [ ] Crowdin is file-based, not string-based
- [ ] `CROWDIN_PERSONAL_TOKEN` is stored as a GitHub secret
- [ ] Repository Actions permissions allow PR creation
- [ ] `Bootstrap ECR` has been run successfully
- [ ] `Build` has been run successfully
- [ ] README and the web app point to `hcp-casdoor` on Crowdin
- [ ] No AWS access keys were added to GitHub secrets

## 10. Troubleshooting

### Crowdin PR is not created

Check the following:

- `CROWDIN_PERSONAL_TOKEN` exists and is correct.
- GitHub Actions permissions allow write access.
- GitHub Actions can create pull requests in the repository settings.
- The workflow is running in the right repository.

### Crowdin sync uses the wrong project

Check the following:

- `CROWDIN_PROJECT_ID` is still `892410` in the workflow.
- `crowdin.yml` and `web/crowdin.yml` still use `project_id_env`.

### ECR bootstrap fails

Check the following:

- The workflow inputs still point to the correct role ARNs.
- The AWS role trust policy allows Doormat to assume the role.
- The workflow is running in `us-east-2`.

### Build publishes nothing

Check the following:

- Semantic release detected a new release.
- The release job completed successfully before the Docker job.
- The ECR bootstrap workflow already created the repositories.

### You want to change the GitHub token approach later

You can switch to a personal access token or GitHub App token later if the repository policy changes. For now, `GITHUB_TOKEN` is the easiest path and keeps the setup lightweight.
