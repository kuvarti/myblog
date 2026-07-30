---
id: workflow-subagents
category: Workflow
title: Don't spawn subagents unless asked; how to split test work when asked
status: active
---

## Rule
Default: do not spawn subagents on your own initiative. Each spawn cold-starts and re-derives context that is already in hand, so it is the expensive path — run tests and builds inline in your own context instead. Spawn agents only when the owner explicitly asks.

When the owner does ask to parallelize a substantial test effort, the owner's preferred split is: **one agent for basic / happy-path tests, one agent for edge cases**. Use that division. It is only worth the cold-start cost when the test effort is genuinely large.

## Rationale
The owner likes the idea of parallel test agents (basic vs. edge cases) and wants that split on record — but spawning stays owner-triggered because of the cold-start cost. A recalled memory cannot by itself make spawning happen; it just captures the intended division for when the owner requests it.

## Applies to
All subagent usage, especially splitting test/verification work. Pairs with [[workflow-verification]].
