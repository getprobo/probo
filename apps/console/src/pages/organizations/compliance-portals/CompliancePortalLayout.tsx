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

import { safeOpenUrl } from "@probo/helpers";
import { usePageTitle } from "@probo/hooks";
import { Badge, Button, PageHeader } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, useFragment, usePreloadedQuery } from "react-relay";
import { Navigate, Outlet, useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalLayoutQuery } from "#/__generated__/core/CompliancePortalLayoutQuery.graphql";
import type { compliancePortalSections_compliancePortal$key } from "#/__generated__/core/compliancePortalSections_compliancePortal.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

import {
  compliancePortalSectionsFragment,
  firstVisibleSection,
  sectionPermissionsFrom,
} from "./_lib/compliancePortalSections";

export const compliancePortalLayoutQuery = graphql`
  query CompliancePortalLayoutQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        id
        entityName
        active
        publicUrl
        organization @required(action: THROW) {
          id
        }
        ...compliancePortalSections_compliancePortal
      }
    }
  }
`;

interface CompliancePortalLayoutProps {
  queryRef: PreloadedQuery<CompliancePortalLayoutQuery>;
}

export function CompliancePortalLayout({ queryRef }: CompliancePortalLayoutProps) {
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { pathname } = useLocation();
  const { t } = useTranslation("organizations/compliance-portals");

  usePageTitle(t("portalLayout.title"));

  const { compliancePortal } = usePreloadedQuery<CompliancePortalLayoutQuery>(
    compliancePortalLayoutQuery,
    queryRef,
  );
  const portalKey = compliancePortal.__typename === "CompliancePortal"
    ? compliancePortal
    : null;
  const sectionData = useFragment<compliancePortalSections_compliancePortal$key>(
    compliancePortalSectionsFragment,
    portalKey,
  );
  if (compliancePortal.__typename !== "CompliancePortal" || sectionData == null) {
    throw new Error("invalid type for node");
  }
  // Guards the nested routes too: a portal reached under another organization's
  // URL would mix that organization's navigation with this portal's settings.
  if (compliancePortal.organization.id !== organizationId) {
    throw new Error("compliance portal does not belong to this organization");
  }

  const portalBase = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}`;
  const compliancePortalUrl = compliancePortal.publicUrl;
  const landingSection = firstVisibleSection(sectionPermissionsFrom(sectionData));

  if (pathname === portalBase && landingSection != null) {
    return <Navigate to={`${portalBase}/${landingSection.path}`} replace />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={compliancePortal.entityName}
        description={t("portalLayout.description")}
      >
        <Badge variant={compliancePortal.active ? "success" : "danger"}>
          {compliancePortal.active
            ? t("portalLayout.status.active")
            : t("portalLayout.status.inactive")}
        </Badge>
        {compliancePortal.active && compliancePortalUrl && (
          <Button
            variant="secondary"
            onClick={() => safeOpenUrl(compliancePortalUrl)}
          >
            {t("portalLayout.actions.open")}
          </Button>
        )}
      </PageHeader>

      <Outlet />
    </div>
  );
}
