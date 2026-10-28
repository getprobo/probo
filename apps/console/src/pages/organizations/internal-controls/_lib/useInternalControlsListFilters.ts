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

import { internalControlImplementationStatuses } from "@probo/helpers";
import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

export const internalControlsPageSize = 100;

const listViews = ["all", "attention"] as const;
export type InternalControlsListView = (typeof listViews)[number];

const sorts = ["name", "nextDue"] as const;
export type InternalControlsListSort = (typeof sorts)[number];

export type InternalControlImplementationStatus
  = (typeof internalControlImplementationStatuses)[number];

export interface InternalControlsListVariables {
  first: number | null;
  after: string | null;
  last: number | null;
  before: string | null;
  order: {
    field: "NAME" | "NEXT_DUE";
    direction: "ASC" | "DESC";
  };
  query: string | null;
  implementationStatuses: InternalControlImplementationStatus[] | null;
  category: string | null;
  ownerId: string | null;
  needsAttention: boolean | null;
}

function isStatus(value: string): value is InternalControlImplementationStatus {
  return (internalControlImplementationStatuses as readonly string[]).includes(value);
}

function isView(value: string): value is InternalControlsListView {
  return (listViews as readonly string[]).includes(value);
}

function isSort(value: string): value is InternalControlsListSort {
  return (sorts as readonly string[]).includes(value);
}

function selectedStatuses(values: readonly string[]): InternalControlImplementationStatus[] {
  const seen = new Set<string>();
  const selected: InternalControlImplementationStatus[] = [];
  for (const value of values) {
    if (seen.has(value) || !isStatus(value)) {
      continue;
    }
    seen.add(value);
    selected.push(value);
  }
  return selected;
}

function pageNumber(value: string | null): number {
  const page = Number(value);
  if (!Number.isInteger(page) || page < 1) {
    return 1;
  }
  return page;
}

export function internalControlsListVariables(input: {
  query: string;
  statuses: readonly InternalControlImplementationStatus[];
  category: string | null;
  ownerId: string | null;
  view: InternalControlsListView;
  sort: InternalControlsListSort;
  direction: "ASC" | "DESC";
  after: string | null;
  before: string | null;
}): InternalControlsListVariables {
  const trimmed = input.query.trim();
  const after = input.after || null;
  const before = after ? null : input.before || null;

  return {
    first: before ? null : internalControlsPageSize,
    after: before ? null : after,
    last: before ? internalControlsPageSize : null,
    before,
    order: {
      field: input.sort === "nextDue" ? "NEXT_DUE" : "NAME",
      direction: input.direction,
    },
    query: trimmed === "" ? null : trimmed,
    implementationStatuses: input.statuses.length === 0 ? null : [...input.statuses],
    category: input.category,
    ownerId: input.ownerId,
    needsAttention: input.view === "attention" ? true : null,
  };
}

export function useInternalControlsListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawView = searchParams.get("view") ?? "";
  const view: InternalControlsListView = isView(rawView) ? rawView : "all";
  const rawSort = searchParams.get("sort") ?? "";
  const sort: InternalControlsListSort = isSort(rawSort) ? rawSort : "name";
  const direction = searchParams.get("dir") === "desc" ? "DESC" : "ASC";
  const categoryValue = searchParams.get("category");
  const category = categoryValue === "" ? null : categoryValue;
  const ownerValue = searchParams.get("owner");
  const ownerId = ownerValue === "" ? null : ownerValue;
  const after = searchParams.get("after") || null;
  const before = after ? null : searchParams.get("before") || null;
  // A page index without a cursor cannot be loaded, so it reads as page 1.
  const page = after == null && before == null
    ? 1
    : pageNumber(searchParams.get("page"));
  const statuses = useMemo(
    () => selectedStatuses(searchParams.getAll("status")),
    [searchParams],
  );

  const write = useCallback((
    mutate: (next: URLSearchParams) => void,
    options: { resetPage: boolean; replace: boolean },
  ) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      mutate(next);
      if (options.resetPage) {
        next.delete("after");
        next.delete("before");
        next.delete("page");
      }
      return next;
    }, { replace: options.replace });
  }, [setSearchParams]);

  const setQuery = useCallback((value: string) => {
    write((next) => {
      if (value) {
        next.set("q", value);
      } else {
        next.delete("q");
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setView = useCallback((value: InternalControlsListView) => {
    write((next) => {
      if (value === "attention") {
        next.set("view", "attention");
      } else {
        next.delete("view");
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setStatuses = useCallback((values: readonly InternalControlImplementationStatus[]) => {
    write((next) => {
      next.delete("status");
      for (const value of values) {
        next.append("status", value);
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setCategory = useCallback((value: string | null) => {
    write((next) => {
      if (value) {
        next.set("category", value);
      } else {
        next.delete("category");
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setOwnerId = useCallback((value: string | null) => {
    write((next) => {
      if (value) {
        next.set("owner", value);
      } else {
        next.delete("owner");
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setSort = useCallback((value: InternalControlsListSort) => {
    write((next) => {
      const current = isSort(next.get("sort") ?? "") ? next.get("sort") : "name";
      const currentDirection = next.get("dir") === "desc" ? "desc" : "asc";
      const nextDirection = current === value && currentDirection === "asc" ? "desc" : "asc";
      if (value === "name") {
        next.delete("sort");
      } else {
        next.set("sort", value);
      }
      if (nextDirection === "asc") {
        next.delete("dir");
      } else {
        next.set("dir", "desc");
      }
    }, { resetPage: true, replace: true });
  }, [write]);

  const setPageCursor = useCallback((cursor: {
    after?: string | null;
    before?: string | null;
    page: number;
  }) => {
    write((next) => {
      next.delete("after");
      next.delete("before");
      next.delete("page");
      if (cursor.after) {
        next.set("after", cursor.after);
      }
      if (cursor.before) {
        next.set("before", cursor.before);
      }
      if (cursor.page > 1) {
        next.set("page", String(cursor.page));
      }
    }, { resetPage: false, replace: false });
  }, [write]);

  const clear = useCallback(() => {
    write((next) => {
      next.delete("q");
      next.delete("status");
      next.delete("category");
      next.delete("owner");
      next.delete("view");
      next.delete("sort");
      next.delete("dir");
    }, { resetPage: true, replace: true });
  }, [write]);

  const hasActiveFilters = query.trim() !== ""
    || statuses.length > 0
    || category != null
    || ownerId != null
    || view === "attention";

  return {
    query,
    statuses,
    category,
    ownerId,
    view,
    sort,
    direction,
    after,
    before,
    page,
    hasActiveFilters,
    variables: internalControlsListVariables({
      query,
      statuses,
      category,
      ownerId,
      view,
      sort,
      direction,
      after,
      before,
    }),
    setQuery,
    setView,
    setStatuses,
    setCategory,
    setOwnerId,
    setSort,
    setPageCursor,
    clear,
  };
}
