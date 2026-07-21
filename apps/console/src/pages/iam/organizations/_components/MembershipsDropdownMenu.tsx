// Copyright (c) 2026 Probo Inc <hello@probo.com>.
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

import { DropdownSeparator } from "@probo/ui";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
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
  const { t } = useTranslation();

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
            {t("membershipsDropdownMenu.pendingInvitations")}
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
