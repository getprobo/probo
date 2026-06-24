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

import { isDeletion, parseCookieName, parseMaxAgeSeconds } from "../cookie-utils";
import type { Detector } from "./detector";
import { isExtensionContext } from "./extension-context";
import { getInitiatorURL } from "./initiator";
import type { ReportQueue } from "./report-queue";
import type { DetectedCookieEntry } from "./types";

export class CookieDetector implements Detector {
  private readonly queue: ReportQueue;
  private readonly apiOrigin: string;
  private readonly knownNames: Set<string>;
  private originalDescriptor: PropertyDescriptor | null = null;
  private cookieStoreHandler: ((event: CookieChangeEvent) => void) | null = null;

  constructor(queue: ReportQueue, apiOrigin: string, knownNames: Set<string>) {
    this.queue = queue;
    this.apiOrigin = apiOrigin;
    this.knownNames = knownNames;
  }

  start(): void {
    this.queue.onNotFound(() => this.stop());

    if (isExtensionContext()) return;

    const desc =
      Object.getOwnPropertyDescriptor(Document.prototype, "cookie") ??
      Object.getOwnPropertyDescriptor(HTMLDocument.prototype, "cookie");

    if (!desc?.set || !desc?.get) return;

    this.originalDescriptor = desc;

    const self = this;
    const originalGet = desc.get;
    const originalSet = desc.set;

    Object.defineProperty(document, "cookie", {
      configurable: true,
      get() {
        return originalGet.call(this);
      },
      set(value: string) {
        originalSet.call(this, value);
        self.onCookieSet(value);
      },
    });

    this.scanExisting();
    this.observeCookieStore();
  }

  stop(): void {
    if (this.cookieStoreHandler && typeof cookieStore !== "undefined") {
      cookieStore.removeEventListener("change", this.cookieStoreHandler);
      this.cookieStoreHandler = null;
    }

    if (this.originalDescriptor) {
      Object.defineProperty(document, "cookie", this.originalDescriptor);
      this.originalDescriptor = null;
    }
  }

  private onCookieSet(raw: string): void {
    if (isDeletion(raw)) return;

    const name = parseCookieName(raw);
    if (!name || this.knownNames.has(name)) return;

    const maxAgeSeconds = parseMaxAgeSeconds(raw);
    const { url: initiatorUrl, fromExtension } = getInitiatorURL(this.apiOrigin);

    const entry: DetectedCookieEntry = {
      name,
      max_age_seconds: maxAgeSeconds,
      source: fromExtension ? "extension" : "script",
    };
    if (initiatorUrl) entry.initiator_url = initiatorUrl;
    this.queue.reportCookie(entry);
  }

  private scanExisting(): void {
    const cookieStr = document.cookie;
    if (!cookieStr) return;

    for (const pair of cookieStr.split(";")) {
      const name = pair.split("=")[0]?.trim();
      if (!name || this.knownNames.has(name)) continue;

      this.queue.reportCookie({ name, max_age_seconds: null, source: "pre-existing" });
    }
  }

  private observeCookieStore(): void {
    if (typeof cookieStore === "undefined" || typeof cookieStore.addEventListener !== "function") {
      return;
    }

    this.cookieStoreHandler = (event: CookieChangeEvent) => {
      for (const cookie of event.changed) {
        if (cookie.name === undefined) continue;
        if (this.knownNames.has(cookie.name)) continue;

        const maxAge = cookie.expires
          ? Math.round((cookie.expires - Date.now()) / 1000)
          : null;

        this.queue.reportCookie({
          name: cookie.name,
          max_age_seconds: maxAge && maxAge > 0 ? maxAge : null,
          source: "http",
        });
      }
    };

    cookieStore.addEventListener("change", this.cookieStoreHandler);
  }
}
