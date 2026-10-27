// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import { pushSSE } from "./sse";

const label = "List_query$defer$Status_connector";

describe("pushSSE", () => {
  it("joins events split across chunks", () => {
    const first = pushSSE("", "event: next\ndata: {\"data\":{\"a\"");
    expect(first.payloads).toEqual([]);

    const second = pushSSE(first.buffer, ":1},\"hasNext\":true}\n\n");
    expect(second.payloads).toEqual([{ data: { a: 1 }, hasNext: true }]);
    expect(second.buffer).toBe("");
  });

  it("ignores the prelude comment and keep-alive pings", () => {
    const push = pushSSE(
      "",
      ":\n\n: ping\n\nevent: next\ndata: {\"data\":{},\"hasNext\":true}\n\n",
    );

    expect(push.payloads).toEqual([{ data: {}, hasNext: true }]);
    expect(push.completed).toBe(false);
  });

  it("accepts CRLF framing", () => {
    const push = pushSSE(
      "",
      "event: next\r\ndata: {\"data\":{\"a\":1}}\r\n\r\n",
    );

    expect(push.payloads).toEqual([{ data: { a: 1 } }]);
  });

  it("marks the last deferred payload final and stops at complete", () => {
    const push = pushSSE(
      "",
      "event: next\ndata: {\"data\":{},\"hasNext\":true}\n\n"
      + `event: next\ndata: {"data":{"s":1},"label":"${label}","path":["node"],"hasNext":false}\n\n`
      + "event: complete\n\n"
      + "event: next\ndata: {\"data\":{}}\n\n",
    );

    expect(push.completed).toBe(true);
    expect(push.payloads).toHaveLength(2);
    expect(push.payloads[1]).toEqual({
      data: { s: 1 },
      label,
      path: ["node"],
      hasNext: false,
      extensions: { is_final: true },
    });
  });

  it("leaves intermediate deferred payloads open", () => {
    const push = pushSSE(
      "",
      `event: next\ndata: {"data":{"s":1},"label":"${label}","path":["node"],"hasNext":true}\n\n`,
    );

    expect(push.payloads).toEqual([
      { data: { s: 1 }, label, path: ["node"], hasNext: true },
    ]);
    expect(push.payloads[0]).not.toHaveProperty("extensions");
  });
});
