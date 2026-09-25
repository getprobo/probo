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

import { DropdownGroup } from "@probo/ui/src/v2/Dropdown/DropdownGroup";
import { DropdownGroupLabel } from "@probo/ui/src/v2/Dropdown/DropdownGroupLabel";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useMemo } from "react";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";

import type { OrganizationSwitcherMenuQuery } from "#/__generated__/iam/OrganizationSwitcherMenuQuery.graphql";

import { OrganizationSwitcherInvitingItem } from "./OrganizationSwitcherInvitingItem";
import { OrganizationSwitcherMenuItem } from "./OrganizationSwitcherMenuItem";

export const organizationSwitcherMenuQuery = graphql`
  query OrganizationSwitcherMenuQuery {
    viewer @required(action: THROW) {
      profiles(
        first: 1000
        orderBy: { direction: ASC, field: ORGANIZATION_NAME }
        filter: { states: [ACTIVE] }
      ) @required(action: THROW) {
        edges @required(action: THROW) {
          node @required(action: THROW) {
            id
            organization @required(action: THROW) {
              name
              ...OrganizationSwitcherMenuItem_organization
            }
            membership @required(action: THROW) {
              ...OrganizationSwitcherMenuItem_membership
            }
          }
        }
      }
      invitingOrganizations {
        id
        name
        ...OrganizationSwitcherInvitingItem_organization
      }
    }
  }
`;

export interface OrganizationSwitcherMenuProps {
  search: string;
  queryRef: PreloadedQuery<OrganizationSwitcherMenuQuery>;
}

export function OrganizationSwitcherMenu({ search, queryRef }: OrganizationSwitcherMenuProps) {
  const { t } = useTranslation();

  const {
    viewer: {
      profiles: { edges: allProfiles },
      invitingOrganizations: allInvitingOrganizations,
    },
  } = usePreloadedQuery<OrganizationSwitcherMenuQuery>(organizationSwitcherMenuQuery, queryRef);

  const query = search.trim().toLowerCase();

  const profiles = useMemo(
    () => (query
      ? allProfiles.filter(({ node }) => node.organization.name.toLowerCase().includes(query))
      : allProfiles),
    [allProfiles, query],
  );

  const invitingOrganizations = useMemo(
    () => (query
      ? allInvitingOrganizations.filter(org => org.name.toLowerCase().includes(query))
      : allInvitingOrganizations),
    [allInvitingOrganizations, query],
  );

  if (profiles.length === 0 && invitingOrganizations.length === 0) {
    return (
      <Text size={2} color="faint" className="block px-3 py-2">
        {t("organizationSwitcher.empty")}
      </Text>
    );
  }

  return (
    <>
      {invitingOrganizations.length > 0 && (
        <>
          <DropdownGroup>
            <DropdownGroupLabel>
              {t("membershipsDropdownMenu.pendingInvitations")}
            </DropdownGroupLabel>
            {invitingOrganizations.map(organization => (
              <OrganizationSwitcherInvitingItem
                key={organization.id}
                organizationKey={organization}
              />
            ))}
          </DropdownGroup>
          <DropdownSeparator />
        </>
      )}
      <DropdownGroup>
        {profiles.map(({ node }) => (
          <OrganizationSwitcherMenuItem
            key={node.id}
            membershipKey={node.membership}
            organizationKey={node.organization}
          />
        ))}
      </DropdownGroup>
    </>
  );
}
