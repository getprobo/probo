// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { BannerConfig, ThemeConfig, WidgetStrings } from "../headless/types";
import { defaultTheme } from "../headless/api";
import { renderPreferences } from "./preferences-renderer";
import { themeToCSSVars } from "./theme";

export interface BannerCallbacks {
  onAcceptAll: () => void;
  onRejectAll: () => void;
  onCustomize: (choices: Record<string, boolean>) => void;
}

export interface BannerOptions {
  preview?: boolean;
  theme?: ThemeConfig;
}

export function renderBanner(
  root: ShadowRoot,
  config: BannerConfig,
  currentConsent: Record<string, boolean>,
  callbacks: BannerCallbacks,
  strings: WidgetStrings,
  options?: BannerOptions,
): void {
  root.innerHTML = "";

  const theme = options?.theme ?? config.theme ?? defaultTheme;

  const style = document.createElement("style");
  style.textContent = `
    :host { ${themeToCSSVars(theme)} }
    * { box-sizing: border-box; margin: 0; padding: 0; }
    .probo-banner-overlay {
      position: fixed; bottom: 0; left: 0; right: 0; z-index: 2147483647;
      font-family: var(--probo-font);
    }
    .probo-banner-container {
      background: var(--probo-bg); border-top: 1px solid var(--probo-border);
      box-shadow: 0 -4px 20px rgba(0,0,0,0.1);
      padding: 0 24px; max-width: 100%;
      overflow: hidden;
      transition: max-height 0.3s ease-out, padding 0.3s ease-out;
      max-height: 0;
    }
    .probo-banner-container.open {
      padding: 24px;
      max-height: 800px;
    }
    .probo-banner-content {
      max-width: 1200px; margin: 0 auto;
    }
    .probo-banner-title {
      font-size: 16px; font-weight: 600; color: var(--probo-text); margin-bottom: 8px;
    }
    .probo-banner-desc {
      font-size: 14px; color: var(--probo-text-secondary); line-height: 1.5; margin-bottom: 16px;
    }
    .probo-banner-actions {
      display: flex; gap: 12px; flex-wrap: wrap; align-items: center;
    }
    .probo-btn {
      padding: 10px 24px; border-radius: var(--probo-radius); font-size: 14px;
      font-weight: 500; cursor: pointer; border: none; transition: opacity 0.2s;
      font-family: var(--probo-font);
    }
    .probo-btn:hover { opacity: 0.9; }
    .probo-btn-primary {
      background: var(--probo-primary); color: var(--probo-primary-text);
    }
    .probo-btn-secondary {
      background: var(--probo-secondary); color: var(--probo-secondary-text);
    }
    .probo-btn-outline {
      background: transparent; color: var(--probo-text);
      border: 1px solid var(--probo-border);
    }
    .probo-privacy-link {
      font-size: 13px; color: var(--probo-text-secondary); text-decoration: underline; cursor: pointer;
      margin-left: auto;
    }
    .probo-privacy-link:hover { color: var(--probo-text); }
  `;
  root.appendChild(style);

  const overlay = document.createElement("div");
  overlay.className = "probo-banner-overlay";

  const container = document.createElement("div");
  container.className = "probo-banner-container";

  const content = document.createElement("div");
  content.className = "probo-banner-content";

  const transitionTo = (render: () => void) => {
    content.innerHTML = "";
    render();
  };

  // Main banner view
  const renderMainContent = () => {
    const title = document.createElement("div");
    title.className = "probo-banner-title";
    title.textContent = config.title;
    content.appendChild(title);

    const desc = document.createElement("div");
    desc.className = "probo-banner-desc";
    desc.textContent = config.description;
    content.appendChild(desc);

    const actions = document.createElement("div");
    actions.className = "probo-banner-actions";

    const rejectBtn = document.createElement("button");
    rejectBtn.className = "probo-btn probo-btn-secondary";
    rejectBtn.textContent = config.reject_all_label;
    rejectBtn.addEventListener("click", () => {
      if (!options?.preview) {
        callbacks.onRejectAll();
        overlay.remove();
      }
    });
    actions.appendChild(rejectBtn);

    const acceptBtn = document.createElement("button");
    acceptBtn.className = "probo-btn probo-btn-primary";
    acceptBtn.textContent = config.accept_all_label;
    acceptBtn.addEventListener("click", () => {
      if (!options?.preview) {
        callbacks.onAcceptAll();
        overlay.remove();
      }
    });
    actions.appendChild(acceptBtn);

    const customizeBtn = document.createElement("button");
    customizeBtn.className = "probo-btn probo-btn-outline";
    customizeBtn.textContent = strings.customize;
    customizeBtn.addEventListener("click", () => {
      transitionTo(() => {
        renderPreferences(
          content,
          config.categories,
          currentConsent,
          options?.preview
            ? () => { transitionTo(renderMainContent); }
            : (choices) => {
                callbacks.onCustomize(choices);
                overlay.remove();
              },
          options?.preview
            ? () => { transitionTo(renderMainContent); }
            : () => {
                callbacks.onRejectAll();
                overlay.remove();
              },
          options?.preview
            ? () => { transitionTo(renderMainContent); }
            : () => {
                callbacks.onAcceptAll();
                overlay.remove();
              },
          strings,
          theme,
        );
      });
    });
    actions.appendChild(customizeBtn);

    if (config.privacy_policy_url) {
      const link = document.createElement("a");
      link.className = "probo-privacy-link";
      link.textContent = strings.privacyPolicy;
      link.href = config.privacy_policy_url;
      link.target = "_blank";
      link.rel = "noopener noreferrer";
      actions.appendChild(link);
    }

    content.appendChild(actions);
  };

  renderMainContent();
  container.appendChild(content);
  overlay.appendChild(container);
  root.appendChild(overlay);

  // Trigger grow animation on next frame
  requestAnimationFrame(() => {
    requestAnimationFrame(() => {
      container.classList.add("open");
    });
  });
}
