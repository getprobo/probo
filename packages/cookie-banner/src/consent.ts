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

export type ConsentData = Record<string, boolean>;
type Callback = (consent: ConsentData) => void;

export class ConsentManager {
  private _ready = false;
  private _hasResponse = false;
  private _snapshot: ConsentData = {};
  private readonly _readyListeners: Callback[] = [];
  private readonly _changeListeners: Callback[] = [];

  get ready(): boolean {
    return this._ready;
  }

  get hasResponse(): boolean {
    return this._hasResponse;
  }

  has(category: string): boolean {
    return !!this._snapshot[category];
  }

  getAll(): ConsentData {
    return this._snapshot;
  }

  subscribe(cb: Callback): () => void {
    const offReady = this.onReady(cb);
    const offChange = this.onChange(cb);
    return () => { offReady(); offChange(); };
  }

  onReady(cb: Callback): () => void {
    if (this._ready) {
      cb(this._snapshot);
      return () => {};
    }
    this._readyListeners.push(cb);
    return () => {
      const idx = this._readyListeners.indexOf(cb);
      if (idx !== -1) this._readyListeners.splice(idx, 1);
    };
  }

  onChange(cb: Callback): () => void {
    this._changeListeners.push(cb);
    return () => {
      const idx = this._changeListeners.indexOf(cb);
      if (idx !== -1) this._changeListeners.splice(idx, 1);
    };
  }

  /** @internal Called by CookieBannerClient when consent state is first resolved. */
  _setReady(consent: ConsentData, hasResponse: boolean): void {
    this._snapshot = { ...consent };
    this._hasResponse = hasResponse;
    this._ready = true;
    for (const cb of this._readyListeners.splice(0)) {
      cb(this._snapshot);
    }
    for (const cb of this._changeListeners) {
      cb(this._snapshot);
    }
  }

  /** @internal Called by CookieBannerClient when consent changes after user action. */
  _notify(consent: ConsentData): void {
    this._snapshot = { ...consent };
    this._hasResponse = true;
    for (const cb of this._changeListeners) {
      cb(this._snapshot);
    }
  }
}

const GLOBAL_KEY = "__proboConsentManager";

export function getConsent(): ConsentManager {
  const g = typeof globalThis !== "undefined"
    ? (globalThis as unknown as Record<string, unknown>)
    : (window as unknown as Record<string, unknown>);

  if (!g[GLOBAL_KEY]) {
    g[GLOBAL_KEY] = new ConsentManager();
  }
  return g[GLOBAL_KEY] as ConsentManager;
}
