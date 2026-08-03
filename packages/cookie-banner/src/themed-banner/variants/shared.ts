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

import { interpolate } from "../../i18n";
import { CHEVRON_DOWN } from "../../html";
import type { BannerConfig } from "../../types";

export function esc(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

// floatingCard wraps a variant's inner markup in the positioned dialog shell
// shared by every presentation.
export function floatingCard(
  position: string,
  aria: { labelledby: string; describedby: string },
  inner: string,
): string {
  return `
    <div class="floating" data-position="${esc(position)}">
      <div class="card" role="dialog" aria-modal="true" aria-labelledby="${aria.labelledby}" aria-describedby="${aria.describedby}">
        ${inner}
      </div>
    </div>`;
}

// CATEGORY_LIST is the preference-panel category/cookie template used by
// presentations that expose granular controls (currently opt-in only).
export const CATEGORY_LIST = `
  <probo-category-list>
    <template>
      <button class="cookie-toggle" data-action="toggle-cookies" aria-expanded="false" data-aria-text="aria_show_details">
        ${CHEVRON_DOWN}
      </button>
      <div class="category-header">
        <div class="category-info">
          <div class="category-name" data-slot="name"></div>
          <div class="category-description" data-slot="description"></div>
        </div>
        <probo-category-toggle>
          <label class="toggle">
            <input type="checkbox">
            <span class="toggle-track"></span>
          </label>
        </probo-category-toggle>
      </div>
      <probo-cookie-list hidden>
        <template>
          <div class="cookie-item">
            <span class="cookie-name" data-slot="name"></span>
            <span class="cookie-detail cookie-type" data-label="label_type"><span data-slot="type"></span></span>
            <span class="cookie-detail" data-label="label_description"><span data-slot="description"></span></span>
            <span class="cookie-detail" data-label="label_duration"><span data-slot="duration"></span></span>
          </div>
        </template>
      </probo-cookie-list>
    </template>
  </probo-category-list>`;

// applyTexts fills every [data-text] / [data-aria-text] node in the tree from
// the config. It is variant-agnostic: link interpolation and category
// interpolation are triggered by the placeholders present in the value, not by
// hardcoded key names, so each renderer is free to point at its own keys.
export function applyTexts(root: ParentNode, config: BannerConfig): void {
  const texts = config.texts ?? {};

  const necessaryCategory = config.categories.find(c => c.kind === "NECESSARY");
  const necessaryCategoryName = necessaryCategory?.name ?? "Necessary";

  root.querySelectorAll("[data-text]").forEach(el => {
    const key = el.getAttribute("data-text")!;
    const raw = texts[key] ?? "";
    if (!raw) return;

    if (raw.includes("{{cookie_policy_link}}") || raw.includes("{{privacy_policy_link}}")) {
      el.innerHTML = renderPolicyLinks(raw, config, texts);
    } else if (raw.includes("{{necessary_category}}")) {
      el.textContent = interpolate(raw, { necessary_category: necessaryCategoryName });
    } else {
      el.textContent = raw;
    }
  });

  root.querySelectorAll("[data-aria-text]").forEach(el => {
    const key = el.getAttribute("data-aria-text")!;
    const raw = texts[key] ?? el.getAttribute("aria-label") ?? "";
    if (raw) el.setAttribute("aria-label", raw);
  });
}

function renderPolicyLinks(
  raw: string,
  config: BannerConfig,
  texts: Record<string, string>,
): string {
  let privacyLink = "";
  if (config.privacy_policy_url) {
    const linkText = esc(texts.privacy_policy_link_text ?? "Privacy Policy");
    privacyLink = `<a href="${esc(config.privacy_policy_url)}" target="_blank" rel="noopener noreferrer">${linkText}</a>`;
  }

  let cookieLink = "";
  if (config.cookie_policy_url) {
    const linkText = esc(texts.cookie_policy_link_text ?? "Cookie Policy");
    cookieLink = `<a href="${esc(config.cookie_policy_url)}" target="_blank" rel="noopener noreferrer">${linkText}</a>`;
  }

  return raw
    .split("{{cookie_policy_link}}")
    .map(seg =>
      seg
        .split("{{privacy_policy_link}}")
        .map(p => esc(p))
        .join(privacyLink),
    )
    .join(cookieLink);
}

// ScrollLock pins the page while a full-height panel is open. It is only used by
// presentations that open a preference panel.
export class ScrollLock {
  private locked = false;
  private savedScrollY = 0;
  private prevHtmlOverflow = "";
  private prevBodyPosition = "";
  private prevBodyTop = "";
  private prevBodyLeft = "";
  private prevBodyRight = "";
  private prevBodyWidth = "";
  private prevBodyOverflow = "";
  private prevBodyPaddingRight = "";

  constructor(private readonly host: HTMLElement) {}

  set(locked: boolean): void {
    if (locked === this.locked) return;

    const html = document.documentElement;
    const body = document.body;

    if (locked) {
      this.savedScrollY = window.scrollY;
      const scrollbarWidth = window.innerWidth - html.clientWidth;

      this.prevHtmlOverflow = html.style.overflow;
      this.prevBodyPosition = body.style.position;
      this.prevBodyTop = body.style.top;
      this.prevBodyLeft = body.style.left;
      this.prevBodyRight = body.style.right;
      this.prevBodyWidth = body.style.width;
      this.prevBodyOverflow = body.style.overflow;
      this.prevBodyPaddingRight = body.style.paddingRight;

      // Pin the body in place. Unlike `overflow: hidden`, this removes the
      // viewport's scroll distance entirely, so JS-driven smooth-scroll
      // libraries (Lenis, Locomotive, ...) have nothing left to animate even
      // when they intercept wheel/touch in the capture phase before us.
      html.style.overflow = "hidden";
      body.style.position = "fixed";
      body.style.top = `-${this.savedScrollY}px`;
      body.style.left = "0";
      body.style.right = "0";
      body.style.width = "100%";
      body.style.overflow = "hidden";
      if (scrollbarWidth > 0) {
        body.style.paddingRight = `${scrollbarWidth}px`;
      }

      // Defense in depth for the common case where the body is not the
      // scroller (e.g. an inner scroll container): capture-phase listeners run
      // before bubble-phase smooth-scroll handlers, cancelling page scroll
      // while still letting the panel's own scrollable region scroll.
      window.addEventListener("wheel", this.onScrollLockEvent, { capture: true, passive: false });
      window.addEventListener("touchmove", this.onScrollLockEvent, { capture: true, passive: false });
    } else {
      html.style.overflow = this.prevHtmlOverflow;
      body.style.position = this.prevBodyPosition;
      body.style.top = this.prevBodyTop;
      body.style.left = this.prevBodyLeft;
      body.style.right = this.prevBodyRight;
      body.style.width = this.prevBodyWidth;
      body.style.overflow = this.prevBodyOverflow;
      body.style.paddingRight = this.prevBodyPaddingRight;

      window.removeEventListener("wheel", this.onScrollLockEvent, { capture: true });
      window.removeEventListener("touchmove", this.onScrollLockEvent, { capture: true });

      // Restore the scroll position the body pinning discarded.
      window.scrollTo(0, this.savedScrollY);
    }

    this.locked = locked;
  }

  private onScrollLockEvent = (e: Event): void => {
    // When the gesture happens inside the banner, only stop propagation so a
    // smooth-scroll library does not move the page; the browser still performs
    // the panel's native scroll (bounded by its overscroll-behavior: contain).
    // Anywhere else, cancel the gesture entirely.
    if (e.composedPath().includes(this.host)) {
      e.stopImmediatePropagation();
      return;
    }
    e.preventDefault();
    e.stopImmediatePropagation();
  };
}
