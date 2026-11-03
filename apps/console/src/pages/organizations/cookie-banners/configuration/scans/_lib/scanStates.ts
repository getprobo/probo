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

export const cookieBannerScanStates = ["PENDING", "RUNNING", "PAUSED", "COMPLETED", "FAILED"] as const;
export type CookieBannerScanState = (typeof cookieBannerScanStates)[number];

export const scanStateBadges = {
  PENDING: { color: "neutral" as const, variant: "outline" as const, labelKey: "pending" },
  RUNNING: { color: "sky" as const, variant: "soft" as const, labelKey: "running" },
  PAUSED: { color: "amber" as const, variant: "soft" as const, labelKey: "paused" },
  COMPLETED: { color: "green" as const, variant: "soft" as const, labelKey: "completed" },
  FAILED: { color: "red" as const, variant: "soft" as const, labelKey: "failed" },
};

export const scanFailureReasonKeys = {
  UNREACHABLE: "unreachable",
  BLOCKED: "blocked",
  NO_PAGES_LOADED: "noPagesLoaded",
  INTERRUPTED: "interrupted",
  INTERNAL: "internal",
  BANNER_NOT_INSTALLED: "bannerNotInstalled",
  DIFFERENT_BANNER: "differentBanner",
} as const;

export function isCookieBannerScanState(value: string | null): value is CookieBannerScanState {
  return value != null && (cookieBannerScanStates as readonly string[]).includes(value);
}

export function isActiveScanState(state: string): boolean {
  return state === "PENDING" || state === "RUNNING" || state === "PAUSED";
}

export function scanStateBadge(state: string) {
  if (isCookieBannerScanState(state)) {
    return scanStateBadges[state];
  }
  return scanStateBadges.PENDING;
}

export function scanFailureReasonKey(reason: string): string | null {
  if (reason in scanFailureReasonKeys) {
    return scanFailureReasonKeys[reason as keyof typeof scanFailureReasonKeys];
  }
  return null;
}
