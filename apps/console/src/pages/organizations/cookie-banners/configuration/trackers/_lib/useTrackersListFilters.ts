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

import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";

export const cookieSources = ["SCRIPT", "PRE_EXISTING", "HTTP", "EXTENSION"] as const;
export type CookieSource = (typeof cookieSources)[number];

export const trackerTypes = [
  "COOKIE",
  "LOCAL_STORAGE",
  "SESSION_STORAGE",
  "INDEXED_DB",
  "CACHE_STORAGE",
] as const;
export type TrackerType = (typeof trackerTypes)[number];

export const trackerPatternOrderFields = ["NAME", "SOURCE", "LAST_MATCHED_AT"] as const;
export type TrackerPatternOrderField = (typeof trackerPatternOrderFields)[number];
export type OrderDirection = "ASC" | "DESC";

export type TrackersListOrder = {
  field: TrackerPatternOrderField;
  direction: OrderDirection;
};

export const defaultTrackersListOrder: TrackersListOrder = {
  field: "NAME",
  direction: "ASC",
};

const firstTrackersListDirection: Record<TrackerPatternOrderField, OrderDirection> = {
  NAME: "ASC",
  SOURCE: "ASC",
  LAST_MATCHED_AT: "DESC",
};

export type TrackersListGraphqlFilter = {
  query: string | null;
  source: CookieSource | null;
  trackerType: TrackerType | null;
  cookieCategoryId: string | null;
  thirdPartyId: string | null;
};

export function isCookieSource(value: string): value is CookieSource {
  return (cookieSources as readonly string[]).includes(value);
}

export function isTrackerType(value: string): value is TrackerType {
  return (trackerTypes as readonly string[]).includes(value);
}

export function isTrackerPatternOrderField(value: string): value is TrackerPatternOrderField {
  return (trackerPatternOrderFields as readonly string[]).includes(value);
}

export function isOrderDirection(value: string): value is OrderDirection {
  return value === "ASC" || value === "DESC";
}

export function trackersListHeaderSort(
  field: TrackerPatternOrderField,
  order: TrackersListOrder,
): "ascending" | "descending" | "none" {
  if (order.field !== field) {
    return "none";
  }
  return order.direction === "ASC" ? "ascending" : "descending";
}

function writeTrackersListOrder(
  params: URLSearchParams,
  field: TrackerPatternOrderField,
  direction: OrderDirection,
) {
  if (
    field === defaultTrackersListOrder.field
    && direction === defaultTrackersListOrder.direction
  ) {
    params.delete("sort");
    params.delete("dir");
    return;
  }

  params.set("sort", field);
  params.set("dir", direction);
}

export function trackersListGraphqlFilter(filters: {
  query: string;
  source: CookieSource | null;
  type: TrackerType | null;
  category: string | null;
  party: string | null;
}): TrackersListGraphqlFilter | null {
  const query = filters.query.trim() || null;
  if (
    query == null
    && filters.source == null
    && filters.type == null
    && filters.category == null
    && filters.party == null
  ) {
    return null;
  }

  return {
    query,
    source: filters.source,
    trackerType: filters.type,
    cookieCategoryId: filters.category,
    thirdPartyId: filters.party,
  };
}

export interface TrackersListFilters {
  query: string;
  source: CookieSource | null;
  type: TrackerType | null;
  category: string | null;
  party: string | null;
  graphqlFilter: TrackersListGraphqlFilter | null;
  graphqlOrder: TrackersListOrder;
  hasActiveFilters: boolean;
  setQuery: (value: string) => void;
  setSource: (value: CookieSource | null) => void;
  setType: (value: TrackerType | null) => void;
  setCategory: (value: string | null) => void;
  setParty: (value: string | null) => void;
  setOrder: (field: TrackerPatternOrderField) => void;
}

export function useTrackersListFilters(): TrackersListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawSource = searchParams.get("source") ?? "";
  const rawType = searchParams.get("type") ?? "";
  const rawCategory = searchParams.get("category") ?? "";
  const rawParty = searchParams.get("party") ?? "";
  const rawSort = searchParams.get("sort") ?? "";
  const rawDir = searchParams.get("dir") ?? "";
  const source = isCookieSource(rawSource) ? rawSource : null;
  const type = isTrackerType(rawType) ? rawType : null;
  const category = rawCategory === "" ? null : rawCategory;
  const party = rawParty === "" ? null : rawParty;
  const field = isTrackerPatternOrderField(rawSort)
    ? rawSort
    : defaultTrackersListOrder.field;
  const direction = isTrackerPatternOrderField(rawSort) && isOrderDirection(rawDir)
    ? rawDir
    : firstTrackersListDirection[field];
  const graphqlOrder = useMemo(
    () => ({ field, direction }),
    [direction, field],
  );

  const graphqlFilter = useMemo(
    () => trackersListGraphqlFilter({ query, source, type, category, party }),
    [query, source, type, category, party],
  );

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
  const setSource = useCallback((value: CookieSource | null) => {
    setParam("source", value ?? "");
  }, [setParam]);
  const setType = useCallback((value: TrackerType | null) => {
    setParam("type", value ?? "");
  }, [setParam]);
  const setCategory = useCallback((value: string | null) => {
    setParam("category", value ?? "");
  }, [setParam]);
  const setParty = useCallback((value: string | null) => {
    setParam("party", value ?? "");
  }, [setParam]);

  const setOrder = useCallback((nextField: TrackerPatternOrderField) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      const firstDirection = firstTrackersListDirection[nextField];
      if (nextField !== field) {
        writeTrackersListOrder(next, nextField, firstDirection);
        return next;
      }
      if (direction === firstDirection) {
        writeTrackersListOrder(next, nextField, direction === "ASC" ? "DESC" : "ASC");
        return next;
      }
      writeTrackersListOrder(
        next,
        defaultTrackersListOrder.field,
        defaultTrackersListOrder.direction,
      );
      return next;
    }, { replace: true });
  }, [direction, field, setSearchParams]);

  return {
    query,
    source,
    type,
    category,
    party,
    graphqlFilter,
    graphqlOrder,
    hasActiveFilters:
      query !== ""
      || source != null
      || type != null
      || category != null
      || party != null,
    setQuery,
    setSource,
    setType,
    setCategory,
    setParty,
    setOrder,
  };
}
