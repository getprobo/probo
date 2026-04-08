// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { StoredConsent } from "./types";

const COOKIE_NAME = "probo_consent";
const LS_CONSENT_KEY = "probo_consent";
const LS_QUEUE_KEY = "probo_consent_queue";

// --- Cookie helpers ---

function getCookieConsent(): StoredConsent | null {
  try {
    const match = document.cookie
      .split("; ")
      .find((c) => c.startsWith(`${COOKIE_NAME}=`));
    if (!match) return null;
    return JSON.parse(
      decodeURIComponent(match.split("=").slice(1).join("=")),
    );
  } catch {
    return null;
  }
}

function setCookieConsent(
  consent: StoredConsent,
  maxAgeDays: number,
): boolean {
  try {
    const value = encodeURIComponent(JSON.stringify(consent));
    const maxAge = maxAgeDays * 24 * 60 * 60;
    document.cookie = `${COOKIE_NAME}=${value}; path=/; max-age=${maxAge}; SameSite=Lax`;
    // Verify the cookie was actually set.
    return document.cookie.includes(`${COOKIE_NAME}=`);
  } catch {
    return false;
  }
}

function clearCookieConsent(): void {
  try {
    document.cookie = `${COOKIE_NAME}=; max-age=-1; path=/; SameSite=Lax`;
  } catch {
    // Ignore.
  }
}

// --- localStorage helpers ---

function getLocalStorageConsent(): StoredConsent | null {
  try {
    const raw = localStorage.getItem(LS_CONSENT_KEY);
    if (!raw) return null;
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function setLocalStorageConsent(consent: StoredConsent): void {
  try {
    localStorage.setItem(LS_CONSENT_KEY, JSON.stringify(consent));
  } catch {
    // localStorage may be full or disabled.
  }
}

function clearLocalStorageConsent(): void {
  try {
    localStorage.removeItem(LS_CONSENT_KEY);
  } catch {
    // Ignore.
  }
}

// --- Public API: cookie-first with localStorage fallback ---

export function getStoredConsent(): StoredConsent | null {
  return getCookieConsent() ?? getLocalStorageConsent();
}

export function setStoredConsent(
  consent: StoredConsent,
  maxAgeDays: number,
): void {
  const cookieOk = setCookieConsent(consent, maxAgeDays);
  if (!cookieOk) {
    // Cookie blocked (e.g. Safari ITP, private browsing) — fall back.
    setLocalStorageConsent(consent);
  } else {
    // Keep localStorage in sync as a backup.
    setLocalStorageConsent(consent);
  }
}

export function clearStoredConsent(): void {
  clearCookieConsent();
  clearLocalStorageConsent();
}

export function generateVisitorId(): string {
  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  return Array.from(array, (b) => b.toString(16).padStart(2, "0")).join("");
}

// --- Consent API retry queue ---

export interface QueuedConsent {
  baseUrl: string;
  bannerId: string;
  visitorId: string;
  consentData: Record<string, boolean>;
  action: string;
  timestamp: number;
}

export function enqueueConsent(entry: QueuedConsent): void {
  try {
    const raw = localStorage.getItem(LS_QUEUE_KEY);
    const queue: QueuedConsent[] = raw ? JSON.parse(raw) : [];
    queue.push(entry);
    // Keep at most 20 entries to avoid filling localStorage.
    if (queue.length > 20) {
      queue.splice(0, queue.length - 20);
    }
    localStorage.setItem(LS_QUEUE_KEY, JSON.stringify(queue));
  } catch {
    // localStorage may be full or disabled.
  }
}

export function dequeueAllConsents(): QueuedConsent[] {
  try {
    const raw = localStorage.getItem(LS_QUEUE_KEY);
    if (!raw) return [];
    localStorage.removeItem(LS_QUEUE_KEY);
    return JSON.parse(raw);
  } catch {
    return [];
  }
}
