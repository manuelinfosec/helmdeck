---
title: slides.render
description: Convert Marp markdown to PDF, PPTX, or HTML via the Marp CLI bundled in the sidecar. Returns an artifact key for the rendered file.
keywords: [helmdeck, slides, render, marp, pdf, pptx, MCP]
---

# `slides.render`

The "Marp markdown → PDF/PPTX/HTML" pack. Caller hands in a Marp deck (frontmatter + `---`-delimited slides) and a target `format`; the pack invokes the Marp CLI inside the sidecar (`marp --stdin --allow-local-files`), streams the binary output to the artifact store, and returns the artifact key + size. The deck never touches disk inside the sidecar — input via stdin, output streamed.

For narrated video output (MP4 with TTS over each slide), see [`slides.narrate`](./narrate.md). For just the static deck, this is the right pack.

## Inputs

| Field | Type | Required | Default | Notes |
|---|---|---|---|---|
| `markdown` | `string` | yes | — | Marp deck. Must start with `---\nmarp: true\n---` frontmatter for the directives to apply. Slides separated by `---`. |
| `format` | `string` | no | `"pdf"` | Closed-set: `pdf`, `pptx`, `html`. Picks the Marp output codec. |

## Outputs

| Field | Type | Notes |
|---|---|---|
| `format` | `string` | Echo of the chosen format. |
| `artifact_key` | `string` | `slides.render/<rand>-deck.{pdf|pptx|html}` — pass to `/api/v1/artifacts/<key>` to retrieve. |
| `size` | `number` | Bytes. |

## Vault credentials needed

**None.** Marp runs entirely in-sidecar.

## Use it from your agent (OpenClaw chat-UI worked example)

**Prompt** (sent in OpenClaw chat UI / `openclaw-cli agent`):

> Use helmdeck__slides-render to convert this Marp markdown to PDF: "---\nmarp: true\n---\n# Hello helmdeck\n\n---\n\n# Thanks for listening". Tell me the artifact_key and the size in bytes.

**Tool call** (1 call, no failures):

```json
{
  "name": "helmdeck__slides-render",
  "arguments": {
    "format": "pdf",
    "markdown": "---\nmarp: true\n---\n# Hello helmdeck\n\n---\n\n# Thanks for listening"
  }
}
```

**Pack response** (truncated):

```json
{"pack":"slides.render","version":"v1","output":{"artifact_key":"slides.render/7d549cd238a7bc34-deck.pdf","format":"pdf","size":7902},"artifacts":[{"key":"slides.render/7d549cd238a7bc34-deck.pdf","url":"http://garage:3900/helmdeck-artifacts/slides.render/7d549cd238a7bc34-deck.pdf?X-Amz-Algorithm=AWS4-HMAC-SHA256\u0026X-Amz-Credential=GKaef177ebe9fd814dc103b84b%2F20260508%2Fgarage%2Fs3%2Faws4_request\u0026X-Amz-Date=20260508T152130Z\u0026X-Amz-Expires=900\u0026X-Amz-SignedHeaders=host\u0026X-Amz-Signature=b2f7c01cfc3a39c4365d0bd5f79dfff7d49e972497f4eff7b535b8eaccf3c544","size":7902,"content_type":"application/pdf","created_at":"2026-05-08T15:21:30.754346118Z","pack":"slides.render"}],"duration_ms":18808030860,"session_id":"586c8ee9-3529-4128-9cbf-541e0cc20de0"}
```

**Agent's final reply**:

> The PDF was created successfully.
> 
> - **artifact_key:** `slides.render/7d549cd238a7bc34-deck.pdf`  
> - **size:** `7902` bytes

*Verified via OpenClaw 2026.5.6 + helmdeck v0.9.0-dev + `openrouter/openai/gpt-oss-120b` on 2026-05-07 (cost: $0.0013).*

## Developer reference (`curl`)

```bash
curl -fsS -X POST http://localhost:3000/api/v1/packs/slides.render \
  -H "Authorization: Bearer $JWT" -H 'Content-Type: application/json' \
  -d '{
    "markdown": "---\nmarp: true\n---\n# Hello helmdeck\n\n---\n\n# Thanks for listening",
    "format":   "pdf"
  }'
```

Response shape:

```json
{
  "pack": "slides.render",
  "version": "v1",
  "output": {
    "format":       "pdf",
    "artifact_key": "slides.render/abc123-deck.pdf",
    "size":         98304
  }
}
```

Retrieve the artifact:

```bash
curl -fsS -H "Authorization: Bearer $JWT" \
  "http://localhost:3000/api/v1/artifacts/slides.render/abc123-deck.pdf" \
  -o deck.pdf
```

## Error codes

| Code | Triggers | Captured response |
|---|---|---|
| `invalid_input` | `markdown` empty | `markdown is required` |
| `invalid_input` | `format` outside the closed set | `unsupported format "docx"; use pdf, pptx, or html` |
| `handler_failed` | Marp CLI exit non-zero (malformed deck, missing fonts) | `marp exit N: <stderr>` (truncated to 1024 chars) |
| `artifact_failed` | Object store write failed | `artifact upload failed: …` |

## Session chaining

**Required (creates if absent).** Stateless logically — the deck is in the input, the output is an artifact — but the Marp render runs inside a session sidecar. Typically not chained; one-shot.

## Async behavior

Synchronous. PDF rendering of a 10-slide deck is ~3–6s. PPTX ~5–10s. HTML ~1–2s.

## See also

- Catalog row: [`PACKS.md`](/PACKS) — `slides.render`.
- Source: [`internal/packs/builtin/slides_render.go`](https://github.com/tosin2013/helmdeck/blob/main/internal/packs/builtin/slides_render.go).
- Companion pack: [`slides.narrate`](./narrate.md) (MP4 + TTS narration).
- Marp documentation: <https://marp.app/>.
