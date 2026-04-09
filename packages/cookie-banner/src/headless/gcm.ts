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

import type { BannerCategory } from "./types";

type ConsentSignal = "granted" | "denied";

interface GCMState {
  ad_storage: ConsentSignal;
  ad_user_data: ConsentSignal;
  ad_personalization: ConsentSignal;
  analytics_storage: ConsentSignal;
  functionality_storage: ConsentSignal;
  personalization_storage: ConsentSignal;
  security_storage: ConsentSignal;
}

// Maps category names (lowercased) to the GCM signals they control.
const categorySignalMap: Record<string, (keyof GCMState)[]> = {
  necessary: ["functionality_storage", "security_storage"],
  analytics: ["analytics_storage"],
  marketing: [
    "ad_storage",
    "ad_user_data",
    "ad_personalization",
  ],
  preferences: ["personalization_storage"],
};

function getGtag(): ((...args: unknown[]) => void) | null {
  const w = window as unknown as Record<string, unknown>;

  if (typeof w.gtag === "function") {
    return w.gtag as (...args: unknown[]) => void;
  }

  // Ensure dataLayer exists so gtag calls are queued for when GTM loads.
  if (!Array.isArray(w.dataLayer)) {
    return null;
  }

  // Define a minimal gtag function that pushes to dataLayer.
  const gtag = function (...args: unknown[]) {
    (w.dataLayer as unknown[]).push(args);
  };
  w.gtag = gtag;
  return gtag;
}

function buildGCMState(
  categories: BannerCategory[],
  consent: Record<string, boolean>,
): GCMState {
  const state: GCMState = {
    ad_storage: "denied",
    ad_user_data: "denied",
    ad_personalization: "denied",
    analytics_storage: "denied",
    functionality_storage: "denied",
    personalization_storage: "denied",
    security_storage: "denied",
  };

  for (const cat of categories) {
    const granted = consent[cat.id] === true;
    if (!granted) continue;

    const signals =
      categorySignalMap[cat.name.toLowerCase()];
    if (!signals) continue;

    for (const signal of signals) {
      state[signal] = "granted";
    }
  }

  return state;
}

export function setGCMDefault(
  categories: BannerCategory[],
  consent: Record<string, boolean>,
): void {
  try {
    const gtag = getGtag();
    if (!gtag) return;

    const state = buildGCMState(categories, consent);
    gtag("consent", "default", {
      ...state,
      wait_for_update: 500,
    });
  } catch {
    // Never break the host site.
  }
}

export function updateGCM(
  categories: BannerCategory[],
  consent: Record<string, boolean>,
): void {
  try {
    const gtag = getGtag();
    if (!gtag) return;

    const state = buildGCMState(categories, consent);
    gtag("consent", "update", state);
  } catch {
    // Never break the host site.
  }
}
