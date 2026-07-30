---
id: documentation-maintenance
category: Documentation
title: Docs are never final and Claude maintains them
status: active
---

## Rule
Treat CLAUDE.md (and the docs it references) as living documents that are never final. Whenever a change invalidates something documented — a new route, a renamed file, a completed WIP item, a changed command — update the relevant document in the same session. Add sections when new architecture appears, and prune what no longer holds instead of letting it drift out of date.

## Rationale
The owner expects documentation to evolve continuously alongside the codebase, and expects Claude to be the one keeping it accurate rather than leaving it stale.

## Applies to
CLAUDE.md primarily, plus any document under `memory/` and other repo docs affected by a change.
