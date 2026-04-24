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

import type { Category } from "../client";
import type { ConsentIntegration } from "./integration";

export class GoogleConsentModeIntegration implements ConsentIntegration {
  private hasMapping(categories: Category[]): boolean {
    return categories.some(
      (cat) => cat.gcm_consent_types && cat.gcm_consent_types.length > 0,
    );
  }

  private getConsentFn(): ((...args: unknown[]) => void) | null {
    const w = window as unknown as Record<string, unknown>;

    if (typeof w.gtag === "function") {
      return w.gtag as (...args: unknown[]) => void;
    }

    if (!Array.isArray(w.dataLayer)) return null;

    const dataLayer = w.dataLayer as unknown[];
    return function () {
      dataLayer.push(arguments);
    };
  }

  setDefaults(categories: Category[]): void {
    if (!this.hasMapping(categories)) return;

    const consentFn = this.getConsentFn();
    if (!consentFn) return;

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
    if (!consentFn) return;

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
