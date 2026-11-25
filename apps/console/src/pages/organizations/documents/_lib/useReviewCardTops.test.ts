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

import { reviewCardGap } from "./reviewCardLayout";
import { layoutReviewCards } from "./useReviewCardTops";

const heights: Record<string, number> = { "comment": 60, "local:pending": 80 };

function layout(anchors: Parameters<typeof layoutReviewCards>[0], activeKey: string | null = null) {
  return layoutReviewCards(anchors, activeKey, pos => pos, slot => heights[slot]);
}

describe("layoutReviewCards", () => {
  it("keeps a card's top and height when its key changes under the same slot", () => {
    const pending = layout([
      { key: "comment", pos: 0 },
      { key: "local:pending", pos: 10 },
    ]);
    const settled = layout([
      { key: "comment", pos: 0 },
      { key: "suggestion-1", slot: "local:pending", pos: 10 },
    ]);

    expect(settled).toEqual(pending);
    expect(settled.tops["local:pending"]).toBe(60 + reviewCardGap);
  });

  it("finds the active card through its slot", () => {
    const { tops } = layout(
      [
        { key: "comment", pos: 100 },
        { key: "suggestion-1", slot: "local:pending", pos: 110 },
      ],
      "suggestion-1",
    );

    expect(tops["local:pending"]).toBe(110);
    expect(tops.comment).toBe(110 - reviewCardGap - 60);
  });
});
