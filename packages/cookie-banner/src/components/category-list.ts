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

import type { Category } from "../types";
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
      wrapper.setAttribute("slug", cat.slug);
      wrapper.setAttribute("kind", cat.kind);
      wrapper.setAttribute("description", cat.description);
      wrapper.setAttribute("cookies", JSON.stringify(cat.cookies));

      const clone = this.template.content.cloneNode(true) as DocumentFragment;
      this.fillSlots(clone, {
        name: cat.name,
        description: cat.description,
      });

      const hasCookies = cat.cookies && cat.cookies.length > 0;
      if (!hasCookies) {
        clone.querySelector("[data-action=toggle-cookies]")?.remove();
        clone.querySelector("probo-cookie-list")?.remove();
      }

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
