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

interface DetectedStorageEntry {
  key: string;
  storage_type: "local_storage" | "session_storage" | "indexed_db";
  value_size: number | null;
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
  private readonly reported: Set<string> = new Set();
  private readonly pending: Map<string, DetectedStorageEntry> = new Map();
  private timer: ReturnType<typeof setTimeout> | null = null;
  private originalLocalSetItem: typeof Storage.prototype.setItem | null = null;
  private originalSessionSetItem: typeof Storage.prototype.setItem | null = null;
  private originalIDBOpen: typeof IDBFactory.prototype.open | null = null;

  constructor(baseUrl: URL, bannerId: string) {
    this.reportUrl = new URL(`${bannerId}/report`, baseUrl);
  }

  start(): void {
    this.wrapStorage();
    this.wrapIndexedDB();
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

    if (this.originalLocalSetItem) {
      Storage.prototype.setItem = this.originalLocalSetItem;
      this.originalLocalSetItem = null;
    }
    if (this.originalSessionSetItem) {
      Storage.prototype.setItem = this.originalLocalSetItem ?? this.originalSessionSetItem!;
      this.originalSessionSetItem = null;
    }
    if (this.originalIDBOpen) {
      IDBFactory.prototype.open = this.originalIDBOpen;
      this.originalIDBOpen = null;
    }
  }

  private wrapStorage(): void {
    const originalSetItem = Storage.prototype.setItem;
    this.originalLocalSetItem = originalSetItem;

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

  private onStorageWrite(
    key: string,
    value: string,
    storageType: "local_storage" | "session_storage",
  ): void {
    if (key.startsWith(OWN_KEY_PREFIX)) return;

    const reportKey = `${storageType}:${key}`;
    if (this.reported.has(reportKey)) return;

    this.reported.add(reportKey);
    this.pending.set(reportKey, {
      key,
      storage_type: storageType,
      value_size: value.length * 2,
    });
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
    if (this.timer) return;
    this.timer = setTimeout(() => {
      this.timer = null;
      this.flush();
    }, DEBOUNCE_MS);
  }

  private flush(): void {
    if (this.pending.size === 0) return;

    const entries: DetectedStorageEntry[] = [];
    for (const [key, entry] of this.pending) {
      entries.push(entry);
      this.pending.delete(key);
      if (entries.length >= MAX_ITEMS_PER_REQUEST) break;
    }

    void fetchJSON(this.reportUrl, {
      method: "POST",
      body: { storage: entries },
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
