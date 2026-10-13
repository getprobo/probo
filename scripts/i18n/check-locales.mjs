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

// Checks that every i18next catalog for a locale matches its en-US source:
// valid JSON, identical key set, same {{variables}}, $t() references and
// <tags>, and no empty strings.
//
// Usage:
//   node scripts/i18n/check-locales.mjs [--locale ja-JP] [--apps a,b] [--root dir]
//   node scripts/i18n/check-locales.mjs --pair <source.json> <target.json>

import { readdirSync, readFileSync, existsSync } from "node:fs";
import { join, dirname, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const SOURCE_LOCALE = "en-US";
export const DEFAULT_LOCALE = "ja-JP";
export const DEFAULT_APPS = ["console", "employee-portal", "compliance-portal"];

const VARIABLE = /\{\{\s*([^},\s]+)[^}]*\}\}/g;
const NESTING = /\$t\(([^)]*)\)/g;
const TAG = /<\/?([A-Za-z0-9]+)\s*\/?>/g;

export function flatten(node, prefix = "") {
  return Object.entries(node).flatMap(([key, value]) => {
    const path = prefix ? `${prefix}.${key}` : key;
    return value !== null && typeof value === "object"
      ? flatten(value, path)
      : [[path, value]];
  });
}

function tokens(value, pattern) {
  return [...value.matchAll(pattern)].map(match => match[0].replace(/\s+/g, "")).sort();
}

function sameTokens(a, b) {
  return a.length === b.length && a.every((token, index) => token === b[index]);
}

function compareValue(key, source, target) {
  if (typeof target !== "string") {
    return [{ key, kind: "type", message: `expected string, got ${typeof target}` }];
  }
  if (target.trim() === "" && source.trim() !== "") {
    return [{ key, kind: "empty", message: "translation is empty" }];
  }

  const checks = [
    ["variable", VARIABLE],
    ["nesting", NESTING],
    ["tag", TAG],
  ];

  return checks.flatMap(([kind, pattern]) => {
    const expected = tokens(source, pattern);
    const actual = tokens(target, pattern);
    return sameTokens(expected, actual)
      ? []
      : [{
          key,
          kind,
          message: `expected [${expected.join(", ")}], got [${actual.join(", ")}]`,
        }];
  });
}

export function compareCatalogs(source, target) {
  const sourceEntries = new Map(flatten(source));
  const targetEntries = new Map(flatten(target));
  const issues = [];

  for (const [key, value] of sourceEntries) {
    if (!targetEntries.has(key)) {
      issues.push({ key, kind: "missing", message: "key is missing" });
      continue;
    }
    issues.push(...compareValue(key, String(value), targetEntries.get(key)));
  }

  for (const key of targetEntries.keys()) {
    if (!sourceEntries.has(key)) {
      issues.push({ key, kind: "extra", message: "key does not exist in source" });
    }
  }

  return issues;
}

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

export function checkFiles(sourcePath, targetPath) {
  if (!existsSync(targetPath)) {
    return [{ key: "", kind: "file", message: "catalog is missing" }];
  }

  let source;
  let target;
  try {
    source = readJson(sourcePath);
  } catch (error) {
    return [{ key: "", kind: "syntax", message: `${sourcePath}: ${error.message}` }];
  }
  try {
    target = readJson(targetPath);
  } catch (error) {
    return [{ key: "", kind: "syntax", message: error.message }];
  }

  return compareCatalogs(source, target);
}

export function findSourceCatalogs(directory) {
  return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      return entry.name === "node_modules" ? [] : findSourceCatalogs(path);
    }
    return entry.name === `${SOURCE_LOCALE}.json` && dirname(path).endsWith("_locales")
      ? [path]
      : [];
  });
}

export function checkLocale({ root, locale, apps }) {
  return apps.flatMap((app) => {
    const sourceDirectory = join(root, "apps", app, "src");
    return findSourceCatalogs(sourceDirectory).map((sourcePath) => {
      const targetPath = join(dirname(sourcePath), `${locale}.json`);
      return {
        file: relative(root, targetPath),
        issues: checkFiles(sourcePath, targetPath),
      };
    });
  });
}

function parseArguments(argv) {
  const options = { locale: DEFAULT_LOCALE, apps: DEFAULT_APPS, root: process.cwd(), pair: null };

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    switch (argument) {
      case "--locale":
        options.locale = argv[++index];
        break;
      case "--apps":
        options.apps = argv[++index].split(",").filter(Boolean);
        break;
      case "--root":
        options.root = resolve(argv[++index]);
        break;
      case "--pair":
        options.pair = [argv[++index], argv[++index]];
        break;
      default:
        throw new Error(`unknown argument: ${argument}`);
    }
  }

  if (!options.locale || (options.pair && options.pair.some(path => !path))) {
    throw new Error("missing argument value");
  }

  return options;
}

function report(results) {
  let total = 0;
  for (const { file, issues } of results) {
    if (issues.length === 0) {
      continue;
    }
    total += issues.length;
    console.error(`\n${file}`);
    for (const { key, kind, message } of issues) {
      console.error(`  [${kind}] ${key ? `${key}: ` : ""}${message}`);
    }
  }
  return total;
}

function main(argv) {
  let options;
  try {
    options = parseArguments(argv);
  } catch (error) {
    console.error(error.message);
    return 2;
  }

  const results = options.pair
    ? [{ file: options.pair[1], issues: checkFiles(...options.pair) }]
    : checkLocale(options);

  const total = report(results);
  if (total > 0) {
    console.error(`\n${total} issue(s) in ${results.filter(r => r.issues.length).length} catalog(s)`);
    return 1;
  }

  console.log(`${results.length} catalog(s) OK`);
  return 0;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  process.exitCode = main(process.argv.slice(2));
}
