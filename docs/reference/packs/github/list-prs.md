---
title: github.list_prs
description: List pull requests on a GitHub repository. Filter by state, head branch, and base branch. Returns the GitHub PR array verbatim plus a count.
keywords: [helmdeck, github, pull request, REST, MCP]
---

# `github.list_prs`

The "show me the open PRs" pack. Caller supplies `repo` and optional filters (`state`, `head`, `base`); the pack GETs `/repos/{repo}/pulls` and returns the PR array wrapped in `{prs: [...], count: N}`.

Stateless. Public-repo reads work without a credential.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `repo` | `string` | yes | — | `owner/name` slug. |
| `state` | `string` | no | `"open"` | One of `open`, `closed`, `all`. |
| `head` | `string` | no | `""` | Filter by head branch (`user:branch`). |
| `base` | `string` | no | `""` | Filter by base branch. |
| `credential` | `string` | no | `github-token` | Optional. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `prs` | `array` | GitHub PR objects verbatim. Each carries `number`, `title`, `state`, `head.ref`, `base.ref`, `user.login`, `created_at`, `merged_at`, `html_url`, `draft`, `mergeable_state`, etc. |
| `count` | `number` | `len(prs)`. Paginated upstream — for repos with &gt;30 PRs this is page 1's count. |

## Vault credentials needed

**Optional.** `github-token` raises the rate limit and is required for private repos.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Use helmdeck__github-list_prs against repo \"tosin2013/helmdeck-pack-doc-fixtures\" with state=open and credential=github-token." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/github.list_prs \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "repo":       "tosin2013/helmdeck-pack-doc-fixtures",
    "state":      "open",
    "credential": "github-token"
  }'
```

Captured response (no PRs in fixture repo):

```json
{
  "pack": "github.list_prs",
  "version": "v1",
  "output": {
    "count": 0,
    "prs":   []
  }
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `handler_failed` | `repo` missing | `{"error":"handler_failed","message":"repo is required"}` |
| `handler_failed` | Repo doesn't exist or PAT lacks access | `github API GET …: 404 Not Found` |
| `handler_failed` | Rate-limited | `github API GET …: 403 …` |

## Session chaining

**No session.** Stateless.

## Async behavior

Synchronous.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `github.list_prs`.
- Source: [`internal/packs/builtin/github.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/github.go).
- ADR 034 — Core GitHub pack set.
- Companion packs: [`github.list_issues`](./list-issues.md), [`github.search`](./search.md), [`github.post_comment`](./post-comment.md) (works on PRs too — they're issues to GitHub).
