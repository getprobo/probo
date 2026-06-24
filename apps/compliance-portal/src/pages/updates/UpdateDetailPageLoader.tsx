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
