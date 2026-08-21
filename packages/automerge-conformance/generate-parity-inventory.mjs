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

import fs from "node:fs";
import path from "node:path";
import process from "node:process";
import { execFileSync } from "node:child_process";

const rustRoot = requiredArgument("--rust-root");
const javascriptRoot = requiredArgument("--javascript-root");
const rustTestList = optionalArgument("--rust-test-list");
const mappings = loadMappings(optionalArgument("--mappings"));

verifyGitCheckout(
  rustRoot,
  "a4f584c86358dd07f83f36708573e1c8d1bd8161",
);
verifyGitCheckout(
  javascriptRoot,
  "f8b0911dc9d86265dd62934b7dc782571e3a7fcb",
);

const inventory = {
  schemaVersion: 1,
  sources: {
    rust: {
      package: "automerge",
      version: "0.10.0",
      gitTag: "rust/automerge-0.10.0",
      gitCommit: "a4f584c86358dd07f83f36708573e1c8d1bd8161",
      crateChecksum: "09b78abcbba93428b9465b26cb2816a5b4654cce507f099a84a8c1b311cb3633",
    },
    javascript: {
      package: "@automerge/automerge",
      version: "3.4.0",
      gitTag: "js/automerge-3.4.0",
      gitCommit: "f8b0911dc9d86265dd62934b7dc782571e3a7fcb",
      npmIntegrity: "sha512-THmghtTNGGt2xsI0pM3o1i3PM8oZKcYFgOj25FOzW7l6e94SQOivNtCwy6xc0I8hVJsQSSotoBNs+yk/9hM2dg==",
    },
  },
  tests: [
    ...rustTests(rustRoot, rustTestList),
    ...rustDocumentationTests(rustTestList),
    ...javascriptTests(javascriptRoot),
    ...javascriptPackagingScenarios(),
  ].sort((left, right) => left.id.localeCompare(right.id)),
};

process.stdout.write(`${JSON.stringify(inventory, null, 2)}\n`);

function rustTests(root, testList) {
  if (!testList) {
    throw new Error("--rust-test-list is required for an authoritative inventory");
  }

  const declarations = [];
  for (const file of files(root, ".rs")) {
    const relative = slashPath(path.relative(root, file));
    const source = fs.readFileSync(file, "utf8");
    const expression
      = /#\s*\[\s*test(?:\s*\([^)]*\))?\s*\][\s\S]*?(?:pub(?:\([^)]*\))?\s+)?(?:async\s+)?fn\s+([A-Za-z0-9_]+)/g;
    for (const match of source.matchAll(expression)) {
      const declarationOffset = match.index + match[0].lastIndexOf("fn ");
      declarations.push({
        file: relative,
        line: lineAt(source, declarationOffset),
        name: match[1],
      });
    }
  }

  return rustRuntimeTestNames(testList).map(runtimeName => {
    const name = runtimeName.split("::").at(-1);
    const candidates = declarations.filter(candidate => candidate.name === name);
    const declaration = selectRustDeclaration(runtimeName, candidates);
    const test = testEntry(
      "rust",
      declaration?.file ?? "runtime-test",
      declaration?.line ?? 1,
      name,
    );
    test.id = `rust:${runtimeName}`;
    test.runtimeName = runtimeName;
    return test;
  });
}

