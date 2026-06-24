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

import type { Detector } from "./detector";
import { isExtensionContext } from "./extension-context";
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
      const { fromExtension } = getInitiatorURL(self.apiOrigin);
      self.onCacheStorageOpen(name, fromExtension ? "extension" : "script");
      return originalOpen(name);
    };
  }

  private onStorageWrite(
    key: string,
    value: string,
    storageType: "local_storage" | "session_storage",
  ): void {
    if (key.startsWith(OWN_KEY_PREFIX)) return;

    const { url: initiatorUrl, fromExtension } = getInitiatorURL(this.apiOrigin);

    const entry: DetectedStorageEntry = {
      key,
      storage_type: storageType,
      value_size: value.length * 2,
      source: fromExtension ? "extension" : "script",
    };
    if (initiatorUrl) entry.initiator_url = initiatorUrl;
    this.queue.reportStorage(entry);
  }

  private onIndexedDBOpen(name: string): void {
    const { fromExtension } = getInitiatorURL(this.apiOrigin);

    this.queue.reportStorage({
      key: name,
      storage_type: "indexed_db",
      value_size: null,
      source: fromExtension ? "extension" : "script",
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
