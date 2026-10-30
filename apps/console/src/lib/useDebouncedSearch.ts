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

const SEARCH_DEBOUNCE_MS = 300;

// Keeps a search field responsive while the committed query lags by one
// debounce. `query` / `setQuery` are the page's filter state, usually the `q`
// URL param. A URL change that this hook did not write updates the field.
// `blankIfWhitespace` commits an all-whitespace field as an empty query.
export function useDebouncedSearch(filter: {
  query: string;
  setQuery: (value: string) => void;
  blankIfWhitespace?: boolean;
}): [string, (value: string) => void] {
  const { query, setQuery, blankIfWhitespace = false } = filter;
  const [input, setInput] = useState(query);
  const lastCommittedRef = useRef(query);

  useEffect(() => {
    if (input === query) {
      return;
    }

    const handle = setTimeout(() => {
      const value = blankIfWhitespace && input.trim() === "" ? "" : input;
      lastCommittedRef.current = value;
      setQuery(value);
    }, SEARCH_DEBOUNCE_MS);

    return () => clearTimeout(handle);
  }, [blankIfWhitespace, input, query, setQuery]);

  useEffect(() => {
    if (query !== lastCommittedRef.current) {
      lastCommittedRef.current = query;
      setInput(query);
    }
  }, [query]);

  return [input, setInput];
}
