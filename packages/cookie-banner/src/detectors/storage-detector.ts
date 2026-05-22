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
import { isExtensionCaller, isExtensionContext } from "./extension-context";
import { getInitiatorURL } from "./initiator";
import type { ReportQueue } from "./report-queue";
import type { DetectedStorageEntry, StorageSource } from "./types";

const OWN_KEY_PREFIX = "probo_consent:";

export class StorageDetector implements Detector {
  private readonly queue: ReportQueue;
  private readonly apiOrigin: string;
  private originalSetItem: typeof Storage.prototype.setItem | null = null;
  private originalIDBOpen: typeof IDBFactory.prototype.open | null = null;
  private originalCachesOpen: typeof CacheStorage.prototype.open | null = null;

  constructor(queue: ReportQueue, apiOrigin: string) {
    this.queue = queue;
    this.apiOrigin = apiOrigin;
  }

  start(): void {
    this.queue.onNotFound(() => this.stop());

    this.wrapStorage();
    this.wrapIndexedDB();
    this.wrapCacheStorage();

    if (isExtensionContext()) return;

    this.scanExisting();
    this.scanCacheStorage();
  }

  stop(): void {
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
      if (isExtensionCaller()) return request;
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
      if (!isExtensionCaller()) {
        self.onCacheStorageOpen(name, "script");
      }
      return originalOpen(name);
    };
  }

  private onStorageWrite(
    key: string,
    value: string,
    storageType: "local_storage" | "session_storage",
  ): void {
    if (key.startsWith(OWN_KEY_PREFIX)) return;

    const initiatorUrl = getInitiatorURL(this.apiOrigin);

    const entry: DetectedStorageEntry = {
      key,
      storage_type: storageType,
      value_size: value.length * 2,
      source: "script",
    };
    if (initiatorUrl) entry.initiator_url = initiatorUrl;
    this.queue.reportStorage(entry);
  }

  private onIndexedDBOpen(name: string): void {
    this.queue.reportStorage({
      key: name,
      storage_type: "indexed_db",
      value_size: null,
      source: "script",
    });
  }

  private onCacheStorageOpen(name: string, source: StorageSource): void {
    this.queue.reportStorage({
      key: name,
      storage_type: "cache_storage",
      value_size: null,
      source,
    });
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
          this.onCacheStorageOpen(name, "pre-existing");
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

      const value = storage.getItem(key);
      this.queue.reportStorage({
        key,
        storage_type: storageType,
        value_size: value ? value.length * 2 : null,
        source: "pre-existing",
      });
    }
  }
}
