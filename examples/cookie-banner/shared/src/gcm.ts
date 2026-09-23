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

export interface GCMSnapshot {
  command: string;
  signals: Record<string, string>;
}

type DataLayerWindow = Window & { dataLayer?: unknown[] };

export function getDataLayer(): unknown[] {
  const w = window as DataLayerWindow;
  if (!Array.isArray(w.dataLayer)) {
    w.dataLayer = [];
  }
  return w.dataLayer;
}

export function normalizeDataLayerEntry(entry: unknown): unknown {
  if (
    entry &&
    typeof entry === "object" &&
    typeof (entry as ArrayLike<unknown>).length === "number"
  ) {
    return Array.from(entry as ArrayLike<unknown>);
  }
  return entry;
}

export function parseGCMEvent(entry: unknown): GCMSnapshot | null {
  if (!Array.isArray(entry) || entry[0] !== "consent") {
    return null;
  }
  const command = typeof entry[1] === "string" ? entry[1] : "";
  const raw = entry[2];
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) {
    return { command, signals: {} };
  }
  const signals: Record<string, string> = {};
  for (const [key, value] of Object.entries(raw as Record<string, unknown>)) {
    signals[key] = String(value);
  }
  return { command, signals };
}

export function readGCMSnapshot(): GCMSnapshot | null {
  const dataLayer = getDataLayer();
  for (let i = dataLayer.length - 1; i >= 0; i--) {
    const parsed = parseGCMEvent(normalizeDataLayerEntry(dataLayer[i]));
    if (parsed) {
      return parsed;
    }
  }
  return null;
}
