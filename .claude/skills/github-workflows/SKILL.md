---
name: github-workflows
description: Review pull requests, triage issues, and create releases for the midiwav repo (nobishino/midiwav) using the gh CLI. Use when the user asks to review a PR, check open issues, or cut a release.
allowed-tools: Bash(gh *)
---

# GitHub Workflows (nobishino/midiwav)

Use the `gh` CLI for all GitHub operations against this repo.

## Review a PR
```
gh pr view <PR_NUMBER> --json title,body,files,reviews,comments,additions,deletions
gh pr diff <PR_NUMBER>
```
Check for: correctness bugs, missing test coverage (this repo uses `go test`,
golden files under `testdata/*.wav` updated via `go test -update`), and
whether commit/PR titles follow the repo's existing style (see `git log`).

## List and triage issues
```
gh issue list --state open --json number,title,labels,updatedAt
gh issue view <ISSUE_NUMBER>
```

## Create a release
This repo's convention (see git log) is a dedicated PR titled
`Release for vX.Y.Z`, merged before tagging. To cut a release:
```
gh pr create --title "Release for vX.Y.Z" --body "<summary of changes>"
# after merge:
gh release create vX.Y.Z --title "vX.Y.Z" --generate-notes
```
Confirm the version bump and changelog content with the user before creating
the PR or the release — both are visible, hard-to-reverse actions.

## Check CI / PR status
```
gh pr checks <PR_NUMBER>
gh pr status
```

Always summarize findings for the user before taking any action that creates,
merges, or publishes something (PRs, releases, comments).
