---
title: github.create_issue
description: File an issue on a GitHub repository using a vault-stored PAT. Returns the new issue number, API URL, and html_url.
keywords: [helmdeck, github, issue, REST, MCP]
---

# `github.create_issue`

The "file a GitHub issue" pack. Caller supplies a `repo` (`owner/name`), a `title`, optionally a `body` and `labels`, and the name of a vault credential containing a PAT with `repo` scope. The pack POSTs to `/repos/{repo}/issues` and returns the new issue's number plus URLs.

Stateless — no session needed. Hits `api.github.com` directly via the helmdeck control-plane (no `gh` CLI dependency).

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `repo` | `string` | yes | — | `owner/name` slug. |
| `title` | `string` | yes | — | Issue title. |
| `body` | `string` | no | `""` | Markdown body. |
| `labels` | `array` | no | `[]` | Label names. Labels must already exist on the repo or GitHub silently drops them. |
| `credential` | `string` | no | `github-token` | Vault credential name resolving to a PAT. Defaults to the canonical name; override only if you store the PAT under a different name. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `number` | `number` | The new issue number. |
| `url` | `string` | API URL. |
| `html_url` | `string` | Web URL — the link a human follows. |

The full GitHub issue JSON body is also passed through; the named fields above are the schema-required subset every caller can rely on.

## Vault credentials needed

**`github-token`** — type `api_key`, scoped to host pattern `api.github.com`, value is a Classic PAT with `repo` scope (or fine-grained PAT with **Issues: Read and write** on the target repo).

Add via the *Vault* panel before invoking. If absent, this pack returns `handler_failed` with the GitHub `401 Bad credentials` body — public-repo write operations require auth.

## Use it from your agent (OpenClaw chat-UI worked example)

<!-- TODO(maintainer): paste an OpenClaw chat-UI transcript here.
     Prompt to use: "Use helmdeck__github-create_issue against repo \"tosin2013/helmdeck-pack-doc-fixtures\" with title=\"Demo issue from helmdeck pack capture\", body=\"This issue is a demo artifact.\", labels=[\"doc-capture\"], credential=github-token." -->

> *OpenClaw chat capture pending.*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/github.create_issue \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "repo":       "tosin2013/helmdeck-pack-doc-fixtures",
    "title":      "Demo issue from helmdeck pack capture",
    "body":       "This issue is a demo artifact captured while documenting the github.create_issue pack. Safe to close.",
    "labels":     ["doc-capture"],
    "credential": "github-token"
  }'
```

Captured response (truncated to the schema-required fields):

```json
{
  "pack": "github.create_issue",
  "version": "v1",
  "output": {
    "number":   1,
    "url":      "https://api.github.com/repos/tosin2013/helmdeck-pack-doc-fixtures/issues/1",
    "html_url": "https://github.com/tosin2013/helmdeck-pack-doc-fixtures/issues/1"
  }
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `handler_failed` | `repo` or `title` missing | `{"error":"handler_failed","message":"repo and title are required"}` |
| `handler_failed` | PAT missing or invalid | `{"error":"handler_failed","message":"github API POST /repos/…/issues: 401 Bad credentials"}` |
| `handler_failed` | Repo doesn't exist or PAT lacks access | `{"error":"handler_failed","message":"github API POST /repos/…/issues: 404 Not Found"}` |
| `handler_failed` | A label in `labels` doesn't exist on the repo | GitHub silently drops the label; the issue is still created. Not an error. |

## Session chaining

**No session.** Stateless REST passthrough. Compatible with anything; commonly chained downstream of `web.scrape` ("scrape this page, file an issue summarizing it") or `repo.fetch` ("clone, find a TODO, file an issue against it").

## Async behavior

Synchronous. Round-trip is whatever GitHub's API takes (typically &lt;500 ms).

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `github.create_issue`.
- Source: [`internal/packs/builtin/github.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/github.go).
- ADR 034 — Core GitHub pack set.
- Vault setup: [`tutorials/install-ui-walkthrough.md`](/tutorials/install-ui-walkthrough#add-a-github-pat-github-token).
- Companion packs: [`github.list_issues`](./list-issues.md), [`github.post_comment`](./post-comment.md), [`github.search`](./search.md).
