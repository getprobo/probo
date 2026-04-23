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

import { CookieBannerClient } from "../client";
import type { BannerConfig } from "../client";
import { ProboElement } from "./base";
import type { ProboState, ProboRootElement, ConsentDraft } from "./base";

export class ProboCookieBannerRoot extends ProboElement implements ProboRootElement {
  private _client: CookieBannerClient | null = null;
  private _config: BannerConfig | null = null;
  private _state: ProboState = "loading";
  private _draft: ConsentDraft = {};

  static get observedAttributes(): string[] {
    return ["banner-id", "base-url", "reopen-widget", "lang"];
  }

  get client(): CookieBannerClient {
    if (!this._client) {
      throw new Error("<probo-cookie-banner-root> not loaded yet");
    }
    return this._client;
  }

  get bannerConfig(): BannerConfig {
    if (!this._config) {
      throw new Error("<probo-cookie-banner-root> not loaded yet");
    }
    return this._config;
  }

  get reopenWidget(): string {
    return this.getAttribute("reopen-widget") ?? "floating";
  }

  get state(): ProboState {
    return this._state;
  }

  get consentDraft(): ConsentDraft {
    return this._draft;
  }

  attributeChangedCallback(name: string, oldValue: string | null, newValue: string | null): void {
    if (name === "reopen-widget" && oldValue !== newValue) {
      this.dispatchEvent(
        new CustomEvent("probo-reopen-widget", {
          bubbles: true,
          composed: true,
          detail: { value: newValue ?? "floating" },
        }),
      );
    }
  }

  connectedCallback(): void {
    document.addEventListener("probo-open-preferences", this.onOpenPreferences);
    this.initClient();
  }

  disconnectedCallback(): void {
    document.removeEventListener("probo-open-preferences", this.onOpenPreferences);
    if (this._client) {
      this._client.destroy();
      this._client = null;
    }
  }

  private onOpenPreferences = (): void => {
    this.setState("panel");
  };

  setState(state: ProboState): void {
    const prev = this._state;
    this._state = state;
    this.dispatchEvent(
      new CustomEvent("probo-state", {
        bubbles: true,
        composed: true,
        detail: { state, prev },
      }),
    );
  }

  updateDraft(category: string, value: boolean): void {
    this._draft[category] = value;
  }

  resetDraft(): void {
    if (!this._config) return;
    this._draft = this.buildDraft(this._config);
  }

  private buildDraft(config: BannerConfig): ConsentDraft {
    const draft: ConsentDraft = {};
    const existing = this._client?.visitorConsent?.consent_data;

    for (const cat of config.categories) {
      if (cat.kind === "NECESSARY") {
        draft[cat.slug] = true;
      } else if (existing && (cat.slug in existing || cat.name in existing)) {
        draft[cat.slug] = existing[cat.slug] ?? existing[cat.name];
      } else {
        draft[cat.slug] = config.consent_mode === "OPT_OUT";
      }
    }

    return draft;
  }

  private async initClient(): Promise<void> {
    const bannerId = this.getAttribute("banner-id");
    const baseUrl = this.getAttribute("base-url");

    if (!bannerId || !baseUrl) {
      this.warn("<probo-cookie-banner-root> requires banner-id and base-url attributes");
      return;
    }

    const lang = this.getAttribute("lang") ?? undefined;

    this._client = new CookieBannerClient({ bannerId, baseUrl, lang });

    try {
      await this._client.load();
    } catch (err) {
      this.warn(`failed to load banner config: ${err}`);
      return;
    }

    // Element was disconnected while load() was in-flight.
    if (!this._client) return;

    this._config = this._client.config;
    this._draft = this.buildDraft(this._config);

    this.dispatchEvent(
      new CustomEvent("probo-ready", {
        bubbles: true,
        composed: true,
        detail: { config: this._config },
      }),
    );

    if (this._client.hasConsent) {
      this.setState("hidden");
    } else {
      this.setState("banner");
    }
  }
}
