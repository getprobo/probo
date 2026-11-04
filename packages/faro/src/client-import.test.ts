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

import type { DataRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { FaroClient } from "./client";

vi.mock("./sdk", () => {
  throw new Error("chunk failed");
});

const router = {} as DataRouter;

describe("FaroClient SDK import", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
    vi.restoreAllMocks();
  });

  it("logs a rejected SDK import and still resolves", async () => {
    vi.stubGlobal("document", {
      querySelector(selector: string) {
        const tags: Record<string, string> = {
          "faro-collector-url": "https://collector.example/collect",
          "faro-api-key": "secret",
        };
        const name = selector.match(/name="([^"]+)"/)?.[1];
        if (name === undefined || tags[name] === undefined) {
          return null;
        }
        return { getAttribute: () => tags[name] };
      },
    });
    const error = vi.spyOn(console, "error").mockImplementation(() => {});
    const client = new FaroClient({ appName: "console" });

    await expect(client.start(router)).resolves.toBeUndefined();

    const calls = error.mock.calls as unknown as Array<
      [string, { cause?: { message?: string } }]
    >;
    expect(calls[0]?.[0]).toBe("Cannot start Faro");
    expect(calls[0]?.[1]?.cause?.message).toBe("chunk failed");
  });
});
