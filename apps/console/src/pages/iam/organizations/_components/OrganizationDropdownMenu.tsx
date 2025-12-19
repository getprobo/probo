import { graphql, usePreloadedQuery, type PreloadedQuery } from "react-relay";
import type { OrganizationDropdownMenuQuery } from "./__generated__/OrganizationDropdownMenuQuery.graphql";
import { useMemo } from "react";
import { OrganizationDropdownMenuItem } from "./OrganizationDropdownMenuItem";

export const organizationDropdownMenuQuery = graphql`
  query OrganizationDropdownMenuQuery {
    viewer @required(action: THROW) {
      memberships(first: 1000, orderBy: { direction: DESC, field: CREATED_AT })
        @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            organization @required(action: THROW) {
              name
            }
            ...OrganizationDropdownMenuItemFragment
          }
        }
      }
    }
  }
`;

interface OrganizationDropdownMenuProps {
  queryRef: PreloadedQuery<OrganizationDropdownMenuQuery>;
  search: string;
}

export function OrganizationDropdownMenu(props: OrganizationDropdownMenuProps) {
  const { queryRef, search } = props;

  const {
    viewer: {
      memberships: { edges: initialMemberships },
    },
  } = usePreloadedQuery<OrganizationDropdownMenuQuery>(
    organizationDropdownMenuQuery,
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
        <OrganizationDropdownMenuItem fKey={node} key={node.id} />
      ))}
    </>
  );
}
