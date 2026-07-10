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
  ".codex-plugin/plugin.json",
  ".agents/plugins/marketplace.json",
  ".mcp.json",
  "commands/access-review.md",
  "commands/missing-signatures.md",
  "skills/access-review/SKILL.md",
  "skills/missing-signatures/SKILL.md",
  "skills/open-source-compliance/SKILL.md",
  "skills/access-review/references/mcp-tools.md",
  "skills/access-review/references/decision-rubric.md",
  "skills/access-review/references/notes-format.md",
  "skills/missing-signatures/references/mcp-tools.md",
  "skills/missing-signatures/references/report-format.md",
  "skills/missing-signatures/references/notes-format.md",
  "COMPATIBILITY.md",
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
const codexManifestPath = join(root, ".codex-plugin/plugin.json");

for (const [label, path] of [
  ["plugin.json", manifestPath],
  [".codex-plugin/plugin.json", codexManifestPath],
]) {
  if (!existsSync(path)) {
    continue;
  }
  try {
    const manifest = JSON.parse(readFileSync(path, "utf8"));
    if (typeof manifest.name !== "string" || manifest.name.length === 0) {
      console.error(`${label}: name must be a non-empty string`);
      failed = true;
    }
    if (manifest.repository != null && typeof manifest.repository !== "string") {
      console.error(
        `${label}: repository must be a string URL, not an object`,
      );
      failed = true;
    }
    if (manifest.bugs != null && typeof manifest.bugs !== "string") {
      console.error(`${label}: bugs must be a string URL, not an object`);
      failed = true;
    }
  } catch (error) {
    console.error(`${label} is not valid JSON: ${error.message}`);
    failed = true;
  }
}

const mcpPath = join(root, ".mcp.json");
if (existsSync(mcpPath)) {
  try {
    const mcpConfig = JSON.parse(readFileSync(mcpPath, "utf8"));
    const servers = mcpConfig.mcpServers ?? {};
    for (const [name, config] of Object.entries(servers)) {
      if (config?.headers?.Authorization != null) {
        console.error(
          `.mcp.json: ${name} must use OAuth 2.0, not headers.Authorization`,
        );
        failed = true;
      }
    }
  } catch (error) {
    console.error(`.mcp.json is not valid JSON: ${error.message}`);
    failed = true;
  }
}

if (failed) {
  process.exit(1);
}

console.log("@probo/skills validation passed");
