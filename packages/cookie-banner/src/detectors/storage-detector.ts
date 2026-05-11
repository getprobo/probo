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
import { getInitiatorURL } from "./initiator";

interface DetectedStorageEntry {
  key: string;
  storage_type:
    | "local_storage"
    | "session_storage"
    | "indexed_db"
    | "cache_storage";
  value_size: number | null;
  initiator_url?: string;
}

const DEBOUNCE_MS = 2_000;
const MAX_ITEMS_PER_REQUEST = 100;
const OWN_KEY_PREFIX = "probo_consent:";
const EXTENSION_URL_RE = /(?:chrome|moz|safari-web)-extension:\/\//;

function isExtensionCaller(): boolean {
  const stack = new Error().stack ?? "";
  return EXTENSION_URL_RE.test(stack);
}

export class StorageDetector implements Detector {
  private readonly reportUrl: URL;
  private readonly proboOrigin: string;
  private readonly reported: Set<string> = new Set();
  private readonly pending: Map<string, DetectedStorageEntry> = new Map();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private flushing = false;
  private originalSetItem: typeof Storage.prototype.setItem | null = null;
  private originalIDBOpen: typeof IDBFactory.prototype.open | null = null;
  private originalCachesOpen: typeof CacheStorage.prototype.open | null = null;

  constructor(baseUrl: URL, bannerId: string) {
    this.reportUrl = new URL(`${bannerId}/report`, baseUrl);
    this.proboOrigin = baseUrl.origin;
  }

  start(): void {
    this.wrapStorage();
    this.wrapIndexedDB();
    this.wrapCacheStorage();
    this.scanExisting();
    this.scanCacheStorage();
  }

  stop(): void {
    if (this.timer) {
      clearTimeout(this.timer);
      this.timer = null;
    }

    if (this.pending.size > 0) {
      this.flush();
    }

    if (this.originalSetItem) {
      Storage.prototype.setItem = this.originalSetItem;
      this.originalSetItem = null;
    }
    if (this.originalIDBOpen) {
      IDBFactory.prototype.open = this.originalIDBOpen;
      this.originalIDBOpen = null;
    }
    if (this.originalCachesOpen && typeof caches !== "undefined") {
      caches.open = this.originalCachesOpen;
      this.originalCachesOpen = null;
    }
  }

  private wrapStorage(): void {
    const originalSetItem = Storage.prototype.setItem;
    this.originalSetItem = originalSetItem;

    const self = this;

    Storage.prototype.setItem = function (key: string, value: string) {
      originalSetItem.call(this, key, value);

      if (isExtensionCaller()) return;

      const storageType: "local_storage" | "session_storage" =
        this === localStorage ? "local_storage" : "session_storage";

      self.onStorageWrite(key, value, storageType);
    };
  }

  private wrapIndexedDB(): void {
    if (typeof indexedDB === "undefined") return;

    const originalOpen = IDBFactory.prototype.open;
    this.originalIDBOpen = originalOpen;

    const self = this;

    IDBFactory.prototype.open = function (name: string, version?: number) {
      const request = originalOpen.call(this, name, version);
      self.onIndexedDBOpen(name);
      return request;
    };
  }

  private wrapCacheStorage(): void {
    if (typeof caches === "undefined") return;

    const originalOpen = caches.open.bind(caches);
    this.originalCachesOpen = originalOpen;

    const self = this;

    caches.open = function (name: string): Promise<Cache> {
      self.onCacheStorageOpen(name);
      return originalOpen(name);
    };
  }

  private onStorageWrite(
    key: string,
    value: string,
    storageType: "local_storage" | "session_storage",
  ): void {
    if (key.startsWith(OWN_KEY_PREFIX)) return;

    const reportKey = `${storageType}:${key}`;
    if (this.reported.has(reportKey)) return;

    const initiatorUrl = getInitiatorURL(this.proboOrigin);

    this.reported.add(reportKey);
    const entry: DetectedStorageEntry = {
      key,
      storage_type: storageType,
      value_size: value.length * 2,
    };
    if (initiatorUrl) entry.initiator_url = initiatorUrl;
    this.pending.set(reportKey, entry);
    this.scheduleFlush();
  }

  private onIndexedDBOpen(name: string): void {
    const reportKey = `indexed_db:${name}`;
    if (this.reported.has(reportKey)) return;

    this.reported.add(reportKey);
    this.pending.set(reportKey, {
      key: name,
      storage_type: "indexed_db",
      value_size: null,
    });
    this.scheduleFlush();
  }

  private onCacheStorageOpen(name: string): void {
    const reportKey = `cache_storage:${name}`;
    if (this.reported.has(reportKey)) return;

    this.reported.add(reportKey);
    this.pending.set(reportKey, {
      key: name,
      storage_type: "cache_storage",
      value_size: null,
    });
    this.scheduleFlush();
  }

  // scanCacheStorage enumerates pre-existing cache buckets created
  // before the SDK loaded. Service workers commonly create their
  // caches eagerly on `install`, so without this scan we would miss
  // any cache bucket whose creation predates the banner script.
  private scanCacheStorage(): void {
    if (typeof caches === "undefined") return;
    caches
      .keys()
      .then((names) => {
        for (const name of names) {
          this.onCacheStorageOpen(name);
        }
      })
      .catch(() => {
        // Insecure context or storage partition errors -- ignore.
      });
  }

  private scanExisting(): void {
    this.scanStorage(localStorage, "local_storage");
    this.scanStorage(sessionStorage, "session_storage");
  }

  private scanStorage(
    storage: Storage,
    storageType: "local_storage" | "session_storage",
  ): void {
    for (let i = 0; i < storage.length; i++) {
      const key = storage.key(i);
      if (!key || key.startsWith(OWN_KEY_PREFIX)) continue;

      const reportKey = `${storageType}:${key}`;
      if (this.reported.has(reportKey)) continue;

      const value = storage.getItem(key);
      this.reported.add(reportKey);
      this.pending.set(reportKey, {
        key,
        storage_type: storageType,
        value_size: value ? value.length * 2 : null,
      });
    }

    if (this.pending.size > 0) {
      this.scheduleFlush();
    }
  }

  private scheduleFlush(): void {
    if (this.timer || this.flushing) return;
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
    if (this.flushing) return;
    if (this.pending.size === 0) return;

    const batchKeys: string[] = [];
    const entries: DetectedStorageEntry[] = [];
    for (const [key, entry] of this.pending) {
      batchKeys.push(key);
      entries.push(entry);
      if (entries.length >= MAX_ITEMS_PER_REQUEST) break;
    }

    this.flushing = true;
    void fetchJSON(this.reportUrl, {
      method: "POST",
      body: { storage: entries },
    })
      .then(() => {
        for (const key of batchKeys) this.pending.delete(key);
      })
      .catch((err) => {
        if (err instanceof NotFoundError) {
          this.pending.clear();
          this.stop();
        }
      })
      .finally(() => {
        this.flushing = false;
        if (this.pending.size > 0) this.scheduleFlush();
      });
  }
}
