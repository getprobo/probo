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

import { IconFolder2, IconMedal, IconPageTextLine, TabLink, Tabs } from "@probo/ui";
import { useTranslation } from "react-i18next";
import { type PreloadedQuery, usePreloadedQuery } from "react-relay";
import { Navigate, Outlet, useLocation, useParams } from "react-router";
import { graphql } from "relay-runtime";

import type { CompliancePortalDocumentsLayoutQuery } from "#/__generated__/core/CompliancePortalDocumentsLayoutQuery.graphql";
import { useOrganizationId } from "#/hooks/useOrganizationId";
import { NotFoundError } from "#/lib/relay/errors";

export const compliancePortalDocumentsLayoutQuery = graphql`
  query CompliancePortalDocumentsLayoutQuery($compliancePortalId: ID!) {
    compliancePortal: node(id: $compliancePortalId) {
      __typename
      ... on CompliancePortal {
        canListAudits: permission(action: "core:audit:list")
        canListDocuments: permission(action: "core:document:list")
        canListFiles: permission(action: "compliance-portal:portal-file:list")
      }
    }
  }
`;

interface CompliancePortalDocumentsLayoutProps {
  queryRef: PreloadedQuery<CompliancePortalDocumentsLayoutQuery>;
}

export function CompliancePortalDocumentsLayout({ queryRef }: CompliancePortalDocumentsLayoutProps) {
  const { t } = useTranslation("organizations/compliance-portals");
  const organizationId = useOrganizationId();
  const { compliancePortalId } = useParams<{ compliancePortalId: string }>();
  const { pathname } = useLocation();

  const { compliancePortal } = usePreloadedQuery<CompliancePortalDocumentsLayoutQuery>(
    compliancePortalDocumentsLayoutQuery,
    queryRef,
  );
  if (compliancePortal.__typename !== "CompliancePortal") {
    throw new Error("invalid type for node");
  }

  const documentsBase = `/organizations/${organizationId}/compliance-portals/${compliancePortalId}/documents`;

  if (pathname === documentsBase && !compliancePortal.canListDocuments) {
    if (compliancePortal.canListAudits) {
      return <Navigate to={`${documentsBase}/audits`} replace />;
    }
    if (compliancePortal.canListFiles) {
      return <Navigate to={`${documentsBase}/files`} replace />;
    }
  }

  const isAudits = pathname.startsWith(`${documentsBase}/audits`);
  const isFiles = pathname.startsWith(`${documentsBase}/files`);
  if (isAudits && !compliancePortal.canListAudits) {
    throw new NotFoundError("Compliance portal documents not found");
  }
  if (isFiles && !compliancePortal.canListFiles) {
    throw new NotFoundError("Compliance portal documents not found");
  }
  if (!isAudits && !isFiles && !compliancePortal.canListDocuments) {
    throw new NotFoundError("Compliance portal documents not found");
  }

  return (
    <div className="space-y-6">
      <Tabs>
        {compliancePortal.canListDocuments && (
          <TabLink to={documentsBase} end>
            <IconPageTextLine className="size-4" />
            {t("documentsLayout.tabs.documents")}
          </TabLink>
        )}
        {compliancePortal.canListAudits && (
          <TabLink to={`${documentsBase}/audits`}>
            <IconMedal className="size-4" />
            {t("documentsLayout.tabs.audits")}
          </TabLink>
        )}
        {compliancePortal.canListFiles && (
          <TabLink to={`${documentsBase}/files`}>
            <IconFolder2 className="size-4" />
            {t("documentsLayout.tabs.files")}
          </TabLink>
        )}
      </Tabs>
      <Outlet />
    </div>
  );
}
