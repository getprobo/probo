// Copyright (c) 2025-2026 Probo Inc <hello@probo.com>.
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

import type { usePaginatedMeasuresFragment$key } from "#/__generated__/core/usePaginatedMeasuresFragment.graphql";
import type { usePaginatedMeasuresQuery } from "#/__generated__/core/usePaginatedMeasuresQuery.graphql";
import type { usePaginatedMeasuresQuery_fragment } from "#/__generated__/core/usePaginatedMeasuresQuery_fragment.graphql";

/* eslint-disable relay/unused-fields */

const measuresQuery = graphql`
  query usePaginatedMeasuresQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...usePaginatedMeasuresFragment
      }
    }
  }
`;

const measuresFragment = graphql`
  fragment usePaginatedMeasuresFragment on Organization
  @refetchable(queryName: "usePaginatedMeasuresQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "MeasureOrder", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    measures(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
    ) @connection(key: "usePaginatedMeasuresQuery_measures") {
      edges {
        node {
          id
          name
          state
          description
          category
        }
      }
    }
  }
`;

/**
 * Hook to retrieve measured paginated (used for link dialog and measure selectors)
 */
export function usePaginatedMeasures(organizationId: string) {
  const query = useLazyLoadQuery<usePaginatedMeasuresQuery>(
    measuresQuery,
    {
      organizationId,
    },
    { fetchPolicy: "network-only" },
  );
  return usePaginationFragment<usePaginatedMeasuresQuery_fragment, usePaginatedMeasuresFragment$key>(
    measuresFragment,
    query.organization as usePaginatedMeasuresFragment$key,
  );
}
