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

import { interpolate } from "./i18n";
import type { TrackerType } from "./types";

// Tracker types whose data persists until explicitly cleared. When such a
// tracker has no max-age, its lifetime is "persistent" rather than "session"
// (cookies and session storage are cleared when the session/tab ends).
const PERSISTENT_TRACKER_TYPES: ReadonlySet<TrackerType> = new Set([
  "LOCAL_STORAGE",
  "INDEXED_DB",
  "CACHE_STORAGE",
]);

interface DurationTexts {
  [key: string]: string;
}

const durationTextsByLanguage: Record<string, DurationTexts> = {
  en: {
    duration_year_one: "{{count}} year",
    duration_year_other: "{{count}} years",
    duration_month_one: "{{count}} month",
    duration_month_other: "{{count}} months",
    duration_week_one: "{{count}} week",
    duration_week_other: "{{count}} weeks",
    duration_day_one: "{{count}} day",
    duration_day_other: "{{count}} days",
    duration_hour_one: "{{count}} hour",
    duration_hour_other: "{{count}} hours",
    duration_minute_one: "{{count}} minute",
    duration_minute_other: "{{count}} minutes",
    duration_second_one: "{{count}} second",
    duration_second_other: "{{count}} seconds",
    duration_session: "session",
    duration_persistent: "persistent",
  },
  fr: {
    duration_year_one: "{{count}} an",
    duration_year_other: "{{count}} ans",
    duration_month_one: "{{count}} mois",
    duration_month_other: "{{count}} mois",
    duration_week_one: "{{count}} semaine",
    duration_week_other: "{{count}} semaines",
    duration_day_one: "{{count}} jour",
    duration_day_other: "{{count}} jours",
    duration_hour_one: "{{count}} heure",
    duration_hour_other: "{{count}} heures",
    duration_minute_one: "{{count}} minute",
    duration_minute_other: "{{count}} minutes",
    duration_second_one: "{{count}} seconde",
    duration_second_other: "{{count}} secondes",
    duration_session: "session",
    duration_persistent: "persistant",
  },
  de: {
    duration_year_one: "{{count}} Jahr",
    duration_year_other: "{{count}} Jahre",
    duration_month_one: "{{count}} Monat",
    duration_month_other: "{{count}} Monate",
    duration_week_one: "{{count}} Woche",
    duration_week_other: "{{count}} Wochen",
    duration_day_one: "{{count}} Tag",
    duration_day_other: "{{count}} Tage",
    duration_hour_one: "{{count}} Stunde",
    duration_hour_other: "{{count}} Stunden",
    duration_minute_one: "{{count}} Minute",
    duration_minute_other: "{{count}} Minuten",
    duration_second_one: "{{count}} Sekunde",
    duration_second_other: "{{count}} Sekunden",
    duration_session: "Sitzung",
    duration_persistent: "dauerhaft",
  },
  es: {
    duration_year_one: "{{count}} año",
    duration_year_other: "{{count}} años",
    duration_month_one: "{{count}} mes",
    duration_month_other: "{{count}} meses",
    duration_week_one: "{{count}} semana",
    duration_week_other: "{{count}} semanas",
    duration_day_one: "{{count}} día",
    duration_day_other: "{{count}} días",
    duration_hour_one: "{{count}} hora",
    duration_hour_other: "{{count}} horas",
    duration_minute_one: "{{count}} minuto",
    duration_minute_other: "{{count}} minutos",
    duration_second_one: "{{count}} segundo",
    duration_second_other: "{{count}} segundos",
    duration_session: "sesión",
    duration_persistent: "persistente",
  },
};

function getDurationTexts(lang?: string): DurationTexts {
  if (lang && durationTextsByLanguage[lang]) return durationTextsByLanguage[lang];
  return durationTextsByLanguage.en;
}

// [unitSeconds, textKey, snapBuffer]
// snapBuffer: if the remainder is within this many seconds of the next
// whole unit, round up instead of carrying into smaller units.
const DURATION_UNITS: [number, string, number][] = [
  [365 * 24 * 3600, "duration_year", 21 * 24 * 3600],
  [30 * 24 * 3600, "duration_month", 2 * 24 * 3600],
  [7 * 24 * 3600, "duration_week", 12 * 3600],
  [24 * 3600, "duration_day", 2 * 3600],
  [3600, "duration_hour", 5 * 60],
  [60, "duration_minute", 5],
  [1, "duration_second", 0],
];

