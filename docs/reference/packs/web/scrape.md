---
title: web.scrape
description: Firecrawl-backed clean-markdown scrape. No CSS selectors required — point at any URL and get back parsed content as clean Markdown.
keywords: [helmdeck, web, scrape, firecrawl, MCP]
---

# `web.scrape`

The zero-selectors scrape pack. The agent supplies a URL and gets back clean Markdown — Firecrawl handles the headless render, content extraction, and Markdown conversion. Where [`web.scrape_spa`](./scrape-spa.md) requires CSS selectors and runs against the session's CDP-driven Chromium, this pack hands the URL off to a Firecrawl service that already knows how to deal with SPAs, ad-walls, and most modern web shapes. Use it whenever the agent doesn't already have the DOM mapped.

## Setup prerequisite

`web.scrape` only works when the Firecrawl overlay is running and the env-var toggle is set:

```bash
docker compose -f deploy/compose/compose.yaml \
               -f deploy/compose/compose.firecrawl.yml \
               --env-file deploy/compose/.env.local up -d

# in deploy/compose/.env.local:
HELMDECK_FIRECRAWL_ENABLED=true
```

When the toggle is off the pack returns `invalid_input: web.scrape is disabled (set HELMDECK_FIRECRAWL_ENABLED=true)`.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `url` | `string` | yes | — | Absolute http(s) URL. Egress-guarded — RFC 1918 / metadata IPs / loopback are rejected before Firecrawl is contacted. |
| `formats` | `array` | no | `["markdown"]` | Subset of `markdown`, `html`, `rawHtml`, `links`. `markdown` is required for the output schema's `markdown` field. |
| `wait_ms` | `number` | no | `0` | Settle-delay (ms) Firecrawl applies after navigation before snapshotting. Useful for SPAs that hydrate post-DOMContentLoaded. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `url` | `string` | Echo of the input. |
| `markdown` | `string` | Always present. Cleaned Markdown of the page's primary content. |
| `html` | `string` | Only when `formats` requested it. |
| `links` | `array` | Only when `formats` requested it. Discovered outbound links. |
| `title` | `string` | From Firecrawl's metadata pass. |
| `status` | `number` | Upstream HTTP status. |

## Vault credentials needed

**None.** The Firecrawl service runs on the private `baas-net` and accepts unauthenticated calls from the control plane.

## Use it from your agent (OpenClaw chat-UI worked example)

**Prompt** (sent in OpenClaw chat UI / `openclaw-cli agent`):

> Use the helmdeck__web-scrape tool to scrape https://example.com to clean markdown. Tell me the page title and the first sentence of the markdown.

**Tool call** (1 call, no failures):

```json
{
  "name": "helmdeck__web-scrape",
  "arguments": {
    "url": "https://example.com"
  }
}
```

**Pack response** (truncated):

```json
{"pack":"web.scrape","version":"v1","output":{"markdown":"Example Domain\n==============\n\nThis domain is for use in documentation examples without needing permission. Avoid use in operations.\n\n[Learn more](https://iana.org/domains/example)","status":200,"title":"Example Domain","url":"https://example.com"},"duration_ms":115733635}
```

**Agent's final reply**:

> - **Page title:** Example Domain  
> - **First sentence of the markdown:** “This domain is for use in documentation examples without needing permission.”

*Verified via OpenClaw 2026.5.6 + helmdeck v0.9.0-dev + `openrouter/openai/gpt-oss-120b` on 2026-05-07 (cost: $0.1685).*

## Developer reference (`curl`)

Mint a JWT first:

```bash
ADMIN_PW=$(grep HELMDECK_ADMIN_PASSWORD /root/helmdeck/deploy/compose/.env.local | cut -d= -f2)
JWT=$(curl -fsS -X POST http://localhost:3000/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d "{\"username\":\"admin\",\"password\":\"${ADMIN_PW}\"}" \
  | python3 -c 'import sys,json;print(json.load(sys.stdin)["token"])')
```

Happy path:

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/web.scrape \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com"}'
```

Captured response:

```json
{
  "pack": "web.scrape",
  "version": "v1",
  "output": {
    "markdown": "Example Domain\n==============\n\nThis domain is for use in documentation examples without needing permission. Avoid use in operations.\n\n[Learn more](https://iana.org/domains/example)",
    "status": 200,
    "title": "Example Domain",
    "url": "https://example.com"
  },
  "duration_ms": 104911180
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `url` missing or empty | `{"error":"invalid_input","message":"url is required"}` |
| `invalid_input` | `HELMDECK_FIRECRAWL_ENABLED` is unset/false | `web.scrape is disabled; set HELMDECK_FIRECRAWL_ENABLED=true …` |
| `invalid_input` | `formats` includes a value outside `markdown`/`html`/`rawHtml`/`links` | `{"error":"invalid_input","message":"unsupported format \"pdf\"; use markdown, html, rawHtml, or links"}` |
| `invalid_input` | URL resolves to a blocked range (metadata, RFC 1918, loopback) | `{"error":"invalid_input","message":"egress denied: security: destination is in a blocked address range: 169.254.169.254 is in 169.254.169.254/32"}` |
| `handler_failed` | Firecrawl returns non-200 (incl. `success: false` body) | `{"error":"handler_failed","message":"firecrawl 500: …"}` |
| `handler_failed` | Firecrawl returns empty markdown (bot-challenge, blank body) | `{"error":"handler_failed","message":"firecrawl returned empty markdown for https://… (status=200)"}` |

## Session chaining

**No session.** Stateless. Chains freely upstream of `doc.parse` (download a page bytestream and feed it through layout-aware parsing), `content.ground` (rewrite a Markdown blob the pack just produced), or `slides.narrate` (turn a scraped page into a narrated deck).

## Async behavior

Synchronous. Firecrawl's own per-request timeout is generous; helmdeck caps the round-trip at 90 seconds. Heavy SPAs may approach that cap; for those, use `web.scrape_spa` with explicit `wait_ms`.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `web.scrape`.
- Source: [`internal/packs/builtin/web_scrape.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/web_scrape.go).
- ADR 035 — MCP Server Hosting & Pack Evolution (Firecrawl overlay rationale).
- Companion packs: [`web.scrape_spa`](./scrape-spa.md), [`web.test`](./test.md), [`browser.screenshot_url`](../browser/screenshot-url.md).
