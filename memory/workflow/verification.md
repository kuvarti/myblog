---
id: workflow-verification
category: Workflow
title: Verify before claiming done — run it, show the output
status: active
---

## Rule
Before saying code or work is "done", "working", "fixed", or "passing", actually run the relevant check yourself and show the real output. Evidence before assertions — never claim success from reading the code alone.

Depending on what changed, "run the relevant check" means: `go build` / `go test` for the backend; `npm run type-check`, `npm run lint`, or `npm run test:unit` for the frontend; or launching the app and exercising the affected path.

## Rationale
The owner wants success claims backed by executed proof, not assumed from inspection.

## Applies to
Any code or behavior change in this repository. For splitting a large test effort across agents, see [[workflow-subagents]].
