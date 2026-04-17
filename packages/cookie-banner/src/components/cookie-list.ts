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

import type { CookieItem } from "../client";
import { ProboElement } from "./base";
import type { ProboCategory } from "./category";

export class ProboCookieList extends ProboElement {
  private category: ProboCategory | null = null;
  private template: HTMLTemplateElement | null = null;

  connectedCallback(): void {
    this.template = this.querySelector("template");
    if (!this.template) {
      this.warn("<probo-cookie-list> requires a <template> child");
      return;
    }

    this.category = this.findAncestor<ProboCategory>("probo-category");

    this.scheduleValidation(() => this.stamp());
  }

  private stamp(): void {
    if (!this.template || !this.category) return;

    const cookies = this.category.cookies;
    for (const cookie of cookies) {
      this.stampCookie(cookie);
    }
  }

  private stampCookie(cookie: CookieItem): void {
    if (!this.template) return;

    const wrapper = document.createElement("probo-cookie");
    wrapper.setAttribute("name", cookie.name);
    const clone = this.template.content.cloneNode(true) as DocumentFragment;
    this.fillSlots(clone, {
      name: cookie.name,
      duration: cookie.duration,
      description: cookie.description,
    });

    wrapper.appendChild(clone);
    this.appendChild(wrapper);
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

export class ProboCookie extends ProboElement {}
