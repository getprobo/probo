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

import {
  resolveLanguage,
  type SupportedLanguage,
} from "./resolveLanguage";

// Short tags used in URL path segments (and persisted on Identity.locale).
export const URL_LOCALES = [
  "en",
  "fr",
  "de",
  "es",
  "id",
  "it",
  "ja",
  "ko",
  "pl",
  "pt",
  "tr",
  "uk",
  "zh",
] as const;

export type UrlLocale = (typeof URL_LOCALES)[number];

const DEFAULT_URL_LOCALE: UrlLocale = "en";

const LANGUAGE_TO_URL_LOCALE: Record<SupportedLanguage, UrlLocale> = {
  "en-US": "en",
  "fr-FR": "fr",
  "de-DE": "de",
  "es-ES": "es",
  "id-ID": "id",
  "it-IT": "it",
  "ja-JP": "ja",
  "ko-KR": "ko",
  "pl-PL": "pl",
  "pt-PT": "pt",
  "tr-TR": "tr",
  "uk-UA": "uk",
  "zh-CN": "zh",
};

const URL_LOCALE_TO_LANGUAGE = Object.fromEntries(
  Object.entries(LANGUAGE_TO_URL_LOCALE).map(([language, locale]) => [locale, language]),
) as Record<UrlLocale, SupportedLanguage>;

// Native-language labels for the locale switcher (not translated via i18n).
export const URL_LOCALE_LABELS: Record<UrlLocale, string> = {
  en: "English",
  fr: "Français",
  de: "Deutsch",
  es: "Español",
  id: "Bahasa Indonesia",
  it: "Italiano",
  ja: "日本語",
  ko: "한국어",
  pl: "Polski",
  pt: "Português",
  tr: "Türkçe",
  uk: "Українська",
  zh: "中文",
};

export function isUrlLocale(value: string | undefined | null): value is UrlLocale {
  return value != null && (URL_LOCALES as readonly string[]).includes(value);
}

export function languageToUrlLocale(language: SupportedLanguage): UrlLocale {
  return LANGUAGE_TO_URL_LOCALE[language];
}

export function urlLocaleToLanguage(locale: UrlLocale): SupportedLanguage {
  return URL_LOCALE_TO_LANGUAGE[locale];
}

export function resolveUrlLocale(): UrlLocale {
  return languageToUrlLocale(resolveLanguage());
}

export function parseUrlLocale(value: string | undefined | null): UrlLocale | null {
  if (!isUrlLocale(value)) {
    return null;
  }
  return value;
}

export function defaultUrlLocale(): UrlLocale {
  return DEFAULT_URL_LOCALE;
}

// Prefix an app-absolute path with a locale segment. Paths must start with "/".
// "/" becomes "/en"; "/documents" becomes "/en/documents".
export function localizedPath(locale: UrlLocale, path: string): string {
  if (!path.startsWith("/")) {
    throw new Error(`localizedPath expects an absolute path, got: ${path}`);
  }
  if (path === "/") {
    return `/${locale}`;
  }
  return `/${locale}${path}`;
}

// Swap the locale prefix on a basename-relative pathname (e.g. "/en/documents"
// → "/fr/documents"). Paths without a valid locale prefix are treated as
// unprefixed and get the new locale prepended.
export function replaceLocaleInPathname(pathname: string, locale: UrlLocale): string {
  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) {
    return `/${locale}`;
  }
  if (isUrlLocale(segments[0])) {
    segments[0] = locale;
    return `/${segments.join("/")}`;
  }
  return `/${locale}/${segments.join("/")}`;
}

// Strip a leading locale segment from a basename-relative pathname.
export function stripLocaleFromPathname(pathname: string): string {
  const segments = pathname.split("/").filter(Boolean);
  if (segments.length === 0) {
    return "/";
  }
  if (isUrlLocale(segments[0])) {
    const rest = segments.slice(1);
    return rest.length === 0 ? "/" : `/${rest.join("/")}`;
  }
  return pathname.startsWith("/") ? pathname : `/${pathname}`;
}
