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

import { registerHeadlessComponents } from "../components";
import type { ProboCookieBannerRoot } from "../components/cookie-banner-root";
import type { BannerConfig } from "../types";
import { getGpcLabel, interpolate } from "../i18n";
import { BRANDING, CHEVRON_DOWN, CLOSE_ICON } from "../html";
import { THEMED_STYLES } from "./styles";

export class ProboThemedBanner extends HTMLElement {
  private shadow: ShadowRoot;
  private scrollLocked = false;
  private savedScrollY = 0;
  private prevHtmlOverflow = "";
  private prevBodyPosition = "";
  private prevBodyTop = "";
  private prevBodyLeft = "";
  private prevBodyRight = "";
  private prevBodyWidth = "";
  private prevBodyOverflow = "";
  private prevBodyPaddingRight = "";

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: "open" });
  }

  static get observedAttributes(): string[] {
    return ["banner-id", "base-url", "reopen-widget", "lang"];
  }

  connectedCallback(): void {
    registerHeadlessComponents();

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
                        <span class="cookie-detail cookie-type" data-label="label_type"><span data-slot="type"></span></span>
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
      const detail = (e as CustomEvent).detail;
      const config = detail.config as BannerConfig;
      this.applyTexts(config);
      if (!config.show_branding) {
        this.shadow.querySelectorAll("[data-branding]").forEach(el => {
          (el as HTMLElement).setAttribute("hidden", "");
        });
      }
      if (detail.gpcApplied) {
        const settingsBtn = this.shadow.querySelector("probo-settings-button");
        settingsBtn?.setAttribute("gpc-label", getGpcLabel(config.language));
      }
    });

    root.addEventListener("probo-consent", (e: Event) => {
      const { action } = (e as CustomEvent).detail;
      if (action !== "GPC") {
        const settingsBtn = this.shadow.querySelector("probo-settings-button");
        settingsBtn?.removeAttribute("gpc-label");
      }
    });

    root.addEventListener("probo-state", (e: Event) => {
      const { state } = (e as CustomEvent).detail;
      this.setScrollLock(state === "panel");
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

  disconnectedCallback(): void {
    this.setScrollLock(false);
  }

  private setScrollLock(locked: boolean): void {
    if (locked === this.scrollLocked) return;

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

    this.scrollLocked = locked;
  }

  private onScrollLockEvent = (e: Event): void => {
    // When the gesture happens inside the banner, only stop propagation so a
    // smooth-scroll library does not move the page; the browser still performs
    // the panel's native scroll (bounded by its overscroll-behavior: contain).
    // Anywhere else, cancel the gesture entirely.
    if (e.composedPath().includes(this)) {
      e.stopImmediatePropagation();
      return;
    }
    e.preventDefault();
    e.stopImmediatePropagation();
  };

  private applyTexts(config: BannerConfig): void {
    const texts = config.texts ?? {};

    const necessaryCategory = config.categories.find(c => c.kind === "NECESSARY");
    const necessaryCategoryName = necessaryCategory?.name ?? "Necessary";

    this.shadow.querySelectorAll("[data-text]").forEach(el => {
      const key = el.getAttribute("data-text")!;
      const raw = texts[key] ?? "";
      if (!raw) return;

      if (key === "banner_description") {
        let privacyLink = "";
        if (config.privacy_policy_url) {
          const linkText = this.esc(texts.privacy_policy_link_text ?? "Privacy Policy");
          privacyLink = `<a href="${this.esc(config.privacy_policy_url)}" target="_blank" rel="noopener noreferrer">${linkText}</a>`;
        }
        let cookieLink = "";
        if (config.cookie_policy_url) {
          const linkText = this.esc(texts.cookie_policy_link_text ?? "Cookie Policy");
          cookieLink = `<a href="${this.esc(config.cookie_policy_url)}" target="_blank" rel="noopener noreferrer">${linkText}</a>`;
        }
        const segments = raw.split("{{cookie_policy_link}}");
        const html = segments.map(seg =>
          seg.split("{{privacy_policy_link}}").map(p => this.esc(p)).join(privacyLink),
        ).join(cookieLink);
        el.innerHTML = html;
      } else if (key === "panel_description") {
        el.textContent = interpolate(raw, { necessary_category: necessaryCategoryName });
      } else {
        el.textContent = raw;
      }
    });

    this.shadow.querySelectorAll("[data-aria-text]").forEach(el => {
      const key = el.getAttribute("data-aria-text")!;
      const raw = texts[key] ?? el.getAttribute("aria-label") ?? "";
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
