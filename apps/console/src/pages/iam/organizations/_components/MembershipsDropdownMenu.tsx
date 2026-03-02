import { useMemo } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MembershipsDropdownMenuQuery } from "#/__generated__/iam/MembershipsDropdownMenuQuery.graphql";

import { MembershipsDropdownMenuItem } from "./MembershipsDropdownMenuItem";

export const membershipsDropdownMenuQuery = graphql`
  query MembershipsDropdownMenuQuery {
    viewer @required(action: THROW) {
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { state: ACTIVE }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            organization @required(action: THROW) {
              name
              ...MembershipsDropdownMenuItem_organizationFragment
            }
            membership @required(action: THROW) {
              ...MembershipsDropdownMenuItemFragment
            }
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
      profiles: { edges: initialProfiles },
    },
  } = usePreloadedQuery<MembershipsDropdownMenuQuery>(
    membershipsDropdownMenuQuery,
    queryRef,
  );

  const profiles = useMemo(() => {
    if (!search) {
      return initialProfiles;
    }

    return initialProfiles.filter(({ node: { organization } }) =>
      organization.name.toLowerCase().includes(search.toLowerCase()),
    );
  }, [initialProfiles, search]);

  return (
    <>
      {profiles.map(({ node }) => (
        <MembershipsDropdownMenuItem fKey={node.membership} organizationFragmentRef={node.organization} key={node.id} />
      ))}
    </>
  );
}
