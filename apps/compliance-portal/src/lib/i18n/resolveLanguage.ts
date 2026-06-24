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

export const SUPPORTED_LANGUAGES = ["en-US", "fr-FR"] as const;

export type SupportedLanguage = (typeof SUPPORTED_LANGUAGES)[number];

// Collapse the browser's preferred languages to one of our supported locales:
// any fr* tag maps to fr-FR, any en* tag maps to en-US. en-US is the ultimate
// fallback when nothing matches. Resolving to a canonical supported tag here
// means i18next is never asked to load an unsupported locale; fallbackLng only
// has to cover individual missing keys.
export function resolveLanguage(): SupportedLanguage {
  const candidates = navigator.languages?.length
    ? navigator.languages
    : [navigator.language];

  for (const tag of candidates) {
    const lower = tag.toLowerCase();
    if (lower.startsWith("fr")) {
      return "fr-FR";
    }
    if (lower.startsWith("en")) {
      return "en-US";
    }
  }

  return "en-US";
}
