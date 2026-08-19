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
import { resolveLayout } from "../layout";
import type { BannerConfig, Presentation } from "../types";
import { THEMED_STYLES } from "./styles";
import { applyTexts, esc, ScrollLock } from "./variants/shared";
import { renderNotice } from "./variants/notice";
import { renderOptIn } from "./variants/optin";
import { renderOptOut } from "./variants/optout";

// ProboThemedBanner is a thin dispatcher: it owns the shadow root and the data
// root, then mounts the renderer matching the visitor's presentation once the
// config is known. Each presentation's markup, wording, and behavior live in
// its own renderer, so this component holds no per-variant branching beyond the
// initial dispatch.
export class ProboThemedBanner extends HTMLElement {
  private shadow: ShadowRoot;
  private scrollLock: ScrollLock;

  constructor() {
    super();
    this.shadow = this.attachShadow({ mode: "open" });
    this.scrollLock = new ScrollLock(this);
  }

  static get observedAttributes(): string[] {
    return ["banner-id", "base-url", "lang"];
  }

  connectedCallback(): void {
    registerHeadlessComponents();

    const bannerId = this.getAttribute("banner-id");
    const baseUrl = this.getAttribute("base-url");

    if (!bannerId || !baseUrl) {
      console.warn("[probo] <probo-cookie-banner> requires banner-id and base-url attributes");
      return;
    }

    const lang = this.getAttribute("lang");
    const langAttr = lang ? ` lang="${esc(lang)}"` : "";

    const gcm = this.getAttribute("gcm-enabled");
    const gcmAttr = gcm ? ` gcm-enabled="${esc(gcm)}"` : "";

    this.shadow.innerHTML = `
      <style>${THEMED_STYLES}</style>
      <probo-cookie-banner-root banner-id="${esc(bannerId)}" base-url="${esc(baseUrl)}"${langAttr}${gcmAttr}></probo-cookie-banner-root>
    `;

    const root = this.shadow.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot;

    root.addEventListener(
      "probo-ready",
      (e: Event) => {
        const config = (e as CustomEvent).detail.config as BannerConfig;
        this.mount(root, config);
      },
      { once: true },
    );
  }

  disconnectedCallback(): void {
    this.scrollLock.set(false);
  }

  private mount(root: ProboCookieBannerRoot, config: BannerConfig): void {
    const layout = resolveLayout(config);
    const position = this.getAttribute("position") ?? "bottom-left";

    root.innerHTML = renderFor(layout.presentation, config, position);

    applyTexts(this.shadow, config);

    if (!config.show_branding) {
      this.shadow.querySelectorAll("[data-branding]").forEach(el => {
        (el as HTMLElement).setAttribute("hidden", "");
      });
    }

    // Opt-in opens a preference panel; CCPA opt-out reopens Privacy Choices.
    // Both need scroll-lock while their larger surface is open.
    if (layout.presentation === "OPT_IN") {
      this.wirePanel(root);
    }
    if (layout.presentation === "OPT_OUT") {
      this.wirePrivacyChoices(root);
    }
  }

  private wirePanel(root: ProboCookieBannerRoot): void {
    root.addEventListener("probo-state", (e: Event) => {
      const { state } = (e as CustomEvent).detail;
      this.scrollLock.set(state === "panel");
    });

    this.shadow.querySelector("[data-action=back]")?.addEventListener("click", () => {
      root.setState(
        root.client.hasConsent ? "hidden" : (root.layout?.initial_state ?? "banner"),
      );
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

  private wirePrivacyChoices(root: ProboCookieBannerRoot): void {
    root.addEventListener("probo-state", (e: Event) => {
      const { state } = (e as CustomEvent).detail;
      this.scrollLock.set(state === "privacy_choices");
    });

    this.shadow
      .querySelector("[data-action=close-privacy-choices]")
      ?.addEventListener("click", () => {
        root.setState("hidden");
      });
  }
}

function renderFor(
  presentation: Presentation,
  config: BannerConfig,
  position: string,
): string {
  switch (presentation) {
    case "OPT_OUT":
      return renderOptOut(position);
    case "NOTICE":
      return renderNotice(config, position);
    default:
      return renderOptIn(position);
  }
}
