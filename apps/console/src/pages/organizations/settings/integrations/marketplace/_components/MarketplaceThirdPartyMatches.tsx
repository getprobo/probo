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

import { ErrorBoundary } from "@probo/ui/src/v2/ErrorBoundary/ErrorBoundary";
import { Suspense, useEffect, useMemo, useRef } from "react";
import {
  graphql,
  type PreloadedQuery,
  useFragment,
  usePaginationFragment,
  usePreloadedQuery,
  useQueryLoader,
} from "react-relay";

import type { MarketplaceThirdPartyMatches_organization$key } from "#/__generated__/core/MarketplaceThirdPartyMatches_organization.graphql";
import type { MarketplaceThirdPartyMatches_query$key } from "#/__generated__/core/MarketplaceThirdPartyMatches_query.graphql";
import type { MarketplaceThirdPartyMatchesPaginationQuery } from "#/__generated__/core/MarketplaceThirdPartyMatchesPaginationQuery.graphql";
import type { MarketplaceThirdPartyMatchesQuery } from "#/__generated__/core/MarketplaceThirdPartyMatchesQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  type ConnectorProviderName,
  matchingConnectorProviders,
} from "../_lib/thirdPartyConnectorMatch";

const THIRD_PARTY_PAGE_SIZE = 500;

const marketplaceThirdPartyMatchesQueryFragment = graphql`
  fragment MarketplaceThirdPartyMatches_query on Query {
    connectorProviders {
      provider
      displayName
    }
  }
`;

const marketplaceThirdPartyMatchesQuery = graphql`
  query MarketplaceThirdPartyMatchesQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        ...MarketplaceThirdPartyMatches_organization
      }
    }
  }
`;

const marketplaceThirdPartyMatchesFragment = graphql`
  fragment MarketplaceThirdPartyMatches_organization on Organization
  @refetchable(queryName: "MarketplaceThirdPartyMatchesPaginationQuery")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 500 }
    after: { type: "CursorKey", defaultValue: null }
    filter: { type: "ThirdPartyFilter", defaultValue: { level: 1 } }
  ) {
    # eslint-disable-next-line relay/unused-fields -- required by @refetchable
    id
    thirdParties(first: $first, after: $after, filter: $filter)
      @connection(key: "MarketplaceThirdPartyMatches_thirdParties", filters: ["filter"]) {
      edges {
        node {
          # eslint-disable-next-line relay/unused-fields -- required by @connection
          id
          name
        }
      }
    }
  }
`;

interface MarketplaceThirdPartyMatchesProps {
  queryKey: MarketplaceThirdPartyMatches_query$key;
  onMatched: (providers: ReadonlySet<string>) => void;
}

export function MarketplaceThirdPartyMatches({
  queryKey,
  onMatched,
}: MarketplaceThirdPartyMatchesProps) {
  const organizationId = useOrganizationId();
  const [queryRef, loadQuery]
    = useQueryLoader<MarketplaceThirdPartyMatchesQuery>(marketplaceThirdPartyMatchesQuery);

  useEffect(() => {
    loadQuery({ organizationId });
  }, [loadQuery, organizationId]);

  const currentQueryRef = queryRef != null
    && queryRef.variables.organizationId === organizationId
    ? queryRef
    : null;

  if (currentQueryRef == null) {
    return null;
  }

  return (
    // Best effort. A failed query shows no matches and leaves the marketplace up.
    <ErrorBoundary fallback={null}>
      <Suspense fallback={null}>
        <MarketplaceThirdPartyMatchesContent
          queryRef={currentQueryRef}
          queryKey={queryKey}
          onMatched={onMatched}
        />
      </Suspense>
    </ErrorBoundary>
  );
}

interface MarketplaceThirdPartyMatchesContentProps {
  queryRef: PreloadedQuery<MarketplaceThirdPartyMatchesQuery>;
  queryKey: MarketplaceThirdPartyMatches_query$key;
  onMatched: (providers: ReadonlySet<string>) => void;
}

function MarketplaceThirdPartyMatchesContent({
  queryRef,
  queryKey,
  onMatched,
}: MarketplaceThirdPartyMatchesContentProps) {
  const data = usePreloadedQuery<MarketplaceThirdPartyMatchesQuery>(
    marketplaceThirdPartyMatchesQuery,
    queryRef,
  );
  const { connectorProviders: providers } = useFragment(
    marketplaceThirdPartyMatchesQueryFragment,
    queryKey,
  );
  const connectorProviders = useMemo(
    () => providers.map(provider => ({
      provider: provider.provider,
      displayName: provider.displayName,
    })),
    [providers],
  );

  if (data.organization?.__typename !== "Organization") {
    return null;
  }

  return (
    <MarketplaceThirdPartyMatchPages
      organizationKey={data.organization}
      connectorProviders={connectorProviders}
      onMatched={onMatched}
    />
  );
}

interface MarketplaceThirdPartyMatchPagesProps {
  organizationKey: MarketplaceThirdPartyMatches_organization$key;
  connectorProviders: readonly ConnectorProviderName[];
  onMatched: (providers: ReadonlySet<string>) => void;
}

function MarketplaceThirdPartyMatchPages({
  organizationKey,
  connectorProviders,
  onMatched,
}: MarketplaceThirdPartyMatchPagesProps) {
  const failedRef = useRef(false);
  const { data, loadNext, hasNext, isLoadingNext } = usePaginationFragment<
    MarketplaceThirdPartyMatchesPaginationQuery,
    MarketplaceThirdPartyMatches_organization$key
  >(marketplaceThirdPartyMatchesFragment, organizationKey);

  useEffect(() => {
    if (failedRef.current) {
      return;
    }

    if (hasNext) {
      if (!isLoadingNext) {
        loadNext(THIRD_PARTY_PAGE_SIZE, {
          onComplete: (error) => {
            if (error != null) {
              failedRef.current = true;
            }
          },
        });
      }
      return;
    }

    if (isLoadingNext) {
      return;
    }

    onMatched(matchingConnectorProviders(
      connectorProviders,
      data.thirdParties.edges.map(edge => edge.node.name),
    ));
  }, [connectorProviders, data, hasNext, isLoadingNext, loadNext, onMatched]);

  return null;
}
