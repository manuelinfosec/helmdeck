---
title: github.create_release
description: Create a GitHub release for a tag. Returns the release id, API URL, and html_url. The tag is created automatically if it doesn't exist (against the default branch's tip).
keywords: [helmdeck, github, release, tag, REST, MCP]
---

# `github.create_release`

The "cut a GitHub release" pack. Caller supplies `repo` and `tag` (and optionally `name`, `body`, and a `draft` flag); the pack POSTs to `/repos/{repo}/releases`. The tag is created automatically pointing at the default branch's tip if it doesn't already exist — but for production releases you'll typically push the tag yourself first so the SHA is reproducible.

Stateless.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `repo` | `string` | yes | — | `owner/name`. |
| `tag` | `string` | yes | — | Tag name. Created against the default branch's tip if absent. |
| `name` | `string` | no | `tag` | Release title shown on the GitHub releases page. |
| `body` | `string` | no | `""` | Markdown body — typically the changelog snippet for this version. |
| `draft` | `boolean` | no | `false` | Create as a draft (not visible to non-collaborators). |
| `credential` | `string` | no | `github-token` | PAT with **Contents: Write** + **Webhooks: Write** (fine-grained) or `repo` (Classic). |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `id` | `number` | Release id. |
| `url` | `string` | API URL. |
| `html_url` | `string` | Web URL — the public release page (or draft URL when `draft: true`). |
| `upload_url` | `string` | URL template for uploading release assets (zip/tar/binary). |

## Vault credentials needed

**`github-token`** — write scope.

## Use it from your agent (OpenClaw chat-UI worked example)

**Prompt** (sent in OpenClaw chat UI / `openclaw-cli agent`):

> Use helmdeck__github-create_release against repo "tosin2013/helmdeck-pack-doc-fixtures" with tag="v0.0.1-demo", name="Demo release", body="Captured during helmdeck pack doc PR-B.", credential=github-token. Tell me the release id and html_url.

**Tool call** (1 call, no failures):

```json
{
  "name": "helmdeck__github-create_release",
  "arguments": {
    "repo": "tosin2013/helmdeck-pack-doc-fixtures",
    "tag": "v0.0.1-demo",
    "name": "Demo release",
    "body": "Captured during helmdeck pack doc PR-B.",
    "credential": "github-token"
  }
}
```

**Pack response** (truncated):

```json
{"error":"handler_failed","message":"github API POST /repos/tosin2013/helmdeck-pack-doc-fixtures/releases: 422 Validation Failed"}
```

**Agent's final reply**:

> The release could not be created – the GitHub API returned a **422 Validation Failed** error. This usually means the tag `v0.0.1‑demo` already exists or the request is missing required data.

*Verified via OpenClaw 2026.5.6 + helmdeck v0.9.0-dev + `openrouter/openai/gpt-oss-120b` on 2026-05-07 (cost: $0.1623).*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/github.create_release \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "repo":       "tosin2013/helmdeck-pack-doc-fixtures",
    "tag":        "v0.0.1-demo",
    "name":       "Demo release",
    "body":       "Captured during helmdeck pack doc PR-B.",
    "credential": "github-token"
  }'
```

Captured response (truncated to schema-required fields):

```json
{
  "pack": "github.create_release",
  "version": "v1",
  "output": {
    "id":       319504238,
    "url":      "https://api.github.com/repos/tosin2013/helmdeck-pack-doc-fixtures/releases/319504238",
    "html_url": "https://github.com/tosin2013/helmdeck-pack-doc-fixtures/releases/tag/v0.0.1-demo"
  }
}
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `handler_failed` | `repo` or `tag` missing | `{"error":"handler_failed","message":"repo and tag are required"}` |
| `handler_failed` | Tag already has a release | `github API POST …: 422 Validation Failed` (with body explaining the conflict) |
| `handler_failed` | PAT lacks write access | `github API POST …: 403 …` |

## Session chaining

**No session.** Common chain: `git.commit` → `repo.push` → `github.create_release` (the agent commits and tags within a session, pushes, then cuts the release).

## Async behavior

Synchronous. The tag-on-default-branch resolution and release insert is fast (&lt;500 ms typically).

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `github.create_release`.
- Source: [`internal/packs/builtin/github.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/github.go).
- ADR 034 — Core GitHub pack set.
- Companion packs: [`git.commit`](../git/commit.md), [`repo.push`](/PACKS).
