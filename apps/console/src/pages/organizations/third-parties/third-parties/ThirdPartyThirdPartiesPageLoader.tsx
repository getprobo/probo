// Copyright (c) 2025-2026 Probo Inc <hello@getprobo.com>.
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

import type { ThirdPartyThirdPartiesPageQuery } from "#/__generated__/core/ThirdPartyThirdPartiesPageQuery.graphql";
import { PageSkeleton } from "#/components/skeletons/PageSkeleton";

import ThirdPartyThirdPartiesPage, { thirdPartyThirdPartiesPageQuery } from "./ThirdPartyThirdPartiesPage";

export default function ThirdPartyThirdPartiesPageLoader() {
  const { thirdPartyId } = useParams<{ thirdPartyId: string }>();
  const [queryRef, loadQuery] = useQueryLoader<ThirdPartyThirdPartiesPageQuery>(thirdPartyThirdPartiesPageQuery);

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
      <ThirdPartyThirdPartiesPage queryRef={queryRef} />
    </Suspense>
  );
}
