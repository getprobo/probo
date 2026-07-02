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

import type { Category } from "../types";
import type { ConsentIntegration } from "./integration";

export class GoogleConsentModeIntegration implements ConsentIntegration {
  private hasMapping(categories: Category[]): boolean {
    return categories.some(
      (cat) => cat.gcm_consent_types && cat.gcm_consent_types.length > 0,
    );
  }

  private getConsentFn(): (...args: unknown[]) => void {
    const w = window as unknown as Record<string, unknown>;

    if (typeof w.gtag === "function") {
      return w.gtag as (...args: unknown[]) => void;
    }

    if (!Array.isArray(w.dataLayer)) {
      w.dataLayer = [];
    }

    const dataLayer = w.dataLayer as unknown[];
    return function () {
      dataLayer.push(arguments);
    };
  }

  bootstrap(): void {
    const consentFn = this.getConsentFn();
    consentFn("consent", "default", {
      ad_storage: "denied",
      ad_user_data: "denied",
      ad_personalization: "denied",
      analytics_storage: "denied",
      functionality_storage: "denied",
      personalization_storage: "denied",
      security_storage: "denied",
    });
  }

  setDefaults(categories: Category[]): void {
    if (!this.hasMapping(categories)) return;

    const consentFn = this.getConsentFn();

    const defaults: Record<string, string> = {};
    for (const cat of categories) {
      if (!cat.gcm_consent_types) continue;
      for (const gcmType of cat.gcm_consent_types) {
        defaults[gcmType] = "denied";
      }
    }

    if (Object.keys(defaults).length > 0) {
      consentFn("consent", "default", defaults);
    }
  }

  update(
    categories: Category[],
    consentData: Record<string, boolean>,
  ): void {
    if (!this.hasMapping(categories)) return;

    const consentFn = this.getConsentFn();

    const update: Record<string, string> = {};
    for (const cat of categories) {
      if (!cat.gcm_consent_types) continue;
      const granted = !!consentData[cat.slug];
      for (const gcmType of cat.gcm_consent_types) {
        update[gcmType] = granted ? "granted" : "denied";
      }
    }

    if (Object.keys(update).length > 0) {
      consentFn("consent", "update", update);
    }
  }
}