export function humanizeDuration(
  seconds: number,
  lang?: string,
  trackerType?: TrackerType,
): string {
  const texts = getDurationTexts(lang);
  if (seconds <= 0) {
    return trackerType && PERSISTENT_TRACKER_TYPES.has(trackerType)
      ? texts.duration_persistent
      : texts.duration_session;
  }

  let remaining = seconds;
  const parts: string[] = [];

  for (const [unit, key, snap] of DURATION_UNITS) {
    if (remaining >= unit - snap) {
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
      parts.push(interpolate(texts[tplKey], { count: String(count) }));
    }
  }

  return parts.length > 0 ? parts.join(", ") : texts.duration_session;
}

export function parseCookieName(raw: string): string {
  const eqIdx = raw.indexOf("=");
  if (eqIdx === -1) return raw.trim();
  return raw.substring(0, eqIdx).trim();
}

export function normalizeCookieDomain(raw: string): string | null {
  const trimmed = raw.trim().toLowerCase();
  // A trailing dot is ignored by cookie parsers, so the cookie is
  // host-only. Do not strip it and treat the rest as a Domain.
  if (trimmed === "" || trimmed.endsWith(".")) return null;

  const value = trimmed.replace(/^\./, "");
  return value === "" ? null : value;
}

// domainAppliesToHost reports whether a Domain attribute would be
// accepted for this host. The browser drops a Domain that is not a
// suffix of the current hostname (dot-boundary), a single-label
// domain (e.g. Domain=com), or any Domain on an IP-literal host,
// leaving the cookie host-only. Multi-label public suffixes (co.uk)
// still need the PSL; this file does not ship one.
export function domainAppliesToHost(domain: string, hostname: string): boolean {
  const host = hostname.toLowerCase().replace(/^\[|\]$/g, "");
  if (isIPLiteralHost(host)) return false;
  if (!domain.includes(".")) return false;
  if (host === domain) return true;
  return host.endsWith("." + domain);
}

function isIPLiteralHost(host: string): boolean {
  if (host.includes(":")) return true;
  return /^(?:\d{1,3}\.){3}\d{1,3}$/.test(host);
}

export interface CookieDomainFields {
  cookie_domain?: string;
  host_only: boolean;
}

// parseCookieSetDomain reads the Domain= attribute from a document.cookie
// assignment. No attribute means the cookie is host-only.
export function parseCookieSetDomain(raw: string, hostname?: string): CookieDomainFields {
  const parts = raw.split(";").map((s) => s.trim());

  for (const part of parts) {
    if (!part.toLowerCase().startsWith("domain=")) continue;

    const normalized = normalizeCookieDomain(part.substring(7));
    if (normalized == null) return { host_only: true };

    if (hostname != null && !domainAppliesToHost(normalized, hostname)) {
      return { host_only: true };
    }

    return { cookie_domain: normalized, host_only: false };
  }

  return { host_only: true };
}

// cookieListItemDomain maps a Cookie Store item. A null domain is
// host-only; a string is the Domain attribute.
export function cookieListItemDomain(domain: string | null): CookieDomainFields {
  if (domain == null || domain === "") {
    return { host_only: true };
  }

  const normalized = normalizeCookieDomain(domain);
  if (normalized == null) return { host_only: true };

  return { cookie_domain: normalized, host_only: false };
}

// PostgreSQL INTEGER (int4) is the server column. Values above this
// are omitted so a far-future Max-Age cannot fail the report batch.
const MAX_INT4 = 2_147_483_647;

export function clampMaxAgeSeconds(seconds: number | null): number | null {
  if (seconds == null || !Number.isFinite(seconds)) {
    return null;
  }

  const rounded = Math.round(seconds);
  if (rounded <= 0 || rounded > MAX_INT4) {
    return null;
  }

  return rounded;
}

export function parseMaxAgeSeconds(raw: string): number | null {
  const parts = raw.split(";").map((s) => s.trim());

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("max-age=")) {
      return clampMaxAgeSeconds(parseInt(part.substring(8), 10));
    }
  }

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("expires=")) {
      const dateStr = part.substring(8);
      const expires = new Date(dateStr);
      if (isNaN(expires.getTime())) return null;
      return clampMaxAgeSeconds(
        (expires.getTime() - Date.now()) / 1000,
      );
    }
  }

  return null;
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
