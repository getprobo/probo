// Copyright (c) 2026 MICHISUU Inc.
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

import { test } from "node:test";
import assert from "node:assert/strict";
import { mkdtempSync, mkdirSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

import {
  checkLocale,
  checkStandalone,
  compareFrameworks,
  FRAMEWORKS_DIR,
  STANDALONE_DIR,
  validateFramework,
} from "./check-frameworks.mjs";

const source = {
  id: "ISO/IEC 27001:2022",
  name: "ISO 27001 (2022)",
  controls: [
    { id: "4.1", name: "Context" },
    { id: "A.5.1", name: "Policies" },
  ],
};

const kinds = issues => issues.map(issue => `${issue.kind}:${issue.key}`);

test("matching framework has no issues", () => {
  const target = { ...source, name: "ISO 27001（2022）", controls: [{ id: "4.1", name: "状況" }, { id: "A.5.1", name: "方針" }] };
  assert.deepEqual(compareFrameworks(source, target), []);
});

test("reports a changed framework id", () => {
  const target = { ...source, id: "ISO 27001 ja" };
  assert.deepEqual(kinds(compareFrameworks(source, target)), ["id:id"]);
});

test("reports missing and extra controls", () => {
  const target = { ...source, controls: [{ id: "4.1", name: "状況" }, { id: "A.5.99", name: "余分" }] };
  assert.deepEqual(kinds(compareFrameworks(source, target)), ["missing:A.5.1", "extra:A.5.99"]);
});

test("reports controls out of order", () => {
  const target = { ...source, controls: [{ id: "A.5.1", name: "方針" }, { id: "4.1", name: "状況" }] };
  assert.deepEqual(kinds(compareFrameworks(source, target)), ["order:controls"]);
});

test("reports empty names", () => {
  const target = { ...source, name: " ", controls: [{ id: "4.1", name: "" }, { id: "A.5.1", name: "方針" }] };
  assert.deepEqual(kinds(compareFrameworks(source, target)), ["empty:name", "empty:4.1"]);
});

test("checkLocale reports syntax errors and files without a source", () => {
  const root = mkdtempSync(join(tmpdir(), "check-frameworks-"));
  const directory = join(root, FRAMEWORKS_DIR);
  mkdirSync(join(directory, "ja-JP"), { recursive: true });
  writeFileSync(join(directory, "ISO.json"), JSON.stringify(source));
  writeFileSync(join(directory, "ja-JP", "ISO.json"), "{ \"id\": ");
  writeFileSync(join(directory, "ja-JP", "Unknown.json"), JSON.stringify(source));

  const results = checkLocale({ root, locale: "ja-JP" });
  const byFile = Object.fromEntries(results.map(r => [r.file, r.issues.map(i => i.kind)]));

  assert.deepEqual(byFile, {
    [join(FRAMEWORKS_DIR, "ja-JP", "ISO.json")]: ["syntax"],
    [join(FRAMEWORKS_DIR, "ja-JP", "Unknown.json")]: ["file"],
  });
});

test("validateFramework accepts a standalone framework", () => {
  const framework = { id: "PMARK", name: "Pマーク", controls: [{ id: "J.1.1", name: "状況" }] };
  assert.deepEqual(validateFramework(framework), []);
});

test("validateFramework reports empty and duplicated controls", () => {
  const framework = {
    id: "",
    name: "Pマーク",
    controls: [{ id: "J.1.1", name: "状況" }, { id: "J.1.1", name: "重複" }, { id: "J.1.2", name: " " }],
  };
  assert.deepEqual(kinds(validateFramework(framework)), ["empty:id", "duplicate:J.1.1", "empty:J.1.2"]);
});

test("checkStandalone validates files in the jp directory", () => {
  const root = mkdtempSync(join(tmpdir(), "check-frameworks-jp-"));
  const directory = join(root, FRAMEWORKS_DIR, STANDALONE_DIR);
  mkdirSync(directory, { recursive: true });
  writeFileSync(join(directory, "OK.json"), JSON.stringify({ id: "X", name: "X", controls: [{ id: "1", name: "a" }] }));
  writeFileSync(join(directory, "Broken.json"), "{");

  const byFile = Object.fromEntries(checkStandalone({ root }).map(r => [r.file, r.issues.map(i => i.kind)]));

  assert.deepEqual(byFile, {
    [join(FRAMEWORKS_DIR, STANDALONE_DIR, "Broken.json")]: ["syntax"],
    [join(FRAMEWORKS_DIR, STANDALONE_DIR, "OK.json")]: [],
  });
});
