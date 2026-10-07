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
  { code: "en", label: "English" },
  { code: "de", label: "Deutsch" },
  { code: "es", label: "Español" },
  { code: "fr", label: "Français" },
  { code: "id", label: "Indonesian" },
  { code: "it", label: "Italiano" },
  { code: "ja", label: "日本語" },
  { code: "ko", label: "한국어" },
  { code: "nl", label: "Nederlands" },
  { code: "pl", label: "Polski" },
  { code: "pt", label: "Português" },
  { code: "tr", label: "Türkçe" },
  { code: "uk", label: "Українська" },
  { code: "zh", label: "中文" },
] as const;

const BANNER_KEYS = [
  "banner_title",
  "banner_description",
  "button_accept_all",
  "button_reject_all",
  "button_customize",
  "cookie_policy_link_text",
  "privacy_policy_link_text",
] as const;

const PANEL_KEYS = [
  "panel_title",
  "panel_description",
  "button_save",
  "aria_close",
  "aria_show_details",
  "aria_hide_details",
  "aria_cookie_settings",
] as const;

const PLACEHOLDER_KEYS = [
  "placeholder_text",
  "placeholder_button",
] as const;

export type TranslationKey
  = | (typeof BANNER_KEYS)[number]
    | (typeof PANEL_KEYS)[number]
    | (typeof PLACEHOLDER_KEYS)[number];

export const ALL_KEYS: readonly TranslationKey[] = [
  ...BANNER_KEYS,
  ...PANEL_KEYS,
  ...PLACEHOLDER_KEYS,
];

export type TranslationFormValues = Record<TranslationKey, string> & {
  categories: CategoryTranslations;
};

export interface CategoryInfo {
  id: string;
  name: string;
  slug: string;
  description: string;
  kind: string;
}

export type CategoryTranslations = Record<
  string,
  { name: string; description: string }
>;
