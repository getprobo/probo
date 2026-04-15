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

const STORAGE_KEY_PREFIX = "probo_consent";

export function getOrCreateVisitorId(bannerId: string): string {
  const key = `${STORAGE_KEY_PREFIX}:${bannerId}:vid`;

  try {
    const stored = localStorage.getItem(key);
    if (stored) {
      return stored;
    }
  } catch {
    // localStorage unavailable
  }

  const array = new Uint8Array(16);
  crypto.getRandomValues(array);
  const id = Array.from(array, (b) => b.toString(16).padStart(2, "0")).join("");

  try {
    localStorage.setItem(key, id);
  } catch {
    // localStorage unavailable
  }

  return id;
}
