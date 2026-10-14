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

/* eslint-disable @typescript-eslint/no-floating-promises -- node:test registers suites without awaiting them */

import assert from "node:assert/strict";
import { describe, it } from "node:test";

import {
  overflowInternalControls,
  previewInternalControls,
  visibleInternalControlLimit,
} from "./visibleInternalControls";

const controls = ["a", "b", "c", "d", "e", "f"];

describe("previewInternalControls", () => {
  it("keeps a short list intact", () => {
    const linked = controls.slice(0, 2);

    assert.deepEqual(previewInternalControls(linked), linked);
    assert.equal(linked.length, 2);
  });

  it("shows only the preview limit", () => {
    assert.deepEqual(
      previewInternalControls(controls),
      controls.slice(0, visibleInternalControlLimit),
    );
    assert.equal(controls.length, 6);
  });
});

describe("overflowInternalControls", () => {
  it("hides the remainder until the list is expanded", () => {
    assert.deepEqual(overflowInternalControls(controls, false, true), []);
    assert.deepEqual(overflowInternalControls(controls, false, false), []);
  });

  it("keeps the editor preview in place and reveals the rest", () => {
    assert.deepEqual(
      overflowInternalControls(controls, true, true),
      controls.slice(visibleInternalControlLimit),
    );
    assert.equal(controls.length, 6);
  });

  it("reveals every read-only link", () => {
    assert.deepEqual(overflowInternalControls(controls, true, false), controls);
  });
});
