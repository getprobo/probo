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

import { activateElements, observeAndActivate } from "./activation";
import { getConsentCookie, setConsentCookie } from "./cookie";
import { NotFoundError } from "./errors";
import { fetchJSON } from "./http";
import { enqueue, flush } from "./queue";
import { getOrCreateVisitorId } from "./visitor";

export interface CookieItem {
  name: string;
  duration: string;
  description: string;
}

export interface Category {
  name: string;
  description: string;
  required: boolean;
  cookies: CookieItem[];
}

export interface BannerConfig {
  banner_id: string;
  version: number;
  privacy_policy_url: string;
  consent_expiry_days: number;
  consent_mode: "OPT_IN" | "OPT_OUT";
  categories: Category[];
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
}

export class CookieBannerClient {
  private readonly baseUrl: string;
  private readonly bannerId: string;
  private readonly visitorId: string;

  private bannerConfig: BannerConfig | null = null;
  private consent: VisitorConsent | null = null;
  private observer: MutationObserver | null = null;

  constructor(config: CookieBannerClientOptions) {
    let base = config.baseUrl;
    while (base.endsWith("/")) {
      base = base.slice(0, -1);
    }
    this.baseUrl = base;
    this.bannerId = config.bannerId;
    this.visitorId = getOrCreateVisitorId(config.bannerId);
  }

  async load(): Promise<void> {
    const configUrl = `${this.baseUrl}/${this.bannerId}/config`;
    const config = await fetchJSON<BannerConfig>(configUrl);
    this.bannerConfig = config;

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
      return;
    }

    const consentUrl = `${this.baseUrl}/${this.bannerId}/consents/${this.visitorId}`;
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

  async acceptAll(): Promise<ConsentRecord> {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.name] = true;
    }

    return this.recordConsent("ACCEPT_ALL", consentData);
  }

  async rejectAll(): Promise<ConsentRecord> {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.name] = cat.required;
    }

    return this.recordConsent("REJECT_ALL", consentData);
  }

  async customize(
    categories: Record<string, boolean>,
  ): Promise<ConsentRecord> {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.name] = cat.required || !!categories[cat.name];
    }

    return this.recordConsent("CUSTOMIZE", consentData);
  }

  private async recordConsent(
    action: ConsentAction,
    consentData: Record<string, boolean>,
  ): Promise<ConsentRecord> {
    const cfg = this.config;
    const url = `${this.baseUrl}/${this.bannerId}/consents`;
    const body = {
      visitor_id: this.visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
    };

    let record: ConsentRecord | null = null;
    try {
      record = await fetchJSON<ConsentRecord>(url, {
        method: "POST",
        body,
      });
      void flush(this.bannerId);
    } catch {
      enqueue(this.bannerId, url, body);
    }

    this.consent = {
      visitor_id: this.visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
      created_at: record?.created_at ?? "",
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

    return record ?? {
      id: "",
      visitor_id: this.visitorId,
      action,
      created_at: "",
    };
  }

  private activate(consentData: Record<string, boolean>): void {
    activateElements(consentData);
    if (!this.observer) {
      this.observer = observeAndActivate(consentData);
    }
  }

  destroy(): void {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  }
}
