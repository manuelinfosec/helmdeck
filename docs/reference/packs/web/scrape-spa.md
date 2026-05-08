---
title: web.scrape_spa
description: Selector-based extraction over a session-local Chromium. The deterministic counterpart to web.scrape — caller supplies CSS selectors, pack returns one value per field plus a list of which fields were missing.
keywords: [helmdeck, web, scrape_spa, css selector, CDP, MCP]
---

# `web.scrape_spa`

The selector-based scrape pack. The agent supplies a URL and a `fields` map naming CSS selectors to evaluate against the rendered DOM; the pack drives the session's Chromium via CDP and returns one value per field plus a `missing` list for any selectors that didn't resolve. Where [`web.scrape`](./scrape.md) is "give me the whole page as Markdown", this is "give me these specific 5 things; tell me which couldn't be found".

Partial-result handling matters because SPAs are flaky: one missing selector should not blow up an otherwise-useful scrape. The pack succeeds whenever **at least one field resolves**; the caller decides whether the missing list is acceptable. Total failure (zero fields resolved) surfaces as `handler_failed` so retries make sense.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `url` | `string` | yes | — | Absolute http(s) URL. The session's Chromium navigates to this URL via CDP. |
| `fields` | `object` | yes | — | Map of `{name: {selector, format}}`. Names are caller-chosen; `selector` is a CSS selector; `format` is `text` (default) or `html`. |
| `wait_ms` | `number` | no | `0` | Settle-delay (ms) after navigation before evaluating selectors. Useful for SPAs that hydrate post-DOMContentLoaded. |
| `_session_id` | `string` | yes (chained) | — | Standard chained-pack input. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `url` | `string` | Echo of the input. |
| `data` | `object` | `{name: extracted_value}` for every selector that resolved. |
| `missing` | `array` | Names of fields whose selectors didn't match. **Always present**, even when empty — the caller can branch on `len(missing) == 0` for "complete extraction". |

## Vault credentials needed

**None.** Runs entirely against the session's local Chromium.

## Use it from your agent (OpenClaw chat-UI worked example)

**Prompt** (sent in OpenClaw chat UI / `openclaw-cli agent`):

> Use helmdeck__web-scrape_spa to scrape https://example.com using fields {"heading": {"selector": "h1", "format": "text"}, "intro": {"selector": "p", "format": "text"}}. Tell me what you got back for each field and which (if any) were missing.

**Tool call** (1 call, no failures):

```json
{
  "name": "helmdeck__web-scrape_spa",
  "arguments": {
    "url": "https://example.com",
    "fields": {
      "heading": {
        "selector": "h1",
        "format": "text"
      },
      "intro": {
        "selector": "p",
        "format": "text"
      }
    }
  }
}
```

**Pack response** (truncated):

```json
{"pack":"web.scrape_spa","version":"v1","output":{"data":{"heading":"Example Domain","intro":"This domain is for use in documentation examples without needing permission. Avoid use in operations."},"missing":[],"url":"https://example.com"},"duration_ms":3353888174,"session_id":"f60ccd4b-bb5e-47d8-acef-af63abd80b6c"}
```

**Agent's final reply**:

> **Results**
> 
> - **heading:** “Example Domain”  
> - **intro:** “This domain is for use in documentation examples without needing permission. Avoid use in operations.”  
> 
> All requested fields were successfully extracted; none were missing.

*Verified via OpenClaw 2026.5.6 + helmdeck v0.9.0-dev + `openrouter/openai/gpt-oss-120b` on 2026-05-07 (cost: $0.1717).*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/web.scrape_spa \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "url": "https://example.com",
    "fields": {
      "heading":     {"selector":"h1","format":"text"},
      "intro":       {"selector":"p","format":"text"},
      "missing_one": {"selector":"#nope","format":"text"}
    }
  }'
```

Captured response (note `missing` carries the unresolved field; `data` carries the rest):

```json
{
  "pack": "web.scrape_spa",
  "version": "v1",
  "output": {
    "data": {
      "heading": "Example Domain",
      "intro": "This domain is for use in documentation examples without needing permission. Avoid use in operations."
    },
    "missing": ["missing_one"],
    "url": "https://example.com"
  },
  "duration_ms": 304435549362,
  "session_id": "1a9920e2-a371-4923-8de6-2deaad0fba85"
}
```

> ⚠️ **Cold-start latency observed at capture**: the response above took 304 seconds on a freshly-spun-up session container. Subsequent calls reusing the same session return in 1–3 seconds. The slowness is in CDP-Chromium handshake on first nav, not in selector evaluation. Tracked in issues filed against the repo if reproducible.

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `url` missing or empty | `{"error":"invalid_input","message":"url: must have required properties …"}` |
| `invalid_input` | `fields` missing or empty | `{"error":"invalid_input","message":"fields map must not be empty"}` |
| `invalid_input` | A field's `selector` is empty | `{"error":"invalid_input","message":"field \"name\": selector required"}` |
| `session_unavailable` | Engine has no CDP factory | `{"error":"session_unavailable","message":"engine has no CDP factory"}` |
| `handler_failed` | Navigate failed (network, redirect loop, render crash) | `{"error":"handler_failed","message":"navigate https://…: …"}` |

## Session chaining

**Optional session.** This pack creates a session if none is supplied; chained workflows pass `_session_id` from a previous pack to reuse the same Chromium (and any cookies / login state already present). Especially valuable when a prior `vision.fill_form_by_label` logged into a SPA — `web.scrape_spa` against the same session sees the post-login DOM.

## Async behavior

Synchronous. Caller-supplied `wait_ms` is the only knob; the round-trip cap is the session's overall timeout (default 5 minutes).

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `web.scrape_spa`.
- Source: [`internal/packs/builtin/scrape_spa.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/scrape_spa.go).
- Companion packs: [`web.scrape`](./scrape.md) (no selectors), [`web.test`](./test.md) (NL-driven), [`browser.interact`](../browser/interact.md) (deterministic actions).
