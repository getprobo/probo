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

import type { ContextPageLoaderQuery } from "#/__generated__/core/ContextPageLoaderQuery.graphql";
import { LinkCardSkeleton } from "#/components/skeletons/LinkCardSkeleton";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { CoreRelayProvider } from "#/providers/CoreRelayProvider";

import ContextPage from "./ContextPage";

const contextPageQuery = graphql`
  query ContextPageLoaderQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        ...ContextPageFragment
      }
    }
  }
`;

function ContextPageQueryLoader() {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery] = useQueryLoader<ContextPageLoaderQuery>(contextPageQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [organizationId, loadQuery]);

  if (!queryRef) return <LinkCardSkeleton />;

  return <ContextPageInner queryRef={queryRef} />;
}

function ContextPageInner({ queryRef }: { queryRef: PreloadedQuery<ContextPageLoaderQuery> }) {
  const data = usePreloadedQuery(contextPageQuery, queryRef);

  return <ContextPage organization={data.organization} />;
}

export default function ContextPageLoader() {
  return (
    <CoreRelayProvider>
      <ContextPageQueryLoader />
    </CoreRelayProvider>
  );
}
