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

import { placeReviewCards, reviewCardGap } from "./reviewCardLayout";

describe("placeReviewCards", () => {
  it("keeps cards on their anchors when they do not overlap", () => {
    expect(placeReviewCards([
      { key: "a", anchor: 0, height: 40 },
      { key: "b", anchor: 100, height: 40 },
    ], null)).toEqual({ a: 0, b: 100 });
  });

  it("pushes overlapping cards down in anchor order", () => {
    expect(placeReviewCards([
      { key: "b", anchor: 10, height: 40 },
      { key: "a", anchor: 0, height: 40 },
    ], null)).toEqual({ a: 0, b: 40 + reviewCardGap });
  });

  it("keeps the active card on its anchor and moves earlier cards up", () => {
    expect(placeReviewCards([
      { key: "a", anchor: 100, height: 40 },
      { key: "b", anchor: 110, height: 40 },
    ], "b")).toEqual({ a: 110 - reviewCardGap - 40, b: 110 });
  });

  it("pushes the active card down when earlier cards cannot move up", () => {
    expect(placeReviewCards([
      { key: "a", anchor: 0, height: 40 },
      { key: "b", anchor: 10, height: 40 },
    ], "b")).toEqual({ a: 0, b: 40 + reviewCardGap });
  });

  it("places cards after the active card below it", () => {
    expect(placeReviewCards([
      { key: "a", anchor: 50, height: 40 },
      { key: "b", anchor: 60, height: 40 },
    ], "a")).toEqual({ a: 50, b: 50 + 40 + reviewCardGap });
  });
});
