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

// Save the native descriptor before anyone else can tamper with it.
const nativeDescriptor = Object.getOwnPropertyDescriptor(
  Document.prototype,
  "cookie",
);

let installed = false;
let categories: BannerCategory[] = [];
let currentConsent: Record<string, boolean> = {};

// Our own cookie name — must always be allowed through.
const PROBO_COOKIE = "probo_consent";

function parseCookieName(cookieStr: string): string {
  const eqIdx = cookieStr.indexOf("=");
  if (eqIdx === -1) return cookieStr.trim();
  return cookieStr.substring(0, eqIdx).trim();
}

function matchesPattern(cookieName: string, pattern: string): boolean {
  if (pattern.startsWith("^") || pattern.endsWith("$")) {
    try {
      return new RegExp(pattern).test(cookieName);
    } catch {
      return false;
    }
  }
  return cookieName === pattern;
}

function isBlocked(cookieName: string): boolean {
  for (const cat of categories) {
    // Required categories are always allowed.
    if (cat.required) continue;

    // If this category is consented, its cookies are allowed.
    if (currentConsent[cat.id] === true) continue;

    // Check if this cookie name belongs to a non-consented category.
    for (const cookie of cat.cookies) {
      if (matchesPattern(cookieName, cookie.name)) {
        return true;
      }
    }
  }

  return false;
}

export function installCookieInterceptor(
  bannerCategories: BannerCategory[],
  consent: Record<string, boolean>,
): void {
  categories = bannerCategories;
  currentConsent = { ...consent };

  if (installed) return;
  if (!nativeDescriptor?.set || !nativeDescriptor?.get) return;

  try {
    const nativeSet = nativeDescriptor.set;
    const nativeGet = nativeDescriptor.get;

    Object.defineProperty(document, "cookie", {
      configurable: true,
      enumerable: true,
      get() {
        return nativeGet.call(this);
      },
      set(value: string) {
        const name = parseCookieName(value);

        // Always allow our own consent cookie.
        if (name === PROBO_COOKIE) {
          nativeSet.call(this, value);
          return;
        }

        // Block cookies that belong to non-consented categories.
        if (isBlocked(name)) {
          return;
        }

        // Unknown cookies (not in any category) are allowed through.
        nativeSet.call(this, value);
      },
    });

    installed = true;
  } catch {
    // If we can't override (e.g. frozen prototype), silently give up.
    // Never break the host site.
  }
}

export function updateCookieInterceptor(
  consent: Record<string, boolean>,
): void {
  currentConsent = { ...consent };
}

export function uninstallCookieInterceptor(): void {
  if (!installed) return;
  if (!nativeDescriptor) return;

  try {
    Object.defineProperty(document, "cookie", nativeDescriptor);
    installed = false;
  } catch {
    // Never break the host site.
  }
}
