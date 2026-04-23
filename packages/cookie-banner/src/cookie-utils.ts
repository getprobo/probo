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

import type { BannerTexts } from "./i18n";
import { interpolate } from "./i18n";

// [unitSeconds, textKey, snapBuffer]
// snapBuffer: if the remainder is within this many seconds of the next
// whole unit, round up instead of carrying into smaller units.
const DURATION_UNITS: [number, string, number][] = [
  [365 * 24 * 3600, "duration_year", 21 * 24 * 3600],
  [30 * 24 * 3600, "duration_month", 2 * 24 * 3600],
  [7 * 24 * 3600, "duration_week", 12 * 3600],
  [24 * 3600, "duration_day", 0],
  [3600, "duration_hour", 5 * 60],
  [60, "duration_minute", 15],
];

function humanizeDuration(seconds: number, texts?: BannerTexts): string {
  if (seconds <= 0) return texts?.duration_session ?? "session";

  let remaining = seconds;
  const parts: string[] = [];

  for (const [unit, key, snap] of DURATION_UNITS) {
    if (remaining >= unit) {
      let count = Math.floor(remaining / unit);
      const leftover = remaining - count * unit;

      if (leftover >= unit - snap) {
        count++;
        remaining = 0;
      } else if (leftover <= snap) {
        remaining = 0;
      } else {
        remaining = leftover;
      }

      const tplKey = count === 1 ? `${key}_one` : `${key}_other`;
      const tpl = texts?.[tplKey] ?? (count === 1 ? `{{count}} ${key.replace("duration_", "")}` : `{{count}} ${key.replace("duration_", "")}s`);
      parts.push(interpolate(tpl, { count: String(count) }));
    }
  }

  return parts.length > 0 ? parts.join(", ") : texts?.duration_session ?? "session";
}

export function parseCookieName(raw: string): string {
  const eqIdx = raw.indexOf("=");
  if (eqIdx === -1) return raw.trim();
  return raw.substring(0, eqIdx).trim();
}

export function parseDuration(raw: string, texts?: BannerTexts): string {
  const session = texts?.duration_session ?? "session";
  const parts = raw.split(";").map((s) => s.trim());

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("max-age=")) {
      const val = parseInt(part.substring(8), 10);
      if (isNaN(val) || val <= 0) return session;
      return humanizeDuration(val, texts);
    }
  }

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("expires=")) {
      const dateStr = part.substring(8);
      const expires = new Date(dateStr);
      if (isNaN(expires.getTime())) return session;
      const deltaSeconds = Math.round(
        (expires.getTime() - Date.now()) / 1000,
      );
      if (deltaSeconds <= 0) return session;
      return humanizeDuration(deltaSeconds, texts);
    }
  }

  return session;
}

export function isDeletion(raw: string): boolean {
  const parts = raw.split(";").map((s) => s.trim().toLowerCase());

  for (const part of parts) {
    if (part.startsWith("max-age=")) {
      const val = parseInt(part.substring(8), 10);
      if (val <= 0) return true;
    }
    if (part.startsWith("expires=")) {
      const dateStr = part.substring(8);
      const expires = new Date(dateStr);
      if (!isNaN(expires.getTime()) && expires.getTime() <= Date.now()) {
        return true;
      }
    }
  }

  return false;
}

function getCandidateDomains(hostname: string): string[] {
  const parts = hostname.split(".");
  if (parts.length <= 1) return [];

  const candidates: string[] = [];
  // Try progressively broader parent domains. The browser silently
  // ignores attempts to clear cookies on public suffixes, so
  // over-trying is safe and avoids maintaining a TLD list.
  for (let i = 0; i < parts.length - 1; i++) {
    candidates.push("." + parts.slice(i).join("."));
  }

  return candidates;
}

export function removeCookies(names: string[]): void {
  const domains = getCandidateDomains(location.hostname);

  for (const name of names) {
    document.cookie = `${name}=; path=/; max-age=0`;
    for (const domain of domains) {
      document.cookie = `${name}=; path=/; domain=${domain}; max-age=0`;
    }
  }
}
