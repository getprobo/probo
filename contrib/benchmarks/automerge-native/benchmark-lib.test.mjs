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

import assert from "node:assert/strict";
import test from "node:test";

import {
  aggregateSamples,
  enforceBaseline,
  trimmedAggregate,
} from "./benchmark-lib.mjs";

test("trimmedAggregate removes one outlier from each side", () => {
  assert.equal(trimmedAggregate([1, 2, 3, 4, 5, 6, 100]), 4);
});

test("aggregateSamples requires stable semantic and output values", () => {
  const samples = [10, 20, 30].map((nsPerOp) => ({
    nsPerOp,
    checksum: "semantic",
    outputBytes: 12,
    outputHash: "wire",
    allocsPerOp: 2,
    bytesAllocatedPerOp: 64,
    metrics: { globalOrderFallbacks: 0 },
  }));

  assert.deepEqual(aggregateSamples(samples, "save/10"), {
    checksum: "semantic",
    outputBytes: 12,
    outputHash: "wire",
    outputStable: true,
    ns: 20,
    allocs: 2,
    allocatedBytes: 64,
    metrics: { globalOrderFallbacks: 0 },
  });
  samples[2].outputHash = "changed";
  assert.throws(
    () => aggregateSamples(samples, "save/10"),
    /unstable outputHash/,
  );
});

test("enforceBaseline checks ratios, resources, output, and counters", () => {
  const report = {
    metadata: { host: "test/host" },
    results: [{
      workload: "save",
      size: 10,
      goNsPerOp: 120,
      goToRustRatio: 1.2,
      goAllocsPerOp: 2,
      goBytesAllocatedPerOp: 64,
      goOutputBytes: 12,
      goOutputHash: "wire",
      metrics: { fullColumnEncodings: 0 },
    }],
  };
  const baseline = {
    schemaVersion: 2,
    gates: [{
      workload: "save",
      size: 10,
      maxGoToRustRatio: 1.3,
      maxGoAllocsPerOp: 2,
      maxGoBytesAllocatedPerOp: 64,
      referenceGoNsPerOp: 100,
      maxGoSelfRegression: 1.25,
      selfRegressionHost: "test/host",
      goOutputBytes: 12,
      goOutputHash: "wire",
      maxMetrics: { fullColumnEncodings: 0 },
    }],
  };

  assert.doesNotThrow(() => enforceBaseline(report, baseline));
  baseline.gates[0].maxMetrics.fullColumnEncodings = -1;
  assert.throws(() => enforceBaseline(report, baseline), /fullColumnEncodings/);
});

test("enforceBaseline checks workload scaling", () => {
  const report = {
    metadata: { host: "test/host" },
    results: [
      { workload: "merge", size: 1000, goNsPerOp: 10 },
      { workload: "merge", size: 10000, goNsPerOp: 90 },
    ],
  };
  const baseline = {
    schemaVersion: 2,
    gates: [{
      workload: "merge",
      size: 10000,
      scaleFromSize: 1000,
      maxGoScaleRatio: 10,
    }],
  };

  assert.doesNotThrow(() => enforceBaseline(report, baseline));
  baseline.gates[0].maxGoScaleRatio = 8;
  assert.throws(() => enforceBaseline(report, baseline), /Go scaling/);
});
