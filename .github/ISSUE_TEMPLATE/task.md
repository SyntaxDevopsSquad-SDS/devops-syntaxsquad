---
name: Development Task
about: Standard template for weekly development and migration tasks (Python to Go).
title: "[W_INSERT_WEEK] Task Name"
labels: ["backend"]
---

<!-- Those --- mark hidden settings that are configuration for GitHub -->

## Description
Provide a concise description of the work.
Does this relate to the Python-to-Go migration?
> yes / no

## Sub-tasks
- [ ]  
> e.g. Set up database connection 

- [ ]  
> e.g. Check if database exists 

- [ ] 
> e.g. Initialize database schema 

- [ ] 
> e.g. Implement password hashing 

## Technical Considerations
- New env variables:
> list them, or write "none"

- OpenAPI spec affected:
> yes/no — if yes, describe what changes

- Other notes:
> anything else the reviewer should know

## PR / Branch
- Branch: `feature/`
> specify branch name

- Reviewer: @
> tag a classmate

- Linked PR:
> add PR link once created

## Estimate
- [ ] Small (a few hours)
- [ ] Medium (half a day – full day)
- [ ] Large (multiple days, consider splitting)

## Definition of Done
- [ ] Logic is implemented in Go if task was migration related.
- [ ] Unit tests are written and passing.
- [ ] GitHub Actions CI pipeline is green.
- [ ] Documentation/README has been updated.