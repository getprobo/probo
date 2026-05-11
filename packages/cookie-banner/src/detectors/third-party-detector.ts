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

import type { Detector } from "./detector";
import { NotFoundError } from "../errors";
import { fetchJSON } from "../http";

type ResourceType =
  | "script"
  | "iframe"
  | "image"
  | "stylesheet"
  | "font"
  | "beacon"
  | "fetch"
  | "media"
  | "service_worker";

interface DetectedResourceEntry {
  url: string;
  resource_type: ResourceType;
}

const DEBOUNCE_MS = 2_000;
const MAX_ITEMS_PER_REQUEST = 100;
const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

// Map browser-reported PerformanceResourceTiming.initiatorType to the
// server-side tracker_resource_type. Anything we cannot classify is
// dropped rather than reported as "other" to keep the table tidy.
function mapInitiatorType(it: string): ResourceType | null {
  switch (it) {
    case "script":
      return "script";
    case "iframe":
      return "iframe";
    case "img":
    case "image":
    case "imageset":
    case "input":
      return "image";
    case "css":
    case "link":
      return "stylesheet";
    case "font":
      return "font";
    case "beacon":
    case "ping":
      return "beacon";
    case "fetch":
    case "xmlhttprequest":
      return "fetch";
    case "video":
    case "audio":
    case "track":
    case "embed":
    case "object":
      return "media";
    default:
      return null;
  }
}

export class ThirdPartyDetector implements Detector {
  private readonly reportUrl: URL;
  private readonly pageOrigin: string;
  private readonly proboOrigin: string;
  private readonly reported: Set<string> = new Set();
  private readonly pending: Map<string, DetectedResourceEntry> = new Map();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private observer: MutationObserver | null = null;
  private perfObserver: PerformanceObserver | null = null;
  private originalSWRegister: typeof ServiceWorkerContainer.prototype.register | null = null;

  constructor(baseUrl: URL, bannerId: string) {
    this.reportUrl = new URL(`${bannerId}/report`, baseUrl);
    this.pageOrigin = location.origin;
    this.proboOrigin = baseUrl.origin;
  }

  start(): void {
    this.scanExisting();
    this.observeMutations();
    this.observePerformance();
    this.wrapServiceWorker();
    this.scanServiceWorkers();
  }

  stop(): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }

    if (this.pending.size > 0) {
      this.flush();
    }

    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }

    if (this.perfObserver) {
      this.perfObserver.disconnect();
      this.perfObserver = null;
    }

    if (
      this.originalSWRegister
      && typeof navigator !== "undefined"
      && navigator.serviceWorker
    ) {
      navigator.serviceWorker.register = this.originalSWRegister;
      this.originalSWRegister = null;
    }
  }

  private scanExisting(): void {
    for (const script of document.querySelectorAll<HTMLScriptElement>("script[src]")) {
      this.processResource(script.src, "script");
    }
    for (const iframe of document.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
      this.processResource(iframe.src, "iframe");
    }

    if (this.pending.size > 0) {
      this.scheduleFlush();
    }
  }

  private observeMutations(): void {
    this.observer = new MutationObserver((mutations) => {
      for (const mutation of mutations) {
        for (const node of mutation.addedNodes) {
          if (!(node instanceof HTMLElement)) continue;

          if (node instanceof HTMLScriptElement && node.src) {
            this.processResource(node.src, "script");
          } else if (node instanceof HTMLIFrameElement && node.src) {
            this.processResource(node.src, "iframe");
          }

          for (const script of node.querySelectorAll<HTMLScriptElement>("script[src]")) {
            this.processResource(script.src, "script");
          }
          for (const iframe of node.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
            this.processResource(iframe.src, "iframe");
          }
        }
      }
    });

    this.observer.observe(document.documentElement, {
      childList: true,
      subtree: true,
    });
  }

  // observePerformance picks up resources the DOM scan misses: tracking
  // pixels (<img>), beacons, fetch/XHR call-homes, CSS-loaded fonts and
  // sub-stylesheets, video/audio embeds. `buffered: true` replays any
  // entries that fired before the observer was attached, so we catch
  // bootstrap resources too.
  private observePerformance(): void {
    if (typeof PerformanceObserver === "undefined") return;

    try {
      this.perfObserver = new PerformanceObserver((list) => {
        for (const entry of list.getEntries() as PerformanceResourceTiming[]) {
          const rt = mapInitiatorType(entry.initiatorType);
          if (rt) this.processResource(entry.name, rt);
        }
      });
      this.perfObserver.observe({ type: "resource", buffered: true });
    } catch {
      // Older browsers may not support the `type` option or the
      // `'resource'` entry type. Silently degrade to MutationObserver
      // coverage only.
      this.perfObserver = null;
    }
  }

  // wrapServiceWorker intercepts navigator.serviceWorker.register so
  // each registration -- even ones initiated by third-party SDKs --
  // surfaces as a tracker_resource entry keyed on the worker script
  // origin+path.
  private wrapServiceWorker(): void {
    if (typeof navigator === "undefined" || !navigator.serviceWorker) return;

    const sw = navigator.serviceWorker;
    const originalRegister = sw.register.bind(sw);
    this.originalSWRegister = originalRegister;

    const self = this;

    sw.register = function (
      scriptURL: string | URL,
      options?: RegistrationOptions,
    ): Promise<ServiceWorkerRegistration> {
      const url = typeof scriptURL === "string" ? scriptURL : scriptURL.toString();
      self.processResource(url, "service_worker");
      return originalRegister(scriptURL, options);
    };
  }

  // scanServiceWorkers enumerates registrations that pre-date the SDK
  // (e.g. installed on a previous visit, restored from cache).
  private scanServiceWorkers(): void {
    if (typeof navigator === "undefined" || !navigator.serviceWorker) return;

    navigator.serviceWorker
      .getRegistrations()
      .then((registrations) => {
        for (const r of registrations) {
          const url
            = r.active?.scriptURL
              ?? r.installing?.scriptURL
              ?? r.waiting?.scriptURL;
          if (url) this.processResource(url, "service_worker");
        }
      })
      .catch(() => {
        // Some browsers throw NotSupportedError in insecure contexts.
      });
  }

  private processResource(src: string, resourceType: ResourceType): void {
    if (EXTENSION_URL_RE.test(src)) return;

    let parsed: URL;
    try {
      parsed = new URL(src);
    } catch {
      return;
    }

    if (parsed.protocol !== "http:" && parsed.protocol !== "https:") return;
    if (parsed.origin === this.pageOrigin || parsed.origin === this.proboOrigin) return;

    const identifier = parsed.origin + parsed.pathname;
    const reportKey = `${resourceType}:${identifier}`;
    if (this.reported.has(reportKey)) return;

    this.reported.add(reportKey);
    this.pending.set(reportKey, { url: identifier, resource_type: resourceType });
    this.scheduleFlush();
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

    const entries: DetectedResourceEntry[] = [];
    for (const [key, entry] of this.pending) {
      entries.push(entry);
      this.pending.delete(key);
      if (entries.length >= MAX_ITEMS_PER_REQUEST) break;
    }

    void fetchJSON(this.reportUrl, {
      method: "POST",
      body: { resources: entries },
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
