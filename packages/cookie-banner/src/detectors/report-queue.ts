// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { NotFoundError } from "../errors";
import { fetchJSON } from "../http";
import type {
  DetectedCookieEntry,
  DetectedResourceEntry,
  DetectedStorageEntry,
} from "./types";

const DEBOUNCE_MS = 2_000;

// Must match `maxDetectedTrackersPerRequest` in
// pkg/server/api/cookiebanner/v1/handler.go. Bumping requires a
// coordinated server-side change; sending more in one request will be
// rejected with 400.
const MAX_ITEMS_PER_REQUEST = 100;

type QueuedItem =
  | { kind: "cookie"; entry: DetectedCookieEntry }
  | { kind: "storage"; entry: DetectedStorageEntry }
  | { kind: "resource"; entry: DetectedResourceEntry };

interface Batch {
  cookies?: DetectedCookieEntry[];
  storage?: DetectedStorageEntry[];
  resources?: DetectedResourceEntry[];
}

// ReportQueue centralises debounce, batching, retry, dedup and the
// page-lifecycle drain for the three tracker detectors. Every reported
// item lives in a single `pending` Map keyed with a type-namespaced
// dedup key (`c:`, `s:`, `r:`) so that, e.g., a cookie literally named
// `s:local_storage:foo` cannot collide with a localStorage entry whose
// key is `foo`. The queue owns the only `reported` Set; detectors are
// pure producers that call `reportCookie/Storage/Resource` and don't
// track what they've already sent.
export class ReportQueue {
  private readonly reportUrl: URL;
  private readonly pending: Map<string, QueuedItem> = new Map();
  private readonly reported: Set<string> = new Set();
  private readonly notFoundListeners: Set<() => void> = new Set();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private flushing = false;
  private stopped = false;
  private pageHideHandler: (() => void) | null = null;
  private visibilityHandler: (() => void) | null = null;

  constructor(reportUrl: URL) {
    this.reportUrl = reportUrl;
    this.attachLifecycleListeners();
  }

  reportCookie(entry: DetectedCookieEntry): void {
    this.enqueue(`c:${entry.name}`, { kind: "cookie", entry });
  }

  reportStorage(entry: DetectedStorageEntry): void {
    this.enqueue(`s:${entry.storage_type}:${entry.key}`, { kind: "storage", entry });
  }

  reportResource(entry: DetectedResourceEntry): void {
    this.enqueue(`r:${entry.resource_type}:${entry.url}`, { kind: "resource", entry });
  }

  onNotFound(cb: () => void): void {
    this.notFoundListeners.add(cb);
  }

  stop(): void {
    if (this.stopped) return;
    this.stopped = true;

    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }

    this.detachLifecycleListeners();

