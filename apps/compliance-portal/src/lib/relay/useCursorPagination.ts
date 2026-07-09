// Copyright (c) 2026 Probo Inc <hello@probo.com>.
//
// Permission to use, copy, modify, and/or distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY
// AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM
// LOSS OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR
// OTHER TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR
// PERFORMANCE OF THIS SOFTWARE.

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
