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

import type { RiskObligationsPageQuery } from "#/__generated__/core/RiskObligationsPageQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";

import RiskObligationsPage, {
  riskObligationsPageQuery,
} from "./RiskObligationsPage";

export default function RiskObligationsPageLoader() {
  const { riskId } = useParams<{ riskId: string }>();
  const [queryRef, loadQuery]
    = useQueryLoader<RiskObligationsPageQuery>(riskObligationsPageQuery);

  useEffect(() => {
    if (riskId) {
      loadQuery({ riskId });
    }
  }, [loadQuery, riskId]);

  if (!queryRef) {
    return <LinkCardSkeleton />;
  }

  return (
    <Suspense fallback={<LinkCardSkeleton />}>
      <RiskObligationsPage queryRef={queryRef} />
    </Suspense>
  );
}
