// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");

const requiredPaths = [
  ".claude-plugin/plugin.json",
  ".mcp.json",
  "skills/open-source-compliance/SKILL.md",
];

let failed = false;

for (const relativePath of requiredPaths) {
  const absolutePath = join(root, relativePath);
  if (!existsSync(absolutePath)) {
    console.error(`missing required file: ${relativePath}`);
    failed = true;
  }
}

const manifestPath = join(root, ".claude-plugin/plugin.json");
if (existsSync(manifestPath)) {
  try {
    const manifest = JSON.parse(readFileSync(manifestPath, "utf8"));
    if (typeof manifest.name !== "string" || manifest.name.length === 0) {
      console.error("plugin.json: name must be a non-empty string");
      failed = true;
    }
    if (manifest.repository != null && typeof manifest.repository !== "string") {
      console.error(
        "plugin.json: repository must be a string URL, not an object",
      );
      failed = true;
    }
    if (manifest.bugs != null && typeof manifest.bugs !== "string") {
      console.error("plugin.json: bugs must be a string URL, not an object");
      failed = true;
    }
  } catch (error) {
    console.error(`plugin.json is not valid JSON: ${error.message}`);
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}

console.log("@probo/claude-plugin validation passed");
