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

import type { ConsentChangeCallback } from "../headless/types";
import { StyledBanner } from "./styled-banner";

let _banner: StyledBanner | null = null;

function getScriptInfo(): {
  bannerId: string;
  baseUrl: string;
  lang: string;
} | null {
  const scripts = document.querySelectorAll(
    'script[data-banner-id]',
  );
  const script = scripts[scripts.length - 1] as HTMLScriptElement | undefined;
  if (!script) return null;

  const bannerId = script.getAttribute("data-banner-id");
  if (!bannerId) return null;

  const src = script.getAttribute("src");
  if (!src) return null;

  const lang =
    script.getAttribute("data-lang") ||
    document.documentElement.lang ||
    "en";

  // Derive base URL: remove /widget.js from the script src
  const baseUrl = src.replace(/\/widget\.js(\?.*)?$/, "");
  return { bannerId, baseUrl, lang };
}

async function init(): Promise<void> {
  const info = getScriptInfo();
  if (!info) return;

  _banner = new StyledBanner({
    bannerId: info.bannerId,
    baseUrl: info.baseUrl,
    lang: info.lang,
  });

  await _banner.init();
}

// --- Public API (exposed as window.ProboCookieBanner.*) ---

export function show(): void {
  _banner?.show();
}

export function getConsent(categoryId: string): boolean | null {
  return _banner?.manager.getConsent(categoryId) ?? null;
}

export function getConsents(): Record<string, boolean> {
  return _banner?.manager.getConsents() ?? {};
}

export function onConsentChange(cb: ConsentChangeCallback): void {
  _banner?.manager.onConsentChange(cb);
}

export function removeConsentChangeListener(cb: ConsentChangeCallback): void {
  _banner?.manager.removeConsentChangeListener(cb);
}

export function reset(): void {
  if (_banner) {
    document.getElementById("probo-cookie-revisit")?.remove();
    _banner.manager.reset();
  }
}

// Initialize when DOM is ready
if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", () => {
    init();
  });
} else {
  init();
}
