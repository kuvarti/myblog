---
id: workflow-mobile-testing
category: Workflow
title: Test mobile via DevTools device emulation, not window resize
status: active
---

## Rule
When verifying a mobile / responsive view, use the browser DevTools **device
toolbar** (device emulation) rather than resizing the browser window. Resizing a
desktop window gives an inconsistent result (desktop chrome, scrollbar eating
width, no real mobile viewport / device pixel ratio); device emulation matches a
real phone viewport and is what the owner uses.

## Rationale
The owner tests mobile through DevTools device mode and finds it far more
consistent than a resized window; screenshots should reflect the same.

## Applies to
Any mobile/responsive verification with the browser automation tools.
