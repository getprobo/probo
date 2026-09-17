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

// Record the job start time, retention window, and GOCACHE path for trim.js
// (the post step). Entries used by this run have mtimes at or after this
// instant, modulo Go's touch granularity — see the cutoff slack in trim.js.
// Nested composites give post the ancestor's INPUT_* (actions/runner#2030).

const fs = require("fs");

const DEFAULT_MAX_STALENESS_HOURS = 2;
const HOUR_MS = 60 * 60 * 1000;
const MAX_STALENESS_HOURS = Math.floor(Number.MAX_SAFE_INTEGER / HOUR_MS);
const gocache = process.env.INPUT_GOCACHE;
const rawMaxStalenessHours =
  process.env.INPUT_MAX_STALENESS_HOURS ??
  process.env["INPUT_MAX-STALENESS-HOURS"] ??
  String(DEFAULT_MAX_STALENESS_HOURS);
const maxStalenessHours = Number(rawMaxStalenessHours);

if (!gocache) {
  throw new Error("gocache input is required");
}

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

fs.appendFileSync(
  process.env.GITHUB_STATE,
  `startMs=${Date.now()}\nmaxStalenessHours=${maxStalenessHours}\ngocache=${gocache}\n`,
);
