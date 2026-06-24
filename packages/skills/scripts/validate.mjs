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

import { existsSync, readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");

const requiredPaths = [
  ".claude-plugin/plugin.json",
  ".claude-plugin/marketplace.json",
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

function fail(message) {
  console.error(message);
  failed = true;
}

function readJsonObject(label, path) {
  let value;
  try {
    value = JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    fail(`${label} is not valid JSON: ${error.message}`);
    return undefined;
  }

  if (value == null || typeof value !== "object" || Array.isArray(value)) {
    fail(`${label}: root must be a JSON object`);
    return undefined;
  }

  return value;
}

function requireNonEmptyString(label, field, value) {
  if (typeof value !== "string" || value.length === 0) {
    fail(`${label}: ${field} must be a non-empty string`);
    return false;
  }
  return true;
}

for (const relativePath of requiredPaths) {
  const absolutePath = join(root, relativePath);
  if (!existsSync(absolutePath)) {
    fail(`missing required file: ${relativePath}`);
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
  const manifest = readJsonObject(label, path);
  if (manifest === undefined) {
    continue;
  }
  requireNonEmptyString(label, "name", manifest.name);
  if (manifest.repository != null && typeof manifest.repository !== "string") {
    fail(`${label}: repository must be a string URL, not an object`);
  }
  if (manifest.bugs != null && typeof manifest.bugs !== "string") {
    fail(`${label}: bugs must be a string URL, not an object`);
  }
}

const packageJsonPath = join(root, "package.json");
const packageName = existsSync(packageJsonPath)
  ? readJsonObject("package.json", packageJsonPath)?.name
  : null;

function validateClaudeMarketplace(label, path) {
  if (!existsSync(path)) {
    return;
  }

  const marketplace = readJsonObject(label, path);
  if (marketplace === undefined) {
    return;
  }

  requireNonEmptyString(label, "name", marketplace.name);

  if (!Array.isArray(marketplace.plugins) || marketplace.plugins.length === 0) {
    fail(`${label}: plugins must be a non-empty array`);
    return;
  }

  for (const [index, plugin] of marketplace.plugins.entries()) {
    const pluginLabel = `${label} plugins[${index}]`;
    requireNonEmptyString(pluginLabel, "name", plugin?.name);

    const source = plugin?.source;
    if (typeof source === "string") {
      if (!source.startsWith("./")) {
        fail(`${pluginLabel}: source path must start with "./"`);
      }
      continue;
    }

    if (source == null || typeof source !== "object") {
      fail(`${pluginLabel}: source must be a path string or npm source object`);
      continue;
    }

    if (source.source === "npm") {
      requireNonEmptyString(pluginLabel, "source.package", source.package);
      if (packageName != null && source.package !== packageName) {
        fail(
          `${pluginLabel}: source.package must match package.json name (${packageName})`,
        );
      }
      continue;
    }

    fail(
      `${pluginLabel}: source.source must be "npm" or use a "./" path string`,
    );
  }
}

function validateCodexMarketplace(label, path) {
  if (!existsSync(path)) {
    return;
  }

  const marketplace = readJsonObject(label, path);
  if (marketplace === undefined) {
    return;
  }

  requireNonEmptyString(label, "name", marketplace.name);

  if (!Array.isArray(marketplace.plugins) || marketplace.plugins.length === 0) {
    fail(`${label}: plugins must be a non-empty array`);
    return;
  }

  for (const [index, plugin] of marketplace.plugins.entries()) {
    const pluginLabel = `${label} plugins[${index}]`;
    requireNonEmptyString(pluginLabel, "name", plugin?.name);

    const source = plugin?.source;
    if (source == null || typeof source !== "object") {
      fail(`${pluginLabel}: source must be an object`);
      continue;
    }

    if (source.source !== "local") {
      fail(`${pluginLabel}: source.source must be "local"`);
    }

    if (typeof source.path !== "string" || !source.path.startsWith("./")) {
      fail(`${pluginLabel}: source.path must be a "./"-prefixed string`);
    }

    const policy = plugin?.policy;
    if (policy == null || typeof policy !== "object") {
      fail(`${pluginLabel}: policy must be an object`);
      continue;
    }

    requireNonEmptyString(pluginLabel, "policy.installation", policy.installation);
    requireNonEmptyString(
      pluginLabel,
      "policy.authentication",
      policy.authentication,
    );
    requireNonEmptyString(pluginLabel, "category", plugin.category);
  }
}

validateClaudeMarketplace(
  ".claude-plugin/marketplace.json",
  join(root, ".claude-plugin/marketplace.json"),
);
validateCodexMarketplace(
  ".agents/plugins/marketplace.json",
  join(root, ".agents/plugins/marketplace.json"),
);

const mcpPath = join(root, ".mcp.json");
if (existsSync(mcpPath)) {
  const mcpConfig = readJsonObject(".mcp.json", mcpPath);
  if (mcpConfig !== undefined) {
    const servers = mcpConfig.mcpServers ?? {};
    for (const [name, config] of Object.entries(servers)) {
      if (config?.headers?.Authorization != null) {
        fail(`.mcp.json: ${name} must use OAuth 2.0, not headers.Authorization`);
      }
    }
  }
}

if (failed) {
  process.exit(1);
}

console.log("@probo/skills validation passed");
