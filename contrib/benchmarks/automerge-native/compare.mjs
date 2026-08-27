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

import { execFileSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const directory = path.dirname(fileURLToPath(import.meta.url));
const repository = path.resolve(directory, "../../..");
const temporaryDirectory = fs.mkdtempSync(
  path.join(os.tmpdir(), "probo-automerge-benchmark-"),
);
const goBinary = path.join(temporaryDirectory, "go-benchmark");
const fixture = path.join(temporaryDirectory, "fixture");
const rustManifest = path.join(directory, "rust", "Cargo.toml");
const rustTarget = path.join(directory, "rust", "target");
const rustBinary = path.join(
  rustTarget,
  "release",
  "probo-automerge-native-benchmark",
);

const scenarios = [
  { workload: "create", size: 0, iterations: 100_000 },
  { workload: "map", size: 100, iterations: 500 },
  { workload: "map", size: 1_000, iterations: 30 },
  { workload: "text", size: 100, iterations: 500 },
  { workload: "text", size: 1_000, iterations: 30 },
  { workload: "load", size: 10_000, iterations: 50 },
  { workload: "save", size: 10_000, iterations: 1_000 },
];
const samples = 3;

try {
  const goVersion = execFileSync("go", ["version"], {
    encoding: "utf8",
  }).trim();
  const rustVersion = execFileSync("rustc", ["+1.90.0", "--version"], {
    encoding: "utf8",
  }).trim();

  execFileSync(
    "go",
    [
      "build",
      "-trimpath",
      "-ldflags=-s -w",
      "-o",
      goBinary,
      "./contrib/benchmarks/automerge-native/go",
    ],
    { cwd: repository, stdio: "inherit" },
  );
  execFileSync(
    "cargo",
    [
      "+1.90.0",
      "build",
      "--release",
      "--locked",
      "--manifest-path",
      rustManifest,
    ],
    {
      cwd: repository,
      env: { ...process.env, CARGO_TARGET_DIR: rustTarget },
      stdio: "inherit",
    },
  );
  execFileSync(
    goBinary,
    [
      "--workload",
      "fixture",
      "--size",
      "10000",
      "--fixture",
      fixture,
    ],
    { stdio: "inherit" },
  );

  const rows = [];
  for (const scenario of scenarios) {
    const go = measure(goBinary, scenario);
    const rust = measure(rustBinary, scenario);
    if (go.checksum !== rust.checksum) {
      throw new Error(
        `checksum mismatch for ${scenario.workload}/${scenario.size}: Go ${go.checksum}, Rust ${rust.checksum}`,
      );
    }

    rows.push({
      workload: scenario.workload,
      size: scenario.size,
      goNS: go.ns,
      rustNS: rust.ns,
      ratio: rust.ns / go.ns,
    });
  }

  process.stdout.write(`Host: ${os.platform()}/${os.arch()} — ${os.cpus()[0]?.model ?? "unknown CPU"}\n`);
  process.stdout.write(`Go: ${goVersion}\n`);
  process.stdout.write(`Rust: ${rustVersion}\n`);
  process.stdout.write(`Samples: ${samples} (median reported)\n\n`);
  process.stdout.write("| Workload | Size | Native Go | Native Rust | Rust/Go |\n");
  process.stdout.write("|---|---:|---:|---:|---:|\n");
  for (const row of rows) {
    process.stdout.write(
      `| ${row.workload} | ${row.size || "—"} | ${duration(row.goNS)} | ${duration(row.rustNS)} | ${row.ratio.toFixed(2)}x |\n`,
    );
  }
} finally {
  fs.rmSync(temporaryDirectory, { recursive: true, force: true });
}

function measure(binary, scenario) {
  const values = [];
  let checksum;
  for (let sample = 0; sample < samples; sample++) {
    const commandArguments = [
      "--workload",
      scenario.workload,
      "--size",
      String(scenario.size),
      "--iterations",
      String(scenario.iterations),
      "--warmups",
      "3",
    ];
    if (scenario.workload === "load" || scenario.workload === "save") {
      commandArguments.push("--fixture", fixture);
    }
    const output = execFileSync(
      binary,
      commandArguments,
      { encoding: "utf8" },
    );
    const result = JSON.parse(output);
    values.push(Number(result.nsPerOp));
    checksum ??= result.checksum;
    if (checksum !== result.checksum) {
      throw new Error(
        `${binary} produced unstable checksums for ${scenario.workload}/${scenario.size}`,
      );
    }
  }
  values.sort((left, right) => left - right);
  return {
    checksum,
    ns: values[Math.floor(values.length / 2)],
  };
}

function duration(nanoseconds) {
  if (nanoseconds < 1_000) return `${nanoseconds.toFixed(0)} ns`;
  if (nanoseconds < 1_000_000) return `${(nanoseconds / 1_000).toFixed(2)} µs`;
  if (nanoseconds < 1_000_000_000) {
    return `${(nanoseconds / 1_000_000).toFixed(2)} ms`;
  }
  return `${(nanoseconds / 1_000_000_000).toFixed(2)} s`;
}
