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

function deleteCookie(name: string): void {
  const hostname = window.location.hostname;
  const variations = [
    `${name}=; max-age=-1; path=/`,
    `${name}=; max-age=-1; path=/; domain=${hostname}`,
    `${name}=; max-age=-1; path=/; domain=.${hostname}`,
  ];
  for (const cookie of variations) {
    document.cookie = cookie;
  }
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

function getAllCookieNames(): string[] {
  if (!document.cookie) return [];
  return document.cookie
    .split(";")
    .map((c) => c.trim().split("=")[0])
    .filter((n) => n.length > 0);
}

export function cleanupCookies(
  categories: BannerCategory[],
  previousConsent: Record<string, boolean>,
  newConsent: Record<string, boolean>,
): void {
  try {
    const cookieNames = getAllCookieNames();
    for (const category of categories) {
      if (!previousConsent[category.id] || newConsent[category.id]) continue;
      if (!category.cookies || category.cookies.length === 0) continue;

      for (const cookie of category.cookies) {
        for (const name of cookieNames) {
          if (matchesPattern(name, cookie.name)) {
            deleteCookie(name);
          }
        }
      }
    }
  } catch {
    // Never break the host site due to cookie cleanup errors.
  }
}
