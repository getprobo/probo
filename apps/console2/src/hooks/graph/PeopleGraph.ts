import { graphql } from "relay-runtime";
import type { PeopleGraphQuery } from "./__generated__/PeopleGraphQuery.graphql";
import { useLazyLoadQuery } from "react-relay";
import { useMemo } from "react";

const peopleQuery = graphql`
  query PeopleGraphQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      ... on Organization {
        peoples(first: 100, orderBy: { direction: ASC, field: CREATED_AT }) {
          edges {
            node {
              id
              fullName
              primaryEmailAddress
            }
          }
        }
      }
    }
  }
`;

export function usePeople(organizationId: string) {
  const data = useLazyLoadQuery<PeopleGraphQuery>(peopleQuery, {
    organizationId: organizationId,
  });
  return useMemo(() => {
    return data.organization?.peoples?.edges.map((edge) => edge.node) ?? [];
  }, [data]);
}
