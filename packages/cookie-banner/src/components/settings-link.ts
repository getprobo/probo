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
    this.root.setState(this.root.reopenState);
  };
}
