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

interface DetectedResourceEntry {
  origin: string;
  resource_type: "script" | "iframe";
}

const DEBOUNCE_MS = 2_000;
const MAX_ITEMS_PER_REQUEST = 100;
const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

export class ThirdPartyDetector implements Detector {
  private readonly reportUrl: URL;
  private readonly pageOrigin: string;
  private readonly proboOrigin: string;
  private readonly reported: Set<string> = new Set();
  private readonly pending: Map<string, DetectedResourceEntry> = new Map();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private observer: MutationObserver | null = null;

  constructor(baseUrl: URL, bannerId: string) {
    this.reportUrl = new URL(`${bannerId}/report`, baseUrl);
    this.pageOrigin = location.origin;
    this.proboOrigin = baseUrl.origin;
  }

  start(): void {
    this.scanExisting();
    this.observeMutations();
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
  }

  private scanExisting(): void {
    for (const script of document.querySelectorAll<HTMLScriptElement>("script[src]")) {
      this.processElement(script.src, "script");
    }
    for (const iframe of document.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
      this.processElement(iframe.src, "iframe");
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
            this.processElement(node.src, "script");
          } else if (node instanceof HTMLIFrameElement && node.src) {
            this.processElement(node.src, "iframe");
          }

          for (const script of node.querySelectorAll<HTMLScriptElement>("script[src]")) {
            this.processElement(script.src, "script");
          }
          for (const iframe of node.querySelectorAll<HTMLIFrameElement>("iframe[src]")) {
            this.processElement(iframe.src, "iframe");
          }
        }
      }
    });

    this.observer.observe(document.documentElement, {
      childList: true,
      subtree: true,
    });
  }

  private processElement(src: string, resourceType: "script" | "iframe"): void {
    if (EXTENSION_URL_RE.test(src)) return;

    let origin: string;
    try {
      origin = new URL(src).origin;
    } catch {
      return;
    }

    if (origin === this.pageOrigin || origin === this.proboOrigin) return;

    const reportKey = `${resourceType}:${origin}`;
    if (this.reported.has(reportKey)) return;

    this.reported.add(reportKey);
    this.pending.set(reportKey, { origin, resource_type: resourceType });
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