    if (this.pending.size > 0) {
      this.flushSync();
    }
  }

  private enqueue(key: string, item: QueuedItem): void {
    if (this.stopped) return;
    if (this.reported.has(key)) return;

    this.reported.add(key);
    this.pending.set(key, item);
    this.scheduleFlush();
  }

  private scheduleFlush(): void {
    if (this.timer || this.flushing || this.stopped) return;

    this.timer = setTimeout(() => {
      this.timer = null;
      this.flush();
    }, DEBOUNCE_MS);
  }

  // flush sends one batch from `pending` and only removes entries on
  // success. Transient failures leave entries in `pending` so they are
  // retried on the next flush. `flushing` guards against re-sending an
  // in-flight batch when new entries arrive mid-request.
  private flush(): void {
    if (this.flushing || this.stopped) return;
    if (this.pending.size === 0) return;

    const { keys, body } = this.takeBatch();

    this.flushing = true;
    void fetchJSON(this.reportUrl, {
      method: "POST",
      body,
    })
      .then(() => {
        for (const key of keys) this.pending.delete(key);
      })
      .catch((err) => {
        if (err instanceof NotFoundError) {
          this.pending.clear();
          this.notifyNotFound();
        }
      })
      .finally(() => {
        this.flushing = false;
        if (!this.stopped && this.pending.size > 0) {
          this.scheduleFlush();
        }
      });
  }

  // takeBatch pulls up to MAX_ITEMS_PER_REQUEST items in insertion
  // order (Map preserves it) and partitions them into the three arrays
  // the server expects. Insertion-order is naturally fair: items are
  // sent in the order the detectors observed them.
  private takeBatch(): { keys: string[]; body: Batch } {
    const keys: string[] = [];
    const cookies: DetectedCookieEntry[] = [];
    const storage: DetectedStorageEntry[] = [];
    const resources: DetectedResourceEntry[] = [];

    for (const [key, item] of this.pending) {
      keys.push(key);
      switch (item.kind) {
        case "cookie":
          cookies.push(item.entry);
          break;
        case "storage":
          storage.push(item.entry);
          break;
        case "resource":
          resources.push(item.entry);
          break;
      }
      if (keys.length >= MAX_ITEMS_PER_REQUEST) break;
    }

    const body: Batch = {};
    if (cookies.length > 0) body.cookies = cookies;
    if (storage.length > 0) body.storage = storage;
    if (resources.length > 0) body.resources = resources;

    return { keys, body };
  }

  private notifyNotFound(): void {
    for (const cb of this.notFoundListeners) {
      try {
        cb();
      } catch {
        // Listeners must not throw across the queue; swallow so other
        // detectors still get notified.
      }
    }
  }

  // attachLifecycleListeners wires the queue to `pagehide` and
  // `visibilitychange:hidden` so a tab close, navigation, or
  // mobile-Safari background drains the queue synchronously instead of
  // losing the last 0--DEBOUNCE_MS window of detections. Both events
  // are listened to because:
  //   - `pagehide` is the canonical "page is going away" signal but
  //     does not always fire when an iOS tab is backgrounded;
  //   - `visibilitychange` to `hidden` covers the iOS background case
  //     and most mobile browsers.
  // `flushSync` is idempotent (no-op when pending is empty), so firing
  // twice is harmless.
  //
  // We deliberately avoid `unload`: it disables bfcache and is
  // unreliable on mobile Safari.
  private attachLifecycleListeners(): void {
    if (typeof window === "undefined") return;

    this.pageHideHandler = () => this.flushSync();
    window.addEventListener("pagehide", this.pageHideHandler, { capture: true });

    if (typeof document !== "undefined") {
      this.visibilityHandler = () => {
        if (document.visibilityState === "hidden") this.flushSync();
      };
      document.addEventListener("visibilitychange", this.visibilityHandler);
    }
  }

  private detachLifecycleListeners(): void {
    if (this.pageHideHandler && typeof window !== "undefined") {
      window.removeEventListener("pagehide", this.pageHideHandler, { capture: true });
      this.pageHideHandler = null;
    }
    if (this.visibilityHandler && typeof document !== "undefined") {
      document.removeEventListener("visibilitychange", this.visibilityHandler);
      this.visibilityHandler = null;
    }
  }

  // flushSync drains up to MAX_ITEMS_PER_REQUEST items via a transport
  // that survives the document unloading. `navigator.sendBeacon` is
  // tried first; if it is unavailable or refuses the payload (queue
  // full / size cap exceeded), we fall back to `fetch` with
  // `keepalive: true`. Both transports cap the total in-flight body
  // size at ~64 KB across all requests, so we send at most one batch
  // and accept losing the tail on pages with > 100 pending items at
  // unload time -- still strictly better than the previous behaviour,
  // which lost the entire last debounce window.
  //
  // sendBeacon is always attempted even when an async `flush()` is
  // in-flight (`this.flushing === true`). The server handles
  // duplicates idempotently. The keepalive-fetch fallback is only
  // used when the async flush mutex is free to avoid concurrent
  // fetch lifecycles conflicting over the `flushing` flag.
  //
  // Items are only removed from `pending` once the transport
  // confirms delivery: sendBeacon synchronously returns true when
  // the browser has accepted ownership of the request, and the
  // keepalive fetch removes items in its `.then` on an `ok`
  // response. On async fetch failure we leave the entries queued so
  // they can be retried by the next flush -- this matters when
  // `visibilitychange:hidden` fires but the page is restored from
  // bfcache rather than truly unloading.
  private flushSync(): void {
    if (this.pending.size === 0) return;

    const { keys, body } = this.takeBatch();
    const payload = JSON.stringify(body);

    if (typeof navigator !== "undefined" && typeof navigator.sendBeacon === "function") {
      const blob = new Blob([payload], { type: "application/json" });
      let queued = false;
      try {
        queued = navigator.sendBeacon(this.reportUrl.toString(), blob);
      } catch {
        queued = false;
      }
      if (queued) {
        for (const key of keys) this.pending.delete(key);
        return;
      }
    }

    // Fall back to keepalive fetch only when no async flush owns the
    // mutex — sendBeacon above is the primary unload-safe path.
    if (this.flushing) return;

    if (typeof fetch === "function") {
      this.flushing = true;
      try {
        void fetch(this.reportUrl.toString(), {
          method: "POST",
          mode: "cors",
          credentials: "omit",
          keepalive: true,
          headers: {
            "Content-Type": "application/json",
            Accept: "application/json",
            "X-SDK-Version": __SDK_VERSION__,
          },
          body: payload,
        })
          .then((res) => {
            if (res.ok) {
              for (const key of keys) this.pending.delete(key);
            }
          })
          .catch(() => {
            // Leave entries in `pending`; if the page is restored
            // from bfcache the next flush retries them, and on a
            // true unload the queue itself is discarded with the
            // page.
          })
          .finally(() => {
            this.flushing = false;
            if (!this.stopped && this.pending.size > 0) {
              this.scheduleFlush();
            }
          });
      } catch {
        this.flushing = false;
      }
    }
  }
}
