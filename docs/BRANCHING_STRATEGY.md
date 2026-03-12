# Branching Strategy - WhoKnows

**Team:** SyntaxDevOpsSquad-SDS  
**Project:** Legacy Python (2009) → Go migration  
**Last updated:** Week 6 of 14

---

## Current Strategy: Feature Branching with Upstream Sync

We use a GitHub Flow-inspired feature branching strategy. The core principle is simple: `main` is always deployable, and all new work happens on short-lived branches that are merged back via Pull Requests.

Our variant differs from pure GitHub Flow on one important point: **before opening a PR, we always sync the feature branch with the latest `main` locally and push to remote.** This ensures the PR is conflict-free and that CI runs against the actual integration state — not the baseline from when the branch was created.

---

## Branch Structure

We have one long-lived branch - `main`. All other branches are short-lived with fixed prefixes:

| Prefix      | Purpose            | Example                   |
| ----------- | ------------------ | ------------------------- |
| `feature/`  | New functionality  | `feature/ci-pipeline-v2`  |
| `fix/`      | Bug fixes          | `fix/update-dependencies` |
| `refactor/` | Code restructuring | `refactor/go-middleware`  |
| `docs/`     | Documentation      | `docs/branching-strategy` |

Branch names use lowercase and hyphens (kebab-case), consistent with our Conventional Commits convention.

---

## Enforcement

The strategy is not just a social agreement - it is technically enforced at two levels:

**GitHub Ruleset on `main`**

- Direct push to `main` is blocked for everyone, including administrators
- All code must go through a Pull Request
- Minimum 1 approved review required before merge

**Automated CI pipeline (`ci.yml`)**  
Runs automatically on all `pull_request` events targeting `main`:

| Job                   | What it checks                                                                         |
| --------------------- | -------------------------------------------------------------------------------------- |
| `go-ci`               | `go build`, `go vet`, `gofmt` format check, `go test` with race detection and coverage |
| `database-validation` | Validates `schema.sql` against SQLite — ensures the DB schema is always valid SQL      |
| `quality-gate`        | Aggregates both jobs — merge is automatically blocked if any job fails                 |

The combination of branch protection rules and the CI pipeline makes it structurally impossible to merge code that has not been reviewed and approved.

---

## Workflow - Step by Step

```bash
# 1. Create branch from updated main
git checkout main && git pull
git checkout -b feature/my-feature

# 2. Commit using Conventional Commits
git commit -m "feat: add user authentication endpoint"

# 3. Sync with main before opening PR
git fetch origin
git merge origin/main
git push origin feature/my-feature

# 4. Open Pull Request targeting main
# 5. CI runs automatically — fix any failures
# 6. Request review from at least one team member
# 7. Merge after approved review + green CI
# 8. Delete the branch after merge
```

> **Why sync before PR?**  
> We discovered early on that CI was failing on conflicts that didn't exist locally — because our feature branches had drifted from `main`. Merging `origin/main` into the feature branch before pushing ensures the PR reflects the real integration state. This step is not part of standard GitHub Flow, but has significantly reduced review time and CI noise.

---

## Strategy Roadmap

We plan to trial three strategies across the course, moving from simple to more structured and back to disciplined-simple:

| Week  | Strategy                                   | Focus                                                     |
| ----- | ------------------------------------------ | --------------------------------------------------------- |
| 1–7   | Feature Branching with upstream sync (now) | Establish baseline, enforce CI/CD                         |
| 8–10  | Git Flow                                   | Add `develop` buffer branch, structured releases          |
| 11–13 | Trunk-Based Development                    | Short-lived branches + feature flags, max CI/CD alignment |

### Why this progression?

**Git Flow (weeks 8–10)**  
We will introduce a `develop` branch as a buffer between feature work and `main`, along with dedicated `release/` and `hotfix/` branches. This is intentional as a learning exercise — Git Flow is widely used in industry and worth understanding hands-on, even though it is [less aligned with Agile DevOps due to slower release cycles](https://www.geeksforgeeks.org/branching-strategies-in-git/) and adds overhead we don't strictly need.

Expected branches under Git Flow:

- `develop` — integration branch; features merge here, not directly to `main`
- `release/x.x` — created from `develop` when a version is ready for testing
- `hotfix/` — created directly from `main` for critical production fixes

**Trunk-Based Development (weeks 11–13)**  
Once test coverage on the Go codebase is better, we will experiment with very short-lived branches (max 1–2 days) or direct commits to `main` behind feature flags. This is the most CI/CD-native model and will stress-test our testing discipline.

---

## Why Not the Others?

According to [GeeksForGeeks — Branching Strategies in Git](https://www.geeksforgeeks.org/branching-strategies-in-git/), the recommended strategy for a **small team with continuous deployment** is GitHub Flow or TBD — which matches our project exactly.

| Strategy                    | Why we didn't use it as our primary strategy                                                                                                                                                                                                                           |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Git Flow**                | Strong structure for versioned releases, but (1) we have no fixed release cycles, (2) it is less aligned with Agile DevOps due to slower release cycles, and (3) the overhead is unnecessary for our team size. We will trial it in weeks 8–10 as a learning exercise. |
| **Trunk-Based Development** | The most CI/CD-native model, but requires strong test coverage we don't yet have on the new Go codebase. Planned for weeks 11–13.                                                                                                                                      |
| **GitLab Flow**             | A hybrid with environment branches (staging, production) designed specifically for GitLab. We use GitHub, and the model solves a complexity we don't have yet — we have no dedicated staging branch.                                                                   |

---

## Pros & Cons

### Strategy 1: Feature Branching with upstream sync (Weeks 1–7)

| ✓ Advantages                                                                 | ✗ Disadvantages / challenges                                                   |
| ---------------------------------------------------------------------------- | ------------------------------------------------------------------------------ |
| Low cognitive overhead — everyone knows what to do                           | Deploying directly from `main` without a develop buffer requires discipline    |
| CI gate catches errors automatically before they reach `main`                | Early in the project, long-lived branches caused large merge conflicts         |
| PR reviews have driven knowledge sharing across the team                     | A single PR approver can become a bottleneck when someone is unavailable       |
| Sync step ensures PRs are conflict-free and CI runs against real integration | Insufficient test coverage meant CI only caught build errors, not logic errors |
| `main` always reflects reality — no stale develop branch                     | No formal hotfix process — we improvised when urgent fixes were needed         |

### Strategy 2: Git Flow (Weeks 8–10)

_To be filled in once trialled._

### Strategy 3: Trunk-Based Development (Weeks 11–13)

_To be filled in once trialled._

---

## Quick Reference

| Element                   | Rule / Choice                                                |
| ------------------------- | ------------------------------------------------------------ |
| Current model (weeks 1–7) | Feature Branching with upstream sync                         |
| Next model (weeks 8–10)   | Git Flow — `develop` branch as release buffer                |
| Final model (weeks 11–13) | Trunk-Based Development — short branches, feature flags      |
| Protected branches        | `main` — no direct push allowed                              |
| Branch prefixes           | `feature/`, `fix/`, `refactor/`, `docs/`                     |
| Commit convention         | Conventional Commits (`feat`, `fix`, `refactor`, `docs`...)  |
| PR requirements           | Min. 1 approved review + green CI                            |
| Sync before PR            | Merge `origin/main` → feature branch locally, push to remote |
| CI trigger                | All PRs targeting `main`                                     |
| CI jobs                   | `go-ci` → `database-validation` → `quality-gate`             |
| Branch lifetime           | Max 2–3 days — delete after merge                            |

---

_This document is updated continuously as we trial new strategies throughout the course._
