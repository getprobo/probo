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

import type { BannerConfig, ThemeConfig } from "./types";
import { enqueueConsent, dequeueAllConsents } from "./storage";

export type { BannerCategory, BannerConfig, ThemeConfig } from "./types";

export const defaultTheme: ThemeConfig = {
  primary_color: "#2563eb",
  primary_text_color: "#ffffff",
  secondary_color: "#1a1a1a",
  secondary_text_color: "#ffffff",
  background_color: "#ffffff",
  text_color: "#1a1a1a",
  secondary_text_body_color: "#4b5563",
  border_color: "#e5e7eb",
  font_family:
    "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif",
  border_radius: 8,
  position: "bottom",
  revisit_position: "bottom-left",
};

export async function fetchConfig(
  baseUrl: string,
  bannerId: string,
): Promise<BannerConfig | null> {
  try {
    const resp = await fetch(`${baseUrl}/${bannerId}/config`);
    if (!resp.ok) return null;
    const data = await resp.json();
    return {
      ...data,
      theme: { ...defaultTheme, ...data.theme },
    } as BannerConfig;
  } catch {
    return null;
  }
}

async function sendConsent(
  baseUrl: string,
  bannerId: string,
  visitorId: string,
  consentData: Record<string, boolean>,
  action: string,
): Promise<boolean> {
  try {
    const resp = await fetch(`${baseUrl}/${bannerId}/consents`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        visitor_id: visitorId,
        consent_data: consentData,
        action,
      }),
    });
    return resp.ok;
  } catch {
    return false;
  }
}

export async function postConsent(
  baseUrl: string,
  bannerId: string,
  visitorId: string,
  consentData: Record<string, boolean>,
  action: string,
): Promise<void> {
  const ok = await sendConsent(
    baseUrl,
    bannerId,
    visitorId,
    consentData,
    action,
  );

  if (!ok) {
    enqueueConsent({
      baseUrl,
      bannerId,
      visitorId,
      consentData,
      action,
      timestamp: Date.now(),
    });
  }
}

export async function flushConsentQueue(): Promise<void> {
  const queued = dequeueAllConsents();
  for (let i = 0; i < queued.length; i++) {
    const entry = queued[i];
    const ok = await sendConsent(
      entry.baseUrl,
      entry.bannerId,
      entry.visitorId,
      entry.consentData,
      entry.action,
    );

    if (!ok) {
      // Re-queue the failed entry and all remaining entries.
      for (let j = i; j < queued.length; j++) {
        enqueueConsent(queued[j]);
      }
      break;
    }
  }
}
