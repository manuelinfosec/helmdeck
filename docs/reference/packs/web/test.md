---
title: web.test
description: Natural-language browser testing — the agent describes what to verify in plain English and the pack drives Playwright MCP through it, optionally checking a list of substring assertions against the final accessibility snapshot.
keywords: [helmdeck, web, test, playwright, MCP, accessibility]
---

# `web.test`

The natural-language browser-testing pack. Caller hands in a URL plus a plain-English instruction — *"log in as alice, open Settings, make sure 2FA is enabled"* — plus an LLM model name; the pack drives Playwright MCP through it step-by-step, using the gateway model to decompose each turn's accessibility-tree snapshot into the next action. Optionally, the caller supplies a list of substring `assertions` checked against the final snapshot.

Why Playwright MCP and not selectors? Accessibility-tree snapshots return structured `[ref=eN]` identifiers the model addresses directly — much smaller token footprint than CSS selectors, much more deterministic for weak models, no vision step required.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `url` | `string` | yes | — | Absolute http(s) URL to test against. Egress-guarded. |
| `instruction` | `string` | yes | — | Plain-English description of what the agent should verify. |
| `model` | `string` | yes | — | Provider/model that drives the action loop. e.g. `openrouter/openai/gpt-oss-120b`, `openai/gpt-4o`, `anthropic/claude-sonnet-4.6`. |
| `max_steps` | `number` | no | `8` | Cap on the plan loop (excludes the seed `browser_navigate` + initial snapshot). |
| `assertions` | `array` | no | `[]` | Substrings to check against the final accessibility snapshot. Substring match, case-sensitive. |
| `_session_id` | `string` | yes (chained) | — | Standard chained input. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `url` | `string` | Echo. |
| `completed` | `boolean` | `true` only when the model emitted `done` AND all assertions passed. |
| `steps` | `array` | Per-step trace: `{tool, arguments, result, is_error, reasoning}`. |
| `steps_used` | `number` | Total steps emitted (including seed nav + final snapshot). |
| `final_snapshot` | `string` | The last accessibility-tree dump after the last action. Used for assertions. |
| `assertions_passed` | `boolean` | `true` if every supplied assertion was a substring of `final_snapshot`. Always `true` when `assertions` is empty. |
| `reason` | `string` | Plain-English exit reason (`"model emitted done"`, `"max_steps (8) reached without done"`, `"egress guard blocked mid-test navigation"`, etc.). |

## Vault credentials needed

**None.** The model is selected by the `model` argument and resolved through the AI provider keys configured in the *AI Providers* UI panel.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Use helmdeck__web-test to verify that https://example.com loads correctly. Instruction: \"navigate to the page, confirm the heading 'Example Domain' is present, then exit\". Use model openrouter/openai/gpt-oss-120b. assertions=[\"Example Domain\"]." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/web.test \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "url":         "https://example.com",
    "instruction": "navigate to the page and confirm the heading Example Domain is present",
    "model":       "openrouter/openai/gpt-oss-120b",
    "max_steps":   3,
    "assertions":  ["Example Domain"]
  }'
```

Live capture is dependent on your gateway-model configuration; the response shape is the schema above with `steps` populated by per-action records.

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `url` empty | `{"error":"invalid_input","message":"url is required"}` |
| `invalid_input` | `instruction` empty | `{"error":"invalid_input","message":"instruction is required"}` |
| `invalid_input` | `model` empty | `{"error":"invalid_input","message":"model is required (provider/model)"}` |
| `invalid_input` | URL resolves to a blocked range | `{"error":"invalid_input","message":"egress denied: …"}` |
| `session_unavailable` | Sidecar built without Playwright MCP | `session has no Playwright MCP endpoint; rebuild the sidecar with T807a` |
| `handler_failed` | Playwright MCP `Initialize` failed after retries | `playwright mcp initialize (after retries): …` |
| `handler_failed` | Mid-test action returned an error from MCP | `<tool> (step N): …` |
| `handler_failed` | Model returned an unparseable plan | `plan step N: no parseable plan JSON in model response: …` |
| `internal` | Pack registered without a gateway dispatcher | `web.test registered without a gateway dispatcher` |

## Session chaining

**Required session.** Always chained — each `web.test` run gets a fresh session by default (no cookie leak between tests). Pass `_session_id` to reuse a session; useful for "log in via fill_form_by_label, then web.test the post-login state".

## Async behavior

Synchronous, but the wall-clock scales with `max_steps × per-step LLM latency`. A 5-step run against an open-weight model can take 30–60 seconds. For CI usage, cap `max_steps` aggressively and rely on the substring `assertions` rather than waiting for the model to emit `done` voluntarily.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `web.test`.
- Source: [`internal/packs/builtin/webtest.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/webtest.go).
- ADR 035 — MCP Server Hosting & Pack Evolution.
- Companion packs: [`web.scrape`](./scrape.md) (no LLM), [`web.scrape_spa`](./scrape-spa.md) (selector-based), [`browser.interact`](../browser/interact.md) (deterministic, agent-authored steps).
