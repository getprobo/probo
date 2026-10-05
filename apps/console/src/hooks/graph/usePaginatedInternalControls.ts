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

import type { usePaginatedInternalControlsFragment$key } from "#/__generated__/core/usePaginatedInternalControlsFragment.graphql";
import type { usePaginatedInternalControlsQuery } from "#/__generated__/core/usePaginatedInternalControlsQuery.graphql";
import type { usePaginatedInternalControlsQuery_fragment } from "#/__generated__/core/usePaginatedInternalControlsQuery_fragment.graphql";

/* eslint-disable relay/unused-fields */

const internalControlsQuery = graphql`
  query usePaginatedInternalControlsQuery(
    $organizationId: ID!
    $first: Int = 20
    $order: InternalControlOrder
    $filter: InternalControlFilter
  ) {
    organization: node(id: $organizationId) {
      id
      ... on Organization {
        ...usePaginatedInternalControlsFragment @arguments(first: $first, order: $order, filter: $filter)
      }
    }
  }
`;

const internalControlsFragment = graphql`
  fragment usePaginatedInternalControlsFragment on Organization
  @refetchable(queryName: "usePaginatedInternalControlsQuery_fragment")
  @argumentDefinitions(
    first: { type: "Int", defaultValue: 20 }
    order: { type: "InternalControlOrder", defaultValue: null }
    filter: { type: "InternalControlFilter", defaultValue: null }
    after: { type: "CursorKey", defaultValue: null }
    before: { type: "CursorKey", defaultValue: null }
    last: { type: "Int", defaultValue: null }
  ) {
    internalControls(
      first: $first
      after: $after
      last: $last
      before: $before
      orderBy: $order
      filter: $filter
    ) @connection(key: "usePaginatedInternalControlsQuery_internalControls", filters: ["orderBy", "filter"]) {
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
 * Hook to retrieve internal controls paginated (used for link dialogs and internal control selectors)
 */
export function usePaginatedInternalControls(
  organizationId: string,
  options?: {
    first?: number;
    order?: { field: "CREATED_AT" | "NAME"; direction: "ASC" | "DESC" };
    filter?: { query?: string | null } | null;
  },
) {
  const query = useLazyLoadQuery<usePaginatedInternalControlsQuery>(
    internalControlsQuery,
    {
      organizationId,
      first: options?.first,
      order: options?.order,
      filter: options?.filter,
    },
    { fetchPolicy: "network-only" },
  );
  return usePaginationFragment<usePaginatedInternalControlsQuery_fragment, usePaginatedInternalControlsFragment$key>(
    internalControlsFragment,
    query.organization as usePaginatedInternalControlsFragment$key,
  );
}
