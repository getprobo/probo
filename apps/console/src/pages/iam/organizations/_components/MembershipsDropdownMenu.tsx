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

import { useTranslate } from "@probo/i18n";
import { DropdownSeparator } from "@probo/ui";
import { useMemo } from "react";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { MembershipsDropdownMenuQuery } from "#/__generated__/iam/MembershipsDropdownMenuQuery.graphql";

import { MembershipsDropdownInvitingItem } from "./MembershipsDropdownInvitingItem";
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
      invitingOrganizations {
        id
        name
        ...MembershipsDropdownInvitingItemFragment
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
  const { __ } = useTranslate();

  const {
    viewer: {
      profiles: { edges: initialProfiles },
      invitingOrganizations: initialInvitingOrganizations,
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

  const invitingOrganizations = useMemo(() => {
    if (!search) {
      return initialInvitingOrganizations;
    }

    return initialInvitingOrganizations.filter(organization =>
      organization.name.toLowerCase().includes(search.toLowerCase()),
    );
  }, [initialInvitingOrganizations, search]);

  return (
    <>
      {invitingOrganizations.length > 0 && (
        <>
          <div className="px-3 py-1 text-xs text-txt-tertiary uppercase">
            {__("Pending invitations")}
          </div>
          {invitingOrganizations.map(organization => (
            <MembershipsDropdownInvitingItem key={organization.id} fKey={organization} />
          ))}
          <DropdownSeparator />
        </>
      )}
      {profiles.map(({ node }) => (
        <MembershipsDropdownMenuItem fKey={node.membership} organizationFragmentRef={node.organization} key={node.id} />
      ))}
    </>
  );
}
