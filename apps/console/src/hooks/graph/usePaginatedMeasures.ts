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
