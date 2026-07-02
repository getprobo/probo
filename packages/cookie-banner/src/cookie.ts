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

import type { ConsentAction } from "./types";

export const COOKIE_NAME = "probo_consent";
const SECONDS_PER_DAY = 86400;

export interface ConsentCookie {
  bid: string;
  v: number;
  vid: string;
  action: ConsentAction;
  data: Record<string, boolean>;
}

export function getConsentCookie(): ConsentCookie | null {
  try {
    const prefix = `${COOKIE_NAME}=`;
    const entry = document.cookie
      .split("; ")
      .find((c) => c.startsWith(prefix));

    if (!entry) {
      return null;
    }

    return JSON.parse(decodeURIComponent(entry.substring(prefix.length)));
  } catch {
    return null;
  }
}

export function setConsentCookie(
  value: ConsentCookie,
  expiryDays: number,
): void {
  const maxAge = expiryDays * SECONDS_PER_DAY;
  const encoded = encodeURIComponent(JSON.stringify(value));

  document.cookie = `${COOKIE_NAME}=${encoded}; path=/; max-age=${maxAge}; SameSite=Lax`;
}

export function clearConsentCookie(): void {
  document.cookie = `${COOKIE_NAME}=; path=/; max-age=0; SameSite=Lax`;
}
