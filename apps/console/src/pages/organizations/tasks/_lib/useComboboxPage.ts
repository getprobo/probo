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
import { fetchQuery, useRelayEnvironment } from "react-relay";
import type { GraphQLTaggedNode } from "relay-runtime";

import { useDebounced } from "#/lib/useDebounced";

const searchDebounceMs = 300;

interface PageResult<TItem> {
  items: TItem[];
  hasNextPage: boolean;
  endCursor: string | null;
}

export function useComboboxPage<
  TQuery extends { response: unknown; variables: object },
  TItem,
>({
  query,
  variables,
  readPage,
  reloadKey,
  enabled = true,
}: {
  query: GraphQLTaggedNode;
  variables: (after: string | null, query: string) => TQuery["variables"];
  readPage: (data: TQuery["response"]) => PageResult<TItem>;
  reloadKey: string;
  enabled?: boolean;
}) {
  const environment = useRelayEnvironment();
  const [search, setSearch] = useState("");
  const [debouncedSearch, setDebouncedSearch] = useDebounced(search, searchDebounceMs);
  const [items, setItems] = useState<TItem[]>([]);
  const [hasNext, setHasNext] = useState(false);
  const [endCursor, setEndCursor] = useState<string | null>(null);
  const [request, setRequest] = useState({ after: null as string | null, append: false, id: 0 });
  const [settledId, setSettledId] = useState<number | null>(null);
  const [failedId, setFailedId] = useState<number | null>(null);
  const variablesRef = useRef(variables);
  const readPageRef = useRef(readPage);
  const [scope, setScope] = useState(reloadKey);
  const [searchScope, setSearchScope] = useState(debouncedSearch);
  const loading = enabled && settledId !== request.id;
  const failed = failedId === request.id;

  if (scope !== reloadKey) {
    setScope(reloadKey);
    setSearch("");
    setDebouncedSearch("");
    setSearchScope("");
    setRequest({ after: null, append: false, id: request.id + 1 });
    setItems([]);
    setHasNext(false);
    setEndCursor(null);
  } else if (searchScope !== debouncedSearch) {
    setSearchScope(debouncedSearch);
    setRequest({ after: null, append: false, id: request.id + 1 });
    setItems([]);
    setHasNext(false);
    setEndCursor(null);
  }

  useEffect(() => {
    variablesRef.current = variables;
    readPageRef.current = readPage;
  });

  useEffect(() => {
    if (!enabled) {
      return;
    }

    let active = true;
    const { after, append, id } = request;
    const subscription = fetchQuery<TQuery>(
      environment,
      query,
      variablesRef.current(after, debouncedSearch),
      { fetchPolicy: "network-only" },
    ).subscribe({
      next: (data) => {
        if (!active) {
          return;
        }

        const page = readPageRef.current(data);
        setItems(current => (append ? [...current, ...page.items] : page.items));
        setHasNext(page.hasNextPage && page.endCursor != null);
        setEndCursor(page.endCursor);
        setFailedId(null);
        setSettledId(id);
      },
      error: () => {
        if (!active) {
          return;
        }

        setFailedId(id);
        setSettledId(id);
      },
    });

    return () => {
      active = false;
      subscription.unsubscribe();
    };
  }, [debouncedSearch, enabled, environment, query, reloadKey, request]);

  function onLoadMore() {
    if (endCursor == null || loading) {
      return;
    }

    setRequest(current => ({
      after: endCursor,
      append: true,
      id: current.id + 1,
    }));
  }

  return {
    items,
    loading,
    failed,
    query: search,
    onQueryChange: setSearch,
    hasNext,
    onLoadMore,
  };
}
