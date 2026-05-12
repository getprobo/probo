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
import type { ReportQueue } from "./report-queue";
import type { ResourceType } from "./types";

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

export class ResourceDetector implements Detector {
  private readonly queue: ReportQueue;
  private readonly pageOrigin: string;
  private readonly apiOrigin: string;
  private observer: MutationObserver | null = null;
  private perfObserver: PerformanceObserver | null = null;
  private originalSWRegister: typeof ServiceWorkerContainer.prototype.register | null = null;

  constructor(queue: ReportQueue, apiOrigin: string) {
    this.queue = queue;
    this.pageOrigin = location.origin;
    this.apiOrigin = apiOrigin;
  }

  start(): void {
    this.queue.onNotFound(() => this.stop());

    this.scanExisting();
    this.observeMutations();
    this.observePerformance();
    this.wrapServiceWorker();
    this.scanServiceWorkers();
  }

  stop(): void {
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
    // Service workers are always same-origin per browser security rules,
    // so we never drop them based on the page origin -- they are tracked
    // regardless of where the script lives. The apiOrigin guard still
    // applies so we never report our own SDK assets.
    if (parsed.origin === this.apiOrigin) return;
    if (resourceType !== "service_worker" && parsed.origin === this.pageOrigin) return;

    const identifier = parsed.origin + parsed.pathname;
    this.queue.reportResource({ url: identifier, resource_type: resourceType });
  }
}
