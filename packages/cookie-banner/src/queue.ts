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

import { COOKIE_NAME } from "./cookie";
import { fetchJSON } from "./http";
const MAX_QUEUE_SIZE = 10;
const MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

interface PendingConsent {
  url: string;
  body: unknown;
  timestamp: number;
}

function storageKey(bannerId: string): string {
  return `${COOKIE_NAME}:${bannerId}:queue`;
}

function readQueue(bannerId: string): PendingConsent[] {
  try {
    const raw = localStorage.getItem(storageKey(bannerId));
    if (!raw) {
      return [];
    }
    return JSON.parse(raw) as PendingConsent[];
  } catch {
    return [];
  }
}

function writeQueue(bannerId: string, queue: PendingConsent[]): void {
  try {
    if (queue.length === 0) {
      localStorage.removeItem(storageKey(bannerId));
    } else {
      localStorage.setItem(storageKey(bannerId), JSON.stringify(queue));
    }
  } catch {
    // localStorage unavailable
  }
}

export function enqueue(
  bannerId: string,
  url: string,
  body: unknown,
): void {
  const queue = readQueue(bannerId);
  queue.push({ url, body, timestamp: Date.now() });

  if (queue.length > MAX_QUEUE_SIZE) {
    queue.splice(0, queue.length - MAX_QUEUE_SIZE);
  }

  writeQueue(bannerId, queue);
}

export async function flush(bannerId: string): Promise<void> {
  const now = Date.now();
  let queue = readQueue(bannerId);

  if (queue.length === 0) {
    return;
  }

  queue = queue.filter((entry) => now - entry.timestamp < MAX_AGE_MS);

  if (queue.length === 0) {
    writeQueue(bannerId, []);
    return;
  }

  const sentTimestamps: number[] = [];

  for (const entry of queue) {
    try {
      await fetchJSON(entry.url, { method: "POST", body: entry.body });
      sentTimestamps.push(entry.timestamp);
    } catch {
      // will remain in queue
    }
  }

  const sentSet = new Set(sentTimestamps);
  const cutoff = now - MAX_AGE_MS;
  const current = readQueue(bannerId);
  writeQueue(
    bannerId,
    current.filter(
      (entry) => !sentSet.has(entry.timestamp) && entry.timestamp > cutoff,
    ),
  );
}
