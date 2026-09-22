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

import { useCallback, useTransition } from "react";

// The `pageInfo` shape a Relay connection exposes for bidirectional cursor
// pagination. Structurally compatible with generated connection page info.
export interface CursorPageInfo {
  hasPreviousPage: boolean;
  hasNextPage: boolean;
  startCursor: string | null | undefined;
  endCursor: string | null | undefined;
}

// The connection pagination arguments passed to a refetch.
export interface CursorPaginationVariables {
  first?: number | null;
  after?: string | null;
  last?: number | null;
  before?: string | null;
}

type CursorRefetch = (variables: CursorPaginationVariables) => void;

// Prev/Next pagination for a Relay cursor connection. Drives the refetch inside
// a transition so the current page stays mounted (dimmed) while the next one
// loads; Prev/Next availability comes from the server `pageInfo`, so it stays
// correct however the page is reached. There is no page-number counter: cursor
// pagination encodes a position, not an ordinal, so a reliable page index
// (deep-linkable or refresh-safe) would need offset + totalCount.
export function useCursorPagination(
  refetch: CursorRefetch,
  pageInfo: CursorPageInfo,
  pageSize: number,
) {
  const [isPending, startTransition] = useTransition();

  const goNext = useCallback(() => {
    if (!pageInfo.hasNextPage || pageInfo.endCursor == null) {
      return;
    }
    startTransition(() => {
      refetch({ first: pageSize, after: pageInfo.endCursor, last: null, before: null });
    });
  }, [refetch, pageSize, pageInfo.hasNextPage, pageInfo.endCursor]);

  const goPrevious = useCallback(() => {
    if (!pageInfo.hasPreviousPage || pageInfo.startCursor == null) {
      return;
    }
    startTransition(() => {
      refetch({ first: null, after: null, last: pageSize, before: pageInfo.startCursor });
    });
  }, [refetch, pageSize, pageInfo.hasPreviousPage, pageInfo.startCursor]);

  return { isPending, goPrevious, goNext };
}
