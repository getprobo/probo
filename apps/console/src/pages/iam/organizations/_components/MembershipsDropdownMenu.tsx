import { useMemo } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MembershipsDropdownMenuQuery } from "#/__generated__/iam/MembershipsDropdownMenuQuery.graphql";

import { MembershipsDropdownMenuItem } from "./MembershipsDropdownMenuItem";

export const membershipsDropdownMenuQuery = graphql`
  query MembershipsDropdownMenuQuery {
    viewer @required(action: THROW) {
      memberships(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            organization @required(action: THROW) {
              name
            }
            ...MembershipsDropdownMenuItemFragment
          }
        }
      }
    }
  }
`;

interface MembershipsDropdownMenuProps {
  queryRef: PreloadedQuery<MembershipsDropdownMenuQuery>;
  search: string;
}

export function MembershipsDropdownMenu(props: MembershipsDropdownMenuProps) {
  const { queryRef, search } = props;

  const {
    viewer: {
      memberships: { edges: initialMemberships },
    },
  } = usePreloadedQuery<MembershipsDropdownMenuQuery>(
    membershipsDropdownMenuQuery,
    queryRef,
  );

  const memberships = useMemo(() => {
    if (!search) {
      return initialMemberships;
    }

    return initialMemberships.filter(({ node: { organization } }) =>
      organization.name.toLowerCase().includes(search.toLowerCase()),
    );
  }, [initialMemberships, search]);

  return (
    <>
      {memberships.map(({ node }) => (
        <MembershipsDropdownMenuItem fKey={node} key={node.id} />
      ))}
    </>
  );
}
