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
