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

import {
  activateElements,
  addPlaceholders,
  deactivateElements,
  observeAndActivate,
} from "./activation";
import { COOKIE_NAME, getConsentCookie, setConsentCookie } from "./cookie";
import { CookieDetector } from "./detector";
import { NotFoundError } from "./errors";
import { fetchJSON } from "./http";
import type { BannerTexts } from "./i18n";
import { detectLanguage } from "./i18n";
import { enqueue, flush } from "./queue";
import { getOrCreateVisitorId } from "./visitor";

export interface CookieItem {
  name: string;
  duration: string;
  description: string;
}

export interface Category {
  name: string;
  slug: string;
  description: string;
  kind: string;
  cookies: CookieItem[];
}

export interface BannerConfig {
  banner_id: string;
  version: number;
  language: string;
  default_language: string;
  privacy_policy_url: string;
  consent_expiry_days: number;
  consent_mode: "OPT_IN" | "OPT_OUT";
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

export class CookieBannerClient {
  private readonly baseUrl: URL;
  private readonly bannerId: string;
  private readonly visitorId: string;
  private readonly lang: string;

  private bannerConfig: BannerConfig | null = null;
  private consent: VisitorConsent | null = null;
  private observer: MutationObserver | null = null;
  private detector: CookieDetector | null = null;

  constructor(config: CookieBannerClientOptions) {
    let base = config.baseUrl;
    if (!base.endsWith("/")) {
      base += "/";
    }
    this.baseUrl = new URL(base);
    this.bannerId = config.bannerId;
    this.visitorId = getOrCreateVisitorId(config.bannerId);
    this.lang = detectLanguage(config.lang);
  }

  async load(): Promise<void> {
    const configUrl = new URL(`${this.bannerId}/config`, this.baseUrl);
    if (this.lang) {
      configUrl.searchParams.set("lang", this.lang);
    }
    const config = await fetchJSON<BannerConfig>(configUrl);
    this.bannerConfig = config;

    this.startDetector(config);

    const cookie = getConsentCookie();
    if (cookie && cookie.v === config.version && cookie.vid === this.visitorId) {
      this.consent = {
        visitor_id: cookie.vid,
        version: cookie.v,
        action: cookie.action,
        consent_data: cookie.data,
        created_at: "",
      };
      this.activate(cookie.data);
      void flush(this.bannerId);
      return;
    }

    const consentUrl = new URL(
      `${this.bannerId}/consents/${this.visitorId}`,
      this.baseUrl,
    );
    const apiConsent = await fetchJSON<VisitorConsent>(consentUrl).catch(
      (err) => {
        if (err instanceof NotFoundError) {
          return null;
        }
        throw err;
      },
    );

    if (apiConsent && apiConsent.version === config.version) {
      this.consent = apiConsent;
      setConsentCookie(
        {
          v: apiConsent.version,
          vid: apiConsent.visitor_id,
          action: apiConsent.action,
          data: apiConsent.consent_data,
        },
        config.consent_expiry_days,
      );
      this.activate(apiConsent.consent_data);
    } else {
      this.consent = null;
    }

    void flush(this.bannerId);
  }

  get config(): BannerConfig {
    if (!this.bannerConfig) {
      throw new Error("CookieBannerClient not loaded: call load() first");
    }
    return this.bannerConfig;
  }

  get visitorConsent(): VisitorConsent | null {
    return this.consent;
  }

  get hasConsent(): boolean {
    return this.consent !== null;
  }

  acceptAll(): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = true;
    }

    this.recordConsent("ACCEPT_ALL", consentData);
  }

  rejectAll(): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = cat.kind === "NECESSARY";
    }

    this.recordConsent("REJECT_ALL", consentData);
  }

  customize(categories: Record<string, boolean>): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = cat.kind === "NECESSARY" || !!categories[cat.slug];
    }

    this.recordConsent("CUSTOMIZE", consentData);
  }

  private recordConsent(
    action: ConsentAction,
    consentData: Record<string, boolean>,
  ): void {
    const cfg = this.config;

    this.consent = {
      visitor_id: this.visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
      created_at: "",
    };

    setConsentCookie(
      {
        v: cfg.version,
        vid: this.visitorId,
        action,
        data: consentData,
      },
      cfg.consent_expiry_days,
    );

    this.activate(consentData);

    const url = new URL(`${this.bannerId}/consents`, this.baseUrl);
    const body = {
      visitor_id: this.visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
    };
    void fetchJSON<ConsentRecord>(url, { method: "POST", body })
      .then(() => void flush(this.bannerId))
      .catch(() => enqueue(this.bannerId, url.href, body));
  }

  private activate(consentData: Record<string, boolean>): void {
    const categoryCookies: Record<string, string[]> = {};
    const categoryLabels: Record<string, string> = {};
    for (const cat of this.config.categories) {
      categoryCookies[cat.slug] = cat.cookies.map((c) => c.name);
      categoryLabels[cat.slug] = cat.name;
    }

    const texts = this.config.texts;
    deactivateElements(consentData, categoryCookies, categoryLabels, texts);
    activateElements(consentData);
    addPlaceholders(consentData, categoryLabels, texts);
    if (this.observer) {
      this.observer.disconnect();
    }
    this.observer = observeAndActivate(consentData, categoryLabels, texts);
  }

  private startDetector(config: BannerConfig): void {
    const knownNames = new Set<string>();
    knownNames.add(COOKIE_NAME);
    for (const cat of config.categories) {
      for (const cookie of cat.cookies) {
        knownNames.add(cookie.name);
      }
    }

    if (this.detector) {
      this.detector.stop();
    }
    this.detector = new CookieDetector(this.baseUrl, this.bannerId, knownNames);
    this.detector.start();
  }

  destroy(): void {
    if (this.detector) {
      this.detector.stop();
      this.detector = null;
    }
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  }
}
