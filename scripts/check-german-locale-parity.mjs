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

import { readFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = resolve(dirname(fileURLToPath(import.meta.url)), "..");

const localePairs = [
  ["apps/console/src/_locales/en-US.json", "apps/console/src/_locales/de-DE.json"],
  [
    "apps/console/src/pages/organizations/tasks/_locales/en-US.json",
    "apps/console/src/pages/organizations/tasks/_locales/de-DE.json",
  ],
  [
    "apps/console/src/pages/organizations/compliance-portals/_locales/en-US.json",
    "apps/console/src/pages/organizations/compliance-portals/_locales/de-DE.json",
  ],
  [
    "apps/console/src/pages/organizations/cookie-banners/_locales/en-US.json",
    "apps/console/src/pages/organizations/cookie-banners/_locales/de-DE.json",
  ],
  ["apps/employee-portal/src/_locales/en-US.json", "apps/employee-portal/src/_locales/de-DE.json"],
  [
    "apps/employee-portal/src/pages/enroll/_locales/en-US.json",
    "apps/employee-portal/src/pages/enroll/_locales/de-DE.json",
  ],
  [
    "apps/employee-portal/src/pages/devices/_locales/en-US.json",
    "apps/employee-portal/src/pages/devices/_locales/de-DE.json",
  ],
  [
    "apps/employee-portal/src/pages/bindings/_locales/en-US.json",
    "apps/employee-portal/src/pages/bindings/_locales/de-DE.json",
  ],
  [
    "apps/employee-portal/src/pages/approvals/_locales/en-US.json",
    "apps/employee-portal/src/pages/approvals/_locales/de-DE.json",
  ],
  [
    "apps/employee-portal/src/pages/signatures/_locales/en-US.json",
    "apps/employee-portal/src/pages/signatures/_locales/de-DE.json",
  ],
  ["apps/compliance-portal/src/_locales/en-US.json", "apps/compliance-portal/src/_locales/de-DE.json"],
  ["apps/compliance-portal/src/pages/nda/_locales/en-US.json", "apps/compliance-portal/src/pages/nda/_locales/de-DE.json"],
  ["apps/compliance-portal/src/pages/updates/_locales/en-US.json", "apps/compliance-portal/src/pages/updates/_locales/de-DE.json"],
  ["apps/compliance-portal/src/pages/requests/_locales/en-US.json", "apps/compliance-portal/src/pages/requests/_locales/de-DE.json"],
  ["apps/compliance-portal/src/pages/documents/_locales/en-US.json", "apps/compliance-portal/src/pages/documents/_locales/de-DE.json"],
  ["apps/compliance-portal/src/pages/subprocessors/_locales/en-US.json", "apps/compliance-portal/src/pages/subprocessors/_locales/de-DE.json"],
];

function flatten(value, prefix = "", result = new Map()) {
  if (typeof value === "string") {
    result.set(prefix, value);
    return result;
  }

  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`Locale value at ${prefix || "<root>"} must be an object or string`);
  }

  for (const [key, child] of Object.entries(value)) {
    flatten(child, prefix ? `${prefix}.${key}` : key, result);
  }

  return result;
}

function sortedTokens(value) {
  const tokens = [];
  for (const match of value.matchAll(/{{\s*([^{}]+?)\s*}}/g)) {
    tokens.push(`{{${match[1].trim()}}}`);
  }
  for (const match of value.matchAll(/<\/?([A-Za-z][A-Za-z0-9_-]*)>/g)) {
    tokens.push(`<${match[1]}>`);
  }
  return tokens.sort();
}

function comparableTokens(key, value) {
  const tokens = sortedTokens(value);

  // i18next uses count to select the `_one` plural branch. A localized
  // singular string can therefore omit the literal count while still relying
  // on `count` for plural selection. All other interpolation and markup tokens
  // remain strict.
  return key.endsWith("_one")
    ? tokens.filter((token) => token !== "{{count}}")
    : tokens;
}

async function loadLocale(path) {
  const absolutePath = resolve(repoRoot, path);
  const source = await readFile(absolutePath, "utf8");
  return JSON.parse(source);
}

let failed = false;

for (const [englishPath, germanPath] of localePairs) {
  let english;
  let german;

  try {
    [english, german] = await Promise.all([loadLocale(englishPath), loadLocale(germanPath)]);
  } catch (error) {
    failed = true;
    console.error(`\n${englishPath} -> ${germanPath}`);
    console.error(`  ${error.message}`);
    continue;
  }

  const englishKeys = flatten(english);
  const germanKeys = flatten(german);
  const missing = [...englishKeys.keys()].filter((key) => !germanKeys.has(key));
  const extra = [...germanKeys.keys()].filter((key) => !englishKeys.has(key));
  const placeholderMismatches = [];
  const emptyGermanValues = [];

  for (const [key, englishValue] of englishKeys) {
    if (!germanKeys.has(key)) continue;
    const germanValue = germanKeys.get(key);

    if (germanValue.trim().length === 0) {
      emptyGermanValues.push(key);
    }

    const englishTokens = comparableTokens(key, englishValue);
    const germanTokens = comparableTokens(key, germanValue);
    if (JSON.stringify(englishTokens) !== JSON.stringify(germanTokens)) {
      placeholderMismatches.push({ key, englishTokens, germanTokens });
    }
  }

  if (missing.length || extra.length || placeholderMismatches.length || emptyGermanValues.length) {
    failed = true;
    console.error(`\n${englishPath} -> ${germanPath}`);
    if (missing.length) console.error(`  Missing keys (${missing.length}):\n    ${missing.join("\n    ")}`);
    if (extra.length) console.error(`  Extra keys (${extra.length}):\n    ${extra.join("\n    ")}`);
    if (emptyGermanValues.length) console.error(`  Empty German values (${emptyGermanValues.length}):\n    ${emptyGermanValues.join("\n    ")}`);
    for (const mismatch of placeholderMismatches) {
      console.error(`  Placeholder mismatch: ${mismatch.key}`);
      console.error(`    en-US: ${mismatch.englishTokens.join(", ")}`);
      console.error(`    de-DE: ${mismatch.germanTokens.join(", ")}`);
    }
  } else {
    console.log(`OK ${germanPath} (${germanKeys.size} strings)`);
  }
}

if (failed) {
  process.exitCode = 1;
} else {
  console.log("German locale parity checks passed.");
}
