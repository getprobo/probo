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

export function useSerializedFieldSave(
  save: (value: string) => Promise<void>,
) {
  const saveRef = useRef(save);
  const pendingRef = useRef<string | null>(null);
  const inFlightRef = useRef(false);

  useEffect(() => {
    saveRef.current = save;
  }, [save]);

  return useCallback((value: string) => {
    pendingRef.current = value;

    function pump() {
      if (inFlightRef.current) {
        return;
      }

      const next = pendingRef.current;
      if (next == null) {
        return;
      }

      pendingRef.current = null;
      inFlightRef.current = true;

      void saveRef.current(next)
        .catch(() => undefined)
        .finally(() => {
          inFlightRef.current = false;
          pump();
        });
    }

    pump();
  }, []);
}

export function useDebouncedSerializedFieldSave(
  save: (value: string) => Promise<void>,
  delayMs: number,
) {
  const persist = useSerializedFieldSave(save);
  const persistRef = useRef(persist);
  const [pending, setPending] = useState<string | null>(null);
  const pendingSave = useRef<{ value: string } | null>(null);

  useEffect(() => {
    persistRef.current = persist;
  }, [persist]);

  useLayoutEffect(() => {
    pendingSave.current = pending == null ? null : { value: pending };
  }, [pending]);

  useEffect(() => {
    return () => {
      const queued = pendingSave.current;
      if (queued == null) {
        return;
      }

      persistRef.current(queued.value);
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
    pendingSave.current = { value };
    setPending(value);
  }

  function flush() {
    const queued = pendingSave.current;
    if (queued == null) {
      return;
    }

    persist(queued.value);
    setPending(null);
  }

  function cancel() {
    pendingSave.current = null;
    setPending(null);
  }

  return { schedule, flush, cancel };
}
