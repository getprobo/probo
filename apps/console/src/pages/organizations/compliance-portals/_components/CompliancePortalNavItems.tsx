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

import { useTranslation } from "react-i18next";
import { graphql, type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { useParams } from "react-router";

import type { CompliancePortalNavItemsQuery } from "#/__generated__/core/CompliancePortalNavItemsQuery.graphql";
import type { compliancePortalSections_compliancePortal$key } from "#/__generated__/core/compliancePortalSections_compliancePortal.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NavPanelGroup } from "#/pages/iam/organizations/_components/shell/NavPanelGroup";
import { NavPanelItem } from "#/pages/iam/organizations/_components/shell/NavPanelItem";

import {
  COMPLIANCE_PORTAL_SECTION_GROUPS,
  compliancePortalSectionsFragment,
  sectionPermissionsFrom,
  visibleCompliancePortalSections,
} from "../_lib/compliancePortalSections";

export const compliancePortalNavItemsQuery = graphql`
  query CompliancePortalNavItemsQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        ...compliancePortalSections_compliancePortal
      }
    }
  }
`;

export interface CompliancePortalNavItemsProps {
  queryRef: PreloadedQuery<CompliancePortalNavItemsQuery>;
}

export function CompliancePortalNavItems({ queryRef }: CompliancePortalNavItemsProps) {
  const { t } = useTranslation();
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const data = usePreloadedQuery<CompliancePortalNavItemsQuery>(
    compliancePortalNavItemsQuery,
    queryRef,
  );
  const portalKey = data.compliancePortal.__typename === "CompliancePortal"
    ? data.compliancePortal
    : null;
  const sectionData = useFragment<compliancePortalSections_compliancePortal$key>(
    compliancePortalSectionsFragment,
    portalKey,
  );
  if (sectionData == null || compliancePortalId == null) {
    return null;
  }
  const permissions = sectionPermissionsFrom(sectionData);
  const visible = visibleCompliancePortalSections(permissions);
  const prefix = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}`;

  return (
    <>
      {COMPLIANCE_PORTAL_SECTION_GROUPS.map((group) => {
        const items = visible.filter(section => section.group === group.id);
        if (items.length === 0) {
          return null;
        }
        return (
          <NavPanelGroup key={group.id} label={t(group.labelKey)}>
            {items.map(section => (
              <NavPanelItem
                key={section.id}
                label={t(section.labelKey)}
                to={`${prefix}/${section.path}`}
              />
            ))}
          </NavPanelGroup>
        );
      })}
    </>
  );
}
