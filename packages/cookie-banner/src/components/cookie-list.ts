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

import type { CookieItem } from "../types";
import { humanizeDuration } from "../cookie-utils";
import { getCookieDetailLabels, getTrackerTypeLabel, interpolate } from "../i18n";
import { ProboElement } from "./base";
import type { ProboCategory } from "./category";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

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

    const root = this.findAncestor<ProboCookieBannerRoot>("probo-cookie-banner-root");
    const lang = root?.bannerConfig?.language ?? "en";
    const labels = getCookieDetailLabels(lang);

    const cookies = this.category.cookies;
    for (const cookie of cookies) {
      this.stampCookie(cookie, labels, lang);
    }
  }

  private stampCookie(cookie: CookieItem, labels: Record<string, string>, lang: string): void {
    if (!this.template) return;

    const duration = cookie.max_age_seconds != null
      ? humanizeDuration(cookie.max_age_seconds, lang, cookie.tracker_type)
      : humanizeDuration(0, lang, cookie.tracker_type);

    const type = getTrackerTypeLabel(cookie.tracker_type);

    const wrapper = document.createElement("probo-cookie");
    wrapper.setAttribute("name", cookie.name);
    const clone = this.template.content.cloneNode(true) as DocumentFragment;
    this.fillSlots(clone, {
      name: cookie.name,
      type,
      duration,
      description: cookie.description,
    });
    this.fillLabels(clone, labels, {
      type,
      description: cookie.description,
      duration,
    });

    const typeEl = clone.querySelector<HTMLElement>(".cookie-type");
    if (typeEl && labels.label_type) {
      typeEl.setAttribute("aria-label", interpolate(labels.label_type, { value: type }));
    }

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

  private fillLabels(
    fragment: DocumentFragment,
    labels: Record<string, string>,
    values: Record<string, string>,
  ): void {
    for (const el of fragment.querySelectorAll("[data-label]")) {
      const key = el.getAttribute("data-label")!;
      const slotName = key.replace("label_", "");
      const value = values[slotName] ?? "";
      const tpl = labels[key];
      if (tpl) {
        const slot = el.querySelector(`[data-slot="${slotName}"]`);
        const parts = tpl.split("{{value}}");
        el.textContent = parts[0] ?? "";
        if (slot) {
          slot.textContent = value;
          el.appendChild(slot);
        }
      }
    }
  }
}

export class ProboCookie extends ProboElement {}
