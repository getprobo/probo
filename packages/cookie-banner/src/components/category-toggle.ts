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
import type { ProboCategory } from "./category";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

export class ProboCategoryToggle extends ProboElement {
  private root: ProboRootElement | null = null;
  private category: ProboCategory | null = null;
  private checkbox: HTMLInputElement | null = null;

  connectedCallback(): void {
    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");
    this.category = this.findAncestor<ProboCategory>("probo-category");

    this.scheduleValidation(() => this.setup());
  }

  disconnectedCallback(): void {
    if (this.checkbox) {
      this.checkbox.removeEventListener("change", this.handleChange);
    }
  }

  private setup(): void {
    this.checkbox = this.querySelector<HTMLInputElement>("input[type=checkbox]");

    if (!this.checkbox) {
      const input = document.createElement("input");
      input.type = "checkbox";
      input.part.add("toggle");
      this.appendChild(input);
      this.checkbox = input;
    }

    if (!this.category || !this.root) return;

    const name = this.category.categoryName;
    const slug = this.category.categorySlug;
    this.checkbox.setAttribute("aria-label", name);
    const isRequired = this.category.kind === "NECESSARY";

    if (isRequired) {
      this.checkbox.checked = true;
      this.checkbox.disabled = true;
      return;
    }

    const draft = this.root.consentDraft;
    this.checkbox.checked = !!draft[slug];
    this.checkbox.addEventListener("change", this.handleChange);

    if (this.root) {
      this.root.addEventListener("probo-state", (e: Event) => {
        const { state } = (e as CustomEvent).detail;
        if (state === "panel" && this.checkbox && this.category && this.root) {
          this.checkbox.checked = !!this.root.consentDraft[this.category.categorySlug];
        }
      });
    }
  }

  private handleChange = (): void => {
    if (!this.checkbox || !this.category || !this.root) return;
    this.root.updateDraft(this.category.categorySlug, this.checkbox.checked);
  };
}
