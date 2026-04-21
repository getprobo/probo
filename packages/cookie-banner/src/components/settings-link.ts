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

import type { ProboCookieBannerRoot } from "./cookie-banner-root";

export class ProboSettingsLink extends HTMLElement {
  private root: ProboCookieBannerRoot | null = null;

  connectedCallback(): void {
    this.root = this.findRoot();

    if (this.root) {
      this.attach(this.root);
    } else {
      document.addEventListener("probo-ready", this.onProboReady, { once: true });
    }

    this.addEventListener("click", this.handleClick);
  }

  disconnectedCallback(): void {
    this.removeEventListener("click", this.handleClick);
    document.removeEventListener("probo-ready", this.onProboReady);
    this.root = null;
  }

  private attach(root: ProboCookieBannerRoot): void {
    this.root = root;
    root.setAttribute("reopen-widget", "custom");
  }

  private findRoot(): ProboCookieBannerRoot | null {
    const direct = document.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot | null;
    if (direct) return direct;

    const themed = document.querySelector("probo-cookie-banner");
    if (themed?.shadowRoot) {
      return themed.shadowRoot.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot | null;
    }

    return null;
  }

  private onProboReady = (e: Event): void => {
    const root = (e as CustomEvent).target as ProboCookieBannerRoot | null;
    if (root?.tagName.toLowerCase() === "probo-cookie-banner-root") {
      this.attach(root);
      return;
    }

    const found = this.findRoot();
    if (found) {
      this.attach(found);
    }
  };

  private handleClick = (e: Event): void => {
    if (!this.root) return;
    e.preventDefault();
    this.root.setState("panel");
  };
}
