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
import { interpolate } from "../i18n";
import { BRANDING, CHEVRON_DOWN, CLOSE_ICON } from "../html";
import { THEMED_STYLES } from "./styles";

export class ProboThemedBanner extends HTMLElement {
  private shadow: ShadowRoot;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: "open" });
  }

  static get observedAttributes(): string[] {
    return ["banner-id", "base-url", "reopen-widget", "lang"];
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
    const lang = this.getAttribute("lang");
    const langAttr = lang ? ` lang="${this.esc(lang)}"` : "";

    this.shadow.innerHTML = `
      <style>${THEMED_STYLES}</style>
      <probo-cookie-banner-root banner-id="${this.esc(bannerId)}" base-url="${this.esc(baseUrl)}"${reopenAttr}${langAttr}>
        <probo-banner>
          <div class="floating" data-position="${this.esc(position)}">
            <div class="card" role="dialog" aria-modal="true" aria-labelledby="probo-banner-title" aria-describedby="probo-banner-desc">
              <p class="title" id="probo-banner-title" data-text="banner_title"></p>
              <p class="description" id="probo-banner-desc" data-text="banner_description"></p>
              <div class="buttons">
                <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
                <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
                <probo-customize-button><button class="btn btn-link" data-text="button_customize"></button></probo-customize-button>
              </div>
              ${BRANDING}
            </div>
          </div>
        </probo-banner>

        <probo-preference-panel>
          <div class="floating" data-position="${this.esc(position)}">
            <div class="card" role="dialog" aria-modal="true" aria-labelledby="probo-panel-title" aria-describedby="probo-panel-desc">
              <div class="panel-header">
                <div class="panel-header-title">
                  <p class="title" id="probo-panel-title" style="margin:0" data-text="panel_title"></p>
                  <button class="panel-close" data-action="back" data-aria-text="aria_close">
                    ${CLOSE_ICON}
                  </button>
                </div>
                <p class="description" id="probo-panel-desc" data-text="panel_description"></p>
              </div>
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
                        <span class="cookie-detail" data-label="label_description"><span data-slot="description"></span></span>
                        <span class="cookie-detail" data-label="label_duration"><span data-slot="duration"></span></span>
                      </div>
                    </template>
                  </probo-cookie-list>
                </template>
              </probo-category-list>
              <div class="footer">
                <div class="buttons">
                  <probo-accept-button><button class="btn btn-primary" data-text="button_accept_all"></button></probo-accept-button>
                  <probo-reject-button><button class="btn" data-text="button_reject_all"></button></probo-reject-button>
                  <probo-save-button>
                    <button class="btn btn-link" style="flex:1" data-text="button_save"></button>
                  </probo-save-button>
                </div>
                ${BRANDING}
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
      this.applyTexts(config);
      if (!config.show_branding) {
        this.shadow.querySelectorAll("[data-branding]").forEach(el => {
          (el as HTMLElement).setAttribute("hidden", "");
        });
      }
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
      const texts = root.bannerConfig?.texts;
      const showLabel = texts?.aria_show_details ?? "Show cookie details";
      const hideLabel = texts?.aria_hide_details ?? "Hide cookie details";
      btn.setAttribute("aria-label", open ? hideLabel : showLabel);
    });
  }

  private applyTexts(config: BannerConfig): void {
    const texts = config.texts ?? {};

    const necessaryCategory = config.categories.find(c => c.kind === "NECESSARY");
    const necessaryCategoryName = necessaryCategory?.name ?? "Necessary";

    this.shadow.querySelectorAll("[data-text]").forEach(el => {
      const key = el.getAttribute("data-text")!;
      const raw = texts[key];
      if (!raw) return;

      if (key === "banner_description") {
        let link = "";
        if (config.privacy_policy_url) {
          const linkText = this.esc(texts.privacy_policy_link_text ?? "Privacy Policy");
          link = `<a href="${this.esc(config.privacy_policy_url)}" target="_blank" rel="noopener noreferrer">${linkText}</a>`;
        }
        const parts = raw.split("{{privacy_policy_link}}");
        el.innerHTML = parts.map(p => this.esc(p)).join(link);
      } else if (key === "panel_description") {
        el.textContent = interpolate(raw, { necessary_category: necessaryCategoryName });
      } else {
        el.textContent = raw;
      }
    });

    this.shadow.querySelectorAll("[data-aria-text]").forEach(el => {
      const key = el.getAttribute("data-aria-text")!;
      const raw = texts[key];
      if (raw) el.setAttribute("aria-label", raw);
    });

    const settingsBtn = this.shadow.querySelector("probo-settings-button");
    if (settingsBtn) {
      const ariaText = texts.aria_cookie_settings;
      if (ariaText) {
        settingsBtn.setAttribute("aria-settings-label", ariaText);
      }
    }
  }

  private esc(str: string): string {
    return str.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }
}
