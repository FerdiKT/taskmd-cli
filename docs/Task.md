# Task

<!-- taskmd:version 1 -->

## Todo

### T006 - Verify Homebrew upgrade installs the new preview build
- priority: p1
- assignee: qa-agent
- labels: brew, qa
- created: 2026-04-09T16:54:17+03:00
- updated: 2026-04-09T16:54:17+03:00

#### Notes
Install or upgrade from the tap on a clean environment and confirm List, Board, assignee, and filter behavior are included.

## In Progress

### T008 - Keep this repo's roadmap current in docs/Task.md
- priority: p2
- assignee: main-agent
- labels: process, dogfooding
- created: 2026-04-09T16:54:17+03:00
- updated: 2026-04-09T16:54:20+03:00

#### Notes
Use taskmd as the project's live tracker and keep active work aligned with real implementation steps.

### T005 - Cut v0.2.0 release with preview filters and assignee support
- priority: p1
- assignee: release-agent
- labels: release, brew
- created: 2026-04-09T16:54:17+03:00
- updated: 2026-04-09T17:07:31+03:00

#### Notes
Tag and publish the next release once the preview UI, assignee support, and docs are stable.

## Done

### T004 - Ship initial taskmd public release
- priority: p1
- assignee: release-agent
- labels: release, brew
- created: 2026-04-09T15:20:00+03:00
- updated: 2026-04-09T16:54:04+03:00

#### Notes
Created the public GitHub repo, published `v0.1.0`, added the Homebrew formula, and verified `brew install taskmd`.

### T001 - Polish Jira-style preview layout and compact timestamps
- priority: p1
- assignee: main-agent
- labels: preview, ui
- created: 2026-04-09T16:05:00+03:00
- updated: 2026-04-09T16:54:04+03:00

#### Notes
Lighten the preview UI, wrap long source paths cleanly, and shorten the visible timestamps so cards stay readable.

### T002 - Support assignee field across CLI and Task.md format
- priority: p1
- assignee: main-agent
- labels: schema, agents
- created: 2026-04-09T16:05:00+03:00
- updated: 2026-04-09T16:54:04+03:00

#### Notes
Add assignee support to add/edit/list flows, bulk JSON patches, human output, and canonical Task.md rendering.

### T003 - Add working preview filters for search, label, and assignee
- priority: p1
- assignee: main-agent
- labels: preview, filters
- created: 2026-04-09T16:05:00+03:00
- updated: 2026-04-09T16:54:04+03:00

#### Notes
Make List and Board views preserve filter state while searching issues and narrowing by label or assignee.

### T007 - Add issue detail drawer to the preview UI
- priority: p2
- assignee: ui-agent
- labels: preview, ui
- created: 2026-04-09T16:54:17+03:00
- updated: 2026-04-09T17:07:30+03:00

#### Notes
Open richer task details from the List or Board view without leaving the page.
