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

import { afterEach, describe, expect, it, vi } from "vitest";

import { StorageDetector } from "./storage-detector";
import type { ReportQueue } from "./report-queue";
import type { DetectedStorageEntry } from "./types";

class MemoryStorage {
  private readonly data = new Map<string, string>();

  get length(): number {
    return this.data.size;
  }

  key(index: number): string | null {
    return [...this.data.keys()][index] ?? null;
  }

  getItem(key: string): string | null {
    return this.data.get(key) ?? null;
  }

  setItem(key: string, value: unknown): void {
    if (typeof value === "symbol") {
      throw new TypeError("Cannot convert a Symbol value to a string");
    }

    this.data.set(key, String(value));
  }

  removeItem(key: string): void {
    this.data.delete(key);
  }

  clear(): void {
    this.data.clear();
  }
}

function installFakeStorage(): { local: MemoryStorage; session: MemoryStorage } {
  const local = new MemoryStorage();
  const session = new MemoryStorage();

  function Storage(): void {}
  Storage.prototype.setItem = function (this: MemoryStorage, key: string, value: string) {
    MemoryStorage.prototype.setItem.call(this, key, value);
  };
  Storage.prototype.getItem = function (this: MemoryStorage, key: string) {
    return MemoryStorage.prototype.getItem.call(this, key);
  };
  Storage.prototype.key = function (this: MemoryStorage, index: number) {
    return MemoryStorage.prototype.key.call(this, index);
  };

  vi.stubGlobal("Storage", Storage);
  vi.stubGlobal("localStorage", local);
  vi.stubGlobal("sessionStorage", session);
  vi.stubGlobal("location", { hostname: "example.com", origin: "https://example.com" });
  vi.stubGlobal("window", {
    localStorage: local,
    sessionStorage: session,
  });

  Object.setPrototypeOf(local, Storage.prototype);
  Object.setPrototypeOf(session, Storage.prototype);

  return { local, session };
}

function fakeQueue(): { queue: ReportQueue; reported: DetectedStorageEntry[] } {
  const reported: DetectedStorageEntry[] = [];
  const queue = {
    reportStorage(entry: DetectedStorageEntry) {
      reported.push(entry);
    },
    onNotFound() {},
  } as unknown as ReportQueue;

  return { queue, reported };
}

describe("StorageDetector", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("reports setItem writes as script", () => {
    installFakeStorage();
    const { queue, reported } = fakeQueue();
    const detector = new StorageDetector(queue, "https://api.example.com");

    detector.start();
    localStorage.setItem("loglevel:probo:example:themed-tcf", "DEBUG");

    expect(reported.some((entry) => (
      entry.key === "loglevel:probo:example:themed-tcf" &&
      entry.source === "script" &&
      entry.storage_type === "local_storage"
    ))).toBe(true);

    detector.stop();
  });

  it("reports named property assignment as script", () => {
    installFakeStorage();
    const { queue, reported } = fakeQueue();
    const detector = new StorageDetector(queue, "https://api.example.com");

    detector.start();
    (window.localStorage as unknown as Record<string, string>).loglevel = "DEBUG";

    expect(reported.some((entry) => (
      entry.key === "loglevel" &&
      entry.source === "script" &&
      entry.storage_type === "local_storage"
    ))).toBe(true);

    detector.stop();
  });

  it("rejects Symbol assignment like native Storage", () => {
    installFakeStorage();
    const { queue, reported } = fakeQueue();
    const detector = new StorageDetector(queue, "https://api.example.com");

    detector.start();
    expect(() => {
      (window.localStorage as unknown as Record<string, unknown>).foo = Symbol("x");
    }).toThrow(TypeError);
    expect(reported).toEqual([]);

    detector.stop();
  });
});
