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

import { fetchConfig, postConsent } from "./api";
import { applyConsent } from "./apply";
import { cleanupCookies } from "./cookies";
import {
  addConsentChangeListener,
  removeConsentChangeListener,
  notifyConsentChange,
} from "./events";
import { getStrings } from "./i18n";
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
  WidgetStrings,
} from "./types";

export class ConsentManager {
  private config: BannerConfig | null = null;
  private currentConsent: Record<string, boolean> = {};
  private visitorId: string;
  private strings: WidgetStrings;
  private readyCbs: Array<(config: BannerConfig) => void> = [];
  private consentRequiredCbs: Array<(config: BannerConfig) => void> = [];

  readonly bannerId: string;
  readonly baseUrl: string;

  constructor(private readonly managerConfig: ConsentManagerConfig) {
    this.bannerId = managerConfig.bannerId;
    this.baseUrl = managerConfig.baseUrl;
    this.strings = getStrings(managerConfig.lang ?? "en");
    this.visitorId =
      getStoredConsent()?.visitorId ?? generateVisitorId();
  }

  async init(): Promise<void> {
    const config = await fetchConfig(this.baseUrl, this.bannerId);
    if (!config) return;

    this.config = config;

    const stored = getStoredConsent();

    // Notify ready listeners
    for (const cb of this.readyCbs) {
      try {
        cb(config);
      } catch {
        // Never break the host site.
      }
    }

    // If consent exists, is valid, and version matches — apply silently
    if (stored && stored.version === config.version) {
      this.currentConsent = stored.categories;
      this.visitorId = stored.visitorId;
      applyConsent(stored.categories);
      return;
    }

    // Notify consent required listeners
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
    // Ensure required categories are always true
    for (const cat of this.config.categories) {
      if (cat.required) choices[cat.id] = true;
    }
    this.applyAndStore(choices, previousConsent, "CUSTOMIZE");
  }

  reset(): void {
    clearStoredConsent();
    this.currentConsent = {};
    this.visitorId = generateVisitorId();

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
    // If already initialized, fire immediately
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

  onConsentChange(cb: ConsentChangeCallback): void {
    addConsentChangeListener(cb);
  }

  removeConsentChangeListener(cb: ConsentChangeCallback): void {
    removeConsentChangeListener(cb);
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
    applyConsent(data, previousConsent);
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
