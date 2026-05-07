// Copies the canonical /CHANGELOG.md (at repo root) into src/pages/changelog.md
// so the docs site `/changelog` route stays in sync without duplicating content
// in source control. Runs as a prebuild/prestart hook in package.json.
//
// Two modes:
//   1. Full-repo build (local dev, GitHub Actions, Vercel git-integrated):
//      reads ../CHANGELOG.md directly.
//   2. Isolated build (Vercel CLI deploy with Root Directory = website,
//      where the parent dir is not part of the build context):
//      falls back to fetching the canonical file from GitHub raw on `main`.

import {readFile, writeFile, mkdir} from 'node:fs/promises';
import {existsSync} from 'node:fs';
import {dirname, resolve} from 'node:path';
import {fileURLToPath} from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(here, '..', '..');
const localSrc = resolve(repoRoot, 'CHANGELOG.md');
const dst = resolve(here, '..', 'src', 'pages', 'changelog.md');
const remoteUrl =
  'https://raw.githubusercontent.com/tosin2013/helmdeck/main/CHANGELOG.md';

const frontmatter = `---
title: Changelog
description: Release history for helmdeck.
---

`;

let body;
let source;
if (existsSync(localSrc)) {
  body = await readFile(localSrc, 'utf8');
  source = localSrc;
} else {
  const res = await fetch(remoteUrl);
  if (!res.ok) {
    throw new Error(
      `[copy-changelog] local ${localSrc} missing and GitHub fetch failed (${res.status} ${res.statusText})`,
    );
  }
  body = await res.text();
  source = remoteUrl;
}

await mkdir(dirname(dst), {recursive: true});
await writeFile(dst, frontmatter + body, 'utf8');
console.log(`[copy-changelog] ${source} -> ${dst}`);
