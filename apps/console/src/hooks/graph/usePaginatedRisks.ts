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

import type { usePaginatedRisksFragment$key } from "#/__generated__/core/usePaginatedRisksFragment.graphql";
import type { usePaginatedRisksQuery } from "#/__generated__/core/usePaginatedRisksQuery.graphql";
import type { usePaginatedRisksQuery_fragment } from "#/__generated__/core/usePaginatedRisksQuery_fragment.graphql";

/* eslint-disable relay/unused-fields */

const risksQuery = graphql`
  query usePaginatedRisksQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...usePaginatedRisksFragment
      }
    }
  }
`;

const risksFragment = graphql`
  fragment usePaginatedRisksFragment on Organization
  @refetchable(queryName: "usePaginatedRisksQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 50 }
    order: { type: "RiskOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
    filter: { type: "RiskFilter", defaultValue: null }
  ) {
    risks(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "usePaginatedRisksQuery_risks", filters: ["filter"]) {
      edges {
        node {
          id
          name
          category
          description
        }
      }
    }
  }
`;

/**
 * Hook to retrieve risks paginated (used for risk selectors)
 */
export function usePaginatedRisks(organizationId: string) {
  const query = useLazyLoadQuery<usePaginatedRisksQuery>(
    risksQuery,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  return usePaginationFragment<usePaginatedRisksQuery_fragment, usePaginatedRisksFragment$key>(
    risksFragment,
    query.organization as usePaginatedRisksFragment$key,
  );
}
