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

import type { DataRouter } from "react-router";

import { type FaroSettings, faroSettings } from "./settings";

export interface FaroClientOptions {
  appName: string;
  apiURL?: string;
}

export class FaroClient {
  private readonly options: FaroClientOptions;
  private settings: FaroSettings | null | undefined;
  private started: Promise<void> | undefined;

  constructor(options: FaroClientOptions) {
    this.options = options;
  }

  enabled(): boolean {
    return this.readSettings() !== null;
  }

  // Idempotent. The SDK module is loaded only when both collector meta tags
  // are present. A failed chunk or a thrown init is logged and must not
  // reject: callers start this without awaiting.
  start(router: DataRouter): Promise<void> {
    if (!this.enabled()) {
      return Promise.resolve();
    }
    this.started ??= this.load(router);
    return this.started;
  }

  private async load(router: DataRouter): Promise<void> {
    try {
      const settings = this.readSettings();
      if (!settings) {
        return;
      }

      const { initFaro } = await import("./sdk");
      initFaro({
        appName: this.options.appName,
        collectorURL: settings.url,
        apiKey: settings.apiKey,
        apiURL: this.options.apiURL,
        router,
      });
    } catch (error) {
      console.error("Cannot start Faro", error);
    }
  }

  private readSettings(): FaroSettings | null {
    if (this.settings !== undefined) {
      return this.settings;
    }

    // Callers start Faro without awaiting, including where document is absent.
    if (typeof document === "undefined") {
      this.settings = null;
      return this.settings;
    }

    const collectorURL = document
      .querySelector("meta[name=\"faro-collector-url\"]")
      ?.getAttribute("content") ?? "";
    const apiKey = document
      .querySelector("meta[name=\"faro-api-key\"]")
      ?.getAttribute("content") ?? "";
    this.settings = faroSettings(collectorURL, apiKey);
    return this.settings;
  }
}
