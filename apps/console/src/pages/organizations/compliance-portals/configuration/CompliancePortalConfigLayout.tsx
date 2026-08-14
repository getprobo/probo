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
import { Badge, Button, IconBell2, IconCheckmark1, IconFolder2, IconMedal, IconPageTextLine, IconPencil, IconPeopleAdd, IconSettingsGear2, IconShield, IconStore, PageHeader, TabLink, Tabs } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, Outlet, useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalConfigLayoutQuery } from "#/__generated__/core/CompliancePortalConfigLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";

export const compliancePortalConfigLayoutQuery = graphql`
  query CompliancePortalConfigLayoutQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        id
        entityName
        active
        publicUrl
        organization {
          id
        }
        canListFrameworks: permission(action: "compliance-portal:compliance-framework:list")
        canListCustomLinks: permission(action: "compliance-portal:compliance-custom-link:list")
        canListReferences: permission(action: "compliance-portal:portal-reference:list")
        canListCommitmentGroups: permission(action: "compliance-portal:commitment-group:list")
        canListAudits: permission(action: "core:audit:list")
        canListDocuments: permission(action: "core:document:list")
        canListFiles: permission(action: "compliance-portal:portal-file:list")
        canGetThirdParty: permission(action: "core:thirdParty:get")
        canListAccesses: permission(action: "compliance-portal:portal-access:list")
        canListMailingListSubscribers: permission(
          action: "compliance-portal:mailing-list-subscriber:list"
        )
      }
    }
  }
`;

interface CompliancePortalConfigLayoutProps {
  queryRef: PreloadedQuery<CompliancePortalConfigLayoutQuery>;
}

export default function CompliancePortalConfigLayout({ queryRef }: CompliancePortalConfigLayoutProps) {
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { pathname } = useLocation();
  const { t } = useTranslation("organizations/compliance-portals");

  usePageTitle(t("configLayout.title"));

  const { compliancePortal } = usePreloadedQuery<CompliancePortalConfigLayoutQuery>(
    compliancePortalConfigLayoutQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }
  // Guards the nested routes too: a portal reached under another organization's
  // URL would mix that organization's navigation with this portal's settings.
  if (compliancePortal.organization.id !== organizationId) {
    throw new Error("compliance portal does not belong to this organization");
  }

  const portalBase = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}`;
  const compliancePortalUrl = compliancePortal.publicUrl;

  // Access managers land on the portal root (overview) from the switcher, but
  // lack overview permissions — send them to Access instead of a forbidden page.
  if (
    pathname === portalBase
    && !compliancePortal.canListFrameworks
    && compliancePortal.canListAccesses
  ) {
    return <Navigate to={`${portalBase}/access`} replace />;
  }

  return (
    <div className="space-y-6">
      <PageHeader
        title={compliancePortal.entityName}
        description={t("configLayout.description")}
      >
        <Badge variant={compliancePortal.active ? "success" : "danger"}>
          {compliancePortal.active
            ? t("configLayout.status.active")
            : t("configLayout.status.inactive")}
        </Badge>
        {compliancePortal.active && compliancePortalUrl && (
          <Button
            variant="secondary"
            onClick={() => safeOpenUrl(compliancePortalUrl)}
          >
            {t("configLayout.actions.open")}
          </Button>
        )}
      </PageHeader>

      <Tabs>
        {compliancePortal.canListFrameworks && (
          <TabLink to={portalBase} end>
            <IconSettingsGear2 className="size-4" />
            {t("configLayout.tabs.overview")}
          </TabLink>
        )}
        {compliancePortal.canListCustomLinks && (
          <TabLink to={`${portalBase}/brand`}>
            <IconPencil className="size-4" />
            {t("configLayout.tabs.brand")}
          </TabLink>
        )}
        {compliancePortal.canListReferences && (
          <TabLink to={`${portalBase}/references`}>
            <IconCheckmark1 className="size-4" />
            {t("configLayout.tabs.references")}
          </TabLink>
        )}
        {compliancePortal.canListCommitmentGroups && (
          <TabLink to={`${portalBase}/commitments`}>
            <IconShield className="size-4" />
            {t("configLayout.tabs.commitments")}
          </TabLink>
        )}
        {compliancePortal.canListAudits && (
          <TabLink to={`${portalBase}/audits`}>
            <IconMedal className="size-4" />
            {t("configLayout.tabs.audits")}
          </TabLink>
        )}
        {compliancePortal.canListDocuments && (
          <TabLink to={`${portalBase}/documents`}>
            <IconPageTextLine className="size-4" />
            {t("configLayout.tabs.documents")}
          </TabLink>
        )}
        {compliancePortal.canListFiles && (
          <TabLink to={`${portalBase}/files`}>
            <IconFolder2 className="size-4" />
            {t("configLayout.tabs.files")}
          </TabLink>
        )}
        {compliancePortal.canGetThirdParty && (
          <TabLink to={`${portalBase}/third-parties`}>
            <IconStore className="size-4" />
            {t("configLayout.tabs.subprocessors")}
          </TabLink>
        )}
        {compliancePortal.canListAccesses && (
          <TabLink to={`${portalBase}/access`}>
            <IconPeopleAdd className="size-4" />
            {t("configLayout.tabs.access")}
          </TabLink>
        )}
        {compliancePortal.canListMailingListSubscribers && (
          <TabLink to={`${portalBase}/mailing-list`}>
            <IconBell2 className="size-4" />
            {t("configLayout.tabs.mailingList")}
          </TabLink>
        )}
      </Tabs>

      <Outlet />
    </div>
  );
}
