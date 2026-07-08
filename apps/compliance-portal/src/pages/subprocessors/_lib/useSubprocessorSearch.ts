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

import { useEffect, useRef, useState } from "react";

import { useSubprocessorFilters } from "./useSubprocessorFilters";

const SEARCH_DEBOUNCE_MS = 300;

// Owns the debounced search input for the toolbar. Mount this in exactly ONE
// component (the toolbar) — it is the single writer of the `q` URL param. It
// keeps an immediate local value for the input and, after a debounce, commits it
// to the URL. A ref tracks our own writes so the URL→input sync only reacts to
// *external* changes (clear button, back/forward), never echoing our own commit
// back onto the input (which would drop in-flight keystrokes).
export function useSubprocessorSearch(): [string, (value: string) => void] {
  const { query, setQuery } = useSubprocessorFilters();
  const [input, setInput] = useState(query);
  const lastCommittedRef = useRef(query);

  useEffect(() => {
    if (input === query) {
      return;
    }

    const handle = setTimeout(() => {
      lastCommittedRef.current = input;
      setQuery(input);
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(handle);
  }, [input, query, setQuery]);

  useEffect(() => {
    if (query !== lastCommittedRef.current) {
      lastCommittedRef.current = query;
      setInput(query);
    }
  }, [query]);

  return [input, setInput];
}
