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

import { ProboElement } from "./base";
import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

const REQUIRED_BUTTONS = [
  "probo-accept-button",
  "probo-reject-button",
  "probo-customize-button",
] as const;

export class ProboBanner extends ProboElement {
  private root: ProboRootElement | null = null;
  private onStateChange = (e: Event): void => {
    const { state } = (e as CustomEvent).detail;
    this.hidden = state !== "banner";
  };

  connectedCallback(): void {
    this.shadow.innerHTML = `
      <style>
        :host { display: block; }
        :host([hidden]) { display: none; }
      </style>
      <slot></slot>
    `;

    this.hidden = true;
    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");

    if (this.root) {
      this.root.addEventListener("probo-state", this.onStateChange);
      if (this.root.state === "banner") {
        this.hidden = false;
      }
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
    for (const tag of REQUIRED_BUTTONS) {
      if (!this.querySelector(tag)) {
        missing.push(tag);
      }
    }
    if (missing.length > 0) {
      this.warn(`<probo-banner> is missing required children: ${missing.join(", ")}`);
      this.emitValidation(missing);
    }
  }
}
