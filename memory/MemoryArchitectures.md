# Memory Architecture

This folder stores the project owner's standing preferences as small, individually editable documents, so that CLAUDE.md stays lean and each preference can be changed in isolation. Any Claude session working in this repository should read the preferences here and follow them.

## Layout

```
memory/
  MemoryArchitectures.md   <- this file: explains the scheme
  <category>/              <- one folder per category
    <preference>.md        <- one file per preference
```

Each preference lives in exactly one file, filed under the category folder it belongs to. Splitting by file (not one big list) keeps preferences independently editable and easy to add, remove, or supersede.

## Index (current tree)

Full folder + file tree of every record, so the right file can be located at a glance without listing the directory. Keep this in sync whenever a file or category is added, removed, or renamed.

```
memory/
├── MemoryArchitectures.md            — this file: scheme + this index
├── communication/
│   └── chat-language.md              — [communication-chat-language] Chat in Turkish; keep technical terms in English
├── documentation/
│   ├── language.md                   — [documentation-language] English-only for docs and code comments
│   └── maintenance.md                — [documentation-maintenance] Docs are never final; Claude keeps them updated
├── git/
│   └── commit-flow.md                — [git-commit-flow] Owner-initiated commits; ask the message theme first
└── workflow/
    ├── verification.md               — [workflow-verification] Run tests/build and show output before claiming done
    ├── subagents.md                  — [workflow-subagents] Subagents are a normal tool; skills may spawn them when warranted
    └── mobile-testing.md             — [workflow-mobile-testing] Test mobile via DevTools device emulation, not window resize
```

## Categories

A category groups preferences that share a subject. Current categories:

| Category | Folder | Covers |
|----------|--------|--------|
| Documentation | `documentation/` | How docs and code comments are written and kept up to date — language, tone, upkeep responsibility |
| Communication | `communication/` | How Claude talks with the owner — chat language, terminology |
| Git | `git/` | Version-control workflow — when and how to commit |
| Workflow | `workflow/` | How Claude executes work — verification discipline, subagent usage |

Add a new category by creating a new folder plus a row in this table. Keep categories broad; only split one out when several preferences clearly share a distinct subject.

## Preference file format

Each preference file starts with YAML frontmatter, then the body:

```markdown
---
id: <category>-<short-slug>
category: <category name>
title: <one-line title>
status: active            # active | superseded
---

## Rule
The preference, stated as an instruction.

## Rationale
Why the owner wants it.

## Applies to
Where it takes effect.
```

## Maintaining this folder

- When the owner states a new preference, add a file under the right category (create the category if needed) and update both the **Index tree** and the **Categories table** above.
- When a preference changes, edit its file in place; if fully replaced, set `status: superseded` and note what replaced it rather than silently deleting history.
- Any time a file or category is added, removed, or renamed, update the **Index tree** in the same edit so it never drifts from the real folder contents.
- CLAUDE.md should point here, not duplicate the full text of these preferences.
