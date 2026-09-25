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

import { ReportQueue } from "./report-queue";

describe("ReportQueue", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("sends unload reports as CORS-safelisted JSON", async () => {
    vi.useFakeTimers();

    const sendBeacon = vi.fn((_url: string, _data?: BodyInit | null) => true);
    vi.stubGlobal("navigator", { sendBeacon });

    const reportUrl = new URL("https://api.example.com/banner/report");
    const queue = new ReportQueue(reportUrl);
    queue.reportCookie({
      name: "analytics",
      max_age_seconds: 3600,
      source: "script",
    });

    queue.stop();

    expect(sendBeacon).toHaveBeenCalledOnce();
    expect(sendBeacon).toHaveBeenCalledWith(reportUrl.toString(), expect.any(Blob));
    const sentData: unknown = sendBeacon.mock.calls[0]?.[1];
    expect(sentData).toBeInstanceOf(Blob);
    if (!(sentData instanceof Blob)) {
      throw new Error("expected sendBeacon to receive a Blob");
    }

    expect(sentData.type).toBe("text/plain;charset=utf-8");
    await expect(sentData.text()).resolves.toBe(
      JSON.stringify({
        cookies: [
          {
            name: "analytics",
            max_age_seconds: 3600,
            source: "script",
          },
        ],
      }),
    );
  });
});
