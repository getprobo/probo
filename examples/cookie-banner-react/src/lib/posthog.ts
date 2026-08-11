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

import posthog from "posthog-js";
import type { BannerConfig } from "@probo/cookie-banner";
import { getConsent, type ConsentData } from "@probo/cookie-banner/consent";

export type ConsentMode = BannerConfig["consent_mode"];
export type ExplicitConsentStatus = "granted" | "denied" | "pending";

/**
 * Fallback category slug to gate PostHog when the banner config does not flag
 * any category with `posthog_consent: true`. Most Probo banners ship with an
 * "analytics" category, hence this default.
 */
const FALLBACK_CATEGORY_SLUG = "analytics";

const DEFAULT_FEATURE_FLAG = "example-beta-panel";
const DEFAULT_DEMO_DISTINCT_ID = "cookie-banner-example-demo";

let subscribed = false;
let initialized = false;
let identified = false;
let unsubscribeConsent: (() => void) | null = null;
let unsubscribeFeatureFlags: (() => void) | null = null;
let categorySlug: string = FALLBACK_CATEGORY_SLUG;
let consentMode: ConsentMode | null = null;
const statusListeners = new Set<() => void>();

export interface PosthogStatus {
  initialized: boolean;
  consentMode: ConsentMode | null;
  consentStatus: ExplicitConsentStatus | null;
  optedIn: boolean;
  optedOut: boolean;
  identified: boolean;
  distinctId: string | null;
  featureFlagKey: string;
  featureFlagEnabled: boolean;
}

// Cached snapshot. `useSyncExternalStore` compares references with `Object.is`,
// so `getPosthogStatus()` must return a stable reference until something
// actually changes — otherwise React loops itself into a stack overflow.
let cachedStatus: PosthogStatus = {
  initialized: false,
  consentMode: null,
  consentStatus: null,
  optedIn: false,
  optedOut: false,
  identified: false,
  distinctId: null,
  featureFlagKey: featureFlagKey(),
  featureFlagEnabled: false,
};

/**
 * Wire up the consent subscription so that any future opt-in / opt-out
 * decisions are mirrored to PostHog.
 *
 * Note: this does NOT call `posthog.init()`. Init is deferred to
 * {@link configurePosthogFromBanner}, which the `probo-ready` event handler
 * should invoke with the banner config (so opt-out-by-default can be derived
 * from the consent snapshot before the first capture).
 *
 * Safe to call multiple times; only the first call wires the subscription.
 */
export function initPosthog(): void {
  if (subscribed) return;

  if (!import.meta.env.PUBLIC_POSTHOG_API_KEY) {
    console.warn(
      "[posthog] PUBLIC_POSTHOG_API_KEY is not set; skipping PostHog init. " +
        "Copy .env.example to .env in examples/cookie-banner-react/ and fill it in.",
    );
    return;
  }

  subscribed = true;

  const consent = getConsent();
  unsubscribeConsent = consent.subscribe((data: ConsentData) => {
    syncCapturing(data);
    refreshStatus();
  });
}

/**
 * Initialize PostHog (on first call) and route opt-in / opt-out decisions
 * through the category flagged with `posthog_consent: true` in the banner
 * config. Call this from the `probo-ready` event handler.
 *
 * Always uses `cookieless_mode: "on_reject"` so a later analytics accept can
 * leave cookieless mode, call `identify()`, and evaluate feature flags.
 * `opt_out_capturing_by_default` is derived from the current consent snapshot
 * (persisted answer or regulation default) so an `OPT_OUT` visitor who already
 * rejected does not get a cookied `$pageview` before `opt_out_capturing()`.
 *
 * Subsequent calls leave PostHog initialized and just refresh the category
 * slug / consent mode in the cached status snapshot.
 */
