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

import { Suspense, useEffect } from "react";
import { useQueryLoader } from "react-relay";
import { useParams } from "react-router";

import type { ThirdPartyDetailLayoutQuery } from "#/__generated__/core/ThirdPartyDetailLayoutQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import ThirdPartyDetailLayout, { thirdPartyDetailLayoutQuery } from "./ThirdPartyDetailLayout";

export default function ThirdPartyDetailLayoutLoader() {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<ThirdPartyDetailLayoutQuery>(thirdPartyDetailLayoutQuery);

  useEffect(() => {
    if (thirdPartyId) {
      loadQuery({ thirdPartyId });
    }
  }, [loadQuery, thirdPartyId]);

  if (!queryRef) {
    return <PageSkeleton />;
  }

  return (
    <Suspense fallback={<PageSkeleton />}>
      <ThirdPartyDetailLayout queryRef={queryRef} />
    </Suspense>
  );
}
