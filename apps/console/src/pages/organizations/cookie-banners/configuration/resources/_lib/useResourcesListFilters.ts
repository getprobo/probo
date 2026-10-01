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

export const trackerResourceTypes = [
  "SCRIPT",
  "IFRAME",
  "IMAGE",
  "STYLESHEET",
  "FONT",
  "BEACON",
  "FETCH",
  "MEDIA",
  "SERVICE_WORKER",
] as const;
export type TrackerResourceType = (typeof trackerResourceTypes)[number];

export type ResourcesListGraphqlFilter = {
  query: string | null;
  type: TrackerResourceType | null;
};

export function isTrackerResourceType(value: string): value is TrackerResourceType {
  return (trackerResourceTypes as readonly string[]).includes(value);
}

export function resourcesListGraphqlFilter(filters: {
  query: string;
  type: TrackerResourceType | null;
}): ResourcesListGraphqlFilter | null {
  const query = filters.query.trim() || null;
  if (query == null && filters.type == null) {
    return null;
  }

  return {
    query,
    type: filters.type,
  };
}

export interface ResourcesListFilters {
  query: string;
  type: TrackerResourceType | null;
  graphqlFilter: ResourcesListGraphqlFilter | null;
  hasActiveFilters: boolean;
  setQuery: (value: string) => void;
  setType: (value: TrackerResourceType | null) => void;
}

export function useResourcesListFilters(): ResourcesListFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const query = searchParams.get("q") ?? "";
  const rawType = searchParams.get("type") ?? "";
  const type = isTrackerResourceType(rawType) ? rawType : null;

  const graphqlFilter = useMemo(
    () => resourcesListGraphqlFilter({ query, type }),
    [query, type],
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
  const setType = useCallback((value: TrackerResourceType | null) => {
    setParam("type", value ?? "");
  }, [setParam]);

  return {
    query,
    type,
    graphqlFilter,
    hasActiveFilters: query !== "" || type != null,
    setQuery,
    setType,
  };
}
