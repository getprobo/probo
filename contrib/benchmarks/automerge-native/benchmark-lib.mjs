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

export function trimmedAggregate(values) {
  if (values.length === 0) throw new Error("cannot aggregate zero samples");
  const sorted = [...values].sort((left, right) => left - right);
  const trim = sorted.length >= 5 ? Math.floor(sorted.length * 0.15) : 0;
  const retained = sorted.slice(trim, sorted.length - trim);

  return retained.reduce((sum, value) => sum + value, 0) / retained.length;
}

export function aggregateSamples(samples, key, requireStableOutput = true) {
  const checksum = stableValue(samples, "checksum", key);
  const outputStable = samples.every(
    (sample) =>
      sample.outputBytes === samples[0].outputBytes &&
      sample.outputHash === samples[0].outputHash,
  );
  const outputBytes = requireStableOutput
    ? stableValue(samples, "outputBytes", key)
    : trimmedAggregate(samples.map((sample) => Number(sample.outputBytes ?? 0)));
  const outputHash = requireStableOutput
    ? stableValue(samples, "outputHash", key)
    : outputStable
      ? samples[0].outputHash
      : "";
  const metricNames = Object.keys(samples[0].metrics ?? {});
  const metrics = Object.fromEntries(
    metricNames.map((name) => [
      name,
      Math.max(
        ...samples.map(
          (sample) =>
            Number(sample.metrics?.[name] ?? 0) /
            Number(sample.iterations ?? 1),
        ),
      ),
    ]),
  );

  return {
    checksum,
    outputBytes,
    outputHash,
    outputStable,
    ns: trimmedAggregate(samples.map((sample) => Number(sample.nsPerOp))),
    allocs: trimmedAggregate(
      samples.map((sample) => Number(sample.allocsPerOp ?? 0)),
    ),
    allocatedBytes: trimmedAggregate(
      samples.map((sample) => Number(sample.bytesAllocatedPerOp ?? 0)),
    ),
    metrics,
  };
}

export function enforceBaseline(report, fileOrBaseline) {
  const baseline =
    typeof fileOrBaseline === "string"
      ? JSON.parse(fs.readFileSync(fileOrBaseline, "utf8"))
      : fileOrBaseline;
  if (baseline.schemaVersion !== 2 || !Array.isArray(baseline.gates)) {
    throw new Error(
      "benchmark baseline must have schemaVersion 2 and a gates array",
    );
  }

  const results = new Map(
    report.results.map((result) => [
      `${result.workload}/${result.size}`,
      result,
    ]),
  );
  const failures = [];
  for (const gate of baseline.gates) {
    const key = `${gate.workload}/${gate.size}`;
    const result = results.get(key);
    if (!result) {
      failures.push(`${key}: benchmark result is missing`);
      continue;
    }

    checkMaximum(failures, key, "Go/Rust", result.goToRustRatio, gate.maxGoToRustRatio, "x");
    if (gate.scaleFromSize !== undefined && gate.maxGoScaleRatio !== undefined) {
      const smaller = results.get(`${gate.workload}/${gate.scaleFromSize}`);
      if (!smaller) {
        failures.push(
          `${key}: scaling result ${gate.workload}/${gate.scaleFromSize} is missing`,
        );
      } else {
        checkMaximum(
          failures,
          key,
          "Go scaling",
          result.goNsPerOp / smaller.goNsPerOp,
          gate.maxGoScaleRatio,
          "x",
        );
      }
    }
    checkMaximum(failures, key, "Go allocs/op", result.goAllocsPerOp, gate.maxGoAllocsPerOp);
    checkMaximum(
      failures,
      key,
      "Go allocated bytes/op",
      result.goBytesAllocatedPerOp,
      gate.maxGoBytesAllocatedPerOp,
    );
    if (
      gate.referenceGoNsPerOp !== undefined &&
      gate.maxGoSelfRegression !== undefined &&
      appliesToHost(gate, report.metadata)
    ) {
      checkMaximum(
        failures,
        key,
        "Go self-regression",
        result.goNsPerOp / gate.referenceGoNsPerOp,
        gate.maxGoSelfRegression,
        "x",
      );
    }
    checkExact(
      failures,
      key,
      "Go output bytes",
      result.goOutputBytes,
      gate.goOutputBytes,
    );
    checkExact(
      failures,
      key,
      "Go output hash",
      result.goOutputHash,
      gate.goOutputHash,
    );
    for (const [name, maximum] of Object.entries(gate.maxMetrics ?? {})) {
      checkMaximum(
        failures,
        key,
        name,
        Number(result.metrics?.[name] ?? 0),
        maximum,
      );
    }
  }
  if (failures.length > 0) {
    throw new Error(`benchmark gates failed:\n${failures.join("\n")}`);
  }
}

function stableValue(samples, name, key) {
  const first = samples[0][name];
  if (samples.some((sample) => sample[name] !== first)) {
    throw new Error(`${key} produced unstable ${name} values`);
  }
  return first;
}

function appliesToHost(gate, metadata) {
  return gate.selfRegressionHost === undefined ||
    gate.selfRegressionHost === metadata.host;
}

function checkMaximum(failures, key, name, actual, maximum, suffix = "") {
  if (maximum === undefined) return;
  if (typeof maximum !== "number" || maximum < 0) {
    failures.push(`${key}: maximum for ${name} must be non-negative`);
  } else if (actual > maximum) {
    failures.push(
      `${key}: ${name} ${actual.toFixed(2)}${suffix} exceeds ${maximum.toFixed(2)}${suffix}`,
    );
  }
}

function checkExact(failures, key, name, actual, expected) {
  if (expected !== undefined && actual !== expected) {
    failures.push(`${key}: ${name} ${actual} does not equal ${expected}`);
  }
}
