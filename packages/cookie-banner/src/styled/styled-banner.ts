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

import { ConsentManager } from "../headless/consent-manager";
import type { BannerConfig, ConsentManagerConfig } from "../headless/types";
import { renderBanner } from "./banner-renderer";
import { renderRevisitIcon } from "./revisit-renderer";

export class StyledBanner {
  readonly manager: ConsentManager;

  constructor(config: ConsentManagerConfig) {
    this.manager = new ConsentManager(config);
  }

  async init(): Promise<void> {
    this.manager.onConsentRequired((config) => {
      this.showBanner(config);
    });

    this.manager.onReady((config) => {
      if (!this.manager.needsConsent()) {
        this.showRevisitIcon(config);
      }
    });

    await this.manager.init();
  }

  show(): void {
    const config = this.manager.getConfig();
    if (!config) return;
    document.getElementById("probo-cookie-revisit")?.remove();
    this.showBanner(config);
  }

  destroy(): void {
    document.getElementById("probo-cookie-banner")?.remove();
    document.getElementById("probo-cookie-revisit")?.remove();
    this.manager.destroy();
  }

  private showBanner(config: BannerConfig): void {
    document.getElementById("probo-cookie-banner")?.remove();

    const host = document.createElement("div");
    host.id = "probo-cookie-banner";
    document.body.appendChild(host);
    const shadow = host.attachShadow({ mode: "closed" });

    const strings = this.manager.getStrings();

    renderBanner(
      shadow,
      config,
      this.manager.getConsents(),
      {
        onAcceptAll: () => {
          this.manager.acceptAll();
          host.remove();
          this.showRevisitIcon(config);
        },
        onRejectAll: () => {
          this.manager.rejectAll();
          host.remove();
          this.showRevisitIcon(config);
        },
        onCustomize: (choices) => {
          this.manager.savePreferences(choices);
          host.remove();
          this.showRevisitIcon(config);
        },
      },
      strings,
    );
  }

  private showRevisitIcon(config: BannerConfig): void {
    document.getElementById("probo-cookie-revisit")?.remove();

    const host = document.createElement("div");
    host.id = "probo-cookie-revisit";
    document.body.appendChild(host);
    const shadow = host.attachShadow({ mode: "closed" });

    renderRevisitIcon(
      shadow,
      () => {
        host.remove();
        this.showBanner(config);
      },
      this.manager.getStrings(),
      config.theme,
    );
  }
}
