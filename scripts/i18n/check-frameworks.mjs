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

// Checks that every translated framework file under
// apps/console/public/data/frameworks/<locale>/ matches its English source:
// same framework id, same control ids in the same order, and non-empty names.
//
// Usage:
//   node scripts/i18n/check-frameworks.mjs [--locale ja-JP] [--root dir]

import { existsSync, readdirSync, readFileSync } from "node:fs";
import { join, relative, resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const FRAMEWORKS_DIR = "apps/console/public/data/frameworks";
export const DEFAULT_LOCALE = "ja-JP";

const isBlank = value => typeof value !== "string" || value.trim() === "";

export function compareFrameworks(source, target) {
  const issues = [];

  if (target.id !== source.id) {
    issues.push({ key: "id", kind: "id", message: `expected "${source.id}", got "${target.id}"` });
  }
  if (isBlank(target.name)) {
    issues.push({ key: "name", kind: "empty", message: "framework name is empty" });
  }
  if (!Array.isArray(target.controls)) {
    return [...issues, { key: "controls", kind: "type", message: "controls must be an array" }];
  }

  const sourceIds = source.controls.map(control => control.id);
  const targetIds = target.controls.map(control => control?.id);
  const missing = sourceIds.filter(id => !targetIds.includes(id));
  const extra = targetIds.filter(id => !sourceIds.includes(id));

  issues.push(
    ...missing.map(id => ({ key: id, kind: "missing", message: "control is missing" })),
    ...extra.map(id => ({ key: String(id), kind: "extra", message: "control does not exist in source" })),
  );

  if (missing.length === 0 && extra.length === 0 && sourceIds.some((id, index) => id !== targetIds[index])) {
    issues.push({ key: "controls", kind: "order", message: "controls are not in source order" });
  }

  target.controls.forEach((control, index) => {
    if (isBlank(control?.name)) {
      issues.push({ key: control?.id ?? `#${index}`, kind: "empty", message: "control name is empty" });
    }
  });

  return issues;
}

function readJson(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

export function checkFile(sourcePath, targetPath) {
  if (!existsSync(sourcePath)) {
    return [{ key: "", kind: "file", message: "no English source with the same file name" }];
  }

  try {
    return compareFrameworks(readJson(sourcePath), readJson(targetPath));
  } catch (error) {
    return [{ key: "", kind: "syntax", message: error.message }];
  }
}

export function checkLocale({ root, locale }) {
  const directory = join(root, FRAMEWORKS_DIR);
  const localeDirectory = join(directory, locale);
  if (!existsSync(localeDirectory)) {
    return [];
  }

  return readdirSync(localeDirectory)
    .filter(name => name.endsWith(".json"))
    .map(name => ({
      file: relative(root, join(localeDirectory, name)),
      issues: checkFile(join(directory, name), join(localeDirectory, name)),
    }));
}

function parseArguments(argv) {
  const options = { locale: DEFAULT_LOCALE, root: process.cwd() };

  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    const value = argv[index + 1];
    if ((argument === "--locale" || argument === "--root") && !value) {
      throw new Error(`missing value for ${argument}`);
    }
    if (argument === "--locale") {
      options.locale = value;
    } else if (argument === "--root") {
      options.root = resolve(value);
    } else {
      throw new Error(`unknown argument: ${argument}`);
    }
    index += 1;
  }

  return options;
}

function main(argv) {
  let options;
  try {
    options = parseArguments(argv);
  } catch (error) {
    console.error(error.message);
    return 2;
  }

  const results = checkLocale(options);
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

  if (total > 0) {
    console.error(`\n${total} issue(s)`);
    return 1;
  }

  console.log(`${results.length} framework(s) OK`);
  return 0;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  process.exitCode = main(process.argv.slice(2));
}
