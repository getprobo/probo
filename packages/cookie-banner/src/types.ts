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

import type { BannerTexts } from "./i18n";

export type TrackerType =
  | "COOKIE"
  | "LOCAL_STORAGE"
  | "SESSION_STORAGE"
  | "INDEXED_DB"
  | "CACHE_STORAGE";

export interface CookieItem {
  name: string;
  tracker_type: TrackerType;
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
