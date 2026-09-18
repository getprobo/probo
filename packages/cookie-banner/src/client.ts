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
  deactivateElements,
  observeAndActivate,
} from "./activation";
import { getConsent } from "./consent";
import { COOKIE_NAME, getConsentCookie, setConsentCookie, type ConsentCookie } from "./cookie";
import type { Detector } from "./detectors";
import {
  CookieDetector,
  ReportQueue,
  ResourceDetector,
  resolveResourceReportingEnabled,
  shouldStartResourceDetector,
  StorageDetector,
} from "./detectors";
import { NotFoundError } from "./errors";
import { rewriteLegacyConsoleHost } from "./hosted-url";
import { fetchJSON } from "./http";
import { detectLanguage } from "./i18n";
import type { ConsentIntegration } from "./integrations";
import { createDefaultIntegrations } from "./integrations";
import { resolveLayout } from "./layout";
import { enqueue, flush } from "./queue";
import { getTCFRuntime } from "./addons";
import { projectConsentFromTCF } from "./project-consent";
import type {
  BannerConfig,
  ConsentAction,
  ConsentRecord,
  CookieBannerClientOptions,
  Regulation,
  VisitorConsent,
} from "./types";
import { getOrCreateVisitorId, getVisitorId } from "./visitor";

function tcfProjects(config: BannerConfig): boolean {
  return !!config.tcf?.gvl &&
    (config.regulation === "GDPR" || config.regulation === "UK_GDPR");
}

export type {
  BannerConfig,
  Category,
  ConsentAction,
  ConsentRecord,
  CookieBannerClientOptions,
  CookieItem,
  Regulation,
  VisitorConsent,
} from "./types";

// Normalize fields that older self-hosted backends may omit so the rest of the
// client can read `layout` and `resource_reporting_enabled` without re-checking.
function resolveConfig(config: BannerConfig): BannerConfig {
  return {
    ...config,
    layout: resolveLayout(config),
    resource_reporting_enabled: resolveResourceReportingEnabled(config),
  };
}

export class CookieBannerClient {
  private readonly baseUrl: URL;
  private readonly bannerId: string;
  private visitorId: string | null;
  private readonly lang: string;

  private readonly integrations: ConsentIntegration[];

  private bannerConfig: BannerConfig | null = null;
  private consent: VisitorConsent | null = null;
  private observer: MutationObserver | null = null;
  private detectors: Detector[] = [];
  private reportQueue: ReportQueue | null = null;
  private _gpcApplied = false;

  constructor(config: CookieBannerClientOptions) {
    let base = config.baseUrl;
    if (!base.endsWith("/")) {
      base += "/";
    }
    this.baseUrl = rewriteLegacyConsoleHost(new URL(base));
    this.bannerId = config.bannerId;
    this.visitorId = getVisitorId(config.bannerId);
    this.lang = detectLanguage(config.lang);
    this.integrations = createDefaultIntegrations(config.integrations);
  }

  get loaded(): boolean {
    return this.bannerConfig !== null;
  }

