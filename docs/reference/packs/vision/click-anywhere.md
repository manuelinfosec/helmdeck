---
title: vision.click_anywhere
description: AI-driven click on the visible XFCE4 desktop. Describe a target ("the URL bar", "the Sign In button"), the pack screenshots → asks a vision model for coords → clicks via xdotool → loops until the goal is reached.
keywords: [helmdeck, vision, click, xdotool, computer use, MCP]
---

# `vision.click_anywhere`

The natural-language click pack. Caller supplies a `goal` — *"click the URL bar at the top of the Chromium window"* — plus a vision-capable `model`; the pack screenshots the visible desktop, asks the model for click coordinates, fires `xdotool click` at those coordinates, screenshots again, and loops until the model emits `done` (goal reached) or `max_steps` is hit. Every step's screenshot is recorded as an artifact so the operator can replay the trail in the *Artifacts* UI panel.

This is the agent-loop counterpart to the [`desktop.*` REST primitives](/PACKS) (documented in PR-C) — use `vision.click_anywhere` when the agent doesn't already know the pixel coordinates; use `desktop.click` when it does.

## Setup prerequisite

Vision packs run on a **desktop-mode** session (`HELMDECK_MODE=desktop` — set automatically by the pack via `SessionSpec`). The session boots with Chromium pre-launched on the XFCE4 display visible at `http://localhost:6080/vnc.html` (noVNC). Operators watching that URL will see the cursor move and the click happen.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `goal` | `string` | yes | — | Plain-English description of what to click. Be specific (`"the blue Submit button at the bottom of the form"`), not generic (`"the button"`). |
| `model` | `string` | yes | — | Vision-capable provider/model. e.g. `openrouter/anthropic/claude-haiku-4.5`, `openai/gpt-4o`. **Non-vision models silently fail** because the screenshot is invisible to them. |
| `max_steps` | `number` | no | `6` | Cap on the loop. Each step = one screenshot + one model call + one click (or no-op). |
| `_session_id` | `string` | yes (chained) | — | Pass the session id from the previous desktop-mode pack to keep the same Chromium / cursor state. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `completed` | `boolean` | `true` if the model emitted `done` within `max_steps`. |
| `steps` | `number` | Steps actually taken. |
| `final_action` | `object` | The last `vision.Action` the model emitted. Includes the `action` (`"click"`, `"done"`, `"none"`, etc.), pixel coordinates if any, and a free-text `reason`. |

The pack also writes one `step-NNN.png` artifact per step — accessible via the Artifacts API or the *Artifacts* UI panel keyed off the pack-call id.

## Vault credentials needed

**None** — the AI key for `model` resolves through the *AI Providers* UI panel (or env-var fast path for OpenRouter via `HELMDECK_OPENROUTER_API_KEY`), not the vault.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Use helmdeck__vision-click_anywhere with goal \"click the URL bar at the top of the Chromium window\" and model openrouter/anthropic/claude-haiku-4.5." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/vision.click_anywhere \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "goal":      "click the URL bar at the top of the Chromium window",
    "model":     "openrouter/anthropic/claude-haiku-4.5",
    "max_steps": 4
  }'
```

Live capture is dependent on which app is currently in focus on the desktop. The response shape returns `{completed, steps, final_action}` plus per-step screenshot artifacts.

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `goal` empty | `{"error":"invalid_input","message":"goal must not be empty"}` |
| `invalid_input` | `model` empty | `{"error":"invalid_input","message":"model must not be empty"}` |
| `session_unavailable` | Engine has no session executor | `{"error":"session_unavailable","message":"engine has no session executor"}` |
| `handler_failed` | Vision step failed (screenshot capture, model call, action parse) | `{"error":"handler_failed","message":"…"}` |
| `internal` | Pack registered without a gateway dispatcher | `vision.* registered without a gateway dispatcher` |

## Session chaining

**Required (creates if absent).** Compatible chains — desktop-mode only:
- Upstream: `desktop.run_app_and_screenshot` (launch an app first), `vision.fill_form_by_label` (fill before clicking submit).
- Downstream: `vision.extract_visible_text` (verify what you clicked), `desktop.screenshot` (capture the result).

Always pass `_session_id` from `repo.fetch` or the upstream desktop-mode pack — see the [Session chaining contract](/integrations/SKILLS#session-chaining-contract--read-before-chaining-fs--cmdrun--git) in SKILLS.md.

## Async behavior

Synchronous. Wall-clock = `steps × (screenshot + model_latency + xdotool)`. Each step on a Haiku-tier vision model is ~2–4 seconds; budget accordingly.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `vision.click_anywhere`.
- Source: [`internal/packs/builtin/vision_packs.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/vision_packs.go).
- ADR 027 — Vision pipeline.
- ADR 035 §2026 revision — Native computer-use tool routing (T807f).
- Companion packs: [`vision.extract_visible_text`](./extract-visible-text.md), [`vision.fill_form_by_label`](./fill-form-by-label.md), [`desktop.run_app_and_screenshot`](/PACKS).
