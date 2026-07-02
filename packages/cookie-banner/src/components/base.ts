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

import type { CookieBannerClient } from "../client";
import type { BannerConfig, Regulation } from "../types";

export type ProboState = "loading" | "banner" | "panel" | "hidden";

export interface ConsentDraft {
  [category: string]: boolean;
}

const FOCUSABLE = 'a[href],button:not([disabled]),input:not([disabled]),select:not([disabled]),textarea:not([disabled]),[tabindex]:not([tabindex="-1"])';

export class ProboElement extends HTMLElement {
  protected focusFirst(): void {
    requestAnimationFrame(() => {
      const el = this.querySelector<HTMLElement>(FOCUSABLE);
      el?.focus({ preventScroll: true });
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
  readonly gpcApplied: boolean;
  readonly regulation: Regulation | null;
  readonly consentMode: "OPT_IN" | "OPT_OUT" | null;
  readonly reopenState: ProboState;
  setState(state: ProboState): void;
  updateDraft(category: string, value: boolean): void;
}
