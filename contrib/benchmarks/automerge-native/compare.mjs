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

import {
  aggregateSamples,
  enforceBaseline,
} from "./benchmark-lib.mjs";

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
const baselinePath =
  argumentValue("--baseline") ?? process.env.AUTOMERGE_BENCHMARK_BASELINE;
const reportPath = argumentValue("--report");

const scenarios = [
  { workload: "create", size: 0, iterations: 100_000 },
  { workload: "map", size: 100, iterations: 500 },
  { workload: "map", size: 1_000, iterations: 30 },
  { workload: "map-update", size: 1_000, iterations: 1, warmups: 0 },
  { workload: "text", size: 100, iterations: 500 },
  { workload: "text", size: 1_000, iterations: 30 },
  { workload: "text-edit", size: 1_000, iterations: 1, warmups: 0 },
  { workload: "load", size: 10_000, iterations: 50 },
  { workload: "save", size: 10_000, iterations: 1_000 },
  {
    workload: "save-after-loaded-change",
    size: 10_000,
    iterations: 1,
    warmups: 0,
  },
  { workload: "merge-loaded", size: 1_000, iterations: 1, warmups: 0 },
  { workload: "merge-loaded", size: 10_000, iterations: 1, warmups: 0 },
  {
    workload: "concurrent-tail-reconcile",
    size: 10_000,
    iterations: 1,
    warmups: 0,
  },
  { workload: "merge-reloaded", size: 10_000, iterations: 30 },
  { workload: "sync-initial", size: 1_000, iterations: 1, warmups: 0 },
  { workload: "sync-diverged", size: 1_000, iterations: 1, warmups: 0 },
  { workload: "sync-diverged", size: 10_000, iterations: 1, warmups: 0 },
];
const samples = Number(
  argumentValue("--samples") ??
    process.env.AUTOMERGE_BENCHMARK_SAMPLES ??
    7,
);
if (!Number.isInteger(samples) || samples < 3 || samples > 20) {
  throw new Error("--samples must be an integer between 3 and 20");
}

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
      checksum: go.checksum,
      goNS: go.ns,
      rustNS: rust.ns,
      ratio: rust.ns / go.ns,
      goAllocs: go.allocs,
      goAllocatedBytes: go.allocatedBytes,
      goOutputBytes: go.outputBytes,
      goOutputHash: go.outputHash,
      goOutputStable: go.outputStable,
      rustOutputBytes: rust.outputBytes,
      rustOutputHash: rust.outputHash,
      rustOutputStable: rust.outputStable,
      metrics: go.metrics,
    });
  }

  process.stdout.write(`Host: ${os.platform()}/${os.arch()} — ${os.cpus()[0]?.model ?? "unknown CPU"}\n`);
  process.stdout.write(`Go: ${goVersion}\n`);
  process.stdout.write(`Rust: ${rustVersion}\n`);
  process.stdout.write(`Samples: ${samples} (trimmed mean reported)\n\n`);
  process.stdout.write("| Workload | Size | Native Go | Native Rust | Rust/Go | Go allocs | Go bytes |\n");
  process.stdout.write("|---|---:|---:|---:|---:|---:|---:|\n");
  for (const row of rows) {
    process.stdout.write(
      `| ${row.workload} | ${row.size || "—"} | ${duration(row.goNS)} | ${duration(row.rustNS)} | ${ratio(row.ratio)} | ${row.goAllocs.toFixed(0)} | ${row.goAllocatedBytes.toFixed(0)} |\n`,
    );
  }

  const report = {
    schemaVersion: 2,
    metadata: {
      host: `${os.platform()}/${os.arch()}`,
      cpu: os.cpus()[0]?.model ?? "unknown CPU",
      go: goVersion,
      rust: rustVersion,
      samples,
    },
    results: rows.map((row) => ({
      workload: row.workload,
      size: row.size,
      checksum: row.checksum,
      goNsPerOp: row.goNS,
      rustNsPerOp: row.rustNS,
      goToRustRatio: row.goNS / row.rustNS,
      goAllocsPerOp: row.goAllocs,
      goBytesAllocatedPerOp: row.goAllocatedBytes,
      goOutputBytes: row.goOutputBytes,
      goOutputHash: row.goOutputHash,
      goOutputStable: row.goOutputStable,
      rustOutputBytes: row.rustOutputBytes,
      rustOutputHash: row.rustOutputHash,
      rustOutputStable: row.rustOutputStable,
      metrics: row.metrics,
    })),
  };
  if (reportPath) {
    fs.writeFileSync(reportPath, `${JSON.stringify(report, null, 2)}\n`);
  }
  if (baselinePath) {
    enforceBaseline(report, baselinePath);
  }
} finally {
  fs.rmSync(temporaryDirectory, { recursive: true, force: true });
}

function measure(binary, scenario) {
  const results = [];
  for (let sample = 0; sample < samples; sample++) {
    const commandArguments = [
      "--workload",
      scenario.workload,
      "--size",
      String(scenario.size),
      "--iterations",
      String(scenario.iterations),
      "--warmups",
      String(scenario.warmups ?? 3),
    ];
    if (
      scenario.workload === "load" ||
      scenario.workload === "save" ||
      scenario.workload === "save-after-loaded-change"
    ) {
      commandArguments.push("--fixture", fixture);
    }
    const output = execFileSync(
      binary,
      commandArguments,
      { encoding: "utf8" },
    );
    const result = JSON.parse(output);
    results.push(result);
  }
  return aggregateSamples(
    results,
    `${path.basename(binary)} ${scenario.workload}/${scenario.size}`,
    binary === goBinary,
  );
}

function argumentValue(name) {
  const index = process.argv.indexOf(name);
  if (index < 0) return undefined;
  if (index + 1 >= process.argv.length) {
    throw new Error(`${name} requires a path`);
  }
  return process.argv[index + 1];
}

function duration(nanoseconds) {
  if (nanoseconds < 1_000) return `${nanoseconds.toFixed(0)} ns`;
  if (nanoseconds < 1_000_000) return `${(nanoseconds / 1_000).toFixed(2)} µs`;
  if (nanoseconds < 1_000_000_000) {
    return `${(nanoseconds / 1_000_000).toFixed(2)} ms`;
  }
  return `${(nanoseconds / 1_000_000_000).toFixed(2)} s`;
}

function ratio(value) {
  const precision = value < 0.1 ? 3 : 2;
  return `${value.toFixed(precision)}x`;
}
