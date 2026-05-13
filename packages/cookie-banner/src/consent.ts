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
    if (this._ready) {
      cb(this._snapshot);
    }
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

let instance: ConsentManager | null = null;

export function getConsent(): ConsentManager {
  if (!instance) {
    instance = new ConsentManager();
  }
  return instance;
}
