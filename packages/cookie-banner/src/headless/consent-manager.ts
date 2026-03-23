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

import { fetchConfig, postConsent, flushConsentQueue } from "./api";
import { applyConsent } from "./apply";
import {
  installCookieInterceptor,
  updateCookieInterceptor,
  uninstallCookieInterceptor,
} from "./cookie-interceptor";
import { cleanupCookies } from "./cookies";
import {
  addConsentChangeListener,
  removeConsentChangeListener,
  notifyConsentChange,
} from "./events";
import { setGCMDefault, updateGCM } from "./gcm";
import { isGPCEnabled } from "./gpc";
import { getStrings } from "./i18n";
import {
  startObserver,
  updateObserverConsent,
  stopObserver,
} from "./observer";
import {
  getStoredConsent,
  setStoredConsent,
  clearStoredConsent,
  generateVisitorId,
} from "./storage";
import type {
  BannerConfig,
  ConsentChangeCallback,
  ConsentManagerConfig,
  ConsentMode,
  WidgetStrings,
} from "./types";

export class ConsentManager {
  private config: BannerConfig | null = null;
  private currentConsent: Record<string, boolean> = {};
  private visitorId: string;
  private strings: WidgetStrings;
  private readyCbs: Array<(config: BannerConfig) => void> = [];
  private consentRequiredCbs: Array<(config: BannerConfig) => void> = [];
  private gpcHandledCbs: Array<(config: BannerConfig) => void> = [];

  readonly bannerId: string;
  readonly baseUrl: string;
  readonly consentMode: ConsentMode;
  readonly gpcEnabled: boolean;

  constructor(private readonly managerConfig: ConsentManagerConfig) {
    this.bannerId = managerConfig.bannerId;
    this.baseUrl = managerConfig.baseUrl;
    this.consentMode = managerConfig.consentMode ?? "opt-in";
    this.strings = getStrings(managerConfig.lang ?? "en");
    this.visitorId =
      getStoredConsent()?.visitorId ?? generateVisitorId();
    this.gpcEnabled = isGPCEnabled();
  }

  async init(): Promise<void> {
    const config = await fetchConfig(this.baseUrl, this.bannerId);
    if (!config) return;

    this.config = config;

    // Install cookie interceptor before anything else to block
    // writes from scripts that may run before consent is given.
    const defaultConsent = this.buildDefaultConsent();
    installCookieInterceptor(config.categories, defaultConsent);

    // Set Google Consent Mode defaults.
    setGCMDefault(config.categories, defaultConsent);

    // Flush any queued consent API calls from previous sessions.
    flushConsentQueue();

    // Notify ready listeners.
    for (const cb of this.readyCbs) {
      try {
        cb(config);
      } catch {
        // Never break the host site.
      }
    }

    // GPC signal: auto-reject non-required categories and skip the banner.
    if (this.gpcEnabled) {
      const gpcData = this.buildConsentData(false);
      this.currentConsent = gpcData;
      updateCookieInterceptor(gpcData);
      applyConsent(gpcData);
      updateGCM(config.categories, gpcData);
      startObserver(gpcData);
      for (const cb of this.gpcHandledCbs) {
        try {
          cb(config);
        } catch {
          // Never break the host site.
        }
      }
      return;
    }

    const stored = getStoredConsent();

    // If consent exists, is valid, and version matches — apply silently.
    if (stored && stored.version === config.version) {
      this.currentConsent = stored.categories;
      this.visitorId = stored.visitorId;
      updateCookieInterceptor(stored.categories);
      applyConsent(stored.categories);
      updateGCM(config.categories, stored.categories);
      startObserver(stored.categories);
      return;
    }

    // No stored consent: apply defaults based on consent mode.
    if (this.consentMode === "opt-out") {
      const defaults = this.buildConsentData(true);
      this.currentConsent = defaults;
      updateCookieInterceptor(defaults);
      applyConsent(defaults);
      updateGCM(config.categories, defaults);
      startObserver(defaults);
    } else {
      // opt-in: only required categories are active — interceptor
      // already installed with defaultConsent above.
      applyConsent(defaultConsent);
      startObserver(defaultConsent);
    }

    // Notify consent required listeners.
    for (const cb of this.consentRequiredCbs) {
      try {
        cb(config);
      } catch {
        // Never break the host site.
      }
    }
  }

