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

import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");

const AGENT_PLUGINS_VERSION = "1.0.0";
const PLUGIN_SCHEMA_ID = `https://agent-plugins.org/schemas/${AGENT_PLUGINS_VERSION}/plugin.schema.json`;
const MCP_SCHEMA_ID = `https://agent-plugins.org/schemas/${AGENT_PLUGINS_VERSION}/mcp.schema.json`;

// Agent Plugins 1.0.0 §5.2 and §7.2.1 define closed schemas: any other
// top-level field is a violation.
const MANIFEST_FIELDS = new Set([
  "$schema",
  "name",
  "version",
  "description",
  "author",
  "homepage",
  "repository",
  "license",
  "keywords",
  "extensions",
]);
const AUTHOR_FIELDS = new Set(["name", "email", "url"]);
const REMOTE_SERVER_FIELDS = new Set(["type", "url", "headers"]);
const STDIO_SERVER_FIELDS = new Set(["type", "command", "args", "env", "cwd"]);

// §5.5 plugin name and Agent Skills `name` constraints.
const PLUGIN_NAME_PATTERN = /^(?!.*(?:--|\.\.))[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?$/;
const SKILL_NAME_PATTERN = /^(?!.*--)[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$/;

// Header names that would turn visible package data into a credential
// (§7.2.1 forbids secrets in `headers`).
const CREDENTIAL_HEADERS = new Set([
  "authorization",
  "proxy-authorization",
  "cookie",
  "x-api-key",
  "x-auth-token",
]);

const requiredPaths = [
  "plugin.json",
  "mcp.json",
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
  "skills/compliance-portal-commitments/SKILL.md",
  "skills/access-review/references/mcp-tools.md",
  "skills/access-review/references/decision-rubric.md",
  "skills/access-review/references/notes-format.md",
  "skills/missing-signatures/references/mcp-tools.md",
  "skills/missing-signatures/references/report-format.md",
  "skills/missing-signatures/references/notes-format.md",
  "skills/compliance-portal-commitments/references/voice.md",
  "skills/compliance-portal-commitments/references/portal-mechanics.md",
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

function requireOptionalString(label, field, value) {
  if (value != null && typeof value !== "string") {
    fail(`${label}: ${field} must be a string`);
  }
}

for (const relativePath of requiredPaths) {
  const absolutePath = join(root, relativePath);
  if (!existsSync(absolutePath)) {
    fail(`missing required file: ${relativePath}`);
  }
}

const packageJsonPath = join(root, "package.json");
const packageJson = existsSync(packageJsonPath)
  ? readJsonObject("package.json", packageJsonPath)
  : undefined;
const packageName = packageJson?.name ?? null;
const packageVersion = packageJson?.version ?? null;

// --- Agent Plugins 1.0.0 portable package ---------------------------------

function validateAgentPluginsManifest(label, path) {
  if (!existsSync(path)) {
    return;
  }

  const manifest = readJsonObject(label, path);
  if (manifest === undefined) {
    return;
  }

  for (const field of Object.keys(manifest)) {
    if (!MANIFEST_FIELDS.has(field)) {
      fail(`${label}: unknown top-level field "${field}" (schema is closed)`);
    }
  }

  if (manifest.$schema !== PLUGIN_SCHEMA_ID) {
    fail(`${label}: $schema must be "${PLUGIN_SCHEMA_ID}"`);
  }

  if (requireNonEmptyString(label, "name", manifest.name)) {
    if (manifest.name.length > 64 || !PLUGIN_NAME_PATTERN.test(manifest.name)) {
      fail(
        `${label}: name "${manifest.name}" must be 1-64 lowercase alphanumeric, "-", or "." characters, start and end alphanumeric, without "--" or ".."`,
      );
    }
  }

  for (const field of ["version", "description", "homepage", "repository", "license"]) {
    requireOptionalString(label, field, manifest[field]);
  }

  if (manifest.author != null) {
    if (typeof manifest.author !== "object" || Array.isArray(manifest.author)) {
      fail(`${label}: author must be an object`);
    } else {
      for (const field of Object.keys(manifest.author)) {
        if (!AUTHOR_FIELDS.has(field)) {
          fail(`${label}: author has unknown field "${field}"`);
        } else {
          requireOptionalString(label, `author.${field}`, manifest.author[field]);
        }
      }
    }
  }

  if (manifest.keywords != null) {
    if (
      !Array.isArray(manifest.keywords) ||
      manifest.keywords.some((keyword) => typeof keyword !== "string")
    ) {
      fail(`${label}: keywords must be an array of strings`);
    }
  }

  if (manifest.extensions != null) {
    if (typeof manifest.extensions !== "object" || Array.isArray(manifest.extensions)) {
      fail(`${label}: extensions must be an object keyed by reverse-domain namespace`);
    }
  }

  if (packageVersion != null && manifest.version !== packageVersion) {
    fail(`${label}: version must match package.json version (${packageVersion})`);
  }
}

function validateRemoteURL(label, rawURL) {
  if (rawURL.includes("${")) {
    fail(
      `${label}: url must be a literal URL — Agent Plugins clients do not expand placeholders in remote URLs`,
    );
    return;
  }

  let url;
  try {
    url = new URL(rawURL);
  } catch {
    fail(`${label}: url must be an absolute URL`);
    return;
  }

  if (url.protocol !== "https:" && url.protocol !== "http:") {
    fail(`${label}: url must use http or https`);
    return;
  }

  if (url.username !== "" || url.password !== "") {
    fail(`${label}: url must not contain user information`);
  }

  if (url.hash !== "") {
    fail(`${label}: url must not contain a fragment`);
  }

  const loopback =
    url.hostname === "localhost" ||
    url.hostname === "127.0.0.1" ||
    url.hostname === "[::1]";
  if (url.protocol === "http:" && !loopback) {
    fail(`${label}: non-loopback url must use https`);
  }
}

function validateHeaders(label, headers) {
  if (typeof headers !== "object" || Array.isArray(headers)) {
    fail(`${label}: headers must be an object of strings`);
    return;
  }

  const seen = new Set();
  for (const [name, value] of Object.entries(headers)) {
    const lowercased = name.toLowerCase();
    if (seen.has(lowercased)) {
      fail(`${label}: header "${name}" is declared more than once under different casing`);
    }
    seen.add(lowercased);

    if (typeof value !== "string") {
      fail(`${label}: header "${name}" must be a string`);
    }

    if (CREDENTIAL_HEADERS.has(lowercased)) {
      fail(
        `${label}: header "${name}" would embed a credential in visible package data — Probo MCP uses OAuth 2.0`,
      );
    }
  }
}

function validateAgentPluginsMcp(label, path) {
  if (!existsSync(path)) {
    return;
  }

  const config = readJsonObject(label, path);
  if (config === undefined) {
    return;
  }

  for (const field of Object.keys(config)) {
    if (field !== "$schema" && field !== "mcpServers") {
      fail(`${label}: unknown top-level field "${field}" (schema is closed)`);
    }
  }

  if (config.$schema !== MCP_SCHEMA_ID) {
    fail(`${label}: $schema must be "${MCP_SCHEMA_ID}"`);
  }

  if (
    config.mcpServers == null ||
    typeof config.mcpServers !== "object" ||
    Array.isArray(config.mcpServers)
  ) {
    fail(`${label}: mcpServers must be an object`);
    return undefined;
  }

  for (const [name, server] of Object.entries(config.mcpServers)) {
    const serverLabel = `${label} mcpServers.${name}`;

    if (server == null || typeof server !== "object" || Array.isArray(server)) {
      fail(`${serverLabel}: server configuration must be an object`);
      continue;
    }

    if (server.type === "streamable-http" || server.type === "sse") {
      for (const field of Object.keys(server)) {
        if (!REMOTE_SERVER_FIELDS.has(field)) {
          fail(`${serverLabel}: unknown field "${field}" for a ${server.type} server`);
        }
      }
      if (requireNonEmptyString(serverLabel, "url", server.url)) {
        validateRemoteURL(serverLabel, server.url);
      }
      if (server.headers != null) {
        validateHeaders(serverLabel, server.headers);
      }
      continue;
    }

    if (server.type === "stdio") {
      for (const field of Object.keys(server)) {
        if (!STDIO_SERVER_FIELDS.has(field)) {
          fail(`${serverLabel}: unknown field "${field}" for a stdio server`);
        }
      }
      requireNonEmptyString(serverLabel, "command", server.command);
      continue;
    }

    fail(`${serverLabel}: type must be "stdio", "streamable-http", or "sse"`);
  }

  return config.mcpServers;
}

// --- Agent Skills ---------------------------------------------------------

// Reads one scalar from SKILL.md frontmatter. Supports plain scalars and the
// block forms (`|`, `>`) the skills use, which is all the Agent Skills
// frontmatter fields need — a full YAML parser would add a dependency to a
// package that otherwise ships no code.
function readFrontmatterScalar(frontmatter, key) {
  const lines = frontmatter.split("\n");
  const index = lines.findIndex((line) => line.startsWith(`${key}:`));
  if (index === -1) {
    return undefined;
  }

  const inline = lines[index].slice(`${key}:`.length).trim();
  if (inline !== "" && !inline.startsWith("|") && !inline.startsWith(">")) {
    return inline.replace(/^["'](.*)["']$/, "$1");
  }

  const block = [];
  for (const line of lines.slice(index + 1)) {
    if (line.trim() !== "" && !/^\s/.test(line)) {
      break;
    }
    block.push(line.trim());
  }

  return block.join(" ").trim();
}

function validateSkill(name, skillDirectory) {
  const label = `skills/${name}/SKILL.md`;
  const skillPath = join(skillDirectory, "SKILL.md");

  if (!existsSync(skillPath) || !statSync(skillPath).isFile()) {
    fail(`${label}: every immediate child of skills/ must contain a SKILL.md file`);
    return;
  }

  const content = readFileSync(skillPath, "utf8");
  const match = /^---\n([\s\S]*?)\n---\n/.exec(content);
  if (match === null) {
    fail(`${label}: must start with YAML frontmatter delimited by "---"`);
    return;
  }

  const frontmatter = match[1];
  const skillName = readFrontmatterScalar(frontmatter, "name");
  const description = readFrontmatterScalar(frontmatter, "description");
  const compatibility = readFrontmatterScalar(frontmatter, "compatibility");

  if (skillName === undefined || skillName === "") {
    fail(`${label}: frontmatter must declare a non-empty name`);
  } else {
    if (skillName !== name) {
      fail(`${label}: frontmatter name "${skillName}" must match the directory name "${name}"`);
    }
    if (skillName.length > 64 || !SKILL_NAME_PATTERN.test(skillName)) {
      fail(
        `${label}: name "${skillName}" must be 1-64 lowercase alphanumeric or "-" characters, start and end alphanumeric, without "--"`,
      );
    }
  }

  if (description === undefined || description === "") {
    fail(`${label}: frontmatter must declare a non-empty description`);
  } else if (description.length > 1024) {
    fail(`${label}: description must be at most 1024 characters (found ${description.length})`);
  }

  if (compatibility !== undefined && compatibility.length > 500) {
    fail(`${label}: compatibility must be at most 500 characters (found ${compatibility.length})`);
  }
}

function validateSkills(skillsDirectory) {
  if (!existsSync(skillsDirectory)) {
    fail("missing required directory: skills/");
    return;
  }

  if (!statSync(skillsDirectory).isDirectory()) {
    fail("skills/ must be a directory");
    return;
  }

  const entries = readdirSync(skillsDirectory, { withFileTypes: true }).filter((entry) =>
    entry.isDirectory(),
  );

  if (entries.length === 0) {
    fail("skills/ must contain at least one skill directory");
    return;
  }

  for (const entry of entries) {
    validateSkill(entry.name, join(skillsDirectory, entry.name));
  }
}

// --- Client-specific manifests and catalogs -------------------------------

function validateClientManifest(label, path) {
  if (!existsSync(path)) {
    return;
  }

  const manifest = readJsonObject(label, path);
  if (manifest === undefined) {
    return;
  }

  requireNonEmptyString(label, "name", manifest.name);
  if (manifest.repository != null && typeof manifest.repository !== "string") {
    fail(`${label}: repository must be a string URL, not an object`);
  }
  if (manifest.bugs != null && typeof manifest.bugs !== "string") {
    fail(`${label}: bugs must be a string URL, not an object`);
  }
  if (packageVersion != null && manifest.version !== packageVersion) {
    fail(`${label}: version must match package.json version (${packageVersion})`);
  }
}

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

// The client-specific `.mcp.json` must stay in lockstep with the portable
// `mcp.json`; only the transport keyword differs (`http` is the Claude Code
// and Codex spelling of MCP Streamable HTTP).
function validateClientMcpParity(label, path, portableServers) {
  if (!existsSync(path)) {
    return;
  }

  const config = readJsonObject(label, path);
  if (config === undefined || portableServers === undefined) {
    return;
  }

  const servers = config.mcpServers;
  if (servers == null || typeof servers !== "object" || Array.isArray(servers)) {
    fail(`${label}: mcpServers must be an object`);
    return;
  }

  const portableNames = Object.keys(portableServers).sort().join(", ");
  const clientNames = Object.keys(servers).sort().join(", ");
  if (portableNames !== clientNames) {
    fail(`${label}: servers (${clientNames}) must match mcp.json servers (${portableNames})`);
    return;
  }

  for (const [name, server] of Object.entries(servers)) {
    const serverLabel = `${label} mcpServers.${name}`;
    const portable = portableServers[name];

    if (server?.headers?.Authorization != null) {
      fail(`${serverLabel}: must use OAuth 2.0, not headers.Authorization`);
    }

    if (portable?.type === "streamable-http" && server?.type !== "http") {
      fail(`${serverLabel}: type must be "http" to match the portable streamable-http server`);
    }

    if (server?.url !== portable?.url) {
      fail(`${serverLabel}: url must match mcp.json (${portable?.url})`);
    }
  }
}

validateAgentPluginsManifest("plugin.json", join(root, "plugin.json"));
const portableServers = validateAgentPluginsMcp("mcp.json", join(root, "mcp.json"));
validateSkills(join(root, "skills"));
validateClientManifest(".claude-plugin/plugin.json", join(root, ".claude-plugin/plugin.json"));
validateClientManifest(".codex-plugin/plugin.json", join(root, ".codex-plugin/plugin.json"));
validateClaudeMarketplace(
  ".claude-plugin/marketplace.json",
  join(root, ".claude-plugin/marketplace.json"),
);
validateCodexMarketplace(
  ".agents/plugins/marketplace.json",
  join(root, ".agents/plugins/marketplace.json"),
);
validateClientMcpParity(".mcp.json", join(root, ".mcp.json"), portableServers);

if (failed) {
  process.exit(1);
}

console.log("@probo/skills validation passed");
