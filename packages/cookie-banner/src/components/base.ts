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

import type { CookieBannerClient, BannerConfig } from "../client";

export type ProboState = "loading" | "banner" | "panel" | "hidden";

export interface ConsentDraft {
  [category: string]: boolean;
}

const FOCUSABLE = 'a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])';

export class ProboElement extends HTMLElement {
  protected focusFirst(): void {
    requestAnimationFrame(() => {
      const el = this.querySelector<HTMLElement>(FOCUSABLE);
      el?.focus();
    });
  }

  protected findAncestor<T extends HTMLElement>(tagName: string): T | null {
    let el: HTMLElement | null = this.parentElement;
    while (el) {
      if (el.tagName.toLowerCase() === tagName) {
        return el as T;
      }
      el = el.parentElement;
    }
    return null;
  }

  protected scheduleValidation(fn: () => void): void {
    queueMicrotask(fn);
  }

  protected warn(message: string): void {
    console.warn(`[probo] ${message}`);
  }

  protected emitValidation(missing: string[]): void {
    this.dispatchEvent(
      new CustomEvent("probo-validation", {
        bubbles: true,
        composed: true,
        detail: { missing },
      }),
    );
  }
}

export interface ProboRootElement extends ProboElement {
  readonly client: CookieBannerClient;
  readonly bannerConfig: BannerConfig;
  readonly state: ProboState;
  readonly reopenWidget: string;
  readonly consentDraft: ConsentDraft;
  setState(state: ProboState): void;
  updateDraft(category: string, value: boolean): void;
}
