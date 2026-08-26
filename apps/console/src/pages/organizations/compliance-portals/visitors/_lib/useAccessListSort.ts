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

import { useCallback } from "react";
import { useSearchParams } from "react-router";

export type AccessListSort = "requests" | "joined";

const orders = {
  requests: { field: "PENDING_REQUEST_COUNT", direction: "DESC" },
  joined: { field: "CREATED_AT", direction: "DESC" },
} as const;

export interface AccessListSortState {
  sort: AccessListSort;
  order: (typeof orders)[AccessListSort];
  setSort: (value: AccessListSort) => void;
}

export function useAccessListSort(): AccessListSortState {
  const [searchParams, setSearchParams] = useSearchParams();
  const sort: AccessListSort = searchParams.get("sort") === "joined" ? "joined" : "requests";

  const setSort = useCallback((value: AccessListSort) => {
    setSearchParams((previous) => {
      const next = new URLSearchParams(previous);
      if (value === "requests") {
        next.delete("sort");
      } else {
        next.set("sort", value);
      }
      return next;
    }, { replace: true });
  }, [setSearchParams]);

  return {
    sort,
    order: orders[sort],
    setSort,
  };
}
