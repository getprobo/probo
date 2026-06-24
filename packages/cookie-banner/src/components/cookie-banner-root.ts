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

import { CookieBannerClient } from "../client";
import type { BannerConfig, Regulation } from "../types";
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

  get gpcApplied(): boolean {
    return this._client?.gpcApplied ?? false;
  }

  get regulation(): Regulation | null {
    return this._client?.regulation ?? null;
  }

  get consentMode(): "OPT_IN" | "OPT_OUT" | null {
    const mode = this._config?.consent_mode;
    if (mode === "OPT_IN" || mode === "OPT_OUT") return mode;
    return null;
  }

  get reopenState(): ProboState {
    return this.consentMode === "OPT_OUT" ? "banner" : "panel";
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
    this.setState(this.reopenState);
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

    if (!this._client.loaded) return;

    this._config = this._client.config;
    this._draft = this.buildDraft(this._config);

    this.dispatchEvent(
      new CustomEvent("probo-ready", {
        bubbles: true,
        composed: true,
        detail: { config: this._config, gpcApplied: this.gpcApplied, regulation: this._client.regulation },
      }),
    );

    if (this._client.hasConsent) {
      this.setState("hidden");
    } else {
      this.setState("banner");
    }
  }
}
