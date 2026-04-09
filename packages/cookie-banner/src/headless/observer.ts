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

import { applyConsent } from "./apply";

let observer: MutationObserver | null = null;
let currentConsent: Record<string, boolean> = {};

function handleMutations(mutations: MutationRecord[]): void {
  let hasNewElements = false;

  for (const mutation of mutations) {
    for (const node of Array.from(mutation.addedNodes)) {
      if (node.nodeType !== Node.ELEMENT_NODE) continue;
      const el = node as HTMLElement;

      if (el.hasAttribute("data-cookie-category")) {
        hasNewElements = true;
        break;
      }

      if (el.querySelector("[data-cookie-category]")) {
        hasNewElements = true;
        break;
      }
    }

    if (hasNewElements) break;
  }

  if (hasNewElements) {
    applyConsent(currentConsent);
  }
}

export function startObserver(consent: Record<string, boolean>): void {
  currentConsent = { ...consent };

  if (observer) return;

  try {
    observer = new MutationObserver(handleMutations);
    observer.observe(document.body, {
      childList: true,
      subtree: true,
    });
  } catch {
    // Never break the host site.
  }
}

export function updateObserverConsent(
  consent: Record<string, boolean>,
): void {
  currentConsent = { ...consent };
}

export function stopObserver(): void {
  if (observer) {
    observer.disconnect();
    observer = null;
  }
}
