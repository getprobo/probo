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

import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { usePaginatedRegistryAssetsFragment$key } from "#/__generated__/core/usePaginatedRegistryAssetsFragment.graphql";
import type { usePaginatedRegistryAssetsQuery } from "#/__generated__/core/usePaginatedRegistryAssetsQuery.graphql";
import type { usePaginatedRegistryAssetsQuery_fragment } from "#/__generated__/core/usePaginatedRegistryAssetsQuery_fragment.graphql";

/* eslint-disable relay/unused-fields */

const registryAssetsQuery = graphql`
  query usePaginatedRegistryAssetsQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...usePaginatedRegistryAssetsFragment
      }
    }
  }
`;

const assetsFragment = graphql`
  fragment usePaginatedRegistryAssetsFragment on Organization
  @refetchable(queryName: "usePaginatedRegistryAssetsQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    registryAssets(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: { direction: ASC, field: NAME }
    ) @connection(key: "usePaginatedRegistryAssetsQuery_registryAssets") {
      edges {
        node {
          id
          name
        }
      }
    }
  }
`;

export function usePaginatedRegistryAssets(organizationId: string) {
  const query = useLazyLoadQuery<usePaginatedRegistryAssetsQuery>(
    registryAssetsQuery,
    { organizationId },
    { fetchPolicy: "network-only" },
  );
  return usePaginationFragment<usePaginatedRegistryAssetsQuery_fragment, usePaginatedRegistryAssetsFragment$key>(
    assetsFragment,
    query.organization as usePaginatedRegistryAssetsFragment$key,
  );
}
