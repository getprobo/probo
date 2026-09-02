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

import { useCallback, useEffect, useLayoutEffect, useRef, useState } from "react";

type RequestedSave = {
  id: number;
  value: string;
  save: (value: string) => Promise<void>;
};

export function useSerializedFieldSave(
  save: (value: string) => Promise<void>,
) {
  const [queued, setQueued] = useState<RequestedSave | null>(null);
  const [inFlightId, setInFlightId] = useState<number | null>(null);

  if (queued != null && inFlightId == null) {
    setInFlightId(queued.id);
  }

  useEffect(() => {
    if (inFlightId == null || queued == null || queued.id !== inFlightId) {
      return;
    }

    const { id, value, save: persistValue } = queued;
    void persistValue(value)
      .finally(() => {
        setInFlightId(current => (current === id ? null : current));
        setQueued(current => (current?.id === id ? null : current));
      })
      .catch(() => undefined);
  }, [inFlightId, queued]);

  return useCallback((value: string) => {
    setQueued(current => ({
      id: (current?.id ?? 0) + 1,
      value,
      save,
    }));
  }, [save]);
}

export function useDebouncedSerializedFieldSave(
  save: (value: string) => Promise<void>,
  delayMs: number,
) {
  const persist = useSerializedFieldSave(save);
  const [pending, setPending] = useState<string | null>(null);
  const pendingSave = useRef<{
    value: string;
    save: (value: string) => Promise<void>;
  } | null>(null);

  useLayoutEffect(() => {
    pendingSave.current = pending == null ? null : { value: pending, save };
  }, [pending, save]);

  useEffect(() => {
    return () => {
      const queued = pendingSave.current;
      if (queued == null) {
        return;
      }

      void queued.save(queued.value).catch(() => undefined);
    };
  }, []);

  useEffect(() => {
    if (pending == null) {
      return;
    }

    const timer = window.setTimeout(() => {
      persist(pending);
      setPending(null);
    }, delayMs);

    return () => {
      window.clearTimeout(timer);
    };
  }, [delayMs, pending, persist]);

  function schedule(value: string) {
    setPending(value);
  }

  function flush() {
    if (pending == null) {
      return;
    }

    persist(pending);
    setPending(null);
  }

  return { schedule, flush };
}
