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

import { registerComponents } from "../components";
import type { ProboCookieBannerRoot } from "../components/cookie-banner-root";
import type { BannerConfig } from "../client";
import { THEMED_STYLES } from "./styles";

const CLOSE_ICON = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>`;
const CHEVRON_DOWN = `<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polyline points="6 9 12 15 18 9"/></svg>`;

export class ProboThemedBanner extends HTMLElement {
  private shadow: ShadowRoot;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: "open" });
  }

  static get observedAttributes(): string[] {
    return ["banner-id", "base-url", "reopen-widget"];
  }

  connectedCallback(): void {
    registerComponents();

    const bannerId = this.getAttribute("banner-id");
    const baseUrl = this.getAttribute("base-url");

    if (!bannerId || !baseUrl) {
      console.warn("[probo] <probo-cookie-banner> requires banner-id and base-url attributes");
      return;
    }

    const position = this.getAttribute("position") ?? "bottom-left";
    const reopenWidget = this.getAttribute("reopen-widget");
    const reopenAttr = reopenWidget ? ` reopen-widget="${this.esc(reopenWidget)}"` : "";

    this.shadow.innerHTML = `
      <style>${THEMED_STYLES}</style>
      <probo-cookie-banner-root banner-id="${this.esc(bannerId)}" base-url="${this.esc(baseUrl)}"${reopenAttr}>
        <probo-banner>
          <div class="floating" data-position="${this.esc(position)}">
            <div class="card" role="dialog" aria-modal="true" aria-labelledby="probo-banner-title" aria-describedby="probo-banner-desc">
              <p class="title" id="probo-banner-title">Cookie Preferences</p>
              <p class="description" id="probo-banner-desc" data-description>
                We use cookies to improve your experience and analyze site traffic.
              </p>
              <div class="buttons">
                <probo-accept-button><button class="btn btn-primary">Accept all</button></probo-accept-button>
                <probo-reject-button><button class="btn">Reject all</button></probo-reject-button>
                <probo-customize-button><button class="btn btn-link">Customize</button></probo-customize-button>
              </div>
            </div>
          </div>
        </probo-banner>

        <probo-preference-panel>
          <div class="floating" data-position="${this.esc(position)}">
            <div class="card" role="dialog" aria-modal="true" aria-labelledby="probo-panel-title">
              <div class="panel-header">
                <p class="title" id="probo-panel-title" style="margin:0">Preferences</p>
                <button class="panel-close" data-action="back" aria-label="Close">
                  ${CLOSE_ICON}
                </button>
              </div>
              <probo-category-list>
                <template>
                  <button class="cookie-toggle" data-action="toggle-cookies" aria-expanded="false" aria-label="Show cookie details">
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
                        <span class="cookie-detail"><span class="cookie-label">Description:</span> <span data-slot="description"></span></span>
                        <span class="cookie-detail"><span class="cookie-label">Duration:</span> <span data-slot="duration"></span></span>
                      </div>
                    </template>
                  </probo-cookie-list>
                </template>
              </probo-category-list>
              <div class="buttons">
                <probo-save-button>
                  <button class="btn btn-primary" style="flex:1">Save preferences</button>
                </probo-save-button>
              </div>
            </div>
          </div>
        </probo-preference-panel>

        <probo-settings-button position="${this.esc(position)}"></probo-settings-button>
      </probo-cookie-banner-root>
    `;

    const root = this.shadow.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot;

    root.addEventListener("probo-ready", (e: Event) => {
      const config = (e as CustomEvent).detail.config as BannerConfig;
      this.updateDescription(config);
    });

    this.shadow.querySelector("[data-action=back]")?.addEventListener("click", () => {
      root.setState(root.client.hasConsent ? "hidden" : "banner");
    });

    this.shadow.addEventListener("click", (e: Event) => {
      const btn = (e.target as Element).closest?.("[data-action=toggle-cookies]") as HTMLElement | null;
      if (!btn) return;
      const category = btn.closest("probo-category");
      const cookieList = category?.querySelector("probo-cookie-list") as HTMLElement | null;
      if (!cookieList) return;
      const open = cookieList.hasAttribute("hidden");
      if (open) {
        cookieList.removeAttribute("hidden");
        btn.classList.add("open");
      } else {
        cookieList.setAttribute("hidden", "");
        btn.classList.remove("open");
      }
      btn.setAttribute("aria-expanded", String(open));
      btn.setAttribute("aria-label", open ? "Hide cookie details" : "Show cookie details");
    });
  }

  private updateDescription(config: BannerConfig): void {
    const el = this.shadow.querySelector("[data-description]");
    if (!el) return;

    let html = "We use cookies to improve your experience and analyze site traffic.";
    if (config.privacy_policy_url) {
      html += ` <a href="${this.esc(config.privacy_policy_url)}" target="_blank" rel="noopener noreferrer">Privacy Policy</a>`;
    }
    el.innerHTML = html;
  }

  private esc(str: string): string {
    return str.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
