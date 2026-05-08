---
title: vision.extract_visible_text
description: Screenshot the visible XFCE4 desktop and ask a vision model to transcribe every readable piece of text. Useful for "what's on the screen now" queries and verifying the result of a prior desktop action.
keywords: [helmdeck, vision, OCR, computer use, MCP]
---

# `vision.extract_visible_text`

The "tell me what's on the screen" pack. Caller supplies a vision-capable `model`; the pack screenshots the visible desktop and asks the model to transcribe every piece of readable text, joined with newlines. Use it as a verification step after a `vision.click_anywhere` or `vision.fill_form_by_label` call ("did the post-submit page load? what's the order ID?"), or for general "what app is showing right now" queries.

Single-step — no internal loop. Returns immediately with the transcription text.

## Setup prerequisite

Vision packs run on a **desktop-mode** session (`HELMDECK_MODE=desktop` — set automatically by the pack via `SessionSpec`).

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `model` | `string` | yes | — | Vision-capable provider/model. e.g. `openrouter/anthropic/claude-haiku-4.5`, `openai/gpt-4o`. |
| `_session_id` | `string` | yes (chained) | — | Pass the session id from the upstream desktop-mode pack. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `text` | `string` | Newline-joined transcription of every readable piece of text on the visible desktop. |
| `model` | `string` | Echo — the model used. |

## Vault credentials needed

**None** — the AI key for `model` resolves through the *AI Providers* UI panel.

## Use it from your agent (OpenClaw chat-UI worked example)

**Prompt** (sent in OpenClaw chat UI / `openclaw-cli agent`):

> Use helmdeck__vision-extract_visible_text against the visible desktop. Use model openrouter/anthropic/claude-haiku-4.5. Tell me what text the vision model transcribed (truncate to 200 chars if very long).

**Tool call** (1 call, no failures):

```json
{
  "name": "helmdeck__vision-extract_visible_text",
  "arguments": {
    "model": "openrouter/anthropic/claude-haiku-4.5"
  }
}
```

**Pack response** (truncated):

```json
{"pack":"vision.extract_visible_text","version":"v1","output":{"model":"openrouter/anthropic/claude-haiku-4.5","text":""},"duration_ms":9509607044,"session_id":"09abae0f-3383-4a03-9fde-1655c18eddc9"}
```

**Agent's final reply**:

> The vision extraction returned an empty string, indicating no readable text was detected on the visible desktop.

*Verified via OpenClaw 2026.5.6 + helmdeck v0.9.0-dev + `openrouter/openai/gpt-oss-120b` on 2026-05-07 (cost: $0.1835).*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/vision.extract_visible_text \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"model":"openrouter/anthropic/claude-haiku-4.5"}'
```

Response shape:

```json
{
  "pack": "vision.extract_visible_text",
  "version": "v1",
  "output": {
    "text":  "<every readable line on the desktop, newline-joined>",
    "model": "openrouter/anthropic/claude-haiku-4.5"
  },
  "duration_ms": …,
  "session_id": "…"
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `model` empty | `{"error":"invalid_input","message":"model must not be empty"}` |
| `session_unavailable` | Engine has no session executor | `{"error":"session_unavailable","message":"engine has no session executor"}` |
| `handler_failed` | Screenshot capture failed (Xvfb died) | `{"error":"handler_failed","message":"…"}` |
| `handler_failed` | Model returned unparseable response | `{"error":"handler_failed","message":"could not parse model response: …"}` |

## Session chaining

**Required (creates if absent).** Mostly used as a verification step after a `vision.*` action. Compatible chains:

- Upstream: any `vision.*` or `desktop.*` pack — chain to verify the result.
- Downstream: nothing in particular — the text is the output the agent uses to make the next decision.

## Async behavior

Synchronous. One screenshot + one model call. Wall-clock ≈ 1–3 seconds on a Haiku-tier model.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `vision.extract_visible_text`.
- Source: [`internal/packs/builtin/vision_packs.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/vision_packs.go).
- ADR 027 — Vision pipeline.
- Companion packs: [`vision.click_anywhere`](./click-anywhere.md), [`vision.fill_form_by_label`](./fill-form-by-label.md), [`doc.ocr`](../doc/ocr.md) (image-bytes-in OCR alternative).
