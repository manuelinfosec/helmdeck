---
title: github.search
description: Search GitHub code, issues, repositories, or commits. Returns the top page of results plus a total_count.
keywords: [helmdeck, github, search, REST, MCP]
---

# `github.search`

The "search GitHub" pack. Caller supplies a `query` (GitHub search syntax — `repo:`, `is:issue`, `language:`, `path:`, `extension:`, `stars:>N`, etc.) and a `type` of `issues` (default), `code`, `repositories`, or `commits`. The pack GETs the corresponding `/search/*` endpoint and returns the GitHub response verbatim.

Stateless.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `query` | `string` | yes | — | GitHub search syntax. URL-encoded internally — the agent supplies plain space-separated terms. |
| `type` | `string` | no | `"issues"` | One of `issues`, `code`, `repositories`, `commits`. Closed-set; anything else returns `handler_failed`. |
| `credential` | `string` | no | `github-token` | Optional for `issues`/`repositories`. **Required** for `code` (GitHub gates code search to authenticated users only). |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `total_count` | `number` | Total matches. |
| `items` | `array` | First page of results (max 30 items per call). Each item's shape depends on `type` — issues come back with `title`/`body`/`html_url`, code with `path`/`html_url`/`text_matches[]`. |

> ⚠️ **Indexing latency observed at capture**: GitHub's search index is updated asynchronously. Issues and commits created within the last few minutes may not appear in search results yet. The `github.list_issues` / `gh api repos/.../issues` endpoints are eventually-consistent on a different (faster) timeline. If you need "find this thing I just created", prefer the list endpoints over `github.search`.

## Vault credentials needed

**Optional for `issues`/`repositories`/`commits`.** **Required for `type: "code"`** — GitHub returns `422` for unauthenticated code searches.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Use helmdeck__github-search with query=\"repo:tosin2013/helmdeck-pack-doc-fixtures is:issue\" type=issues credential=github-token." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/github.search \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "query":      "repo:tosin2013/helmdeck-pack-doc-fixtures is:issue",
    "type":       "issues",
    "credential": "github-token"
  }'
```

Captured response (just-created issue may not have indexed yet):

```json
{
  "pack": "github.search",
  "version": "v1",
  "output": {
    "total_count":        0,
    "items":              [],
    "incomplete_results": false
  }
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `handler_failed` | `query` missing | `{"error":"handler_failed","message":"query is required"}` |
| `handler_failed` | `type` outside the closed set | `{"error":"handler_failed","message":"type must be one of: code, issues, repositories, commits"}` |
| `handler_failed` | Unauthenticated `type: "code"` | `github API GET /search/code: 422 Validation Failed` |
| `handler_failed` | Rate-limited (search has tighter limits than core API: 30 req/min authenticated) | `github API GET …: 403 API rate limit exceeded …` |

## Session chaining

**No session.** Stateless. Common chain — search to discover something, then act on it: `github.search` (find issues mentioning X) → `github.post_comment` (reply to each).

## Async behavior

Synchronous.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `github.search`.
- Source: [`internal/packs/builtin/github.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/github.go).
- ADR 034 — Core GitHub pack set.
- GitHub search syntax: <https://docs.github.com/en/search-github/searching-on-github>.
- Companion packs: [`github.list_issues`](./list-issues.md), [`github.list_prs`](./list-prs.md).
