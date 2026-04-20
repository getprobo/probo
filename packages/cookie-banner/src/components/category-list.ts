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

import type { Category } from "../client";
import { ProboElement } from "./base";
import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

export class ProboCategoryList extends ProboElement {
  private root: ProboRootElement | null = null;
  private template: HTMLTemplateElement | null = null;
  private onReady = (e: Event): void => {
    const { config } = (e as CustomEvent).detail;
    this.stamp(config.categories);
  };

  connectedCallback(): void {
    this.template = this.querySelector("template");
    if (!this.template) {
      this.warn("<probo-category-list> requires a <template> child");
      return;
    }

    this.root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");
    if (!this.root) return;

    this.validateTemplate();

    try {
      const config = this.root.bannerConfig;
      this.stamp(config.categories);
    } catch {
      this.root.addEventListener("probo-ready", this.onReady, { once: true });
    }
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.removeEventListener("probo-ready", this.onReady);
    }
  }

  private stamp(categories: Category[]): void {
    if (!this.template) return;

    for (const cat of categories) {
      const wrapper = document.createElement("probo-category");
      wrapper.setAttribute("name", cat.name);
      wrapper.setAttribute("kind", cat.kind);
      wrapper.setAttribute("description", cat.description);
      wrapper.setAttribute("cookies", JSON.stringify(cat.cookies));

      const clone = this.template.content.cloneNode(true) as DocumentFragment;
      this.fillSlots(clone, {
        name: cat.name,
        description: cat.description,
      });

      wrapper.appendChild(clone);
      this.appendChild(wrapper);
    }
  }

  private validateTemplate(): void {
    if (!this.template) return;
    const content = this.template.content;
    const missing: string[] = [];
    if (!content.querySelector("probo-category-toggle")) {
      missing.push("probo-category-toggle");
    }
    if (!content.querySelector("probo-cookie-list")) {
      missing.push("probo-cookie-list");
    }
    if (missing.length > 0) {
      this.warn(`<probo-category-list> template is missing required elements: ${missing.join(", ")}`);
      this.emitValidation(missing);
    }
  }

  private fillSlots(
    fragment: DocumentFragment,
    data: Record<string, string>,
  ): void {
    for (const [key, value] of Object.entries(data)) {
      const els = fragment.querySelectorAll(`[data-slot="${key}"]`);
      for (const el of els) {
        el.textContent = value;
      }
    }
  }
}