  destroy(): void {
    this.readyCbs = [];
    this.consentRequiredCbs = [];
    this.gpcHandledCbs = [];
    stopObserver();
    uninstallCookieInterceptor();
  }

  getConfig(): BannerConfig | null {
    return this.config;
  }

  getStrings(): WidgetStrings {
    return this.strings;
  }

  getConsent(categoryId: string): boolean | null {
    if (categoryId in this.currentConsent) {
      return this.currentConsent[categoryId];
    }
    return null;
  }

  getConsents(): Record<string, boolean> {
    return { ...this.currentConsent };
  }

  needsConsent(): boolean {
    if (!this.config) return false;
    if (this.gpcEnabled) return false;
    const stored = getStoredConsent();
    return !stored || stored.version !== this.config.version;
  }

  acceptAll(): void {
    if (!this.config) return;
    const previousConsent = { ...this.currentConsent };
    const data = this.buildConsentData(true);
    this.applyAndStore(data, previousConsent, "ACCEPT_ALL");
  }

  rejectAll(): void {
    if (!this.config) return;
    const previousConsent = { ...this.currentConsent };
    const data = this.buildConsentData(false);
    this.applyAndStore(data, previousConsent, "REJECT_ALL");
  }

  acceptCategory(categoryId: string): void {
    if (!this.config) return;
    const previousConsent = { ...this.currentConsent };
    const data = { ...this.currentConsent };
    for (const cat of this.config.categories) {
      if (cat.required) data[cat.id] = true;
    }
    data[categoryId] = true;
    this.applyAndStore(data, previousConsent, "ACCEPT_CATEGORY");
  }

  savePreferences(choices: Record<string, boolean>): void {
    if (!this.config) return;
    const previousConsent = { ...this.currentConsent };
    // Ensure required categories are always true.
    for (const cat of this.config.categories) {
      if (cat.required) choices[cat.id] = true;
    }
    this.applyAndStore(choices, previousConsent, "CUSTOMIZE");
  }

  reset(): void {
    clearStoredConsent();
    this.currentConsent = {};
    this.visitorId = generateVisitorId();
    stopObserver();

    if (this.config) {
      for (const cb of this.consentRequiredCbs) {
        try {
          cb(this.config);
        } catch {
          // Never break the host site.
        }
      }
    }
  }

  onReady(cb: (config: BannerConfig) => void): void {
    this.readyCbs.push(cb);
    // If already initialized, fire immediately.
    if (this.config) {
      try {
        cb(this.config);
      } catch {
        // Never break the host site.
      }
    }
  }

  onConsentRequired(cb: (config: BannerConfig) => void): void {
    this.consentRequiredCbs.push(cb);
  }

  onGPCHandled(cb: (config: BannerConfig) => void): void {
    this.gpcHandledCbs.push(cb);
  }

  onConsentChange(cb: ConsentChangeCallback): void {
    addConsentChangeListener(cb);
  }

  removeConsentChangeListener(cb: ConsentChangeCallback): void {
    removeConsentChangeListener(cb);
  }

  private buildDefaultConsent(): Record<string, boolean> {
    const data: Record<string, boolean> = {};
    if (!this.config) return data;
    for (const cat of this.config.categories) {
      data[cat.id] = cat.required;
    }
    return data;
  }

  private buildConsentData(accepted: boolean): Record<string, boolean> {
    const data: Record<string, boolean> = {};
    for (const cat of this.config!.categories) {
      data[cat.id] = cat.required || accepted;
    }
    return data;
  }

  private applyAndStore(
    data: Record<string, boolean>,
    previousConsent: Record<string, boolean>,
    action: string,
  ): void {
    this.currentConsent = data;
    setStoredConsent(
      {
        visitorId: this.visitorId,
        version: this.config!.version,
        categories: data,
        timestamp: Date.now(),
      },
      this.config!.consent_expiry_days,
    );
    updateCookieInterceptor(data);
    applyConsent(data, previousConsent);
    updateObserverConsent(data);
    updateGCM(this.config!.categories, data);
    cleanupCookies(this.config!.categories, previousConsent, data);
    notifyConsentChange(data);
    postConsent(
      this.baseUrl,
      this.bannerId,
      this.visitorId,
      data,
      action,
    );
  }
}
