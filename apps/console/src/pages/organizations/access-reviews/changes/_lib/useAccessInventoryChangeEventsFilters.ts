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

export const accessInventoryChangeTypes = [
  "APPEARED",
  "DISAPPEARED",
  "ATTRIBUTE_CHANGED",
] as const;

export type AccessInventoryChangeTypeFilter
  = (typeof accessInventoryChangeTypes)[number];

const changeTypeParam = {
  APPEARED: "appeared",
  DISAPPEARED: "disappeared",
  ATTRIBUTE_CHANGED: "attribute-changed",
} as const;

const isoDatePattern = /^\d{4}-\d{2}-\d{2}$/;

function isChangeTypeFilter(
  value: string,
): value is AccessInventoryChangeTypeFilter {
  return (accessInventoryChangeTypes as readonly string[]).includes(value);
}

function changeTypeFromParam(
  value: string,
): AccessInventoryChangeTypeFilter | null {
  if (value === "appeared") {
    return "APPEARED";
  }
  if (value === "disappeared") {
    return "DISAPPEARED";
  }
  if (value === "attribute-changed") {
    return "ATTRIBUTE_CHANGED";
  }
  return null;
}

function isoDateFromParam(value: string | null): string | null {
  if (value == null || value === "") {
    return null;
  }
  if (!isoDatePattern.test(value)) {
    return null;
  }
  const parsed = Date.parse(`${value}T00:00:00.000Z`);
  if (Number.isNaN(parsed)) {
    return null;
  }
  return value;
}

function startOfUtcDay(date: string): string {
  return `${date}T00:00:00.000Z`;
}

function endOfUtcDay(date: string): string {
  return `${date}T23:59:59.999Z`;
}

function accessInventoryChangeEventsGraphqlFilter(filters: {
  sourceId: string | null;
  changeType: AccessInventoryChangeTypeFilter | null;
  observedAfter: string | null;
  observedBefore: string | null;
}) {
  return {
    accessReviewSourceId: filters.sourceId,
    changeTypes: filters.changeType != null ? [filters.changeType] : null,
    observedAtOnOrAfter: filters.observedAfter != null
      ? startOfUtcDay(filters.observedAfter)
      : null,
    observedAtOnOrBefore: filters.observedBefore != null
      ? endOfUtcDay(filters.observedBefore)
      : null,
  };
}

export type AccessInventoryChangeEventsGraphqlFilter = ReturnType<
  typeof accessInventoryChangeEventsGraphqlFilter
>;

export interface AccessInventoryChangeEventsFilters {
  sourceId: string | null;
  changeType: AccessInventoryChangeTypeFilter | null;
  observedAfter: string | null;
  observedBefore: string | null;
  graphqlFilter: AccessInventoryChangeEventsGraphqlFilter;
  hasActiveFilters: boolean;
  setSourceId: (value: string | null) => void;
  setChangeType: (value: AccessInventoryChangeTypeFilter | null) => void;
  setObservedAfter: (value: string | null) => void;
  setObservedBefore: (value: string | null) => void;
}

export function useAccessInventoryChangeEventsFilters(): AccessInventoryChangeEventsFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const sourceId = searchParams.get("source");
  const changeType = changeTypeFromParam(searchParams.get("changeType") ?? "");
  const observedAfter = isoDateFromParam(searchParams.get("observedAfter"));
  const observedBefore = isoDateFromParam(searchParams.get("observedBefore"));

  const graphqlFilter = useMemo(
    () => accessInventoryChangeEventsGraphqlFilter({
      sourceId,
      changeType,
      observedAfter,
      observedBefore,
    }),
    [sourceId, changeType, observedAfter, observedBefore],
  );

  const setSourceId = useCallback((value: string | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("source", value);
      } else {
        next.delete("source");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setChangeType = useCallback((
    value: AccessInventoryChangeTypeFilter | null,
  ) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value != null && isChangeTypeFilter(value)) {
        next.set("changeType", changeTypeParam[value]);
      } else {
        next.delete("changeType");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setObservedAfter = useCallback((value: string | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      const date = isoDateFromParam(value);
      if (date != null) {
        next.set("observedAfter", date);
      } else {
        next.delete("observedAfter");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setObservedBefore = useCallback((value: string | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      const date = isoDateFromParam(value);
      if (date != null) {
        next.set("observedBefore", date);
      } else {
        next.delete("observedBefore");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    sourceId,
    changeType,
    observedAfter,
    observedBefore,
    graphqlFilter,
    hasActiveFilters: sourceId != null
      || changeType != null
      || observedAfter != null
      || observedBefore != null,
    setSourceId,
    setChangeType,
    setObservedAfter,
    setObservedBefore,
  };
}
