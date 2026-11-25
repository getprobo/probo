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

import { useEffect, useRef, useState, useTransition } from "react";

export const cardListPageSize = 10;

const searchDelayMs = 300;

// The search box of a sidebar card list. Typing commits the trimmed query
// after a pause, inside a transition so the card stays on screen while the
// list refetches.
export function useCardListSearch(onSearch: (query: string) => void) {
  const [query, setQuery] = useState("");
  const [isPending, startTransition] = useTransition();
  const committedQuery = useRef("");
  const search = useRef(onSearch);

  useEffect(() => {
    search.current = onSearch;
  }, [onSearch]);

  useEffect(() => {
    const handle = setTimeout(() => {
      const next = query.trim();
      if (next === committedQuery.current) {
        return;
      }
      committedQuery.current = next;
      startTransition(() => {
        search.current(next);
      });
    }, searchDelayMs);
    return () => {
      clearTimeout(handle);
    };
  }, [query]);

  return { query, setQuery, filtering: query.trim() !== "", isPending };
}
