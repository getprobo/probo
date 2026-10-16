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

import { ComboboxPaged } from "@probo/ui/src/v2/Combobox/ComboboxPaged";
import { type ReactNode, useEffect, useRef, useState } from "react";
import { useTranslation } from "react-i18next";
import { fetchQuery, useRelayEnvironment } from "react-relay";
import type { GraphQLTaggedNode } from "relay-runtime";

const searchDebounceMs = 300;

interface PageResult<TItem> {
  items: TItem[];
  hasNextPage: boolean;
  endCursor: string | null;
}

export interface RemoteComboboxProps<TQuery extends { response: unknown; variables: object }, TItem> {
  query: GraphQLTaggedNode;
  variables: (after: string | null, query: string) => TQuery["variables"];
  readPage: (data: TQuery["response"]) => PageResult<TItem>;
  reloadKey: string;
  selectedId: string | null;
  selectedLabel: string;
  ariaLabel: string;
  placeholder: string;
  emptyLabel: string;
  enabled?: boolean;
  disabled?: boolean;
  pinned?: {
    id: string;
    label: string;
  };
  onSelect: (item: TItem) => void;
  onPinnedSelect?: () => void;
  onFirstItem?: (item: TItem) => void;
  itemId: (item: TItem) => string;
  itemLabel: (item: TItem) => string;
  itemLeading?: (item: TItem) => ReactNode;
  pinnedLeading?: ReactNode;
}

function useDebounced(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const timeout = setTimeout(() => setDebounced(value), delay);
    return () => clearTimeout(timeout);
  }, [value, delay]);

  return [debounced, setDebounced] as const;
}

export function RemoteCombobox<TQuery extends { response: unknown; variables: object }, TItem>({
  query,
  variables,
  readPage,
  reloadKey,
  selectedId,
  selectedLabel,
  ariaLabel,
  placeholder,
  emptyLabel,
  enabled = true,
  disabled,
  pinned,
  onSelect,
  onPinnedSelect,
  onFirstItem,
  itemId,
  itemLabel,
  itemLeading,
  pinnedLeading,
}: RemoteComboboxProps<TQuery, TItem>) {
  const { t } = useTranslation("organizations/tasks");
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
  const onFirstItemRef = useRef(onFirstItem);
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
    onFirstItemRef.current = onFirstItem;
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
        if (!append && debouncedSearch === "" && page.items[0] != null) {
          onFirstItemRef.current?.(page.items[0]);
        }
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

  return (
    <ComboboxPaged
      items={items}
      loading={loading}
      failed={failed}
      query={search}
      onQueryChange={setSearch}
      hasNext={hasNext}
      onLoadMore={() => {
        if (endCursor == null || loading) {
          return;
        }

        setRequest(current => ({
          after: endCursor,
          append: true,
          id: current.id + 1,
        }));
      }}
      selectedId={selectedId}
      selectedLabel={selectedLabel}
      ariaLabel={ariaLabel}
      placeholder={placeholder}
      emptyLabel={emptyLabel}
      errorLabel={t("detailsPage.linear.errors.load")}
      loadMoreLabel={t("detailsPage.linear.loadMore")}
      disabled={disabled}
      pinned={pinned == null
        ? undefined
        : { id: pinned.id, label: pinned.label, leading: pinnedLeading }}
      onPinnedSelect={onPinnedSelect}
      onSelect={onSelect}
      itemId={itemId}
      itemLabel={itemLabel}
      itemLeading={itemLeading}
    />
  );
}
