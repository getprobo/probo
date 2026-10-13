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

import { checkLocale, compareCatalogs, flatten } from "./check-locales.mjs";

const kinds = issues => issues.map(issue => `${issue.kind}:${issue.key}`);

test("flatten produces dotted paths for nested keys", () => {
  assert.deepEqual(flatten({ a: { b: "x", c: { d: "y" } } }), [
    ["a.b", "x"],
    ["a.c.d", "y"],
  ]);
});

test("matching catalogs have no issues", () => {
  const source = { a: { title: "Hello {{name}}", count_one: "{{count}} item", count_other: "{{count}} items" } };
  const target = { a: { title: "{{name}}さん、こんにちは", count_one: "{{count}}件", count_other: "{{count}}件" } };
  assert.deepEqual(compareCatalogs(source, target), []);
});

test("reports missing and extra keys", () => {
  const issues = compareCatalogs({ a: "A", b: "B" }, { a: "あ", c: "し" });
  assert.deepEqual(kinds(issues), ["missing:b", "extra:c"]);
});

test("reports dropped or renamed variables", () => {
  const issues = compareCatalogs({ a: "Hi {{name}}" }, { a: "こんにちは {{nom}}" });
  assert.deepEqual(kinds(issues), ["variable:a"]);
});

test("ignores variable formatting options and whitespace", () => {
  const issues = compareCatalogs(
    { a: "On {{date, datetime}}" },
    { a: "{{ date, datetime }}に" },
  );
  assert.deepEqual(issues, []);
});

test("reports broken markup and nesting references", () => {
  const issues = compareCatalogs(
    { a: "Read the <link>policy</link>", b: "$t(common.save) now" },
    { a: "<link>ポリシー</link を読む", b: "今すぐ保存" },
  );
  assert.deepEqual(kinds(issues), ["tag:a", "nesting:b"]);
});

test("reports empty translations and type mismatches", () => {
  const issues = compareCatalogs({ a: "A", b: "B" }, { a: "  ", b: { nested: "x" } });
  assert.deepEqual(kinds(issues).sort(), ["extra:b.nested", "missing:b", "empty:a"].sort());
});

test("checkLocale reports missing catalogs and syntax errors per file", () => {
  const root = mkdtempSync(join(tmpdir(), "check-locales-"));
  const ok = join(root, "apps/app/src/_locales");
  const broken = join(root, "apps/app/src/pages/broken/_locales");
  const missing = join(root, "apps/app/src/pages/missing/_locales");
  for (const directory of [ok, broken, missing]) {
    mkdirSync(directory, { recursive: true });
    writeFileSync(join(directory, "en-US.json"), JSON.stringify({ a: "A" }));
  }
  writeFileSync(join(ok, "ja-JP.json"), JSON.stringify({ a: "あ" }));
  writeFileSync(join(broken, "ja-JP.json"), "{ \"a\": ");

  const results = checkLocale({ root, locale: "ja-JP", apps: ["app"] });
  const byFile = Object.fromEntries(results.map(r => [r.file, r.issues.map(i => i.kind)]));

  assert.deepEqual(byFile, {
    "apps/app/src/_locales/ja-JP.json": [],
    "apps/app/src/pages/broken/_locales/ja-JP.json": ["syntax"],
    "apps/app/src/pages/missing/_locales/ja-JP.json": ["file"],
  });
});
