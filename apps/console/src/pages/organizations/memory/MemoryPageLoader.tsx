// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
import { type PreloadedQuery, usePreloadedQuery, useQueryLoader } from "react-relay";
import { graphql } from "relay-runtime";

import type { MemoryPageLoaderQuery } from "#/__generated__/core/MemoryPageLoaderQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import MemoryPage from "./MemoryPage";

const memoryPageQuery = graphql`
  query MemoryPageLoaderQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...MemoryPageFragment
      }
    }
  }
`;

function MemoryPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<MemoryPageLoaderQuery>(memoryPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [organizationId, loadQuery]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <MemoryPageInner queryRef={queryRef} />;
}

function MemoryPageInner({ queryRef }: { queryRef: PreloadedQuery<MemoryPageLoaderQuery> }) {
  const data = usePreloadedQuery(memoryPageQuery, queryRef);

  return <MemoryPage organization={data.organization} />;
}

export default function MemoryPageLoader() {
  return (
    <CoreRelayProvider>
      <MemoryPageQueryLoader />
    </CoreRelayProvider>
  );
}
