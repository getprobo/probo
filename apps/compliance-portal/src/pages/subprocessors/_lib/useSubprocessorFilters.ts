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

import { useCallback, useEffect, useState } from "react";
import { useSearchParams } from "react-router";

const SEARCH_DEBOUNCE_MS = 300;

export interface SubprocessorFilters {
  // Debounced search term written to the URL (drives the query variables).
  query: string;
  category: string;
  country: string;
  // Immediate search input value (updates on every keystroke).
  queryInput: string;
  hasActiveFilters: boolean;
  setQueryInput: (value: string) => void;
  setCategory: (value: string) => void;
  setCountry: (value: string) => void;
  clear: () => void;
}

// Subprocessor filter state, persisted in the URL so it is shareable and
// survives reloads. The search term is debounced before it is committed to the
// URL (and therefore before it triggers a refetch).
export function useSubprocessorFilters(): SubprocessorFilters {
  const [searchParams, setSearchParams] = useSearchParams();

  const category = searchParams.get("category") ?? "";
  const country = searchParams.get("country") ?? "";
  const query = searchParams.get("q") ?? "";

  const [queryInput, setQueryInput] = useState(query);

  useEffect(() => {
    if (queryInput === query) {
      return;
    }

    const handle = setTimeout(() => {
      setSearchParams((previous) => {
        const next = new URLSearchParams(previous);
        if (queryInput) {
          next.set("q", queryInput);
        } else {
          next.delete("q");
        }
        return next;
      }, { replace: true });
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(handle);
  }, [queryInput, query, setSearchParams]);

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

  const setCategory = useCallback((value: string) => setParam("category", value), [setParam]);
  const setCountry = useCallback((value: string) => setParam("country", value), [setParam]);

  const clear = useCallback(() => {
    setQueryInput("");
    setSearchParams({}, { replace: true });
  }, [setSearchParams]);

  return {
    query,
    category,
    country,
    queryInput,
    hasActiveFilters: query !== "" || category !== "" || country !== "",
    setQueryInput,
    setCategory,
    setCountry,
    clear,
  };
}
