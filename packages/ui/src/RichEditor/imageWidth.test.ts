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

import { describe, expect, it } from "vitest";

import { dragImageWidth, imageWidthPercent, keyboardImageWidth, snapImageWidth } from "./imageWidth";

describe("imageWidthPercent", () => {
  it("reads a column percentage", () => {
    expect(imageWidthPercent("50%")).toBe(50);
    expect(imageWidthPercent(" 12% ")).toBe(12);
  });

  it("reads a full column and drops other values", () => {
    expect(imageWidthPercent("100%")).toBe(100);
    expect(imageWidthPercent("0%")).toBeNull();
    expect(imageWidthPercent("300")).toBeNull();
    expect(imageWidthPercent("40px")).toBeNull();
    expect(imageWidthPercent(null)).toBeNull();
  });
});

describe("dragImageWidth", () => {
  const field = { slotLeft: 100, slotRight: 500, slotWidth: 400, targets: [] as number[] };

  it("uses the full column while the pointer is outside the field", () => {
    expect(dragImageWidth({
      side: "right",
      clientX: 520,
      startX: 300,
      startPercent: 50,
      ...field,
    })).toBe(100);

    expect(dragImageWidth({
      side: "right",
      clientX: 300,
      startX: 300,
      startPercent: 50,
      ...field,
    })).toBe(50);
  });

  it("uses the full column past the left edge", () => {
    expect(dragImageWidth({
      side: "left",
      clientX: 80,
      startX: 200,
      startPercent: 50,
      ...field,
    })).toBe(100);
  });

  it("keeps tracking while the pointer stays inside", () => {
    expect(dragImageWidth({
      side: "right",
      clientX: 280,
      startX: 300,
      startPercent: 50,
      ...field,
    })).toBe(40);
  });
});

describe("keyboardImageWidth", () => {
  it("grows on enter and space from either handle", () => {
    expect(keyboardImageWidth({ side: "left", key: "Enter", width: 40, targets: [] })).toBe(50);
    expect(keyboardImageWidth({ side: "right", key: " ", width: 40, targets: [] })).toBe(50);
  });

  it("follows the handle direction for arrow keys", () => {
    expect(keyboardImageWidth({ side: "right", key: "ArrowRight", width: 40, targets: [] })).toBe(50);
    expect(keyboardImageWidth({ side: "right", key: "ArrowLeft", width: 40, targets: [] })).toBe(30);
    expect(keyboardImageWidth({ side: "left", key: "ArrowLeft", width: 40, targets: [] })).toBe(50);
    expect(keyboardImageWidth({ side: "left", key: "ArrowRight", width: 40, targets: [] })).toBe(30);
  });

  it("ignores other keys and stays in range", () => {
    expect(keyboardImageWidth({ side: "right", key: "Tab", width: 40, targets: [] })).toBeNull();
    expect(keyboardImageWidth({ side: "right", key: "Enter", width: 96, targets: [] })).toBe(100);
    expect(keyboardImageWidth({ side: "left", key: "ArrowRight", width: 12, targets: [] })).toBe(10);
  });
});

describe("snapImageWidth", () => {
  it("snaps to another image and to the full column", () => {
    expect(snapImageWidth(52, [50])).toBe(50);
    expect(snapImageWidth(98, [])).toBe(100);
    expect(snapImageWidth(40, [50])).toBe(40);
  });

  it("stays inside the drag range", () => {
    expect(snapImageWidth(1, [])).toBe(10);
    expect(snapImageWidth(140, [])).toBe(100);
  });
});
