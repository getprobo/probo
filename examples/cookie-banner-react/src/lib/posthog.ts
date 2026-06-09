// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import posthog from "posthog-js";
import type { BannerConfig } from "@probo/cookie-banner";
import { getConsent, type ConsentData } from "@probo/cookie-banner/consent";

export type ConsentMode = BannerConfig["consent_mode"];

/**
 * Fallback category slug to gate PostHog when the banner config does not flag
 * any category with `posthog_consent: true`. Most Probo banners ship with an
 * "analytics" category, hence this default.
 */
const FALLBACK_CATEGORY_SLUG = "analytics";

let subscribed = false;
let initialized = false;
let unsubscribeConsent: (() => void) | null = null;
let categorySlug: string = FALLBACK_CATEGORY_SLUG;
let consentMode: ConsentMode | null = null;
const statusListeners = new Set<() => void>();

export interface PosthogStatus {
  initialized: boolean;
  consentMode: ConsentMode | null;
  optedIn: boolean;
  optedOut: boolean;
  distinctId: string | null;
}

// Cached snapshot. `useSyncExternalStore` compares references with `Object.is`,
// so `getPosthogStatus()` must return a stable reference until something
// actually changes — otherwise React loops itself into a stack overflow.
let cachedStatus: PosthogStatus = {
  initialized: false,
  consentMode: null,
  optedIn: false,
  optedOut: false,
  distinctId: null,
};

/**
 * Wire up the consent subscription so that any future opt-in / opt-out
 * decisions are mirrored to PostHog.
 *
 * Note: this does NOT call `posthog.init()`. We can't choose `cookieless_mode`
 * until the banner config tells us the regulation's `consent_mode`, so the
 * actual SDK init is deferred to {@link configurePosthogFromBanner}, which
 * the `probo-ready` event handler should invoke with the banner config.
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
 * The init options are derived from the current consent snapshot, which the
 * banner client has already populated with either the persisted answer (cookie
 * or API) or the per-regulation default (`true` for non-necessary categories
 * under `OPT_OUT`, `false` under `OPT_IN`):
 * - Analytics allowed → cookies and capture on from the start; the
 *   subscription downgrades to cookieless if the visitor later rejects.
 * - Analytics denied → fully cookieless and opted-out; the subscription opts
 *   in (still cookieless for the rest of the session) if the visitor later
 *   accepts.
 *
 * Driving the init off the snapshot rather than `consent_mode` plugs the race
 * where `posthog.init()` fires a `$pageview` (and sets the posthog cookie)
 * before we can call `opt_out_capturing()` for an `OPT_OUT` visitor who
 * already rejected on a prior visit.
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
      cookieless_mode: analyticsAllowed ? "on_reject" : "always",
      opt_out_capturing_by_default: !analyticsAllowed,
      person_profiles: "identified_only",
      respect_dnt: true,
      debug: import.meta.env.DEV,
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
  subscribed = false;
}

/** Subscribe to PostHog status changes (init / opt-in / opt-out / category). */
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
  } else {
    posthog.opt_out_capturing();
  }
}

function refreshStatus(): void {
  const next: PosthogStatus = initialized
    ? {
        initialized: true,
        consentMode,
        optedIn: posthog.has_opted_in_capturing(),
        optedOut: posthog.has_opted_out_capturing(),
        distinctId: posthog.get_distinct_id?.() ?? null,
      }
    : {
        initialized: false,
        consentMode,
        optedIn: false,
        optedOut: false,
        distinctId: null,
      };

  if (statusEqual(cachedStatus, next)) return;
  cachedStatus = next;
  for (const cb of statusListeners) cb();
}

function statusEqual(a: PosthogStatus, b: PosthogStatus): boolean {
  return (
    a.initialized === b.initialized &&
    a.consentMode === b.consentMode &&
    a.optedIn === b.optedIn &&
    a.optedOut === b.optedOut &&
    a.distinctId === b.distinctId
  );
}
