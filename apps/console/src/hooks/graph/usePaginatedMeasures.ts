import { graphql, useLazyLoadQuery, usePaginationFragment } from "react-relay";

import type { usePaginatedMeasuresFragment$key } from "#/__generated__/core/usePaginatedMeasuresFragment.graphql";
import type { usePaginatedMeasuresQuery } from "#/__generated__/core/usePaginatedMeasuresQuery.graphql";

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
  return usePaginationFragment(
    measuresFragment,
    query.organization as usePaginatedMeasuresFragment$key,
  );
}
