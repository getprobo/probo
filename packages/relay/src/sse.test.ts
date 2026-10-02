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

import { pushSSE } from "./sse";

describe("pushSSE", () => {
  it("reassembles an event split across chunks", () => {
    const first = pushSSE("", "event: next\ndata: {\"data\":{\"name\":\"rea");
    expect(first.payloads).toEqual([]);
    expect(first.completed).toBe(false);

    const second = pushSSE(first.buffer, "dy\"}}\n\n");
    expect(second.completed).toBe(false);
    expect(second.payloads).toEqual([{ data: { name: "ready" } }]);
    expect(second.buffer).toBe("");
  });

  it("ignores the prelude comment and keep-alive pings", () => {
    const pushed = pushSSE("", ":\n\n: ping\n\nevent: next\ndata: {\"data\":{\"name\":\"ready\"}}\n\n");
    expect(pushed.payloads).toEqual([{ data: { name: "ready" } }]);
    expect(pushed.completed).toBe(false);
  });

  it("normalizes CRLF frames", () => {
    const pushed = pushSSE("", "event: next\r\ndata: {\"data\":{\"ok\":true}}\r\n\r\n");
    expect(pushed.payloads).toEqual([{ data: { ok: true } }]);
  });

  it("stops at event complete and leaves nothing pending", () => {
    const pushed = pushSSE(
      "",
      "event: next\ndata: {\"data\":{\"name\":\"ready\"},\"hasNext\":true}\n\n"
      + "event: next\ndata: {\"data\":{\"name\":\"later\"},\"label\":\"$defer$Status\",\"path\":[\"node\"],\"hasNext\":false}\n\n"
      + "event: complete\n\n"
      + "event: next\ndata: {\"data\":{\"name\":\"ignored\"}}\n\n",
    );

    expect(pushed.completed).toBe(true);
    expect(pushed.payloads).toHaveLength(2);
    expect(pushed.payloads[0]).toEqual({
      data: { name: "ready" },
      hasNext: true,
    });
    expect(pushed.payloads[1]).toMatchObject({
      data: { name: "later" },
      label: "$defer$Status",
      path: ["node"],
      hasNext: false,
      extensions: { is_final: true },
    });
  });

  it("leaves an incremental payload open while more events are coming", () => {
    const pushed = pushSSE(
      "",
      "event: next\ndata: {\"data\":{\"name\":\"part\"},\"label\":\"$defer$Status\",\"path\":[\"node\",0],\"hasNext\":true}\n\n",
    );

    expect(pushed.payloads[0]?.extensions).toBeUndefined();
  });
});
