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

import { COOKIE_NAME } from "../cookie";
import { CookieDetector } from "./cookie-detector";
import { ReportQueue } from "./report-queue";
import type { DetectedCookieEntry } from "./types";

interface CookieChangeLike {
  changed: Array<{ name: string; expires: number | null; domain: string | null }>;
}

function installFakeCookieDocument(): {
  document: { cookie: string };
  emitCookieStoreChange: (event: CookieChangeLike) => void;
} {
  let store = "";

  const proto: { cookie: string } = Object.defineProperty({}, "cookie", {
    configurable: true,
    enumerable: true,
    get(): string {
      return store;
    },
    set(value: string) {
      const pair = String(value).split(";")[0] ?? "";
      const name = pair.split("=")[0]?.trim() ?? "";
      if (!name) return;

      if (/(?:^|;\s*)max-age=0(?:;|$)/i.test(value) || /(?:^|;\s*)expires=Thu, 01 Jan 1970/i.test(value)) {
        store = store
          .split("; ")
          .filter((entry) => entry && !entry.startsWith(`${name}=`))
          .join("; ");
        return;
      }

      const next = store
        .split("; ")
        .filter((entry) => entry && !entry.startsWith(`${name}=`));
      next.push(pair);
      store = next.join("; ");
    },
  }) as { cookie: string };

  function FakeDocument(): void {}
  FakeDocument.prototype = proto;

  const doc = Object.create(proto) as { cookie: string };

  const listeners = new Set<(event: CookieChangeLike) => void>();

  vi.stubGlobal("Document", FakeDocument);
  vi.stubGlobal("HTMLDocument", FakeDocument);
  vi.stubGlobal("document", doc);
  vi.stubGlobal("location", { hostname: "example.com", origin: "https://example.com" });
  vi.stubGlobal("cookieStore", {
    addEventListener(_type: string, handler: (event: CookieChangeLike) => void) {
      listeners.add(handler);
    },
    removeEventListener(_type: string, handler: (event: CookieChangeLike) => void) {
      listeners.delete(handler);
    },
    getAll: async () => [],
  });

  return {
    document: doc,
    emitCookieStoreChange(event: CookieChangeLike) {
      for (const listener of listeners) {
        listener(event);
      }
    },
  };
}

function fakeQueue(): { queue: ReportQueue; reported: DetectedCookieEntry[] } {
  const reported: DetectedCookieEntry[] = [];
  const queue = {
    reportCookie(entry: DetectedCookieEntry) {
      reported.push(entry);
    },
    onNotFound() {},
  } as unknown as ReportQueue;

  return { queue, reported };
}

describe("CookieDetector", () => {
  afterEach(() => {
    vi.useRealTimers();
    vi.unstubAllGlobals();
  });

  it("reports document.cookie writes as script via the prototype setter", () => {
    const { document } = installFakeCookieDocument();
    const { queue, reported } = fakeQueue();
    const detector = new CookieDetector(queue, "https://api.example.com", new Set([COOKIE_NAME]));

    detector.start();
    document.cookie = "foo=bar; path=/";

    expect(reported).toEqual([
      expect.objectContaining({ name: "foo", source: "script" }),
    ]);

    detector.stop();
  });

  it("does not report the CMP consent cookie", () => {
    const { document } = installFakeCookieDocument();
    const { queue, reported } = fakeQueue();
    const detector = new CookieDetector(queue, "https://api.example.com", new Set([COOKIE_NAME]));

    detector.start();
    document.cookie = `${COOKIE_NAME}=abc; path=/`;

    expect(reported).toEqual([]);

    detector.stop();
  });

  it("keeps script when cookieStore echoes the same write as http", async () => {
    vi.useFakeTimers();

    const { document, emitCookieStoreChange } = installFakeCookieDocument();

    const sendBeacon = vi.fn((_url: string, _data?: BodyInit | null) => true);
    vi.stubGlobal("navigator", { sendBeacon });

    const queue = new ReportQueue(new URL("https://api.example.com/banner/report"));
    const detector = new CookieDetector(queue, "https://api.example.com", new Set([COOKIE_NAME]));

    detector.start();
    document.cookie = "foo=bar; path=/";
    emitCookieStoreChange({
      changed: [{ name: "foo", expires: null, domain: null }],
    });
    queue.stop();

    expect(sendBeacon).toHaveBeenCalledOnce();
    const sentData: unknown = sendBeacon.mock.calls[0]?.[1];
    expect(sentData).toBeInstanceOf(Blob);
    if (!(sentData instanceof Blob)) {
      throw new Error("expected sendBeacon to receive a Blob");
    }

    const body = JSON.parse(await sentData.text()) as {
      cookies: Array<{ name: string; source: string }>;
    };
    expect(body.cookies).toEqual([
      expect.objectContaining({ name: "foo", source: "script" }),
    ]);

    detector.stop();
  });
});
