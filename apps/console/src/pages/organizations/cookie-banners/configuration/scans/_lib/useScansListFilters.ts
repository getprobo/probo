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

import type { CursorPaginationVariables } from "#/lib/relay/useCursorPagination";

import { SCANS_PAGE_SIZE } from "./pageSize";
import { type CookieBannerScanState, isCookieBannerScanState } from "./scanStates";

export type OrderDirection = "ASC" | "DESC";

export type ScansListOrder = {
  field: "CREATED_AT";
  direction: OrderDirection;
};

export type ScansListGraphqlFilter = {
  state: CookieBannerScanState;
} | null;

const defaultDirection: OrderDirection = "DESC";

export function scansListGraphqlPagination(
  after: string | null,
  before: string | null,
): CursorPaginationVariables {
  if (after != null) {
    return {
      first: SCANS_PAGE_SIZE,
      after,
      last: null,
      before: null,
    };
  }
  if (before != null) {
    return {
      first: null,
      after: null,
      last: SCANS_PAGE_SIZE,
      before,
    };
  }
  return {
    first: SCANS_PAGE_SIZE,
    after: null,
    last: null,
    before: null,
  };
}

function clearPagination(params: URLSearchParams) {
  params.delete("after");
  params.delete("before");
}

function isOrderDirection(value: string | null): value is OrderDirection {
  return value === "ASC" || value === "DESC";
}

export function useScansListFilters() {
  const [searchParams, setSearchParams] = useSearchParams();

  const state = useMemo(() => {
    const value = searchParams.get("state");
    return isCookieBannerScanState(value) ? value : null;
  }, [searchParams]);

  const direction = useMemo(() => {
    const value = searchParams.get("dir");
    return isOrderDirection(value) ? value : defaultDirection;
  }, [searchParams]);

  const after = searchParams.get("after");
  const before = searchParams.get("before");

  const graphqlFilter = useMemo<ScansListGraphqlFilter>(
    () => (state == null ? null : { state }),
    [state],
  );

  const graphqlOrder = useMemo<ScansListOrder>(
    () => ({ field: "CREATED_AT", direction }),
    [direction],
  );

  const graphqlPagination = useMemo(
    () => scansListGraphqlPagination(after, before),
    [after, before],
  );

  const setState = useCallback((value: CookieBannerScanState | null) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value == null) {
        next.delete("state");
      } else {
        next.set("state", value);
      }
      clearPagination(next);
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const toggleDirection = useCallback(() => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      const current = isOrderDirection(previous.get("dir")) ? previous.get("dir") : defaultDirection;
      if (current === "DESC") {
        next.set("dir", "ASC");
      } else {
        next.delete("dir");
      }
      clearPagination(next);
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setAfter = useCallback((cursor: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.set("after", cursor);
      next.delete("before");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const setBefore = useCallback((cursor: string) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      next.set("before", cursor);
      next.delete("after");
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  const resetPagination = useCallback(() => {
    setSearchParams((previous) => {
      if (!previous.has("after") && !previous.has("before")) {
        return previous;
      }
      const next = new URLSearchParams(previous);
      clearPagination(next);
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    state,
    direction,
    graphqlFilter,
    graphqlOrder,
    graphqlPagination,
    hasStateFilter: state != null,
    setState,
    toggleDirection,
    setAfter,
    setBefore,
    resetPagination,
  };
}