export function configurePosthogFromBanner(config: BannerConfig): void {
  const flagged = config.categories.find((c) => c.posthog_consent);
  const slug = flagged?.slug ?? FALLBACK_CATEGORY_SLUG;
  const consent = getConsent();

  if (!initialized) {
    const apiKey = import.meta.env.PUBLIC_POSTHOG_API_KEY;
    if (!apiKey) return;

    const analyticsAllowed = consent.getAll()[slug] === true;
    posthog.init(apiKey, {
      api_host: "https://t.probo.com",
      ui_host: "https://us.posthog.com",
      cookieless_mode: "on_reject",
      opt_out_capturing_by_default: !analyticsAllowed,
      person_profiles: "identified_only",
      respect_dnt: true,
      debug: import.meta.env.DEV,
    });
    unsubscribeFeatureFlags = posthog.onFeatureFlags(() => {
      refreshStatus();
    });
    initialized = true;
  }

  consentMode = config.consent_mode;
  categorySlug = slug;

  syncCapturing(consent.getAll());
  refreshStatus();
}

/** Tear down the consent subscription. Does not un-initialize PostHog itself. */
export function teardownPosthog(): void {
  if (unsubscribeConsent) {
    unsubscribeConsent();
    unsubscribeConsent = null;
  }
  if (unsubscribeFeatureFlags) {
    unsubscribeFeatureFlags();
    unsubscribeFeatureFlags = null;
  }
  subscribed = false;
}

/** Subscribe to PostHog status changes (init / opt-in / opt-out / flags). */
export function subscribePosthogStatus(cb: () => void): () => void {
  statusListeners.add(cb);
  return () => statusListeners.delete(cb);
}

export function getPosthogStatus(): PosthogStatus {
  return cachedStatus;
}

function syncCapturing(data: ConsentData): void {
  if (!initialized) return;
  if (data[categorySlug]) {
    posthog.opt_in_capturing();
    posthog.identify(demoDistinctId());
    identified = true;
  } else {
    // reset() clears stored consent back to the init default — call it
    // before opt_out so a boot with analytics already allowed does not
    // leave capturing on after revoke.
    posthog.reset();
    posthog.opt_out_capturing();
    identified = false;
  }
}

function isFlagEligible(): boolean {
  return (
    initialized &&
    posthog.get_explicit_consent_status() === "granted" &&
    identified
  );
}

function readDemoFlag(): boolean {
  if (!isFlagEligible()) return false;
  return posthog.isFeatureEnabled(featureFlagKey()) === true;
}

function featureFlagKey(): string {
  return import.meta.env.PUBLIC_POSTHOG_FEATURE_FLAG || DEFAULT_FEATURE_FLAG;
}

function demoDistinctId(): string {
  return (
    import.meta.env.PUBLIC_POSTHOG_DEMO_DISTINCT_ID || DEFAULT_DEMO_DISTINCT_ID
  );
}

function refreshStatus(): void {
  const next: PosthogStatus = initialized
    ? {
        initialized: true,
        consentMode,
        consentStatus: posthog.get_explicit_consent_status(),
        optedIn: posthog.has_opted_in_capturing(),
        optedOut: posthog.has_opted_out_capturing(),
        identified,
        distinctId: posthog.get_distinct_id?.() ?? null,
        featureFlagKey: featureFlagKey(),
        featureFlagEnabled: readDemoFlag(),
      }
    : {
        initialized: false,
        consentMode,
        consentStatus: null,
        optedIn: false,
        optedOut: false,
        identified: false,
        distinctId: null,
        featureFlagKey: featureFlagKey(),
        featureFlagEnabled: false,
      };

  if (statusEqual(cachedStatus, next)) return;
  cachedStatus = next;
  for (const cb of statusListeners) cb();
}

function statusEqual(a: PosthogStatus, b: PosthogStatus): boolean {
  return (
    a.initialized === b.initialized &&
    a.consentMode === b.consentMode &&
    a.consentStatus === b.consentStatus &&
    a.optedIn === b.optedIn &&
    a.optedOut === b.optedOut &&
    a.identified === b.identified &&
    a.distinctId === b.distinctId &&
    a.featureFlagKey === b.featureFlagKey &&
    a.featureFlagEnabled === b.featureFlagEnabled
  );
}
