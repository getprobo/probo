// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

export interface SubprocessorFilters {
  query: string;
  category: string;
  country: string;
  hasActiveFilters: boolean;
  setQuery: (value: string) => void;
  setCategory: (value: string) => void;
  setCountry: (value: string) => void;
  clear: () => void;
}

// Subprocessor filter state, persisted in the URL so it is shareable and
// survives reloads. This hook is pure URL state (no local component state or
// effects), so it can be read from any number of components without them
// fighting over the search params. The debounced search *input* lives in a
// single-owner hook (`useSubprocessorSearch`) to avoid write-back loops.
export function useSubprocessorFilters(): SubprocessorFilters {
  const [searchParams, setSearchParams] = useSearchParams();

  const category = searchParams.get("category") ?? "";
  const country = searchParams.get("country") ?? "";
  const query = searchParams.get("q") ?? "";

  const setParam = useCallback((key: string, value: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set(key, value);
      } else {
        next.delete(key);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setQuery = useCallback((value: string) => setParam("q", value), [setParam]);
  const setCategory = useCallback((value: string) => setParam("category", value), [setParam]);
  const setCountry = useCallback((value: string) => setParam("country", value), [setParam]);

  const clear = useCallback(() => {
    setSearchParams({}, { replace: true });
  }, [setSearchParams]);

  return {
    query,
    category,
    country,
    hasActiveFilters: query !== "" || category !== "" || country !== "",
    setQuery,
    setCategory,
    setCountry,
    clear,
  };
}
