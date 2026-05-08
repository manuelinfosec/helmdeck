---
title: vision.fill_form_by_label
description: Fill a form on the visible XFCE4 desktop by matching field labels to their values. The pack iterates one field at a time, asking a vision model to locate each label, then xdotool-types the value into the matched field.
keywords: [helmdeck, vision, form, xdotool, computer use, MCP]
---

# `vision.fill_form_by_label`

The "fill in this form" pack. Caller supplies a `fields` map of `{label: value}` pairs and a vision-capable `model`; the pack iterates each label in alphabetical order, screenshots the desktop, asks the model to locate the matching field, types the value via xdotool, and moves to the next field. Returns the list of fields that were successfully filled and the total step count.

This is the messiest of the three vision packs because the action loop must track per-field progress. Pairs naturally with [`vision.click_anywhere`](./click-anywhere.md) (to submit afterward) and [`vision.extract_visible_text`](./extract-visible-text.md) (to verify the post-submit state).

## Setup prerequisite

Vision packs run on a **desktop-mode** session. The Chromium window must already be at the form URL — typically achieved by `vision.click_anywhere` to focus the URL bar, then `desktop.type` + `desktop.key` Enter to navigate.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `fields` | `object` | yes | — | `{label: value}` map. Labels are alphabetized internally for deterministic iteration. |
| `model` | `string` | yes | — | Vision-capable provider/model. |
| `max_steps` | `number` | no | `12` | Cap on the **total** step count (across all fields). Forms with N fields typically take ~N+1 steps. |
| `_session_id` | `string` | yes (chained) | — | Pass the session id from the upstream desktop-mode pack. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `completed` | `boolean` | `true` only when **every** field was filled within `max_steps`. |
| `fields_filled` | `array` | Labels successfully filled, in alphabetical order. |
| `steps` | `number` | Total steps used. |

## Vault credentials needed

**None** — the AI key for `model` resolves through the *AI Providers* UI panel.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Navigate to https://httpbin.org/forms/post on the visible desktop, then use helmdeck__vision-fill_form_by_label with fields={\"customer name\":\"Alice\",\"telephone\":\"555-0100\"} and model=openrouter/anthropic/claude-haiku-4.5." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/vision.fill_form_by_label \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "fields": {"customer name":"Alice","telephone":"555-0100"},
    "model":  "openrouter/anthropic/claude-haiku-4.5"
  }'
```

Response shape:

```json
{
  "pack": "vision.fill_form_by_label",
  "version": "v1",
  "output": {
    "completed":     true,
    "fields_filled": ["customer name","telephone"],
    "steps":         3
  },
  "duration_ms": …,
  "session_id": "…"
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `model` empty | `{"error":"invalid_input","message":"model must not be empty"}` |
| `invalid_input` | `fields` map empty | `{"error":"invalid_input","message":"fields must contain at least one entry"}` |
| `session_unavailable` | Engine has no session executor | `{"error":"session_unavailable","message":"engine has no session executor"}` |
| `handler_failed` | Vision step failed (screenshot, model call, parse) | `{"error":"handler_failed","message":"…"}` |

When `completed: false` is returned without a top-level error, the pack ran out of steps before all fields were filled. Inspect `fields_filled` to see how far it got and consider raising `max_steps`.

## Session chaining

**Required (creates if absent).** Typical chain:

```
desktop.screenshot → vision.click_anywhere (focus URL bar) → desktop.type (URL) →
desktop.key (Enter) → vision.fill_form_by_label → vision.click_anywhere (Submit) →
vision.extract_visible_text (verify success)
```

Always pass `_session_id` through every step — see the [Session chaining contract](/integrations/SKILLS#session-chaining-contract--read-before-chaining-fs--cmdrun--git).

## Async behavior

Synchronous. Wall-clock = `len(fields) × (screenshot + model_latency + xdotool)`. A 5-field form on a Haiku-tier model is typically 15–25 seconds.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `vision.fill_form_by_label`.
- Source: [`internal/packs/builtin/vision_packs.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/vision_packs.go).
- ADR 027 — Vision pipeline.
- Companion packs: [`vision.click_anywhere`](./click-anywhere.md), [`vision.extract_visible_text`](./extract-visible-text.md).
