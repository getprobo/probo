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

// Page size for the cursor-paginated updates list. Matches the Figma list frame.
export const UPDATES_PAGE_SIZE = 10;

export interface UpdatesPageInfo {
  hasPreviousPage: boolean;
  hasNextPage: boolean;
  startCursor: string | null | undefined;
  endCursor: string | null | undefined;
}

export interface UpdatesPaginationVariables {
  first?: number | null;
  after?: string | null;
  last?: number | null;
  before?: string | null;
}

type RefetchUpdates = (variables: UpdatesPaginationVariables) => void;

// Cursor-based Prev/Next pagination for the updates list. Drives the connection
// refetch inside a transition so the current page stays mounted (dimmed) while
// the next one loads. Enabled/disabled state comes from the server `pageInfo`,
// which stays correct however the page is reached. No page-number counter:
// cursor pagination encodes a position, not an ordinal, so a reliable page
// index (deep-linkable or refresh-safe) would need offset + totalCount, which
// this API does not expose.
export function useUpdatesPagination(refetch: RefetchUpdates, pageInfo: UpdatesPageInfo) {
  const [isPending, startTransition] = useTransition();

  const goNext = useCallback(() => {
    if (!pageInfo.hasNextPage || pageInfo.endCursor == null) {
      return;
    }
    startTransition(() => {
      refetch({ first: UPDATES_PAGE_SIZE, after: pageInfo.endCursor, last: null, before: null });
    });
  }, [refetch, pageInfo.hasNextPage, pageInfo.endCursor]);

  const goPrevious = useCallback(() => {
    if (!pageInfo.hasPreviousPage || pageInfo.startCursor == null) {
      return;
    }
    startTransition(() => {
      refetch({ first: null, after: null, last: UPDATES_PAGE_SIZE, before: pageInfo.startCursor });
    });
  }, [refetch, pageInfo.hasPreviousPage, pageInfo.startCursor]);

  return { isPending, goPrevious, goNext };
}
