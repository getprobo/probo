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

import type { ProboRootElement } from "./base";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

const COOKIE_ICON = `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><circle cx="8" cy="9" r="1" fill="currentColor"/><circle cx="15" cy="11" r="1" fill="currentColor"/><circle cx="10" cy="15" r="1" fill="currentColor"/><circle cx="13" cy="7" r="1" fill="currentColor"/></svg>`;

export class ProboSettingsButton extends HTMLElement {
  private shadow: ShadowRoot;
  private root: ProboRootElement | null = null;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: "open" });
  }

  static get observedAttributes(): string[] {
    return ["position"];
  }

  private get position(): string {
    return this.getAttribute("position") ?? "bottom-left";
  }

  connectedCallback(): void {
    const pos = this.position;
    const isRight = pos === "bottom-right";

    this.shadow.innerHTML = `
      <style>
        :host {
          position: fixed;
          bottom: 16px;
          ${isRight ? "right" : "left"}: var(--probo-settings-offset, 16px);
          z-index: var(--probo-z-index, 2147483646);
        }
        :host([hidden]) { display: none; }
        button {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 6px;
          border: none;
          border-radius: var(--probo-settings-radius, 9999px);
          background: var(--probo-settings-bg, var(--probo-accent, #1a1a1a));
          color: var(--probo-settings-color, var(--probo-accent-text, #ffffff));
          padding: var(--probo-settings-padding, 10px);
          cursor: pointer;
          font-family: inherit;
          font-size: var(--probo-settings-font-size, 14px);
          box-shadow: var(--probo-settings-shadow, 0 2px 8px rgba(0, 0, 0, 0.15));
          transition: opacity 0.2s;
        }
        button:hover { opacity: 0.85; }
        .icon { display: flex; flex-shrink: 0; }
        ::slotted(*) { display: contents; }
      </style>
      <button part="button">
        <span class="icon" part="icon">${COOKIE_ICON}</span>
        <slot></slot>
      </button>
    `;

    this.hidden = true;
    this.root = this.findRoot();

    if (this.root) {
      if (this.root.reopenWidget === "custom") {
        return;
      }

      this.root.addEventListener("probo-state", this.onStateChange);
      this.root.addEventListener("probo-reopen-widget", this.onReopenWidgetChange);
      if (this.root.state === "hidden") {
        this.hidden = false;
      }
    }

    const btn = this.shadow.querySelector("button");
    btn?.addEventListener("click", this.handleClick);
  }

  disconnectedCallback(): void {
    if (this.root) {
      this.root.removeEventListener("probo-state", this.onStateChange);
      this.root.removeEventListener("probo-reopen-widget", this.onReopenWidgetChange);
    }
  }

  private findRoot(): ProboCookieBannerRoot | null {
    let el: HTMLElement | null = this.parentElement;
    while (el) {
      if (el.tagName.toLowerCase() === "probo-cookie-banner-root") {
        return el as ProboCookieBannerRoot;
      }
      el = el.parentElement;
    }
    return null;
  }

  private onStateChange = (e: Event): void => {
    const { state } = (e as CustomEvent).detail;
    this.hidden = state !== "hidden";
  };

  private onReopenWidgetChange = (e: Event): void => {
    const { value } = (e as CustomEvent).detail;
    if (value === "custom") {
      this.hidden = true;
      this.root?.removeEventListener("probo-state", this.onStateChange);
      this.root?.removeEventListener("probo-reopen-widget", this.onReopenWidgetChange);
    }
  };

  private handleClick = (): void => {
    if (!this.root) return;
    this.root.setState("panel");
  };
}
