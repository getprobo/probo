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

export const accessInventoryGrantPresences = ["PRESENT", "DISAPPEARED"] as const;
export type AccessInventoryGrantPresenceFilter
  = (typeof accessInventoryGrantPresences)[number];

const presenceParam = {
  PRESENT: "present",
  DISAPPEARED: "disappeared",
} as const;

function isPresenceFilter(value: string): value is AccessInventoryGrantPresenceFilter {
  return (accessInventoryGrantPresences as readonly string[]).includes(value);
}

function presenceFromParam(value: string): AccessInventoryGrantPresenceFilter | null {
  if (value === "present") {
    return "PRESENT";
  }
  if (value === "disappeared") {
    return "DISAPPEARED";
  }
  return null;
}

function accessInventoryGrantsGraphqlFilter(filters: {
  sourceId: string | null;
  presence: AccessInventoryGrantPresenceFilter | null;
  includeDisappeared: boolean;
}) {
  return {
    accessReviewSourceId: filters.sourceId,
    principalId: null,
    resourceId: null,
    presence: filters.presence,
    includeDisappeared: filters.includeDisappeared ? true : null,
  };
}

export type AccessInventoryGrantsGraphqlFilter = ReturnType<
  typeof accessInventoryGrantsGraphqlFilter
>;

export interface AccessInventoryGrantsFilters {
  sourceId: string | null;
  presence: AccessInventoryGrantPresenceFilter | null;
  includeDisappeared: boolean;
  graphqlFilter: AccessInventoryGrantsGraphqlFilter;
  hasActiveFilters: boolean;
  setSourceId: (value: string | null) => void;
  setPresence: (value: AccessInventoryGrantPresenceFilter | null) => void;
  setIncludeDisappeared: (value: boolean) => void;
}

export function useAccessInventoryGrantsFilters(): AccessInventoryGrantsFilters {
  const [searchParams, setSearchParams] = useSearchParams();
  const sourceId = searchParams.get("source");
  const presence = presenceFromParam(searchParams.get("presence") ?? "");
  const includeDisappeared = searchParams.get("includeDisappeared") === "1";

  const graphqlFilter = useMemo(
    () => accessInventoryGrantsGraphqlFilter({
      sourceId,
      presence,
      includeDisappeared,
    }),
    [sourceId, presence, includeDisappeared],
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

  const setPresence = useCallback((value: AccessInventoryGrantPresenceFilter | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value != null && isPresenceFilter(value)) {
        next.set("presence", presenceParam[value]);
      } else {
        next.delete("presence");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setIncludeDisappeared = useCallback((value: boolean) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value) {
        next.set("includeDisappeared", "1");
      } else {
        next.delete("includeDisappeared");
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    sourceId,
    presence,
    includeDisappeared,
    graphqlFilter,
    hasActiveFilters: sourceId != null || presence != null || includeDisappeared,
    setSourceId,
    setPresence,
    setIncludeDisappeared,
  };
}
