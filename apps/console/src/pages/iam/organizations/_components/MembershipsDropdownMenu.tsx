// Copyright (c) 2026 Probo Inc <hello@getprobo.com>.
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
