---
id: workflow-subagents
category: Workflow
title: Subagents are a normal tool; skills may spawn them when warranted
status: active
---

## Rule
Subagents are a normal tool, not an owner-only escalation. When a skill or process is designed around subagents (e.g. `subagent-driven-development`) or the work genuinely benefits from parallel/isolated execution, spawn them when it is warranted — no need to ask permission each time. The owner has authorized skill-driven and process-warranted spawning.

Still apply judgment about cost: each spawn cold-starts and re-derives context, so don't spawn for trivial work that is cheaper to do inline. When parallelizing a large test effort, a sensible split is one agent for basic / happy-path tests and one for edge cases.

## Rationale
An earlier version of this memory restricted spawning to explicit owner requests. The owner decided that was a mistake — it fought against skills whose whole design is subagent-based. Subagents should be available whenever the work calls for them.

## Applies to
All subagent usage. Pairs with [[workflow-verification]].