function javascriptTests(root) {
  const tests = [];
  for (const file of files(root, ".ts")) {
    const relative = slashPath(path.relative(root, file));
    const source = fs.readFileSync(file, "utf8");
    const executableSource = stripComments(source);
    const expression
      = /\b(?:it|test)(?:\.(?:only|skip|todo))?\s*\(\s*(["'`])((?:\\.|(?!\1)[\s\S])*)\1/g;
    for (const match of executableSource.matchAll(expression)) {
      tests.push(
        testEntry(
          "javascript",
          relative,
          lineAt(source, match.index),
          match[2].replaceAll(/\s+/g, " ").trim(),
        ),
      );
    }
  }
  return tests;
}

function stripComments(source) {
  return source
    .replaceAll(/\/\*[\s\S]*?\*\//g, comment =>
      comment.replaceAll(/[^\n]/g, " ")
    )
    .replaceAll(/^\s*\/\/.*$/gm, comment =>
      comment.replaceAll(/[^\n]/g, " ")
    );
}

function rustDocumentationTests(testList) {
  if (!testList) return [];

  const tests = [];
  const lines = fs.readFileSync(testList, "utf8").split("\n");
  for (let index = 0; index < lines.length; index++) {
    const match = /^(automerge\/.+)\s*:\s*test$/.exec(lines[index]);
    if (!match) continue;

    const name = match[1];
    const sourceMatch = /^(.+?)\s+-\s+.+?\s+\(line\s+(\d+)\)$/.exec(name);
    tests.push(
      testEntry(
        "rust-doc",
        sourceMatch?.[1] ?? "unknown",
        Number(sourceMatch?.[2] ?? index + 1),
        name,
      ),
    );
  }
  return tests;
}

function rustRuntimeTestNames(testList) {
  const tests = [];
  for (const line of fs.readFileSync(testList, "utf8").split("\n")) {
    const match = /^(.+): test$/.exec(line);
    if (!match || match[1].startsWith("automerge/")) continue;
    tests.push(match[1]);
  }
  return tests;
}

function selectRustDeclaration(runtimeName, candidates) {
  if (candidates.length <= 1) return candidates[0];

  const namespace = runtimeName.toLowerCase().replaceAll("_", "");
  return candidates.find(candidate => {
    const file = candidate.file
      .toLowerCase()
      .replaceAll("_", "")
      .replaceAll(".rs", "");
    return namespace.includes(file.split("/").at(-1));
  }) ?? candidates[0];
}

function javascriptPackagingScenarios() {
  const scenarios = [
    "webpack_cjs_fullfat",
    "webpack_cjs_slim",
    "webpack_esm_fullfat",
    "webpack_esm_slim",
    "node_cjs_fullfat",
    "node_cjs_slim",
    "node_esm_fullfat",
    "node_esm_slim",
    "vite_fullfat:vite_dev_server_fullfat",
    "vite_fullfat:vite_build_fullfat",
    "vite_slim:vite_dev_server_slim",
    "vite_slim:vite_build_slim",
    "vite_iife_fullfat",
    "workerd",
    "workerd_slim",
    "iife",
  ];

  return scenarios.map(name => ({
    ...testEntry(
      "javascript-packaging",
      "packaging_tests/run.mjs",
      360,
      name,
    ),
    classification: "language-specific",
    requirement: "language-specific",
    rationale: "Exercises JavaScript package exports, WASM loading, or bundler/runtime integration rather than Go CRDT behavior.",
  }));
}

function testEntry(source, file, line, name) {
  const languageSpecific = languageSpecificRationale(source, file);
  const mapping = mappings.find(candidate =>
    candidate.source === source
    && candidate.file === file
    && candidate.name === name
    && (candidate.line === undefined || candidate.line === line)
  ) ?? builtInCoverage(source, file, name);
  const classification = mapping?.classification
    ?? (languageSpecific ? "language-specific" : "pending");
  return {
    id: `${source}:${file}:${line}:${name}`,
    source,
    file,
    line,
    name,
    classification,
    requirement: classification === "language-specific"
      ? "language-specific"
      : interoperabilityRequirement(source, file, name),
    localTests: mapping?.localTests ?? [],
    rationale: mapping?.rationale ?? languageSpecific,
  };
}

function builtInCoverage(source, file, name) {
  if (source !== "rust" || file !== "tests/batch_insert.rs") {
    return null;
  }
  if (
    /(patch|merges_correctly|scalar_fails|with_transaction|multiple_batch|batch_insert_into_existing_map|batch_put_overwrite_with_nested_structure)/.test(name)
  ) {
    return null;
  }
  if (name.startsWith("splice_")) {
    return {
      classification: "covered",
      localTests: ["TestDocument_HydrateSpliceMatchesReference"],
      rationale: "Hydrated list splice deletion, insertion, nested objects, text, replacement, and Rust parity are exercised.",
    };
  }

  return {
    classification: "covered",
    localTests: [
      "TestDocument_HydrateMatchesReference",
      "TestDocument_HydrateRollback",
    ],
    rationale: "Recursive map/list/text/scalar hydration, save/load, empty values, deep nesting, and rollback execute against native and Rust engines.",
  };
}

function interoperabilityRequirement(source, file, name) {
  if (source === "javascript") {
    if (
      [
        "block_test.ts",
        "change_time.ts",
        "cursors.ts",
        "marks.ts",
        "text_test.ts",
      ].includes(file)
    ) {
      return "interop-required";
    }
    if (
      file === "basic_test.ts"
      && /(ImmutableString|RawString|ints and floats|handle text)/i.test(name)
    ) {
      return "interop-required";
    }
    if (
      file === "legacy_tests.ts"
      && /(Date objects|strings as initial values)/i.test(name)
    ) {
      return "interop-required";
    }

    return "api-convenience";
  }

  return "interop-required";
}

function languageSpecificRationale(source, file) {
  if (
    source === "rust"
    && (
      file === "src/sequence_tree.rs"
      || file === "src/change_graph.rs"
      || file === "src/clock.rs"
      || file === "src/exid.rs"
      || file === "src/legacy/serde_impls/op.rs"
      || file === "src/storage/bundle.rs"
      || file === "src/storage/change/change_op_columns.rs"
      || file === "src/storage/columns/column_specification.rs"
      || file === "src/transaction/inner.rs"
      || file.startsWith("src/columnar/")
      || file.startsWith("src/op_set2/")
      || file.startsWith("src/storage/parse/")
      || file.startsWith("src/text_diff/")
    )
  ) {
    return "Exercises a private Rust data structure or algorithm rather than observable Automerge behavior.";
  }
  if (
    source === "javascript"
    && (
      file === "bundle_test.ts"
      || file === "error.ts"
      || file === "next_test.ts"
      || file === "proxies.ts"
    )
  ) {
    return "Exercises JavaScript packaging or Proxy object semantics that do not exist in the Go API.";
  }
  return "";
}

function files(root, extension) {
  const result = [];
  for (const entry of fs.readdirSync(root, { withFileTypes: true })) {
    const file = path.join(root, entry.name);
    if (entry.isDirectory()) {
      result.push(...files(file, extension));
    } else if (entry.isFile() && file.endsWith(extension)) {
      result.push(file);
    }
  }
  return result;
}

function requiredArgument(name) {
  const index = process.argv.indexOf(name);
  if (index < 0 || index + 1 >= process.argv.length) {
    throw new Error(`missing required argument ${name}`);
  }
  return path.resolve(process.argv[index + 1]);
}

function optionalArgument(name) {
  const index = process.argv.indexOf(name);
  if (index < 0) return null;
  if (index + 1 >= process.argv.length) {
    throw new Error(`missing value for argument ${name}`);
  }
  return path.resolve(process.argv[index + 1]);
}

function loadMappings(file) {
  if (!file) return [];

  const mappings = JSON.parse(fs.readFileSync(file, "utf8"));
  if (!Array.isArray(mappings)) {
    throw new Error("parity mappings must be an array");
  }
  return mappings;
}

function verifyGitCheckout(root, expectedCommit) {
  const actualCommit = execFileSync(
    "git",
    ["-C", root, "rev-parse", "HEAD"],
    { encoding: "utf8" },
  ).trim();
  if (actualCommit !== expectedCommit) {
    throw new Error(
      `source ${root} is at ${actualCommit}, expected ${expectedCommit}`,
    );
  }
}

function slashPath(value) {
  return value.split(path.sep).join("/");
}

function lineAt(source, offset) {
  return source.slice(0, offset).split("\n").length;
}
