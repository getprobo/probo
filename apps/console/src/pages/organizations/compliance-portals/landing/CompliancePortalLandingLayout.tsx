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

import { IconBook, IconImage, TabLink, Tabs } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, Outlet, useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalLandingLayoutQuery } from "#/__generated__/core/CompliancePortalLandingLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const compliancePortalLandingLayoutQuery = graphql`
  query CompliancePortalLandingLayoutQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        canListCustomLinks: permission(action: "compliance-portal:compliance-custom-link:list")
        canListFrameworks: permission(action: "compliance-portal:compliance-framework:list")
        canListCommitmentGroups: permission(action: "compliance-portal:commitment-group:list")
        canListReferences: permission(action: "compliance-portal:portal-reference:list")
      }
    }
  }
`;

interface CompliancePortalLandingLayoutProps {
  queryRef: PreloadedQuery<CompliancePortalLandingLayoutQuery>;
}

export function CompliancePortalLandingLayout({ queryRef }: CompliancePortalLandingLayoutProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { pathname } = useLocation();

  const { compliancePortal } = usePreloadedQuery<CompliancePortalLandingLayoutQuery>(
    compliancePortalLandingLayoutQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  const landingBase = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}/landing`;
  const canBranding = compliancePortal.canListCustomLinks;
  const canContent
    = compliancePortal.canListFrameworks
      || compliancePortal.canListCommitmentGroups
      || compliancePortal.canListReferences;

  if (pathname === landingBase && !canBranding && canContent) {
    return <Navigate to={`${landingBase}/content`} replace />;
  }

  return (
    <div className="space-y-6">
      <Tabs>
        {canBranding && (
          <TabLink to={landingBase} end>
            <IconImage className="size-4" />
            {t("landingLayout.tabs.branding")}
          </TabLink>
        )}
        {canContent && (
          <TabLink to={`${landingBase}/content`}>
            <IconBook className="size-4" />
            {t("landingLayout.tabs.content")}
          </TabLink>
        )}
      </Tabs>
      <Outlet />
    </div>
  );
}
