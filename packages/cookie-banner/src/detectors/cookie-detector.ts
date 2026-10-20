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

import {
  cookieListItemDomain,
  isDeletion,
  parseCookieName,
  parseCookieSetDomain,
  clampMaxAgeSeconds,
  parseMaxAgeSeconds,
} from "../cookie-utils";
import type { Detector } from "./detector";
import { isExtensionContext } from "./extension-context";
import { getInitiatorURL } from "./initiator";
import type { ReportQueue } from "./report-queue";
import type { DetectedCookieEntry } from "./types";

const COOKIE_HOOK_MARKER = Symbol("proboCookieHook");

interface PrototypeHook {
  proto: object;
  original: PropertyDescriptor;
}

export class CookieDetector implements Detector {
  private readonly queue: ReportQueue;
  private readonly apiOrigin: string;
  private readonly knownNames: Set<string>;
  private started = false;
  private prototypeHooks: PrototypeHook[] = [];
  private cookieStoreHandler: ((event: CookieChangeEvent) => void) | null = null;

  constructor(queue: ReportQueue, apiOrigin: string, knownNames: Set<string>) {
    this.queue = queue;
    this.apiOrigin = apiOrigin;
    this.knownNames = knownNames;
  }

  start(): void {
    this.queue.onNotFound(() => this.stop());

    if (isExtensionContext()) return;
    if (this.started) return;
    this.started = true;

    // Another SDK instance may already own the prototype hook. Still
    // observe cookieStore so this instance reports HTTP-source
    // changes; the client scans leftovers after published names land.
    this.installPrototypeHooks();
    this.observeCookieStore();
  }

  stop(): void {
    if (this.cookieStoreHandler && typeof cookieStore !== "undefined") {
      cookieStore.removeEventListener("change", this.cookieStoreHandler);
      this.cookieStoreHandler = null;
    }

    for (const hook of this.prototypeHooks) {
      Object.defineProperty(hook.proto, "cookie", hook.original);
    }
    this.prototypeHooks = [];
    this.started = false;
  }

  // Chrome ignores an instance override of the cookie IDL attribute
  // and keeps calling Document.prototype. A document.cookie write
  // then only shows up on cookieStore "change" and is classified as
  // HTTP. Hook the prototypes the browser actually uses.
  private installPrototypeHooks(): void {
    if (typeof Document === "undefined") return;

    const seen = new Set<object>();

    this.tryHookPrototype(Document.prototype, seen);
    if (typeof HTMLDocument !== "undefined") {
      this.tryHookPrototype(HTMLDocument.prototype, seen);
    }
  }

  private tryHookPrototype(proto: object, seen: Set<object>): void {
    if (seen.has(proto)) return;

    const desc = Object.getOwnPropertyDescriptor(proto, "cookie");
    if (!desc?.set || !desc?.get) return;
    if (COOKIE_HOOK_MARKER in desc.set) return;

    seen.add(proto);

    const self = this;
    const originalGet = desc.get;
    const originalSet = desc.set;

    const get = function (this: Document): string {
      return originalGet.call(this);
    };

    const set = function (this: Document, value: string): void {
      originalSet.call(this, value);
      self.onCookieSet(value);
    };
    Object.defineProperty(set, COOKIE_HOOK_MARKER, { value: true });

    Object.defineProperty(proto, "cookie", {
      configurable: true,
      enumerable: desc.enumerable,
      get,
      set,
    });

    this.prototypeHooks.push({ proto, original: desc });
  }

  private onCookieSet(raw: string): void {
    if (isDeletion(raw)) return;

    const name = parseCookieName(raw);
    if (!name || this.knownNames.has(name)) return;

    const maxAgeSeconds = parseMaxAgeSeconds(raw);
    const { url: initiatorUrl, fromExtension } = getInitiatorURL(this.apiOrigin);

    const domain = parseCookieSetDomain(raw, location.hostname);
    const entry: DetectedCookieEntry = {
      name,
      max_age_seconds: maxAgeSeconds,
      source: fromExtension ? "extension" : "script",
      host_only: domain.host_only,
    };
    if (domain.cookie_domain != null) entry.cookie_domain = domain.cookie_domain;
    if (initiatorUrl) entry.initiator_url = initiatorUrl;
    this.queue.reportCookie(entry);
  }

  // scanExisting reports leftovers as PRE_EXISTING. Call after
  // published snapshot names are in knownNames so inventoried
  // cookies are not re-queued on every page load.
  scanExisting(): void {
    if (!this.started) return;

    // Queue names from document.cookie immediately so a tab close
    // during getAll() still reports pre-existing cookies. getAll is
    // an enrichment pass that re-sends the same names with a domain.
    this.scanDocumentCookie();

    if (typeof cookieStore === "undefined" || typeof cookieStore.getAll !== "function") {
      return;
    }

    cookieStore
      .getAll()
      .then((cookies) => {
        for (const cookie of cookies) {
          this.reportExistingCookie(cookie.name, cookie.expires, cookieListItemDomain(cookie.domain));
        }
      })
      .catch(() => {
        // document.cookie already ran; a rejected getAll must not
        // drop those names or scan them twice.
      });
  }

  private scanDocumentCookie(): void {
    const cookieStr = document.cookie;
    if (!cookieStr) return;

    for (const pair of cookieStr.split(";")) {
      const name = pair.split("=")[0]?.trim();
      this.reportExistingCookie(name, null, null);
    }
  }

  private reportExistingCookie(
    name: string | undefined,
    expires: number | null,
    domain: ReturnType<typeof cookieListItemDomain> | null,
  ): void {
    if (!name || this.knownNames.has(name)) return;

    const maxAge = expires == null
      ? null
      : clampMaxAgeSeconds((expires - Date.now()) / 1000);

    const entry: DetectedCookieEntry = {
      name,
      max_age_seconds: maxAge,
      source: "pre-existing",
    };
    if (domain != null) {
      entry.host_only = domain.host_only;
      if (domain.cookie_domain != null) entry.cookie_domain = domain.cookie_domain;
    }
    this.queue.reportCookie(entry);
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
          ? clampMaxAgeSeconds((cookie.expires - Date.now()) / 1000)
          : null;

        const domain = cookieListItemDomain(cookie.domain);
        const entry: DetectedCookieEntry = {
          name: cookie.name,
          max_age_seconds: maxAge,
          source: "http",
          host_only: domain.host_only,
        };
        if (domain.cookie_domain != null) entry.cookie_domain = domain.cookie_domain;
        this.queue.reportCookie(entry);
      }
    };

    cookieStore.addEventListener("change", this.cookieStoreHandler);
  }
}
