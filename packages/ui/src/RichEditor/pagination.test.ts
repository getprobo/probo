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

import { computePageBreaks, type PageLayout } from "./pagination";

// A 100px content area per page: 140 tall with 20px margins, 10px apart.
const layout: PageLayout = { pageHeight: 140, marginTop: 20, marginBottom: 20, gap: 10 };
const stride = layout.pageHeight + layout.gap;

describe("computePageBreaks", () => {
  it("keeps one page while the content fits", () => {
    expect(computePageBreaks([
      { pos: 0, top: 0, height: 40 },
      { pos: 10, top: 50, height: 40 },
    ], layout)).toEqual({ breaks: [], pages: 1 });
  });

  it("moves a block that crosses the bottom to the next page", () => {
    expect(computePageBreaks([
      { pos: 0, top: 0, height: 70 },
      { pos: 10, top: 80, height: 40 },
    ], layout)).toEqual({
      breaks: [{ pos: 10, spacer: stride - 80 }],
      pages: 2,
    });
  });

  it("keeps later blocks after the moved one", () => {
    const result = computePageBreaks([
      { pos: 0, top: 0, height: 70 },
      { pos: 10, top: 80, height: 40 },
      { pos: 20, top: 130, height: 40 },
    ], layout);

    expect(result.breaks).toEqual([{ pos: 10, spacer: stride - 80 }]);
    expect(result.pages).toBe(2);
  });

  it("lets a block taller than a page run past it and starts the next one on a new page", () => {
    const result = computePageBreaks([
      { pos: 0, top: 0, height: 20 },
      { pos: 10, top: 30, height: 250 },
      { pos: 20, top: 290, height: 20 },
    ], layout);

    expect(result.breaks[0]).toEqual({ pos: 10, spacer: stride - 30 });
    const tallEnd = stride + 250;
    const thirdPage = stride * 3;
    expect(result.breaks[1]).toEqual({ pos: 20, spacer: thirdPage - (tallEnd + 10) });
    expect(result.pages).toBe(4);
  });

  it("counts the pages a tall last block covers", () => {
    expect(computePageBreaks([
      { pos: 0, top: 0, height: 260 },
    ], layout).pages).toBe(3);
  });
});
