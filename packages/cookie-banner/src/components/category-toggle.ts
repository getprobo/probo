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
