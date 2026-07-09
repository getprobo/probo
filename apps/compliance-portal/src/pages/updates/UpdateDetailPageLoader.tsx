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

import { useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { UpdateDetailPageQuery } from "./__generated__/UpdateDetailPageQuery.graphql";
import { UpdateDetailPage, updateDetailPageQuery } from "./UpdateDetailPage";
import { UpdateDetailPageSkeleton } from "./UpdateDetailPageSkeleton";

export default function UpdateDetailPageLoader() {
  const { updateId } = useParams<{ updateId: string }>();
  const [queryRef, loadQuery, disposeQuery] = useQueryLoader<UpdateDetailPageQuery>(updateDetailPageQuery);

  // Dispose on updateId change so navigating between updates shows the skeleton
  // during the transition instead of the previous update's content.
  useEffect(() => {
    if (updateId) {
      loadQuery({ updateId });
    }
    return () => disposeQuery();
  }, [loadQuery, disposeQuery, updateId]);

  if (!queryRef) {
    return <UpdateDetailPageSkeleton />;
  }

  return <UpdateDetailPage queryRef={queryRef} />;
}
