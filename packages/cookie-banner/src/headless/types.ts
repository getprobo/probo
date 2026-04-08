// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

export interface BannerCookieItem {
  name: string;
  duration: string;
  description: string;
}

export interface BannerCategory {
  id: string;
  name: string;
  description: string;
  required: boolean;
  rank: number;
  cookies: BannerCookieItem[];
}

export interface ThemeConfig {
  primary_color: string;
  primary_text_color: string;
  secondary_color: string;
  secondary_text_color: string;
  background_color: string;
  text_color: string;
  secondary_text_body_color: string;
  border_color: string;
  font_family: string;
  border_radius: number;
  position: "bottom" | "bottom-left" | "bottom-right" | "center";
  revisit_position: "bottom-left" | "bottom-right";
}

export interface BannerConfig {
  id: string;
  title: string;
  description: string;
  accept_all_label: string;
  reject_all_label: string;
  save_preferences_label: string;
  privacy_policy_url: string;
  consent_expiry_days: number;
  consent_mode: ConsentMode;
  version: number;
  categories: BannerCategory[];
  theme: ThemeConfig;
}

export interface StoredConsent {
  visitorId: string;
  version: number;
  categories: Record<string, boolean>;
  timestamp: number;
}

export type ConsentChangeCallback = (
  consents: Record<string, boolean>,
) => void;

export type ConsentMode = "opt-in" | "opt-out";

export interface ConsentManagerConfig {
  bannerId: string;
  baseUrl: string;
  lang?: string;
  consentMode?: ConsentMode;
}

export interface WidgetStrings {
  customize: string;
  cookiePreferences: string;
  back: string;
  rejectAll: string;
  acceptAll: string;
  savePreferences: string;
  privacyPolicy: string;
  cookiePreferencesTooltip: string;
  contextualBlockedMessage: string;
  contextualAllowButton: string;
}
