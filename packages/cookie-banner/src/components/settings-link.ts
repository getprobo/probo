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

import { CCPA_OPT_OUT_ICON, CCPA_PRIVACY_CHOICES_LABEL } from "../html";
import { getGpcLabel } from "../i18n";
import type { ProboCookieBannerRoot } from "./cookie-banner-root";

const AUTO_ATTR = "data-probo-auto";

export class ProboSettingsLink extends HTMLElement {
  private root: ProboCookieBannerRoot | null = null;
  private attached = false;

  connectedCallback(): void {
    this.root = this.findRoot();

    if (this.root) {
      this.attach(this.root);
    } else {
      document.addEventListener("probo-ready", this.onProboReady, { once: true });
    }

    this.addEventListener("click", this.handleClick);
    this.addEventListener("keydown", this.handleKeydown);
  }

  disconnectedCallback(): void {
    this.removeEventListener("click", this.handleClick);
    this.removeEventListener("keydown", this.handleKeydown);
    document.removeEventListener("probo-ready", this.onProboReady);
    if (this.root) {
      this.root.removeEventListener("probo-ready", this.onRootReady);
      this.root.removeEventListener("probo-consent", this.onConsent);
    }
    this.root = null;
    this.attached = false;
  }

  private attach(root: ProboCookieBannerRoot): void {
    this.root = root;

    if (!this.attached) {
      this.attached = true;
      root.addEventListener("probo-consent", this.onConsent);
    }

    try {
      this.applyContent(root);
      this.applyGpc(root.gpcApplied, root.bannerConfig.language);
    } catch {
      root.addEventListener("probo-ready", this.onRootReady, { once: true });
    }
  }

  private onRootReady = (e: Event): void => {
    if (!this.root) return;
    const detail = (e as CustomEvent).detail as {
      gpcApplied?: boolean;
      config?: { language?: string };
    };
    this.applyContent(this.root);
    this.applyGpc(!!detail.gpcApplied, detail.config?.language ?? "en");
  };

  private applyContent(root: ProboCookieBannerRoot): void {
    this.style.display = "inline-flex";
    this.style.alignItems = "center";
    this.style.gap = "6px";
    this.style.cursor = "pointer";

    if (!this.hasAttribute("role")) {
      this.setAttribute("role", "button");
    }
    if (!this.hasAttribute("tabindex")) {
      this.tabIndex = 0;
    }

    // CCPA always uses the statutory Alternative Opt-out Link label + icon,
    // even when the integrator provided children (those are for other regs).
    if (root.regulation === "CCPA") {
      this.replaceChildren();
      this.setAttribute(AUTO_ATTR, "");

      const label = document.createElement("span");
      label.textContent = CCPA_PRIVACY_CHOICES_LABEL;
      this.styleInheritedText(label);
      this.append(label);

      const icon = document.createElement("span");
      icon.style.display = "inline-flex";
      icon.style.flexShrink = "0";
      icon.style.height = "1em";
      icon.style.width = "auto";
      icon.innerHTML = CCPA_OPT_OUT_ICON;
      const svg = icon.querySelector("svg");
      if (svg) {
        svg.style.height = "100%";
        svg.style.width = "auto";
        svg.removeAttribute("width");
        svg.removeAttribute("height");
      }
      this.append(icon);
      return;
    }

    if (this.hasConsumerChildren()) return;

    this.replaceChildren();
    this.setAttribute(AUTO_ATTR, "");

    const text =
      root.bannerConfig.texts?.aria_cookie_settings ?? "Cookie settings";
    const label = document.createElement("span");
    label.textContent = text;
    this.styleInheritedText(label);
    this.append(label);
    this.setAttribute("aria-label", text);
  }

  /** Auto-filled text inherits typography/color from the host. */
  private styleInheritedText(el: HTMLElement): void {
    el.style.color = "inherit";
    el.style.font = "inherit";
    el.style.lineHeight = "inherit";
  }

  private hasConsumerChildren(): boolean {
    if (this.hasAttribute(AUTO_ATTR)) return false;
    return this.childNodes.length > 0;
  }

  private applyGpc(applied: boolean, language: string): void {
    this.querySelector("[data-probo-gpc]")?.remove();
    if (!applied) return;

    const badge = document.createElement("span");
    badge.setAttribute("data-probo-gpc", "");
    badge.textContent = getGpcLabel(language);
    badge.style.fontSize = "11px";
    badge.style.fontWeight = "600";
    badge.style.whiteSpace = "nowrap";
    this.append(badge);
  }

  private onConsent = (e: Event): void => {
    if (!this.root) return;
    const { action } = (e as CustomEvent).detail as { action?: string };
    if (action === "GPC") {
      this.applyGpc(true, this.root.bannerConfig.language);
    } else {
      this.applyGpc(false, this.root.bannerConfig.language);
    }
  };

  private findRoot(): ProboCookieBannerRoot | null {
    const direct = document.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot | null;
    if (direct) return direct;

    const themed = document.querySelector("probo-cookie-banner");
    if (themed?.shadowRoot) {
      return themed.shadowRoot.querySelector("probo-cookie-banner-root") as ProboCookieBannerRoot | null;
    }

    return null;
  }

  private onProboReady = (e: Event): void => {
    const root = (e as CustomEvent).target as ProboCookieBannerRoot | null;
    if (root?.tagName.toLowerCase() === "probo-cookie-banner-root") {
      this.attach(root);
      return;
    }

    const found = this.findRoot();
    if (found) {
      this.attach(found);
    }
  };

  private handleClick = (e: Event): void => {
    if (!this.root) return;
    e.preventDefault();
    this.root.setState(this.root.reopenState);
  };

  private handleKeydown = (e: KeyboardEvent): void => {
    if (e.key !== "Enter" && e.key !== " ") return;
    e.preventDefault();
    this.handleClick(e);
  };
}
