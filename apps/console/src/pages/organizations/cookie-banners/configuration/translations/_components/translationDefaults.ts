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

export const SUPPORTED_LANGUAGES = [
  { code: "en", label: "English" },
  { code: "fr", label: "Français" },
  { code: "de", label: "Deutsch" },
  { code: "es", label: "Español" },
] as const;

export const BANNER_KEYS = [
  "banner_title",
  "banner_description",
  "button_accept_all",
  "button_reject_all",
  "button_customize",
  "cookie_policy_link_text",
  "privacy_policy_link_text",
] as const;

export const PANEL_KEYS = [
  "panel_title",
  "panel_description",
  "button_save",
  "aria_close",
  "aria_show_details",
  "aria_hide_details",
  "aria_cookie_settings",
] as const;

export const PLACEHOLDER_KEYS = [
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

export const TRANSLATION_LABELS: Record<string, string> = {
  banner_title: "Banner title",
  banner_description: "Banner description",
  button_accept_all: "Accept all button",
  button_reject_all: "Reject all button",
  button_customize: "Customize button",
  cookie_policy_link_text: "Cookie policy link text",
  privacy_policy_link_text: "Privacy policy link text",
  panel_title: "Panel title",
  panel_description: "Panel description",
  button_save: "Save button",
  aria_close: "Close (ARIA)",
  aria_show_details: "Show details (ARIA)",
  aria_hide_details: "Hide details (ARIA)",
  aria_cookie_settings: "Cookie settings (ARIA)",
  placeholder_text: "Placeholder text",
  placeholder_button: "Placeholder button",
};
