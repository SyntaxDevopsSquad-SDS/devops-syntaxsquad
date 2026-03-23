---
name: Bug Report
about: Report a problem that needs to be fixed before the system meets quality standards.
title: "[BUG][W_INSERT_WEEK] Short description of the problem"
labels: ["bug", "critical", "backend"]
---

<!-- Those --- mark hidden settings that are configuration for GitHub -->

## Description
What is the bug? Provide a clear and concise description of the problem.

## Steps to Reproduce

> Be as specific as possible — the person fixing it needs to be able to reproduce the bug before they can fix it.

1. Go to '...'
2. Do '...'
3. See error

## Expected Behaviour
What should have happened in a successful scenario?

## Actual Behaviour
What actually happened instead?

## Technical Considerations
- Affects API / OpenAPI spec:
> yes/no — if yes, which endpoint?

- Related to Python-to-Go migration:
> yes / no

- Any error logs or stack traces:
> paste here or write "none"

## PR / Branch
- Branch: `fix/`
> name it: fix/short-description

- Reviewer: @
> tag a classmate

- Linked PR:
> add PR link once created

## Estimate
- [ ] Small (a few hours)
- [ ] Medium (half a day – full day)
- [ ] Large (multiple days, consider splitting)

## Definition of Done
- [ ] Bug can no longer be reproduced by following the steps above.
- [ ] Unit tests are written and passing.
- [ ] GitHub Actions CI pipeline is green.
- [ ] Documentation/README has been updated if needed.