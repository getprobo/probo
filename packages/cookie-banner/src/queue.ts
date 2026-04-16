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

import { fetchJSON } from "./http";

const STORAGE_KEY_PREFIX = "probo_consent";
const MAX_QUEUE_SIZE = 10;
const MAX_AGE_MS = 30 * 24 * 60 * 60 * 1000;

interface PendingConsent {
  url: string;
  body: unknown;
  timestamp: number;
}

function storageKey(bannerId: string): string {
  return `${STORAGE_KEY_PREFIX}:${bannerId}:queue`;
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

  const remaining: PendingConsent[] = [];

  for (const entry of queue) {
    try {
      await fetchJSON(entry.url, { method: "POST", body: entry.body });
    } catch {
      remaining.push(entry);
    }
  }

  writeQueue(bannerId, remaining);
}
