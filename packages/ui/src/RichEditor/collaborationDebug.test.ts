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

import { afterEach, describe, expect, it, vi } from "vitest";

import { collaborationDebug } from "./collaborationDebug";

describe("collaboration debug", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("does not build details when debugging is disabled", () => {
    expect(() => collaborationDebug("edit", () => {
      throw new Error("details should stay lazy");
    })).not.toThrow();
  });

  it("treats inaccessible storage as debugging disabled", () => {
    vi.stubGlobal("window", {
      localStorage: {
        getItem() {
          throw new DOMException("denied", "SecurityError");
        },
      },
    });

    expect(() => collaborationDebug("edit", () => {
      throw new Error("details should stay lazy");
    })).not.toThrow();
  });

  it("logs details when debugging is enabled", () => {
    vi.stubGlobal("window", {
      localStorage: {
        getItem: () => "1",
      },
    });
    const info = vi.spyOn(console, "info").mockImplementation(() => {});

    collaborationDebug("edit", () => ({ heads: 1 }));

    expect(info).toHaveBeenCalledWith(
      "[probo:collaboration]",
      "edit",
      JSON.stringify({ heads: 1 }),
    );
  });
});
