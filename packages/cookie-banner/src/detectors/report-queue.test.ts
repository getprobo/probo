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

  it("re-sends when a later observation adds a domain", async () => {
    vi.useFakeTimers();

    const fetch = vi.fn().mockResolvedValue({ ok: true, status: 204 });
    vi.stubGlobal("fetch", fetch);

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    queue.reportCookie({ name: "sid", max_age_seconds: null, source: "pre-existing" });

    await vi.advanceTimersByTimeAsync(2_000);
    await Promise.resolve();
    await Promise.resolve();

    expect(fetch).toHaveBeenCalledTimes(1);
    const firstBody = JSON.parse(String(fetch.mock.calls[0]?.[1]?.body)) as {
      cookies: Array<{ cookie_domain?: string }>;
    };
    expect(firstBody.cookies[0]?.cookie_domain).toBeUndefined();

    queue.reportCookie({
      name: "sid",
      max_age_seconds: null,
      source: "pre-existing",
      cookie_domain: "example.com",
      host_only: false,
    });

    await vi.advanceTimersByTimeAsync(2_000);
    await Promise.resolve();
    await Promise.resolve();

    expect(fetch).toHaveBeenCalledTimes(2);
    const secondBody = JSON.parse(String(fetch.mock.calls[1]?.[1]?.body)) as {
      cookies: Array<{ cookie_domain?: string; source: string }>;
    };
    expect(secondBody.cookies[0]?.source).toBe("pre-existing");
    expect(secondBody.cookies[0]?.cookie_domain).toBe("example.com");
  });

  it("replaces a weaker observation's max-age when source promotes", async () => {
    vi.useFakeTimers();

    const sendBeacon = vi.fn((_url: string, _data?: BodyInit | null) => true);
    vi.stubGlobal("navigator", { sendBeacon });

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    queue.reportCookie({ name: "sid", max_age_seconds: 3600, source: "pre-existing" });
    queue.reportCookie({ name: "sid", max_age_seconds: null, source: "script" });
    queue.stop();

    expect(sendBeacon).toHaveBeenCalledOnce();
    const sentData: unknown = sendBeacon.mock.calls[0]?.[1];
    expect(sentData).toBeInstanceOf(Blob);
    if (!(sentData instanceof Blob)) {
      throw new Error("expected sendBeacon to receive a Blob");
    }
    const body = JSON.parse(await sentData.text()) as {
      cookies: Array<{ source: string; max_age_seconds: number | null }>;
    };
    expect(body.cookies[0]?.source).toBe("script");
    expect(body.cookies[0]?.max_age_seconds).toBeNull();
  });

  it("promotes a pre-existing cookie to script", async () => {
    vi.useFakeTimers();

    const sendBeacon = vi.fn((_url: string, _data?: BodyInit | null) => true);
    vi.stubGlobal("navigator", { sendBeacon });

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    queue.reportCookie({ name: "sid", max_age_seconds: null, source: "pre-existing" });
    queue.reportCookie({ name: "sid", max_age_seconds: 3600, source: "script" });
    queue.stop();

    expect(sendBeacon).toHaveBeenCalledOnce();
    const sentData: unknown = sendBeacon.mock.calls[0]?.[1];
    expect(sentData).toBeInstanceOf(Blob);
    if (!(sentData instanceof Blob)) {
      throw new Error("expected sendBeacon to receive a Blob");
    }
    const body = JSON.parse(await sentData.text()) as {
      cookies: Array<{ source: string; max_age_seconds: number | null }>;
    };
    expect(body.cookies[0]?.source).toBe("script");
    expect(body.cookies[0]?.max_age_seconds).toBe(3600);
  });

  it("keeps extension when an HTTP echo arrives later", async () => {
    vi.useFakeTimers();

    const sendBeacon = vi.fn((_url: string, _data?: BodyInit | null) => true);
    vi.stubGlobal("navigator", { sendBeacon });

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    queue.reportCookie({ name: "ext", max_age_seconds: null, source: "extension" });
    queue.reportCookie({ name: "ext", max_age_seconds: null, source: "http" });
    queue.stop();

    expect(sendBeacon).toHaveBeenCalledOnce();
    const sentData: unknown = sendBeacon.mock.calls[0]?.[1];
    expect(sentData).toBeInstanceOf(Blob);
    if (!(sentData instanceof Blob)) {
      throw new Error("expected sendBeacon to receive a Blob");
    }
    const body = JSON.parse(await sentData.text()) as {
      cookies: Array<{ source: string }>;
    };
    expect(body.cookies[0]?.source).toBe("extension");
  });

  it("sends a richer replacement that arrives during an in-flight flush", async () => {
    vi.useFakeTimers();

    let resolveFirst: ((value: { ok: boolean; status: number }) => void) | undefined;
    const first = new Promise<{ ok: boolean; status: number }>((resolve) => {
      resolveFirst = resolve;
    });
    const fetch = vi.fn()
      .mockImplementationOnce(() => first)
      .mockResolvedValue({ ok: true, status: 204 });
    vi.stubGlobal("fetch", fetch);

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    queue.reportCookie({ name: "sid", max_age_seconds: null, source: "pre-existing" });

    await vi.advanceTimersByTimeAsync(2_000);
    await Promise.resolve();

    queue.reportCookie({ name: "sid", max_age_seconds: null, source: "script" });
    resolveFirst?.({ ok: true, status: 204 });
    await Promise.resolve();
    await Promise.resolve();
    await vi.advanceTimersByTimeAsync(2_000);
    await Promise.resolve();

    expect(fetch).toHaveBeenCalledTimes(2);
    const secondBody = JSON.parse(String(fetch.mock.calls[1]?.[1]?.body)) as {
      cookies: Array<{ source: string }>;
    };
    expect(secondBody.cookies[0]?.source).toBe("script");
  });
});