  async load(): Promise<void> {
    for (const integration of this.integrations) {
      integration.bootstrap();
    }

    const configUrl = new URL(`${this.bannerId}/config`, this.baseUrl);
    if (this.lang) {
      configUrl.searchParams.set("lang", this.lang);
    }

    let config: BannerConfig;
    try {
      config = resolveConfig(await fetchJSON<BannerConfig>(configUrl));
    } catch {
      // Discovery mode: no published banner config, but detectors still run
      // so admins can inventory trackers. Grant GCM so GTM-managed tags can
      // fire; otherwise bootstrap's deny-all would hide them from discovery.
      for (const integration of this.integrations) {
        integration.grantAll();
      }
      this.startDetector();
      if (this.observer) {
        this.observer.disconnect();
      }
      this.observer = observeAndActivate({}, {});
      getConsent()._setReady({}, false);
      return;
    }
    this.bannerConfig = config;

    for (const integration of this.integrations) {
      integration.setDefaults(config.categories);
    }
    this.startDetector(config);

    if (this.visitorId) {
      const cookie = getConsentCookie();
      if (cookie && cookie.bid === this.bannerId && cookie.v === config.version && cookie.vid === this.visitorId) {
        this.consent = {
          visitor_id: cookie.vid,
          version: cookie.v,
          action: cookie.action,
          consent_data: cookie.data,
          created_at: "",
          tc: cookie.tc,
        };
        this._gpcApplied = cookie.action === "GPC";
        const cookieData = this.projectedConsent(cookie.data, cookie.tc);
        this.activate(cookieData);
        getConsent()._setReady(cookieData, true);
        getTCFRuntime()?.onConfig(config, cookie.tc);
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
        this._gpcApplied = apiConsent.action === "GPC";
        const restored: ConsentCookie = {
          bid: this.bannerId,
          v: apiConsent.version,
          vid: apiConsent.visitor_id,
          action: apiConsent.action,
          data: apiConsent.consent_data,
        };
        if (apiConsent.tc) {
          restored.tc = apiConsent.tc;
        }
        setConsentCookie(restored, config.consent_expiry_days);
        const restoredData = this.projectedConsent(apiConsent.consent_data, apiConsent.tc);
        this.activate(restoredData);
        getConsent()._setReady(restoredData, true);
        getTCFRuntime()?.onConfig(config, apiConsent.tc);
      } else {
        this.consent = null;
        getTCFRuntime()?.onConfig(config);
      }
    } else {
      getTCFRuntime()?.onConfig(config);
    }

    if (!this.consent && this.gpcDetected) {
      const gpcData: Record<string, boolean> = {};
      for (const cat of config.categories) {
        gpcData[cat.slug] = cat.kind === "NECESSARY";
      }
      getConsent()._setReady(gpcData, false);
      this.gpc();
      this._gpcApplied = true;
    } else if (!this.consent) {
      const defaults = this.buildDefaultConsentData();
      this.activate(defaults);
      getConsent()._setReady(defaults, false);
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

  get gpcDetected(): boolean {
    return typeof navigator !== "undefined" &&
      (navigator as Navigator & { globalPrivacyControl?: boolean }).globalPrivacyControl === true;
  }

  get gpcApplied(): boolean {
    return this._gpcApplied;
  }

  get regulation(): Regulation | null {
    return this.bannerConfig?.regulation ?? null;
  }

  gpc(): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = cat.kind === "NECESSARY";
    }

    this.recordConsent("GPC", consentData);
  }

