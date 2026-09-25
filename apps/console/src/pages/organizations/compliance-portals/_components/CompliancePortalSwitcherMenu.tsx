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

import { PlusIcon } from "@phosphor-icons/react";
import { DropdownItem } from "@probo/ui/src/v2/Dropdown/DropdownItem";
import { DropdownSeparator } from "@probo/ui/src/v2/Dropdown/DropdownSeparator";
import { Text } from "@probo/ui/src/v2/typography/Text";
import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Link } from "react-router";

import type { CompliancePortalSwitcherMenuQuery } from "#/__generated__/core/CompliancePortalSwitcherMenuQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { navPanelSwitcher } from "#/pages/organizations/_components/NavPanelSwitcher";

import { CompliancePortalSwitcherListItem } from "./CompliancePortalSwitcherListItem";

export const compliancePortalSwitcherMenuQuery = graphql`
  query CompliancePortalSwitcherMenuQuery($organizationId: ID!) {
    organization: node(id: $organizationId) {
      __typename
      ... on Organization {
        canCreateCompliancePortal: permission(action: "compliance-portal:portal:create")
        compliancePortals(first: 50, orderBy: { field: CREATED_AT, direction: DESC })
          @connection(key: "CompliancePortalSwitcherMenu_compliancePortals", filters: [])
          @required(action: THROW) {
          edges {
            node {
              id
              ...CompliancePortalSwitcherListItem_compliancePortal
            }
          }
        }
      }
    }
  }
`;

export interface CompliancePortalSwitcherMenuProps {
  queryRef: PreloadedQuery<CompliancePortalSwitcherMenuQuery>;
}

export function CompliancePortalSwitcherMenu({ queryRef }: CompliancePortalSwitcherMenuProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const slots = navPanelSwitcher();

  const { organization } = usePreloadedQuery<CompliancePortalSwitcherMenuQuery>(
    compliancePortalSwitcherMenuQuery,
    queryRef,
  );
  if (organization.__typename !== "Organization") {
    throw new Error("invalid type for node");
  }

  const portals = organization.compliancePortals.edges.map(edge => edge.node);
  const newPortalHref = `/organizations/${organizationId}/compliance-portals/new`;

  return (
    <>
      <div className={slots.list()}>
        {portals.length === 0
          ? (
              <Text size={2} color="faint" className={slots.empty()}>
                {t("nav.compliancePortalSwitcher.empty")}
              </Text>
            )
          : portals.map(portal => (
              <CompliancePortalSwitcherListItem key={portal.id} compliancePortalKey={portal} />
            ))}
      </div>
      {organization.canCreateCompliancePortal && (
        <>
          <DropdownSeparator />
          <DropdownItem
            iconStart={<PlusIcon />}
            render={<Link to={newPortalHref} />}
          >
            {t("nav.compliancePortalSwitcher.create")}
          </DropdownItem>
        </>
      )}
    </>
  );
}
