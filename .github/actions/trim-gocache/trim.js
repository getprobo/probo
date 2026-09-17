// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Trim GOCACHE to this run's working set before actions/cache saves it.
//
// Go touches an entry's mtime whenever it's used, so after the job's
// build/test steps every entry this run read or wrote has mtime >= job start.
// Everything older went unused this run and is safe to drop. That keeps the
// saved cache the size of the working set.
//
// The configurable cutoff slack defaults to 2h. Go skips the mtime touch when
// the file was already touched within the last hour (mtimeInterval), so
// entries restored from a cache saved very recently may retain pre-job mtimes
// even though they were used. Over-keeping slightly is harmless;
// under-keeping forces recompiles.
//
// Failed jobs cannot poison the cache: actions/cache's save step is gated on
// post-if: success(), so a partial run's trimmed disk state is never uploaded.
//
// This runs as the composite action's post-hook in reverse declaration order —
// after the consumer workflow's build/lint/test steps populate GOCACHE, before
// actions/cache's post-save tars and uploads it.

const fs = require("fs");
const path = require("path");

const FALLBACK_AGE_MS = 5 * 24 * 60 * 60 * 1000;
const DEFAULT_MAX_STALENESS_HOURS = 2;
const HOUR_MS = 60 * 60 * 1000;
const MAX_STALENESS_HOURS = Math.floor(Number.MAX_SAFE_INTEGER / HOUR_MS);

const dir = process.env.STATE_gocache;

if (!dir || !fs.existsSync(dir)) {
  console.log(`GOCACHE ${dir || "(unset)"} does not exist; nothing to trim.`);
  process.exit(0);
}

const startMs = parseInt(process.env.STATE_startMs || "", 10);
const rawMaxStalenessHours =
  process.env.STATE_maxStalenessHours ??
  process.env.INPUT_MAX_STALENESS_HOURS ??
  process.env["INPUT_MAX-STALENESS-HOURS"] ??
  String(DEFAULT_MAX_STALENESS_HOURS);
const maxStalenessHours = Number(rawMaxStalenessHours);
if (
  !/^\d+$/.test(rawMaxStalenessHours) ||
  !Number.isSafeInteger(maxStalenessHours) ||
  maxStalenessHours < 1 ||
  maxStalenessHours > MAX_STALENESS_HOURS
) {
  throw new Error(
    "max-staleness-hours must be an integer >= 1 (Go's mtime touch interval) that fits in a JavaScript timestamp",
  );
}
const cutoff = Number.isFinite(startMs)
  ? startMs - maxStalenessHours * HOUR_MS
  : Date.now() - FALLBACK_AGE_MS;
if (!Number.isFinite(startMs)) {
  console.warn(
    "job start time missing from state; falling back to age-based cutoff",
  );
}

let kept = 0;
let keptBytes = 0;
let removed = 0;
let removedBytes = 0;

function walk(d) {
  for (const ent of fs.readdirSync(d, { withFileTypes: true })) {
    const p = path.join(d, ent.name);
    if (ent.isDirectory()) {
      walk(p);
      continue;
    }
    if (!ent.isFile()) {
      continue;
    }
    let s;
    try {
      s = fs.statSync(p);
    } catch {
      continue;
    }
    if (s.mtimeMs >= cutoff) {
      kept++;
      keptBytes += s.size;
      continue;
    }
    try {
      fs.unlinkSync(p);
      removed++;
      removedBytes += s.size;
    } catch (e) {
      console.warn(`unlink ${p}: ${e.message}`);
    }
  }
}

walk(dir);

const mb = (b) => Math.round(b / 1024 / 1024);
console.log(
  `GOCACHE trim: cutoff=${new Date(cutoff).toISOString()}, ` +
    `kept ${kept} files (${mb(keptBytes)}MB), ` +
    `removed ${removed} files (${mb(removedBytes)}MB)`,
);
