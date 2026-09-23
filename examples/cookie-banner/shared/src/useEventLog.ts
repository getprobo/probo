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

import { useCallback, useEffect, useState } from "react";
import {
  getDataLayer,
  normalizeDataLayerEntry,
  parseGCMEvent,
} from "./gcm";
import { getExampleLogger } from "./logger";

const gcmLogger = getExampleLogger("gcm");

export interface EventEntry {
  id: number;
  time: string;
  type: string;
  detail: unknown;
}

const MAX_EVENTS = 50;

export function useEventLog(): {
  events: EventEntry[];
  pushEvent: (type: string, detail: unknown) => void;
} {
  const [events, setEvents] = useState<EventEntry[]>([]);
  const pushEvent = useCallback((type: string, detail: unknown) => {
    setEvents((prev) => {
      const next: EventEntry = {
        id: (prev[0]?.id ?? 0) + 1,
        time: new Date().toISOString(),
        type,
        detail,
      };
      return [next, ...prev].slice(0, MAX_EVENTS);
    });
  }, []);

  useEffect(() => {
    let seen = 0;
    let dataLayer = getDataLayer();
    let disposed = false;

    const ingest = (dl: unknown[]): void => {
      if (dl !== dataLayer) {
        dataLayer = dl;
        seen = 0;
      }
      while (seen < dl.length) {
        const parsed = parseGCMEvent(normalizeDataLayerEntry(dl[seen++]));
        if (!parsed) {
          continue;
        }
        gcmLogger.debug("[gcm]", parsed);
        pushEvent("gcm-consent", parsed);
      }
    };

    let wrappedLayer: unknown[] | null = null;
    let wrappedOriginal: typeof Array.prototype.push | null = null;

    const attach = (): void => {
      const dl = getDataLayer();
      ingest(dl);
      const current = dl.push as typeof dl.push & { __proboExampleGcm?: boolean };
      if (current.__proboExampleGcm) {
        return;
      }
      const original = dl.push;
      const wrapped = function (this: unknown[], ...items: unknown[]): number {
        const n = original.apply(this, items);
        if (!disposed) {
          ingest(this);
        }
        return n;
      } as typeof dl.push & { __proboExampleGcm?: boolean };
      wrapped.__proboExampleGcm = true;
      dl.push = wrapped;
      wrappedLayer = dl;
      wrappedOriginal = original;
    };

    attach();
    const id = window.setInterval(attach, 1000);
    return () => {
      disposed = true;
      window.clearInterval(id);
      if (
        wrappedLayer &&
        wrappedOriginal &&
        (wrappedLayer.push as { __proboExampleGcm?: boolean }).__proboExampleGcm
      ) {
        wrappedLayer.push = wrappedOriginal;
      }
    };
  }, [pushEvent]);

  return { events, pushEvent };
}
