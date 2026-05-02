// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

import { isDeletion, parseCookieName, parseMaxAgeSeconds } from "./cookie-utils";
import { NotFoundError } from "./errors";
import { fetchJSON } from "./http";

interface DetectedCookieEntry {
  name: string;
  max_age_seconds: number | null;
  source: "script" | "pre-existing";
}

const DEBOUNCE_MS = 2_000;
const MAX_COOKIES_PER_REQUEST = 100;
const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

function isExtensionCaller(): boolean {
  const stack = new Error().stack ?? "";
  return EXTENSION_URL_RE.test(stack);
}

export class CookieDetector {
  private readonly reportUrl: URL;
  private readonly knownNames: Set<string>;
  private readonly reported: Set<string> = new Set();
  private readonly pending: Map<string, DetectedCookieEntry> = new Map();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private originalDescriptor: PropertyDescriptor | null = null;

  constructor(baseUrl: URL, bannerId: string, knownNames: Set<string>) {
    this.reportUrl = new URL(`${bannerId}/detected-cookies`, baseUrl);
    this.knownNames = knownNames;
  }

  start(): void {
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
  }

  stop(): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }

    if (this.pending.size > 0) {
      this.flush();
    }

    if (this.originalDescriptor) {
      Object.defineProperty(document, "cookie", this.originalDescriptor);
      this.originalDescriptor = null;
    }
  }

  private onCookieSet(raw: string): void {
    if (isDeletion(raw)) return;
    if (isExtensionCaller()) return;

    const name = parseCookieName(raw);
    if (!name || this.knownNames.has(name) || this.reported.has(name)) return;

    const maxAgeSeconds = parseMaxAgeSeconds(raw);

    this.reported.add(name);
    this.pending.set(name, { name, max_age_seconds: maxAgeSeconds, source: "script" });
    this.scheduleFlush();
  }

  private scanExisting(): void {
    const cookieStr = document.cookie;
    if (!cookieStr) return;

    for (const pair of cookieStr.split(";")) {
      const name = pair.split("=")[0]?.trim();
      if (!name || this.knownNames.has(name) || this.reported.has(name)) {
        continue;
      }
      this.reported.add(name);
      this.pending.set(name, { name, max_age_seconds: null, source: "pre-existing" });
    }

    if (this.pending.size > 0) {
      this.scheduleFlush();
    }
  }

  private scheduleFlush(): void {
    if (this.timer) return;
    this.timer = setTimeout(() => {
      this.timer = null;
      this.flush();
    }, DEBOUNCE_MS);
  }

  private flush(): void {
    if (this.pending.size === 0) return;

    const iter = this.pending.entries();
    const entries: DetectedCookieEntry[] = [];
    for (const [key, entry] of iter) {
      entries.push(entry);
      this.pending.delete(key);
      if (entries.length >= MAX_COOKIES_PER_REQUEST) break;
    }

    void fetchJSON(this.reportUrl, {
      method: "POST",
      body: { cookies: entries },
    }).catch((err) => {
      if (err instanceof NotFoundError) {
        this.pending.clear();
        this.stop();
      }
    });

    if (this.pending.size > 0) {
      this.scheduleFlush();
    }
  }
}