  acceptAll(): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = true;
    }

    this.recordConsent("ACCEPT_ALL", consentData);
  }

  // acknowledge records dismissal of a notice-only banner. Under implied
  // consent all categories are already granted, so it persists that state with
  // the ACKNOWLEDGE action to distinguish it from an explicit accept-all.
  acknowledge(): void {
    const cfg = this.config;

    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = true;
    }

    this.recordConsent("ACKNOWLEDGE", consentData);
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
    const pendingChoices = getTCFRuntime()?.getPendingChoices?.();
    const consentData = tcfProjects(cfg) && pendingChoices
      ? projectConsentFromTCF(cfg.categories, pendingChoices)
      : Object.fromEntries(
          cfg.categories.map(cat => [
            cat.slug,
            cat.kind === "NECESSARY" || !!categories[cat.slug],
          ]),
        );

    this.recordConsent("CUSTOMIZE", consentData);
  }

  private ensureVisitorId(): string {
    if (!this.visitorId) {
      this.visitorId = getOrCreateVisitorId(this.bannerId);
    }
    return this.visitorId;
  }

  private recordConsent(
    action: ConsentAction,
    consentData: Record<string, boolean>,
  ): void {
    this._gpcApplied = action === "GPC";

    const cfg = this.config;
    const visitorId = this.ensureVisitorId();

    const tc = getTCFRuntime()?.onConsent(action, cfg);

    this.consent = {
      visitor_id: visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
      created_at: "",
    };
    if (tc) {
      this.consent.tc = tc;
    }

    const cookie: ConsentCookie = {
      bid: this.bannerId,
      v: cfg.version,
      vid: visitorId,
      action,
      data: consentData,
    };
    if (tc) {
      cookie.tc = tc;
    }

    setConsentCookie(cookie, cfg.consent_expiry_days);

    this.activate(consentData);
    getConsent()._notify(consentData);

    const url = new URL(`${this.bannerId}/consents`, this.baseUrl);
    const body: {
      visitor_id: string;
      version: number;
      action: ConsentAction;
      consent_data: Record<string, boolean>;
      tc?: string;
    } = {
      visitor_id: visitorId,
      version: cfg.version,
      action,
      consent_data: consentData,
    };
    if (tc) {
      body.tc = tc;
    }
    void fetchJSON<ConsentRecord>(url, { method: "POST", body })
      .then(() => void flush(this.bannerId))
      .catch(() => enqueue(this.bannerId, url.href, body));
  }

  private projectedConsent(
    stored: Record<string, boolean>,
    tc?: string,
  ): Record<string, boolean> {
    if (!tcfProjects(this.config) || !getTCFRuntime()?.decodeChoices) {
      return stored;
    }

    const choices = tc ? getTCFRuntime()?.decodeChoices?.(tc) ?? null : null;
    return projectConsentFromTCF(this.config.categories, choices);
  }

  private buildDefaultConsentData(): Record<string, boolean> {
    const cfg = this.config;
    const defaultGranted = cfg.layout.default_non_necessary_granted;
    const consentData: Record<string, boolean> = {};
    for (const cat of cfg.categories) {
      consentData[cat.slug] = defaultGranted || cat.kind === "NECESSARY";
    }
    return consentData;
  }

  private activate(consentData: Record<string, boolean>): void {
    for (const integration of this.integrations) {
      integration.update(this.config.categories, consentData);
    }

    const categoryCookies: Record<string, string[]> = {};
    const categoryLabels: Record<string, string> = {};
    for (const cat of this.config.categories) {
      categoryCookies[cat.slug] = cat.cookies.map((c) => c.name);
      categoryLabels[cat.slug] = cat.name;
    }

    const texts = this.config.texts;
    deactivateElements(consentData, categoryCookies, categoryLabels, texts);
    if (this.observer) {
      this.observer.disconnect();
    }
    this.observer = observeAndActivate(consentData, categoryLabels, texts);
  }

  private startDetector(config?: BannerConfig): void {
    this.stopDetectors();

    const knownNames = new Set<string>();
    knownNames.add(COOKIE_NAME);
    if (config) {
      for (const cat of config.categories) {
        for (const cookie of cat.cookies) {
          knownNames.add(cookie.name);
        }
      }
    }

    const reportUrl = new URL(`${this.bannerId}/report`, this.baseUrl);
    this.reportQueue = new ReportQueue(reportUrl);

    const apiOrigin = this.baseUrl.origin;
    this.detectors = [
      new CookieDetector(this.reportQueue, apiOrigin, knownNames),
      new StorageDetector(this.reportQueue, apiOrigin),
    ];

    if (shouldStartResourceDetector(config)) {
      this.detectors.push(new ResourceDetector(this.reportQueue, apiOrigin));
    }

    for (const d of this.detectors) {
      d.start();
    }
  }

  private stopDetectors(): void {
    for (const d of this.detectors) {
      d.stop();
    }
    this.detectors = [];
    if (this.reportQueue) {
      this.reportQueue.stop();
      this.reportQueue = null;
    }
  }

  destroy(): void {
    this.stopDetectors();
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
  }
}
