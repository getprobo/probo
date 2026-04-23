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

export interface BannerTexts {
  [key: string]: string;
}

function normalizeLocale(locale: string): string {
  return locale.split("-")[0].toLowerCase();
}

export function detectLanguage(explicit?: string): string {
  if (explicit) return normalizeLocale(explicit);

  if (typeof document !== "undefined" && document.documentElement) {
    const htmlLang = document.documentElement.lang;
    if (htmlLang) return normalizeLocale(htmlLang);
  }

  if (typeof navigator !== "undefined" && navigator.language) {
    return normalizeLocale(navigator.language);
  }

  return "";
}

export function interpolate(
  template: string,
  vars: Record<string, string>,
): string {
  return template.replace(/\{\{(\w+)\}\}/g, (_, key) => vars[key] ?? "");
}

const COOKIE_DETAIL_LABELS: Record<string, Record<string, string>> = {
  en: { label_description: "Description: {{value}}", label_duration: "Duration: {{value}}" },
  fr: { label_description: "Description : {{value}}", label_duration: "Durée : {{value}}" },
  de: { label_description: "Beschreibung: {{value}}", label_duration: "Dauer: {{value}}" },
  es: { label_description: "Descripción: {{value}}", label_duration: "Duración: {{value}}" },
};

export function getCookieDetailLabels(lang: string): Record<string, string> {
  return COOKIE_DETAIL_LABELS[lang] ?? COOKIE_DETAIL_LABELS.en;
}
