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

import type { BannerTexts } from "./i18n";

export interface CookieItem {
  name: string;
  max_age_seconds: number | null;
  description: string;
}

export interface Category {
  name: string;
  slug: string;
  description: string;
  kind: string;
  cookies: CookieItem[];
  gcm_consent_types: string[];
  posthog_consent: boolean;
}

export type Regulation =
  | "GDPR"
  | "UK_GDPR"
  | "FADP"
  | "CCPA"
  | "PIPEDA"
  | "LGPD"
  | "LFPDPPP"
  | "POPIA"
  | "PDPA"
  | "PIPL"
  | "PIPA"
  | "APPI"
  | "DPDP"
  | "PDPL";

export interface BannerConfig {
  banner_id: string;
  version: number;
  language: string;
  default_language: string;
  privacy_policy_url?: string;
  cookie_policy_url: string;
  consent_expiry_days: number;
  consent_mode: "OPT_IN" | "OPT_OUT";
  regulation: Regulation | null;
  show_branding: boolean;
  categories: Category[];
  texts: BannerTexts;
}

export type ConsentAction = "ACCEPT_ALL" | "REJECT_ALL" | "CUSTOMIZE" | "GPC";

export interface VisitorConsent {
  visitor_id: string;
  version: number;
  action: ConsentAction;
  consent_data: Record<string, boolean>;
  created_at: string;
}

export interface ConsentRecord {
  id: string;
  visitor_id: string;
  action: string;
  created_at: string;
}

export interface CookieBannerClientOptions {
  bannerId: string;
  baseUrl: string;
  lang?: string;
}
