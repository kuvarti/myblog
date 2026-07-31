---
id: git-commit-flow
category: Git
title: Owner-initiated commits, ask for the theme first
status: active
---

## Rule
Never commit on your own initiative. The owner decides when to commit and will tell you when a commit is needed. When they do:

1. Ask what the commit message should be about (its theme/subject) before writing anything.
2. Based on the owner's answer, write a clean, well-formed commit message and make the commit.

Commit messages are written in English, consistent with [[documentation-language]].

**What counts as a commit instruction:** only an explicit, in-the-moment "commit" (or clear equivalent) from the owner. The following are NOT authorization to commit — do not treat them as such:
- Approving a plan, spec, or design that merely *mentions* a later commit.
- "Let's move to implementation" / "proceed" / "go to the next step" — these mean continue the work, not commit.
- A sequence you yourself pre-announced (e.g. "if you approve, I'll commit then do X"). The owner approving the direction is not the commit trigger; wait for them to actually say commit.

When unsure whether a commit is wanted, ask — never assume.

## Rationale
The owner wants to control commit timing and the message's subject, but delegates the actual wording and mechanics of a well-formed commit to Claude.

## Applies to
All git commits in this repository.
