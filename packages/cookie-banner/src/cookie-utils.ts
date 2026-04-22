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

// [unitSeconds, singular, plural, snapBuffer]
// snapBuffer: if the remainder is within this many seconds of the next
// whole unit, round up instead of carrying into smaller units.
const DURATION_UNITS: [number, string, string, number][] = [
  [365 * 24 * 3600, "year", "years", 21 * 24 * 3600],
  [30 * 24 * 3600, "month", "months", 2 * 24 * 3600],
  [7 * 24 * 3600, "week", "weeks", 12 * 3600],
  [24 * 3600, "day", "days", 0],
  [3600, "hour", "hours", 5 * 60],
  [60, "minute", "minutes", 15],
];

function humanizeDuration(seconds: number): string {
  if (seconds <= 0) return "session";

  let remaining = seconds;
  const parts: string[] = [];

  for (const [unit, singular, plural, snap] of DURATION_UNITS) {
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

      parts.push(count === 1 ? `1 ${singular}` : `${count} ${plural}`);
    }
  }

  return parts.length > 0 ? parts.join(", ") : "session";
}

export function parseCookieName(raw: string): string {
  const eqIdx = raw.indexOf("=");
  if (eqIdx === -1) return raw.trim();
  return raw.substring(0, eqIdx).trim();
}

export function parseDuration(raw: string): string {
  const parts = raw.split(";").map((s) => s.trim());

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("max-age=")) {
      const val = parseInt(part.substring(8), 10);
      if (isNaN(val) || val <= 0) return "session";
      return humanizeDuration(val);
    }
  }

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("expires=")) {
      const dateStr = part.substring(8);
      const expires = new Date(dateStr);
      if (isNaN(expires.getTime())) return "session";
      const deltaSeconds = Math.round(
        (expires.getTime() - Date.now()) / 1000,
      );
      if (deltaSeconds <= 0) return "session";
      return humanizeDuration(deltaSeconds);
    }
  }

  return "session";
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
