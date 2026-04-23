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

import { interpolate } from "./i18n";

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
    duration_session: "session",
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
    duration_session: "session",
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
    duration_session: "Sitzung",
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
    duration_session: "sesión",
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
  [24 * 3600, "duration_day", 0],
  [3600, "duration_hour", 5 * 60],
  [60, "duration_minute", 15],
];

function humanizeDuration(seconds: number, lang?: string): string {
  const texts = getDurationTexts(lang);
  if (seconds <= 0) return texts.duration_session;

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

export function parseDuration(raw: string, lang?: string): string {
  const texts = getDurationTexts(lang);
  const session = texts.duration_session;
  const parts = raw.split(";").map((s) => s.trim());

  for (const part of parts) {
    const lower = part.toLowerCase();
    if (lower.startsWith("max-age=")) {
      const val = parseInt(part.substring(8), 10);
      if (isNaN(val) || val <= 0) return session;
      return humanizeDuration(val, lang);
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
      return humanizeDuration(deltaSeconds, lang);
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
