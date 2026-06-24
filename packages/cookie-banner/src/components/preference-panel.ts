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

import { ProboElement } from "./base";
import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

const REQUIRED_CHILDREN = [
  "probo-category-list",
  "probo-save-button",
] as const;

export class ProboPreferencePanel extends ProboElement {
  private root: ProboRootElement | null = null;
  private onStateChange = (e: Event): void => {
    const { state } = (e as CustomEvent).detail;
    this.hidden = state !== "panel";
    if (state === "panel") {
      (this.root as ProboCookieBannerRoot).resetDraft();
      this.focusFirst();
    }
  };

  connectedCallback(): void {
    this.hidden = true;
    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");

    if (this.root) {
      this.root.addEventListener("probo-state", this.onStateChange);
    }

    this.scheduleValidation(() => this.validate());
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.removeEventListener("probo-state", this.onStateChange);
    }
  }

  private validate(): void {
    const missing: string[] = [];
    for (const tag of REQUIRED_CHILDREN) {
      if (!this.querySelector(tag)) {
        missing.push(tag);
      }
    }
    if (missing.length > 0) {
      this.warn(`<probo-preference-panel> is missing required children: ${missing.join(", ")}`);
      this.emitValidation(missing);
    }
  }
}

export class ProboSaveButton extends ProboElement {
  private root: ProboRootElement | null = null;

  connectedCallback(): void {
    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");
    this.addEventListener("click", this.handleClick);
  }

  disconnectedCallback(): void {
    this.removeEventListener("click", this.handleClick);
  }

  private handleClick = (): void => {
    if (!this.root) return;
    const draft = { ...this.root.consentDraft };
    this.root.client.customize(draft);
    this.root.setState("hidden");
    this.root.dispatchEvent(
      new CustomEvent("probo-consent", {
        bubbles: true,
        composed: true,
        detail: { action: "CUSTOMIZE", consent_data: draft },
      }),
    );
  };
}
