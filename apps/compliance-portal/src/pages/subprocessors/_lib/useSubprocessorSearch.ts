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
