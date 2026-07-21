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

export const SUPPORTED_LANGUAGES = [
  "en-US",
  "fr-FR",
  "de-DE",
  "es-ES",
  "id-ID",
  "it-IT",
  "ja-JP",
  "ko-KR",
  "pl-PL",
  "pt-PT",
  "tr-TR",
  "uk-UA",
  "zh-CN",
] as const;

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number];

// Maps a two-letter language prefix (lowercased) to its canonical supported
// tag, e.g. any "de*" browser tag resolves to "de-DE".
const PREFIX_TO_LANGUAGE: Record<string, SupportedLanguage> = {
  en: "en-US",
  fr: "fr-FR",
  de: "de-DE",
  es: "es-ES",
  id: "id-ID",
  it: "it-IT",
  ja: "ja-JP",
  ko: "ko-KR",
  pl: "pl-PL",
  pt: "pt-PT",
  tr: "tr-TR",
  uk: "uk-UA",
  zh: "zh-CN",
};

// Collapse the browser's preferred languages to one of our supported locales
// by matching the two-letter language prefix (e.g. any "fr*" tag maps to
// fr-FR). en-US is the ultimate fallback when nothing matches. Resolving to a
// canonical supported tag here means i18next is never asked to load an
// unsupported locale; fallbackLng only has to cover individual missing keys.
export function resolveLanguage(): SupportedLanguage {
  const candidates = navigator.languages?.length
    ? navigator.languages
    : [navigator.language];

  for (const tag of candidates) {
    const prefix = tag.toLowerCase().split("-")[0];
    const language = PREFIX_TO_LANGUAGE[prefix];
    if (language) {
      return language;
    }
  }

  return "en-US";
}
